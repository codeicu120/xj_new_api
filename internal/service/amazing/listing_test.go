package amazing

import (
	"context"
	"testing"
	"time"

	amazingRepo "xj_comp/internal/repository/amazing"
)

type fakeSoftwareStore struct {
	filter  amazingRepo.SoftwareFilter
	total   int
	page    int
	size    int
	orderBy string
	rows    []map[string]interface{}
}

func (s *fakeSoftwareStore) CountActive(_ context.Context, filter amazingRepo.SoftwareFilter) (int, error) {
	s.filter = filter
	if s.total == 0 {
		return len(s.rows), nil
	}
	return s.total, nil
}

func (s *fakeSoftwareStore) ListActive(_ context.Context, filter amazingRepo.SoftwareFilter, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error) {
	s.filter = filter
	s.total = total
	s.page = page
	s.size = pageSize
	s.orderBy = orderBy
	return s.rows, nil
}

func TestListingServiceList(t *testing.T) {
	store := &fakeSoftwareStore{
		rows: []map[string]interface{}{
			{"id": "11", "category_id": "3", "icon": "/icon.png", "image": "image.png"},
		},
	}
	service := NewListingService(store, "https://res.example.test/")
	service.now = func() time.Time { return time.Unix(1000, 0) }

	data, err := service.List(context.Background(), ListingRequest{
		Action:     "listing",
		PathParams: "3-0-0",
		QueryPage:  "2",
	})
	if err != nil {
		t.Fatalf("list amazing: %v", err)
	}

	if data.Now != 1000 {
		t.Fatalf("unexpected now %d", data.Now)
	}
	if store.filter.CategoryID != 3 {
		t.Fatalf("unexpected category id %d", store.filter.CategoryID)
	}
	if store.page != 2 {
		t.Fatalf("expected query page fallback, got %d", store.page)
	}
	if store.size != 20 {
		t.Fatalf("unexpected page size %d", store.size)
	}
	if store.orderBy != "id DESC" {
		t.Fatalf("unexpected order %q", store.orderBy)
	}
	row := data.Rows[0]
	if row["icon"] != "https://res.example.test//icon.png" {
		t.Fatalf("unexpected icon %v", row["icon"])
	}
	if row["image"] != "https://res.example.test/image.png" {
		t.Fatalf("unexpected image %v", row["image"])
	}
}

func TestListingServiceActionRules(t *testing.T) {
	tests := []struct {
		action      string
		order       string
		isRecommend bool
	}{
		{action: "recommend", order: "id DESC", isRecommend: true},
		{action: "hot", order: "dl_count DESC"},
		{action: "latest", order: "id DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			store := &fakeSoftwareStore{}
			service := NewListingService(store, "https://res.example.test")
			_, err := service.List(context.Background(), ListingRequest{Action: tt.action})
			if err != nil {
				t.Fatalf("list amazing: %v", err)
			}
			if store.orderBy != tt.order {
				t.Fatalf("expected order %q, got %q", tt.order, store.orderBy)
			}
			if store.filter.IsRecommend != tt.isRecommend {
				t.Fatalf("expected recommend %v, got %v", tt.isRecommend, store.filter.IsRecommend)
			}
		})
	}
}
