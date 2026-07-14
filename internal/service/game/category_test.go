package game

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

func TestCategoryServiceListPrefixesResources(t *testing.T) {
	service := NewCategoryService(fakeCategoryStore{
		rows: []map[string]interface{}{
			{
				"id":    "-1",
				"icon":  "/202407/icon.png",
				"image": "/202407/image.png",
			},
		},
	}, "https://image.xjdev.one/")

	data, err := service.List(context.Background(), 0)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}

	row := data.Data[0]
	if row["icon"] != "https://image.xjdev.one/202407/icon.png" {
		t.Fatalf("unexpected icon %v", row["icon"])
	}
	if row["image"] != "https://image.xjdev.one/202407/image.png" {
		t.Fatalf("unexpected image %v", row["image"])
	}
}
