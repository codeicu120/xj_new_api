package so

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"xj_comp/internal/domain"
)

type ConfigStore interface {
	FindValue(ctx context.Context, version int, arm string, channel string) (string, error)
}

type ConfigService struct {
	store ConfigStore
}

func NewConfigService(store ConfigStore) *ConfigService {
	return &ConfigService{store: store}
}

func (s *ConfigService) List(ctx context.Context, version int, arm string, channel string) (domain.SOListData, error) {
	value, err := s.store.FindValue(ctx, version, sanitizeInput(arm), sanitizeInput(channel))
	if err != nil {
		return domain.SOListData{}, fmt.Errorf("find so config: %w", err)
	}

	data := decodePHPJSON(value)
	return domain.SOListData{Data: data}, nil
}

func sanitizeInput(value string) string {
	value = strings.ReplaceAll(value, "\x00", "")
	value = strings.ReplaceAll(value, "<", "")
	value = strings.ReplaceAll(value, ">", "")
	return value
}

func decodePHPJSON(value string) interface{} {
	if value == "" {
		return nil
	}

	var data interface{}
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return nil
	}
	return data
}
