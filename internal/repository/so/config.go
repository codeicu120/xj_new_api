package so

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type ConfigRepository struct {
	db *sql.DB
}

func NewConfigRepository(db *sql.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

func (r *ConfigRepository) FindValue(ctx context.Context, version int, arm string, channel string) (string, error) {
	if r.db == nil {
		return "", nil
	}

	var value string
	err := r.db.QueryRowContext(
		ctx,
		"SELECT `value` FROM server_so_config WHERE arm=? AND version > ? AND channel=? LIMIT 1",
		arm,
		version,
		channel,
	).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("query so config: %w", err)
	}
	return value, nil
}
