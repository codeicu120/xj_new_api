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
