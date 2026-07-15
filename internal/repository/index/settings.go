package index

import (
	"context"
	"database/sql"
	"fmt"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) SettingValue(ctx context.Context) (string, error) {
	if r.db == nil {
		return "", nil
	}
	var value sql.NullString
	if err := r.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE uuid='setting'").Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("query setting: %w", err)
	}
	return value.String, nil
}

func (r *SettingsRepository) SettingByUUID(ctx context.Context, uuid string) (map[string]interface{}, error) {
	if r.db == nil || uuid == "" {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT uuid,value,type FROM settings WHERE uuid=?", uuid)
	if err != nil {
		return nil, fmt.Errorf("query setting %s: %w", uuid, err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *SettingsRepository) CalldataByUUID(ctx context.Context, uuid string) (map[string]interface{}, error) {
	if r.db == nil || uuid == "" {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM maintain_calldata WHERE uuid=?", uuid)
	if err != nil {
		return nil, fmt.Errorf("query calldata %s: %w", uuid, err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *SettingsRepository) queryRows(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
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
	return result, rows.Err()
}

func normalizeSQLValue(value interface{}) interface{} {
	switch v := value.(type) {
	case nil:
		return ""
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}
