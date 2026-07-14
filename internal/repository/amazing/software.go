package amazing

import (
	"context"
	"database/sql"
	"fmt"
)

type SoftwareFilter struct {
	CategoryID  int
	IsRecommend bool
}

type SoftwareRepository struct {
	db *sql.DB
}

func NewSoftwareRepository(db *sql.DB) *SoftwareRepository {
	return &SoftwareRepository{db: db}
}

func (r *SoftwareRepository) CountActive(ctx context.Context, filter SoftwareFilter) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	where, args := buildSoftwareWhere(filter)
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM amazing WHERE 1=1 "+where, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count amazing software: %w", err)
	}
	return total, nil
}

func (r *SoftwareRepository) ListActive(ctx context.Context, filter SoftwareFilter, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	where, args := buildSoftwareWhere(filter)
	offset := limitOffset(total, pageSize, page)
	query := "SELECT * FROM amazing WHERE 1=1 " + where + " ORDER BY " + orderBy + " LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query amazing software: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func buildSoftwareWhere(filter SoftwareFilter) (string, []interface{}) {
	where := " AND status=1"
	args := []interface{}{}
	if filter.CategoryID != 0 {
		where += " AND category_id=?"
		args = append(args, filter.CategoryID)
	}
	if filter.IsRecommend {
		where += " AND is_recommend=1"
	}
	return where, args
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
