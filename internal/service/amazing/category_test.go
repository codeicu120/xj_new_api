package amazing

import (
	"context"
	"testing"
)

type fakeCategoryStore struct {
	rows []map[string]interface{}
}

func (s fakeCategoryStore) ListActive(context.Context, int) ([]map[string]interface{}, error) {
	return s.rows, nil
}

func TestCategoryServiceList(t *testing.T) {
	service := NewCategoryService(fakeCategoryStore{
		rows: []map[string]interface{}{
			{"id": "4", "title": "漫画"},
		},
	})

	data, err := service.List(context.Background(), 0)
	if err != nil {
		t.Fatalf("list amazing categories: %v", err)
	}
	if len(data.Rows) != 1 {
		t.Fatalf("expected one row, got %d", len(data.Rows))
	}
	if data.Rows[0]["title"] != "漫画" {
		t.Fatalf("unexpected title %v", data.Rows[0]["title"])
	}
}
