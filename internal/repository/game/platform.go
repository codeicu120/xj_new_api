package game

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PlatformRepository struct {
	db *sql.DB
}

func NewPlatformRepository(db *sql.DB) *PlatformRepository {
	return &PlatformRepository{db: db}
}

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

type GameRepository struct {
	db *sql.DB
}

func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{db: db}
}

type BroadcastRepository struct {
	db *sql.DB
}

func NewBroadcastRepository(db *sql.DB) *BroadcastRepository {
	return &BroadcastRepository{db: db}
}

func (r *CategoryRepository) ListActive(ctx context.Context, parentID int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}

	query := "SELECT * FROM game_category WHERE 1=1 AND status=1"
	args := []interface{}{}
	if parentID != 0 {
		query += " AND parent_id=?"
		args = append(args, parentID)
	}
	query += " ORDER BY `order` DESC LIMIT 100"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query game categories: %w", err)
	}
	defer rows.Close()

	return scanRows(rows)
}

func (r *PlatformRepository) ListActive(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}

	rows, err := r.db.QueryContext(ctx, "SELECT * FROM game_platform WHERE 1=1 AND status=1 ORDER BY `order` DESC LIMIT 100")
	if err != nil {
		return nil, fmt.Errorf("query game platforms: %w", err)
	}
	defer rows.Close()

	return scanRows(rows)
}

func (r *PlatformRepository) PlatformByID(ctx context.Context, id int) (map[string]interface{}, error) {
	if r.db == nil || id <= 0 {
		return map[string]interface{}{}, nil
	}

	rows, err := r.db.QueryContext(ctx, "SELECT id,name,slug,status,json AS config_json FROM game_platform WHERE id=?", id)
	if err != nil {
		return nil, fmt.Errorf("query game platform: %w", err)
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

func (r *GameRepository) ListActive(ctx context.Context, platformID int, categoryID int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}

	query := "SELECT * FROM game WHERE 1=1 AND status=1"
	args := []interface{}{}
	if platformID != 0 {
		query += " AND platform_id=?"
		args = append(args, platformID)
	}
	if categoryID != 0 {
		query += " AND category_id=?"
		args = append(args, categoryID)
	}
	query += " ORDER BY `order` DESC LIMIT 100"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query games: %w", err)
	}
	defer rows.Close()

	return scanRows(rows)
}

func (r *BroadcastRepository) ListActive(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}

	rows, err := r.db.QueryContext(ctx, "SELECT * FROM game_broadcast WHERE 1=1 AND status=1 ORDER BY id DESC LIMIT 100")
	if err != nil {
		return nil, fmt.Errorf("query game broadcasts: %w", err)
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
			if column == "json" {
				continue
			}
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
