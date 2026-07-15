package community

import (
	"context"
	"testing"
	"time"

	communityRepo "xj_comp/internal/repository/community"
)

type fakeStore struct {
	filter       communityRepo.TopicFilter
	order        string
	missingTopic bool
	visitCount   int
	commentOrder string
}

func (s *fakeStore) CountTopics(_ context.Context, filter communityRepo.TopicFilter) (int, error) {
	s.filter = filter
	return 1, nil
}

func (s *fakeStore) ListTopics(_ context.Context, filter communityRepo.TopicFilter, _ int, _ int, _ int, orderBy string) ([]map[string]interface{}, error) {
	s.filter = filter
	s.order = orderBy
	return []map[string]interface{}{{"tid": "9", "content": `<p><img src="a.jpg"></p>`, "image_srvid": "0", "video_srvid": "0"}}, nil
}

func (s *fakeStore) Servers(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}
func (s *fakeStore) ImagesByTIDs(context.Context, []int) (map[int][]map[string]interface{}, error) {
	return map[int][]map[string]interface{}{9: []map[string]interface{}{{"tid": "9", "image_path": "img.jpg"}}}, nil
}
func (s *fakeStore) VideosByTIDs(context.Context, []int) (map[int][]map[string]interface{}, error) {
	return map[int][]map[string]interface{}{}, nil
}
func (s *fakeStore) FavoriteTopicIDs(context.Context, int, []int) (map[int]int, error) {
	return map[int]int{9: 1}, nil
}
func (s *fakeStore) UpTopicIDs(context.Context, int, []int) (map[int]int, error) {
	return map[int]int{9: 1}, nil
}
func (s *fakeStore) TopicByID(context.Context, int) (map[string]interface{}, error) {
	if s.missingTopic {
		return map[string]interface{}{}, nil
	}
	return map[string]interface{}{"tid": "9", "content": `<p><img src="a.jpg"></p>`, "image_srvid": "0", "video_srvid": "0"}, nil
}
func (s *fakeStore) IncrementTopicVisit(context.Context, int) error {
	s.visitCount++
	return nil
}
func (s *fakeStore) CountComments(context.Context, int) (int, error) { return 1, nil }
func (s *fakeStore) ListComments(_ context.Context, _ int, _ int, _ int, _ int, orderBy string) ([]map[string]interface{}, error) {
	s.commentOrder = orderBy
	return []map[string]interface{}{{"id": "1", "tid": "9", "uid": "7", "addtime": "1699999940", "subrows": []map[string]interface{}{}}}, nil
}
func (s *fakeStore) UpCommentIDs(context.Context, int, []int) (map[int]int, error) {
	return map[int]int{1: 1}, nil
}

type fakeAuth struct{ user map[string]interface{} }

func (a fakeAuth) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return a.user, nil
}

func TestListingBuildsParamsAndRows(t *testing.T) {
	store := &fakeStore{}
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, store, "https://res.test")
	service.now = func() time.Time { return time.Unix(1700000000, 0) }

	data, err := service.Listing(context.Background(), ListingRequest{Action: "favorite", PathParams: "3-2-0-1", Token: "abc"})
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if store.filter.FavoriteUID != 7 || store.filter.CategoryID != 3 || store.filter.Type != 2 {
		t.Fatalf("filter=%#v", store.filter)
	}
	if data.PageInfo["page_url"] != "/community/favorite-3-2-0-[?]" || len(data.Rows) != 1 {
		t.Fatalf("data=%#v", data)
	}
	if data.Rows[0]["is_favorite"] != 1 || data.Rows[0]["is_up"] != 1 {
		t.Fatalf("row=%#v", data.Rows[0])
	}
}

func TestFavoriteRequiresLogin(t *testing.T) {
	service := NewService(fakeAuth{}, &fakeStore{}, "https://res.test")
	_, err := service.Listing(context.Background(), ListingRequest{Action: "favorite"})
	if err != ErrLoginRequired {
		t.Fatalf("expected ErrLoginRequired, got %v", err)
	}
}

func TestCommentListing(t *testing.T) {
	store := &fakeStore{}
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, store, "https://res.test")
	service.now = func() time.Time { return time.Unix(1700000000, 0) }

	data, err := service.CommentListing(context.Background(), CommentListingRequest{TID: 9, QueryPage: "2", QueryOrder: "1", Token: "abc"})
	if err != nil {
		t.Fatalf("clisting: %v", err)
	}
	if data.PageInfo["page_url"] != "/community/clisting-1-[?]" || data.Rows[0]["is_up"] != 1 {
		t.Fatalf("data=%#v", data)
	}
}

func TestShowReturnsTopicAndComments(t *testing.T) {
	store := &fakeStore{}
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, store, "https://res.test")

	data, err := service.Show(context.Background(), ShowRequest{TID: 9, QueryOrder: "1", Token: "abc"})
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if store.visitCount != 1 || store.commentOrder != "a.upnum DESC" {
		t.Fatalf("visit=%d order=%q", store.visitCount, store.commentOrder)
	}
	row := data["row"].(map[string]interface{})
	if row["is_favorite"] != 1 || row["is_up"] != 1 {
		t.Fatalf("row=%#v", row)
	}
	if data["totalCommentCount"] != 1 || len(data["comments"].([]map[string]interface{})) != 1 {
		t.Fatalf("data=%#v", data)
	}
}

func TestShowMissingTopic(t *testing.T) {
	service := NewService(fakeAuth{}, &fakeStore{missingTopic: true}, "https://res.test")

	_, err := service.Show(context.Background(), ShowRequest{TID: 404})
	if err != ErrTopicNotFound {
		t.Fatalf("expected ErrTopicNotFound, got %v", err)
	}
}
