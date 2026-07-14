package game

import (
	"context"
	"strings"
	"testing"
)

type fakeGameStore struct {
	platformID int
	categoryID int
	rows       []map[string]interface{}
}

func (s *fakeGameStore) ListActive(_ context.Context, platformID int, categoryID int) ([]map[string]interface{}, error) {
	s.platformID = platformID
	s.categoryID = categoryID
	return s.rows, nil
}

type fakeBroadcastStore struct {
	rows []map[string]interface{}
}

func (s fakeBroadcastStore) ListActive(context.Context) ([]map[string]interface{}, error) {
	return s.rows, nil
}

func TestListingServiceList(t *testing.T) {
	store := &fakeGameStore{
		rows: []map[string]interface{}{
			{
				"id":    "1",
				"icon":  "/icon.png",
				"image": "image.png",
				"cover": "https://static.example.test/cover.png",
			},
		},
	}
	service := NewListingService(store, "https://res.example.test")

	data, err := service.List(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("list games: %v", err)
	}
	if store.platformID != 1 || store.categoryID != 2 {
		t.Fatalf("unexpected filters %d %d", store.platformID, store.categoryID)
	}
	row := data.Data[0]
	if row["icon"] != "https://res.example.test/icon.png" {
		t.Fatalf("unexpected icon %v", row["icon"])
	}
	if row["image"] != "https://res.example.test/image.png" {
		t.Fatalf("unexpected image %v", row["image"])
	}
	if row["cover"] != "https://static.example.test/cover.png" {
		t.Fatalf("unexpected cover %v", row["cover"])
	}
}

func TestBroadcastServiceList(t *testing.T) {
	service := NewBroadcastService(fakeBroadcastStore{
		rows: []map[string]interface{}{
			{
				"msg":       "用户{user}赢得{amount}金币",
				"min_value": "100",
				"max_value": "100",
			},
		},
	})

	data, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("list broadcasts: %v", err)
	}
	if len(data.Data) != 1 {
		t.Fatalf("expected one message, got %d", len(data.Data))
	}
	msg := data.Data[0]
	if strings.Contains(msg, "{user}") || strings.Contains(msg, "{amount}") {
		t.Fatalf("template placeholders were not replaced: %q", msg)
	}
	if !strings.Contains(msg, "100金币") {
		t.Fatalf("unexpected message %q", msg)
	}
}
