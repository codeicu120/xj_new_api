package starlive

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"xj_comp/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Info(ctx context.Context) (domain.StarLiveInfo, error) {
	if r.db == nil {
		return domain.StarLiveInfo{}, nil
	}
	row, err := r.queryOne(ctx, "SELECT appId, secKey, apiHost, env, src, liveHost FROM starlive_info LIMIT 1")
	if err != nil {
		return domain.StarLiveInfo{}, fmt.Errorf("query starlive info: %w", err)
	}
	if len(row) == 0 {
		return domain.StarLiveInfo{}, nil
	}
	return domain.StarLiveInfo{
		AppID:    fmt.Sprint(row["appId"]),
		SecKey:   fmt.Sprint(row["secKey"]),
		APIHost:  fmt.Sprint(row["apiHost"]),
		Env:      row["env"],
		Src:      fmt.Sprint(row["src"]),
		LiveHost: fmt.Sprint(row["liveHost"]),
	}, nil
}

func (r *Repository) queryOne(ctx context.Context, query string, args ...interface{}) (map[string]interface{}, error) {
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
		return map[string]interface{}{}, nil
	}
	return items[0], nil
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
