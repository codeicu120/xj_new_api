package minivod

import (
	"context"
	"testing"
	"time"

	minivodRepo "xj_comp/internal/repository/minivod"
)

type fakeStore struct {
	filter     minivodRepo.Filter
	order      string
	randomUsed bool
	actionKey  string
	user       map[string]interface{}
	vod        map[string]interface{}
}

func (s *fakeStore) Categories(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"cateid": "1", "parentid": "0", "catename": "主类"}, {"cateid": "2", "parentid": "1", "catename": "子类"}}, nil
}

func (s *fakeStore) Areas(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"areaid": "3", "areaname": "日本"}}, nil
}

func (s *fakeStore) Years(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"yearid": "4", "yearname": "2025"}}, nil
}

func (s *fakeStore) Servers(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (s *fakeStore) TagsByNames(context.Context, []string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (s *fakeStore) Count(_ context.Context, filter minivodRepo.Filter, _ int64) (int, error) {
	s.filter = filter
	return 1, nil
}

func (s *fakeStore) List(_ context.Context, filter minivodRepo.Filter, _ int, _ int, _ int, orderBy string, _ int64) ([]map[string]interface{}, error) {
	s.filter = filter
	s.order = orderBy
	return []map[string]interface{}{{"vodid": "9", "authorid": "7", "title": "mini"}}, nil
}

func (s *fakeStore) CountByAuthor(context.Context, int) (int, error) {
	return 1, nil
}

func (s *fakeStore) ListByAuthor(context.Context, int, int, int, int, string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"vodid": "9", "authorid": "7", "title": "mini"}}, nil
}

func (s *fakeStore) Random(context.Context, int) ([]map[string]interface{}, error) {
	s.randomUsed = true
	return []map[string]interface{}{{"vodid": "9", "authorid": "7", "title": "mini"}}, nil
}

func (s *fakeStore) VODByID(context.Context, int) (map[string]interface{}, error) {
	if s.vod != nil {
		return s.vod, nil
	}
	return map[string]interface{}{"vodid": "9", "authorid": "7", "title": "mini", "showtype": "1", "cateid": "2", "tags": "tag1", "actor_tags": ""}, nil
}

func (s *fakeStore) UserByID(context.Context, int) (map[string]interface{}, error) {
	if s.user != nil {
		return s.user, nil
	}
	return map[string]interface{}{"uid": "7", "username": "u", "nickname": "n", "avatar": "avatar.jpg", "gender": "1"}, nil
}

func (s *fakeStore) SimilarVODsByTagIDs(context.Context, []int, int, int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"vodid": "10", "authorid": "8", "title": "similar"}}, nil
}

func (s *fakeStore) RandomVODsExcept(_ context.Context, pageSize int, _ int, _ int) ([]map[string]interface{}, error) {
	rows := []map[string]interface{}{}
	for i := 0; i < pageSize; i++ {
		rows = append(rows, map[string]interface{}{"vodid": "20", "authorid": "8", "title": "random"})
	}
	return rows, nil
}

func (s *fakeStore) Setting(_ context.Context, key string) (string, error) {
	s.actionKey = key
	return "9,8", nil
}

func (s *fakeStore) UsersByIDs(context.Context, []int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"uid": "7", "username": "u", "nickname": "n", "avatar": "avatar.jpg", "gender": "1"}}, nil
}

type fakeProcessor struct{}

func (fakeProcessor) ProcessRowsFullPrice(_ context.Context, rows []map[string]interface{}, _ bool) ([]map[string]interface{}, error) {
	for _, row := range rows {
		row["processed"] = "1"
	}
	return rows, nil
}

func (fakeProcessor) ProcessMiniRowsFullPrice(_ context.Context, rows []map[string]interface{}, _ bool) ([]map[string]interface{}, error) {
	for _, row := range rows {
		row["processed"] = "1"
	}
	return rows, nil
}

func TestListingFiltersAndPageURL(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, fakeProcessor{}, "https://res.test")
	service.now = func() time.Time { return time.Unix(1700000000, 0) }

	data, err := service.Listing(context.Background(), ListingRequest{Action: "listing", PathParams: "1-3-4-5-2-1-1-2-3-2-2"})
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if store.order != "playcount_total DESC" || len(store.filter.CateIDs) != 2 || store.filter.TagIDs[0] != 5 || !store.filter.FreeOnly {
		t.Fatalf("filter=%#v order=%q", store.filter, store.order)
	}
	if data.PageInfo["page_url"] != "/minivod/listing-1-3-4-5-2-1-1-2-3-2-[?]" {
		t.Fatalf("pageinfo = %#v", data.PageInfo)
	}
	if len(data.Rows) != 1 || len(data.VODRows) != 1 || data.VODRows[0]["processed"] != "1" {
		t.Fatalf("data rows=%#v vodrows=%#v", data.Rows, data.VODRows)
	}
}

func TestRecommendUsesRandomRows(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, fakeProcessor{}, "https://res.test")

	data, err := service.Listing(context.Background(), ListingRequest{Action: "recommend"})
	if err != nil {
		t.Fatalf("recommend: %v", err)
	}
	if !store.randomUsed || data.PageInfo["total"] != 0 {
		t.Fatalf("random=%v pageinfo=%#v", store.randomUsed, data.PageInfo)
	}
}

func TestTopRowsIncludeUser(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, fakeProcessor{}, "https://res.test")

	data, err := service.Listing(context.Background(), ListingRequest{Action: "topzan"})
	if err != nil {
		t.Fatalf("topzan: %v", err)
	}
	if store.actionKey != "minivod.zan_vodids" || store.order != "FIELD(vodid, 9,8)" {
		t.Fatalf("setting=%q order=%q", store.actionKey, store.order)
	}
	if len(data.Rows) != 1 || data.Rows[0]["user"] == nil {
		t.Fatalf("rows=%#v", data.Rows)
	}
}

func TestShowReturnsDetailRowsAndUser(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, fakeProcessor{}, "https://res.test")

	data, err := service.Show(context.Background(), 9, false)
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if data.VODRow["vodid"] != "9" || data.VODRow["processed"] != "1" {
		t.Fatalf("vodrow=%#v", data.VODRow)
	}
	if len(data.Categories) != 2 || data.Categories[0]["cateid"] != "1" || data.Categories[1]["cateid"] != "2" {
		t.Fatalf("categories=%#v", data.Categories)
	}
	if len(data.SimilarRows) != 10 || len(data.LikeRows) != 5 {
		t.Fatalf("similar=%d like=%d", len(data.SimilarRows), len(data.LikeRows))
	}
	if data.VODUser["uid"] != "7" || data.VODUser["avatar_url"] != "https://res.test/C1/avatar/avatar.jpg" {
		t.Fatalf("voduser=%#v", data.VODUser)
	}
}

func TestShowRejectsMissingAuthor(t *testing.T) {
	store := &fakeStore{user: map[string]interface{}{}}
	service := NewService(store, fakeProcessor{}, "https://res.test")

	_, err := service.Show(context.Background(), 9, false)
	if err != ErrAuthorNotFound {
		t.Fatalf("expected ErrAuthorNotFound, got %v", err)
	}
}

func TestAuthorListingReturnsUserRowsAndMiniVODRows(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, fakeProcessor{}, "https://res.test")
	service.now = func() time.Time { return time.Unix(1700000000, 0) }

	data, err := service.AuthorListing(context.Background(), 7, 1, false)
	if err != nil {
		t.Fatalf("author listing: %v", err)
	}
	if data.UserRow["uid"] != "7" || data.PageInfo["total"] != 1 {
		t.Fatalf("data=%#v", data)
	}
	if len(data.VODRows) != 1 || data.VODRows[0]["processed"] != "1" {
		t.Fatalf("vodrows=%#v", data.VODRows)
	}
}
