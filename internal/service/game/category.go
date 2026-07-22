package game

import (
	"context"
	"fmt"
	"strings"

	"xj_comp/internal/domain"
	"xj_comp/internal/service/resourceurl"
)

type CategoryStore interface {
	ListActive(ctx context.Context, parentID int) ([]map[string]interface{}, error)
}

type CategoryService struct {
	store           CategoryStore
	resourceBaseURL string
	resources       *resourceurl.Resolver
}

func (s *CategoryService) WithResourceResolver(r *resourceurl.Resolver) *CategoryService {
	s.resources = r
	return s
}

func NewCategoryService(store CategoryStore, resourceBaseURL string) *CategoryService {
	return &CategoryService{
		store:           store,
		resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"),
	}
}

func (s *CategoryService) List(ctx context.Context, parentID int) (domain.GameCategoriesData, error) {
	return s.ListForRequest(ctx, parentID, resourceurl.Request{})
}

func (s *CategoryService) ListForRequest(ctx context.Context, parentID int, req resourceurl.Request) (domain.GameCategoriesData, error) {
	rows, err := s.store.ListActive(ctx, parentID)
	if err != nil {
		return domain.GameCategoriesData{}, fmt.Errorf("list game categories: %w", err)
	}
	resolved := resourceurl.Resolved{BaseURL: s.resourceBaseURL}
	if s.resources != nil {
		resolved, err = s.resources.Resolve(ctx, req)
		if err != nil {
			return domain.GameCategoriesData{}, err
		}
	}
	for _, row := range rows {
		row["icon"] = resolved.GetRes(strings.TrimLeft(fmt.Sprint(row["icon"]), "/"), "")
		row["image"] = resolved.GetRes(strings.TrimLeft(fmt.Sprint(row["image"]), "/"), "")
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
