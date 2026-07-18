package aiundress

import (
	"context"
	"database/sql"
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

func (r *Repository) Count(ctx context.Context, uid int, module int) (int, error) {
	if r.db == nil || uid <= 0 {
		return 0, nil
	}
	query, args := listingWhere(uid, module)
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ai_undress WHERE "+query, args...).Scan(&total); err != nil {
		if isMissingTable(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("count ai undress rows: %w", err)
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, uid int, module int, total int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return []map[string]interface{}{}, nil
	}
	where, args := listingWhere(uid, module)
	offset := limitOffset(total, pageSize, page)
	args = append(args, pageSize, offset)
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM ai_undress WHERE "+where+" ORDER BY id DESC LIMIT ? OFFSET ?", args...)
	if err != nil {
		if isMissingTable(err) {
			return []map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("list ai undress rows: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) ByID(ctx context.Context, id int) (map[string]interface{}, error) {
	if r.db == nil || id <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM ai_undress WHERE id=? LIMIT 1", id)
	if err != nil {
		if isMissingTable(err) {
			return map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("query ai undress row: %w", err)
	}
	defer rows.Close()
	result, err := scanRows(rows)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return map[string]interface{}{}, nil
	}
	return result[0], nil
}

func (r *Repository) ByUIDImage(ctx context.Context, uid int, image string) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 || image == "" {
		return map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM ai_undress WHERE uid=? AND image=? LIMIT 1", uid, image)
	if err != nil {
		if isMissingTable(err) {
			return map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("query ai undress row by image: %w", err)
	}
	defer rows.Close()
	result, err := scanRows(rows)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return map[string]interface{}{}, nil
	}
	return result[0], nil
}

func (r *Repository) MarkDeleted(ctx context.Context, id int, updateTime int64) error {
	if r.db == nil || id <= 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE ai_undress SET status=5, update_time=? WHERE id=?", updateTime, id); err != nil {
		if isMissingTable(err) {
			return nil
		}
		return fmt.Errorf("mark ai undress row deleted: %w", err)
	}
	return nil
}

func (r *Repository) SettingByUUID(ctx context.Context, uuid string) (string, error) {
	if r.db == nil || uuid == "" {
		return "", nil
	}
	var value string
	if err := r.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE uuid=?", uuid).Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("query setting %s: %w", uuid, err)
	}
	return value, nil
}

func listingWhere(uid int, module int) (string, []interface{}) {
	where := "uid=? AND status <= 4 AND status > 0"
	args := []interface{}{uid}
	if module != 0 {
		where += " AND module=?"
		args = append(args, module)
	}
	return where, args
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

func limitOffset(total int, pageSize int, page int) int {
	if pageSize < 1 {
		pageSize = 1
	}
	totalPage := (total + pageSize - 1) / pageSize
	if totalPage < 1 {
		totalPage = 1
	}
	if page < 1 {
		page = 1
	}
	if page > totalPage {
		page = totalPage
	}
	return (page - 1) * pageSize
}

func isMissingTable(err error) bool {
	return err != nil && strings.Contains(err.Error(), "doesn't exist")
}
