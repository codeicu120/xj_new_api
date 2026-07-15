package vod

import (
	"context"
	"testing"
	"time"

	vodRepo "xj_comp/internal/repository/vod"
)

type fakeSpecialVoteStore struct {
	row        map[string]interface{}
	guest      map[string]interface{}
	keyCount   int
	votedField string
	keySet     string
}

func (s *fakeSpecialVoteStore) Categories(context.Context) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s *fakeSpecialVoteStore) Areas(context.Context) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s *fakeSpecialVoteStore) Years(context.Context) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s *fakeSpecialVoteStore) Servers(context.Context) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s *fakeSpecialVoteStore) TagsByNames(context.Context, []string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s *fakeSpecialVoteStore) CountSpecials(context.Context, vodRepo.SpecialFilter) (int, error) {
	return 0, nil
}
func (s *fakeSpecialVoteStore) ListSpecials(context.Context, vodRepo.SpecialFilter, int, int, int, string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s *fakeSpecialVoteStore) ListActorSpecials(context.Context, int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s *fakeSpecialVoteStore) SpecialByID(context.Context, int) (map[string]interface{}, error) {
	return s.row, nil
}
func (s *fakeSpecialVoteStore) VODsByIDs(context.Context, []int, string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s *fakeSpecialVoteStore) UpdateSpecialRand(context.Context, int) error { return nil }
func (s *fakeSpecialVoteStore) IncrementSpecialViews(context.Context, int, int64, int64) error {
	return nil
}
func (s *fakeSpecialVoteStore) GuestBySID(context.Context, string) (map[string]interface{}, error) {
	return s.guest, nil
}
func (s *fakeSpecialVoteStore) KeylimitCount(context.Context, string) (int, error) {
	return s.keyCount, nil
}
func (s *fakeSpecialVoteStore) SetKeylimit(_ context.Context, key string, _ int, _ string, _ int64) error {
	s.keySet = key
	return nil
}
func (s *fakeSpecialVoteStore) IncrementSpecialVote(_ context.Context, _ int, field string) error {
	s.votedField = field
	return nil
}

type fakeSpecialAuthStore struct {
	user map[string]interface{}
}

func (s fakeSpecialAuthStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func TestSpecialVoteNotFound(t *testing.T) {
	service := NewSpecialService(&fakeSpecialVoteStore{}, nil, "", 100)

	retcode, errmsg, err := service.Vote(context.Background(), "", 1, "up")
	if err != ErrSpecialNotFound {
		t.Fatalf("expected ErrSpecialNotFound, got %v", err)
	}
	if retcode != -1 || errmsg != "记录不存在或已被删除" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestSpecialVoteGuestRequiresKnownGuest(t *testing.T) {
	store := &fakeSpecialVoteStore{row: map[string]interface{}{"spid": "3", "showtype": "0"}}
	service := NewSpecialService(store, nil, "", 100)

	retcode, errmsg, err := service.Vote(context.Background(), "", 3, "up")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作，客户端游客请先携带信息" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestSpecialVoteRejectsDuplicate(t *testing.T) {
	store := &fakeSpecialVoteStore{row: map[string]interface{}{"spid": "3", "showtype": "0"}, keyCount: 1}
	service := NewSpecialService(store, fakeSpecialAuthStore{user: map[string]interface{}{"uid": "8", "sid": "sid"}}, "", 100)

	retcode, errmsg, err := service.Vote(context.Background(), "sid", 3, "down")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if retcode != -1 || errmsg != "您已经赞/踩过了" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestSpecialVoteSuccess(t *testing.T) {
	store := &fakeSpecialVoteStore{row: map[string]interface{}{"spid": "3", "showtype": "0"}}
	service := NewSpecialService(store, fakeSpecialAuthStore{user: map[string]interface{}{"uid": "8", "sid": "sid"}}, "", 100)
	service.processor.now = func() time.Time { return time.Unix(1234, 0) }

	retcode, errmsg, err := service.Vote(context.Background(), "sid", 3, "down")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if retcode != 0 || errmsg != "已赞" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if store.votedField != "downnum" {
		t.Fatalf("expected downnum increment, got %q", store.votedField)
	}
	if store.keySet != "special.updown.3.8" {
		t.Fatalf("unexpected key %q", store.keySet)
	}
}
