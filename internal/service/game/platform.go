package game

import (
	"context"
	"fmt"

	"xj_comp/internal/domain"
)

type PlatformStore interface {
	ListActive(ctx context.Context) ([]map[string]interface{}, error)
}

type PlatformService struct {
	store PlatformStore
}

func NewPlatformService(store PlatformStore) *PlatformService {
	return &PlatformService{store: store}
}

func (s *PlatformService) List(ctx context.Context) (domain.GamePlatformsData, error) {
	rows, err := s.store.ListActive(ctx)
	if err != nil {
		return domain.GamePlatformsData{}, fmt.Errorf("list game platforms: %w", err)
	}
	return domain.GamePlatformsData{Data: rows}, nil
}
