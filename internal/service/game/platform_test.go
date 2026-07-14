package game

import (
	"context"
	"testing"
)

type fakePlatformStore struct {
	rows []map[string]interface{}
}

func (s fakePlatformStore) ListActive(context.Context) ([]map[string]interface{}, error) {
	return s.rows, nil
}

func TestPlatformServiceList(t *testing.T) {
	service := NewPlatformService(fakePlatformStore{
		rows: []map[string]interface{}{
			{
				"id":     "1",
				"name":   "瓦力游戏",
				"status": "1",
			},
		},
	})

	data, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("list platforms: %v", err)
	}
	if len(data.Data) != 1 {
		t.Fatalf("expected one row, got %d", len(data.Data))
	}
	if data.Data[0]["id"] != "1" {
		t.Fatalf("unexpected id %v", data.Data[0]["id"])
	}
}
