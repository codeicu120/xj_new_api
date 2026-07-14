package amazing

import (
	"context"
	"fmt"

	"xj_comp/internal/domain"
)

type CategoryStore interface {
	ListActive(ctx context.Context, parentID int) ([]map[string]interface{}, error)
}

type CategoryService struct {
	store CategoryStore
}

func NewCategoryService(store CategoryStore) *CategoryService {
	return &CategoryService{store: store}
}

func (s *CategoryService) List(ctx context.Context, parentID int) (domain.AmazingCategoriesData, error) {
	rows, err := s.store.ListActive(ctx, parentID)
	if err != nil {
		return domain.AmazingCategoriesData{}, fmt.Errorf("list amazing categories: %w", err)
	}
	return domain.AmazingCategoriesData{Rows: rows}, nil
}
