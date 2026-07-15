package activity

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CurrentActivities(ctx context.Context, now int64, page int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	if page < 1 {
		page = 1
	}
	offset := page - 1
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM activity WHERE 1=1 AND reward_expire_time>? ORDER BY id DESC LIMIT 1 OFFSET ?", now, offset)
	if err != nil {
		return nil, fmt.Errorf("query current activities: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) ActivityByID(ctx context.Context, id int) (map[string]interface{}, error) {
	if r.db == nil || id <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM activity WHERE id=?", id)
	if err != nil {
		return nil, fmt.Errorf("query activity: %w", err)
	}
	defer rows.Close()
	items, err := scanRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return map[string]interface{}{}, nil
	}
	return items[0], nil
}

func (r *Repository) PrizesByActivityID(ctx context.Context, id int) ([]map[string]interface{}, error) {
	if r.db == nil || id <= 0 {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM activity_prizes WHERE aid=? ORDER BY ranking ASC", id)
	if err != nil {
		return nil, fmt.Errorf("query activity prizes: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) PrizeLogs(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return []map[string]interface{}{}, nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM activity_prizelogs WHERE uid=? ORDER BY createtime DESC LIMIT ? OFFSET ?", uid, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query activity prize logs: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) CountActivityRecords(ctx context.Context, aid int) (int, error) {
	if r.db == nil || aid <= 0 {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM activity_records WHERE 1=1 AND aid=?", aid).Scan(&total); err != nil {
		return 0, fmt.Errorf("count activity records: %w", err)
	}
	return total, nil
}

func (r *Repository) ActivityRecords(ctx context.Context, aid int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil || aid <= 0 {
		return []map[string]interface{}{}, nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `SELECT a.*, b.username, b.avatar
FROM activity_records a
LEFT JOIN users b ON b.uid=a.uid
WHERE 1=1 AND a.aid=?
ORDER BY a.score DESC LIMIT ? OFFSET ?`, aid, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query activity records: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) ActivityRanking(ctx context.Context, aid int, uid int) (map[string]interface{}, error) {
	if r.db == nil || aid <= 0 || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM activity_records WHERE aid=? ORDER BY score DESC", aid)
	if err != nil {
		return nil, fmt.Errorf("query activity ranking: %w", err)
	}
	defer rows.Close()
	items, err := scanRows(rows)
	if err != nil {
		return nil, err
	}
	for index, row := range items {
		if fmt.Sprint(row["uid"]) == fmt.Sprint(uid) {
			out := make(map[string]interface{}, len(row)+1)
			for key, value := range row {
				out[key] = value
			}
			out["ranking"] = fmt.Sprint(index + 1)
			return out, nil
		}
	}
	return map[string]interface{}{}, nil
}

func (r *Repository) BotByID(ctx context.Context, uid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM bot_users WHERE uid=?", uid)
	if err != nil {
		return nil, fmt.Errorf("query bot user: %w", err)
	}
	defer rows.Close()
	items, err := scanRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return map[string]interface{}{}, nil
	}
	return items[0], nil
}

func (r *Repository) CountRecommendedUsers(ctx context.Context, recommenderUID int, start int64, end int64) (int, error) {
	if r.db == nil || recommenderUID <= 0 {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM users a
WHERE 1=1
  AND a.regtime>=?
  AND a.regtime<=?
  AND a.uid IN (SELECT uid FROM user_recommend WHERE recommend_uid=?)`, start, end, recommenderUID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count activity recommended users: %w", err)
	}
	return total, nil
}

func (r *Repository) RecommendedUsers(ctx context.Context, recommenderUID int, start int64, end int64, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil || recommenderUID <= 0 {
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
WHERE 1=1
  AND a.regtime>=?
  AND a.regtime<=?
  AND a.uid IN (SELECT uid FROM user_recommend WHERE recommend_uid=?)
ORDER BY a.uid DESC LIMIT ? OFFSET ?`, start, end, recommenderUID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query activity recommended users: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) UserGroups(ctx context.Context) ([]map[string]interface{}, error) {
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
