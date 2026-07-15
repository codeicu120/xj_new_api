package explore

import (
	"context"
	"database/sql"
	"encoding/json"
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

func (r *Repository) Tabs(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM explore_tabs WHERE 1=1 AND showtype=0 ORDER BY sortnum ASC LIMIT ? OFFSET 0", 100)
}

func (r *Repository) UpdateUserNotificationAll(ctx context.Context, uid int, value string) error {
	if r.db == nil || uid <= 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE users SET notification_all=? WHERE uid=?", value, uid); err != nil {
		return fmt.Errorf("update user notification_all: %w", err)
	}
	return nil
}

func (r *Repository) UpdateGuestNotificationAll(ctx context.Context, sid string, value string) error {
	if r.db == nil || sid == "" {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE user_guests SET notification_all=? WHERE sid=?", value, sid); err != nil {
		return fmt.Errorf("update guest notification_all: %w", err)
	}
	return nil
}

func (r *Repository) VodTaskByID(ctx context.Context, vid int) (map[string]interface{}, error) {
	if r.db == nil || vid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM explore_vods WHERE vid=?", vid)
	if err != nil {
		return nil, fmt.Errorf("query explore vod: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) UserVodTaskLog(ctx context.Context, uid int, today int64, vid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 || vid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM explore_vodlogs WHERE uid=? AND addtime>? AND vid=? LIMIT 1", uid, today, vid)
	if err != nil {
		return nil, fmt.Errorf("query explore vod log: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) GuestVodTaskLog(ctx context.Context, sid string, today int64, vid int) (map[string]interface{}, error) {
	if r.db == nil || sid == "" || vid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM explore_guestvodlogs WHERE sid=? AND addtime>=? AND vid=? LIMIT 1", sid, today, vid)
	if err != nil {
		return nil, fmt.Errorf("query explore guest vod log: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) CreateUserVodTaskLog(ctx context.Context, uid int, vid int, addtime int64, reqcoin int) (int, error) {
	if r.db == nil || uid <= 0 || vid <= 0 {
		return 0, nil
	}
	result, err := r.db.ExecContext(ctx, "INSERT INTO explore_vodlogs(uid, vid, addtime, reqcoin) VALUES(?, ?, ?, ?)", uid, vid, addtime, reqcoin)
	if err != nil {
		return 0, fmt.Errorf("insert explore vod log: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert explore vod log id: %w", err)
	}
	return int(id), nil
}

func (r *Repository) CreateGuestVodTaskLog(ctx context.Context, sid string, vid int, addtime int64, reqcoin int) (int, error) {
	if r.db == nil || sid == "" || vid <= 0 {
		return 0, nil
	}
	result, err := r.db.ExecContext(ctx, "INSERT INTO explore_guestvodlogs(sid, vid, addtime, reqcoin) VALUES(?, ?, ?, ?)", sid, vid, addtime, reqcoin)
	if err != nil {
		return 0, fmt.Errorf("insert explore guest vod log: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert explore guest vod log id: %w", err)
	}
	return int(id), nil
}

func (r *Repository) queryRows(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

func DecodeJSON(value interface{}) interface{} {
	raw := strings.TrimSpace(fmt.Sprint(value))
	if raw == "" {
		return nil
	}
	var decoded interface{}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return nil
	}
	return decoded
}
