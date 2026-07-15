package community

import (
	"context"
	"testing"
	"time"

	"xj_comp/internal/domain"
	communityRepo "xj_comp/internal/repository/community"
)

type fakeStore struct {
	filter        communityRepo.TopicFilter
	order         string
	missingTopic  bool
	visitCount    int
	commentOrder  string
	topicFavorite bool
	favoriteDelta int
	topicUp       bool
	commentUp     bool
	topicDelta    int
	commentDelta  int
	recentByUID   []map[string]interface{}
	recentByIP    []map[string]interface{}
	created       *domain.CommunityCommentCreateInput
	commentCount  int
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
	if s.topicFavorite {
		return map[int]int{9: 1}, nil
	}
	return map[int]int{}, nil
}
func (s *fakeStore) SetTopicFavorite(_ context.Context, _ int, _ int, favorite bool, _ int64) (int, error) {
	if !favorite && s.topicFavorite {
		s.topicFavorite = false
		return 1, nil
	}
	if favorite && !s.topicFavorite {
		s.topicFavorite = true
		return 1, nil
	}
	return 0, nil
}
func (s *fakeStore) IncrementTopicFavorite(_ context.Context, _ int, delta int) error {
	s.favoriteDelta += delta
	return nil
}
func (s *fakeStore) UpTopicIDs(context.Context, int, []int) (map[int]int, error) {
	if s.topicUp {
		return map[int]int{9: 1}, nil
	}
	return map[int]int{}, nil
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
func (s *fakeStore) SetTopicUp(_ context.Context, _ int, _ int, up bool, _ int64) error {
	s.topicUp = up
	return nil
}
func (s *fakeStore) IncrementTopicUp(_ context.Context, _ int, delta int) error {
	s.topicDelta += delta
	return nil
}
func (s *fakeStore) CountComments(context.Context, int) (int, error) { return 1, nil }
func (s *fakeStore) ListComments(_ context.Context, _ int, _ int, _ int, _ int, orderBy string) ([]map[string]interface{}, error) {
	s.commentOrder = orderBy
	return []map[string]interface{}{{"id": "1", "tid": "9", "uid": "7", "addtime": "1699999940", "subrows": []map[string]interface{}{}}}, nil
}
func (s *fakeStore) UpCommentIDs(context.Context, int, []int) (map[int]int, error) {
	if s.commentUp {
		return map[int]int{1: 1}, nil
	}
	return map[int]int{}, nil
}
func (s *fakeStore) CommentByID(context.Context, int) (map[string]interface{}, error) {
	return map[string]interface{}{"id": "1", "tid": "9"}, nil
}
func (s *fakeStore) SetCommentUp(_ context.Context, _ int, _ int, up bool, _ int64) error {
	s.commentUp = up
	return nil
}
func (s *fakeStore) IncrementCommentUp(_ context.Context, _ int, delta int) error {
	s.commentDelta += delta
	return nil
}
func (s *fakeStore) RecentCommentsByUID(context.Context, int, int64) ([]map[string]interface{}, error) {
	return s.recentByUID, nil
}
func (s *fakeStore) RecentCommentsByIP(context.Context, string, int64) ([]map[string]interface{}, error) {
	return s.recentByIP, nil
}
func (s *fakeStore) CreateComment(_ context.Context, input domain.CommunityCommentCreateInput, _ map[string]interface{}) (int, error) {
	s.created = &input
	return 99, nil
}
func (s *fakeStore) IncrementTopicCommentCount(context.Context, int) error {
	s.commentCount++
	return nil
}

type fakeAuth struct{ user map[string]interface{} }

func (a fakeAuth) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return a.user, nil
}

func TestListingBuildsParamsAndRows(t *testing.T) {
	store := &fakeStore{topicFavorite: true, topicUp: true}
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
	store := &fakeStore{commentUp: true}
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
	store := &fakeStore{topicFavorite: true, topicUp: true, commentUp: true}
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

func TestUpTopicRequiresLogin(t *testing.T) {
	service := NewService(fakeAuth{}, &fakeStore{}, "https://res.test")

	retcode, errmsg, err := service.UpTopic(context.Background(), "", 9)
	if err != nil {
		t.Fatalf("up topic: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作" {
		t.Fatalf("response=%d %q", retcode, errmsg)
	}
}

func TestAttentionRequiresLogin(t *testing.T) {
	service := NewService(fakeAuth{}, &fakeStore{}, "https://res.test")

	retcode, errmsg, err := service.Attention(context.Background(), "", 9, nil)
	if err != nil {
		t.Fatalf("attention: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作" {
		t.Fatalf("response=%d %q", retcode, errmsg)
	}
}

func TestAttentionTogglesFavorite(t *testing.T) {
	store := &fakeStore{}
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, store, "https://res.test")

	retcode, errmsg, err := service.Attention(context.Background(), "abc", 9, nil)
	if err != nil {
		t.Fatalf("attention: %v", err)
	}
	if retcode != 0 || errmsg != "收藏成功" || !store.topicFavorite || store.favoriteDelta != 1 {
		t.Fatalf("response=%d %q favorite=%v delta=%d", retcode, errmsg, store.topicFavorite, store.favoriteDelta)
	}
}

func TestAttentionBatchCancelsFavorites(t *testing.T) {
	store := &fakeStore{topicFavorite: true}
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, store, "https://res.test")

	retcode, errmsg, err := service.Attention(context.Background(), "abc", 0, []int{8, 9})
	if err != nil {
		t.Fatalf("attention batch: %v", err)
	}
	if retcode != 0 || errmsg != "批量取消收藏成功" || store.topicFavorite || store.favoriteDelta != -1 {
		t.Fatalf("response=%d %q favorite=%v delta=%d", retcode, errmsg, store.topicFavorite, store.favoriteDelta)
	}
}

func TestUpTopicToggles(t *testing.T) {
	store := &fakeStore{}
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, store, "https://res.test")

	retcode, errmsg, err := service.UpTopic(context.Background(), "abc", 9)
	if err != nil {
		t.Fatalf("up topic: %v", err)
	}
	if retcode != 0 || errmsg != "已赞" || !store.topicUp || store.topicDelta != 1 {
		t.Fatalf("response=%d %q up=%v delta=%d", retcode, errmsg, store.topicUp, store.topicDelta)
	}
}

func TestUpCommentCancelsExisting(t *testing.T) {
	store := &fakeStore{commentUp: true}
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, store, "https://res.test")

	retcode, errmsg, err := service.UpComment(context.Background(), "abc", 1)
	if err != nil {
		t.Fatalf("up comment: %v", err)
	}
	if retcode != 0 || errmsg != "取消赞成功" || store.commentUp || store.commentDelta != -1 {
		t.Fatalf("response=%d %q up=%v delta=%d", retcode, errmsg, store.commentUp, store.commentDelta)
	}
}

func TestCommentRequiresLogin(t *testing.T) {
	service := NewService(fakeAuth{}, &fakeStore{}, "https://res.test")

	retcode, errmsg, err := service.Comment(context.Background(), "", 9, 0, "hello", "127.0.0.1")
	if err != nil {
		t.Fatalf("comment: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作" {
		t.Fatalf("response=%d %q", retcode, errmsg)
	}
}

func TestCommentRejectsInvalidContent(t *testing.T) {
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7", "nickname": "nick", "perms": map[string]interface{}{}}}, &fakeStore{}, "https://res.test")

	retcode, errmsg, err := service.Comment(context.Background(), "abc", 9, 0, "", "127.0.0.1")
	if err != nil {
		t.Fatalf("comment: %v", err)
	}
	if retcode != 4 || errmsg != "评论允许1-30字之间" {
		t.Fatalf("response=%d %q", retcode, errmsg)
	}
}

func TestCommentRejectsDuplicateContent(t *testing.T) {
	store := &fakeStore{recentByUID: []map[string]interface{}{{"content": "这是一条社区评论"}}}
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7", "nickname": "nick", "perms": map[string]interface{}{}}}, store, "https://res.test")

	retcode, errmsg, err := service.Comment(context.Background(), "abc", 9, 0, "这是一条社区评论", "127.0.0.1")
	if err != nil {
		t.Fatalf("comment: %v", err)
	}
	if retcode != 10 || errmsg != "请勿发布重复内容1" {
		t.Fatalf("response=%d %q", retcode, errmsg)
	}
}

func TestCommentCreatesPendingComment(t *testing.T) {
	store := &fakeStore{}
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7", "nickname": "nick", "perms": map[string]interface{}{}}}, store, "https://res.test")
	service.now = func() time.Time { return time.Unix(2000, 0) }

	retcode, errmsg, err := service.Comment(context.Background(), "abc", 9, 0, "不错", "127.0.0.1")
	if err != nil {
		t.Fatalf("comment: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("response=%d %q", retcode, errmsg)
	}
	if store.created == nil || store.created.TID != 9 || store.created.UID != 7 || store.created.ShowType != 4 || store.created.Content != "不错" {
		t.Fatalf("created=%#v", store.created)
	}
	if store.commentCount != 1 {
		t.Fatalf("comment count=%d", store.commentCount)
	}
}
