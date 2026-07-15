package bought

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

func (r *Repository) Items(ctx context.Context, uid int, page int, pageSize int) (int, []map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return 0, []map[string]interface{}{}, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_bought WHERE uid=?", uid).Scan(&total); err != nil {
		return 0, nil, fmt.Errorf("count bought vods: %w", err)
	}
	offset := limitOffset(total, pageSize, page)
	rows, err := r.queryRows(
		ctx,
		"SELECT a.*, b.* FROM user_bought a LEFT JOIN vods b ON b.vodid=a.vodid WHERE a.uid=? AND b.showtype=0 ORDER BY a.buytime DESC LIMIT ? OFFSET ?",
		uid,
		pageSize,
		offset,
	)
	if err != nil {
		return 0, nil, fmt.Errorf("list bought vods: %w", err)
	}
	return total, rows, nil
}

func (r *Repository) Delete(ctx context.Context, uid int, vodid int) error {
	if r.db == nil || uid <= 0 || vodid <= 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "DELETE FROM user_bought WHERE uid=? AND vodid=?", uid, vodid); err != nil {
		return fmt.Errorf("delete bought vod: %w", err)
	}
	return nil
}

func (r *Repository) queryRows(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
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
