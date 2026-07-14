package user

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UserBySession(ctx context.Context, sid string) (map[string]interface{}, error) {
	if r.db == nil || !validSID(sid) {
		return nil, nil
	}

	session, err := r.queryOne(ctx, "SELECT sid,token,uid,type,timestamp FROM sessions WHERE sid=?", sid)
	if err != nil {
		return nil, fmt.Errorf("query session: %w", err)
	}
	if len(session) == 0 {
		return nil, nil
	}
	if fmt.Sprint(session["type"]) != "0" {
		return nil, nil
	}

	user, err := r.queryOne(ctx, "SELECT * FROM users WHERE uid=?", session["uid"])
	if err != nil {
		return nil, fmt.Errorf("query session user: %w", err)
	}
	if len(user) == 0 {
		return nil, nil
	}
	expected := md5.Sum([]byte(fmt.Sprintf("%s_%s", user["password"], user["salt"])))
	if fmt.Sprint(session["token"]) != hex.EncodeToString(expected[:]) {
		return nil, nil
	}
	user["sid"] = sid
	return user, nil
}

func (r *Repository) Groups(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM user_groups ORDER BY gid ASC")
	if err != nil {
		return nil, fmt.Errorf("query user groups: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) CountRecommended(ctx context.Context, uid int) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users a WHERE 1=1 AND a.recommend_uid=?", uid).Scan(&total); err != nil {
		return 0, fmt.Errorf("count recommended users: %w", err)
	}
	return total, nil
}

func (r *Repository) RecommendedUsers(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `
SELECT a.*, b.gname AS sysgname, c.gname
FROM users a
LEFT JOIN user_groups b ON b.gid=a.sysgid
LEFT JOIN user_groups c ON c.gid=a.gid
WHERE 1=1 AND a.recommend_uid=?
ORDER BY a.uid DESC
LIMIT ? OFFSET ?`, uid, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query recommended users: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func validSID(sid string) bool {
	if len(sid) != 32 {
		return false
	}
	for _, ch := range sid {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') {
			return false
		}
	}
	return true
}

func scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("read columns: %w", err)
	}
	result := []map[string]interface{}{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		row := make(map[string]interface{}, len(columns))
		for i, column := range columns {
			row[column] = normalizeSQLValue(values[i])
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}
	return result, nil
}

func normalizeSQLValue(value interface{}) interface{} {
	switch v := value.(type) {
	case nil:
		return nil
	case []byte:
		return string(v)
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprint(v)
	}
}

func (r *Repository) queryOne(ctx context.Context, query string, args ...interface{}) (map[string]interface{}, error) {
	if r.db == nil {
		return nil, nil
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return items[0], nil
}

func CleanToken(token string) string {
	token = strings.TrimSpace(token)
	decoded, err := hex.DecodeString(token)
	if err == nil && len(decoded) == 32 {
		return string(decoded)
	}
	return token
}
