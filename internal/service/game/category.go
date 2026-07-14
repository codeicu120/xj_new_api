package game

import (
	"context"
	"fmt"
	"strings"

	"xj_comp/internal/domain"
)

type CategoryStore interface {
	ListActive(ctx context.Context, parentID int) ([]map[string]interface{}, error)
}

type CategoryService struct {
	store           CategoryStore
	resourceBaseURL string
}

func NewCategoryService(store CategoryStore, resourceBaseURL string) *CategoryService {
	return &CategoryService{
		store:           store,
		resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"),
	}
}

func (s *CategoryService) List(ctx context.Context, parentID int) (domain.GameCategoriesData, error) {
	rows, err := s.store.ListActive(ctx, parentID)
	if err != nil {
		return domain.GameCategoriesData{}, fmt.Errorf("list game categories: %w", err)
	}
	for _, row := range rows {
		prefixResource(row, "icon", s.resourceBaseURL)
		prefixResource(row, "image", s.resourceBaseURL)
	}
	return domain.GameCategoriesData{Data: rows}, nil
}

func prefixResource(row map[string]interface{}, field string, baseURL string) {
	value, ok := row[field].(string)
	if !ok || value == "" || baseURL == "" {
		return
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return
	}
	row[field] = baseURL + "/" + strings.TrimLeft(value, "/")
}
