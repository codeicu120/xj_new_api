package art

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

func (r *Repository) Categories(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT cateid,parentid,uuid,catename FROM art_categories ORDER BY sort ASC, cateid ASC")
	if err != nil {
		return nil, fmt.Errorf("query art categories: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) CountByCategory(ctx context.Context, cateID int) (int, error) {
	if r.db == nil || cateID <= 0 {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM arts WHERE 1=1 AND cateid=? AND showtype=0", cateID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count arts: %w", err)
	}
	return total, nil
}

func (r *Repository) ListByCategory(ctx context.Context, cateID int, total int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil || cateID <= 0 {
		return []map[string]interface{}{}, nil
	}
	offset := limitOffset(total, pageSize, page)
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM arts WHERE 1=1 AND cateid=? AND showtype=0 ORDER BY utimestamp DESC LIMIT ? OFFSET ?", cateID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query arts: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) ArtByID(ctx context.Context, artID int) (map[string]interface{}, error) {
	if r.db == nil || artID <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT a.*, c.content, c.aids FROM arts a LEFT JOIN arts_content c ON c.artid=a.artid WHERE a.artid=?", artID)
	if err != nil {
		return nil, fmt.Errorf("query art: %w", err)
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
