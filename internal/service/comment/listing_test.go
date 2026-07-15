package comment

import (
	"context"
	"testing"
	"time"
)

type fakeStore struct {
	lastOrder string
	vod       map[string]interface{}
	comment   map[string]interface{}
	voted     []string
}

func (s *fakeStore) VODByID(context.Context, int) (map[string]interface{}, error) {
	if s.vod != nil {
		return s.vod, nil
	}
	return map[string]interface{}{"vodid": "61494", "showtype": "0"}, nil
}

func (s *fakeStore) UserGroups(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"gid": "1", "gicon": "V1"}, {"gid": "6", "gicon": "V6"}}, nil
}

func (s *fakeStore) CountRoots(context.Context, int) (int, error) {
	return 31, nil
}

func (s *fakeStore) RootComments(_ context.Context, _ int, _ int, _ int, _ int, orderBy string) ([]map[string]interface{}, error) {
	s.lastOrder = orderBy
	return []map[string]interface{}{
		{
			"id":             "1",
			"rootid":         "0",
			"parentid":       "0",
			"lft":            "1",
			"rgt":            "4",
			"depth":          "0",
			"vodid":          "61494",
			"uid":            "5",
			"sid":            "5",
			"username":       "~user",
			"nickname":       "nick",
			"avatar":         "",
			"gender":         "1",
			"sysgid":         "6",
			"sysgid_exptime": "3000",
			"gid":            "1",
			"content":        "hello",
			"upnum":          "2",
			"downnum":        "0",
			"addtime":        "1000",
			"showtype":       "0",
			"__closenum__":   "1",
			"subrows":        []map[string]interface{}{{"id": "2", "rootid": "1", "parentid": "1", "lft": "2", "rgt": "3", "depth": "1", "vodid": "61494", "uid": "6", "sid": "6", "username": "~sub", "nickname": "", "avatar": "", "gender": "0", "sysgid": "0", "sysgid_exptime": "0", "gid": "1", "content": "sub", "upnum": "0", "downnum": "0", "addtime": "1000", "showtype": "0", "__closenum__": "0"}},
		},
	}, nil
}

func (s *fakeStore) CommentByID(context.Context, int) (map[string]interface{}, error) {
	if s.comment != nil {
		return s.comment, nil
	}
	return map[string]interface{}{}, nil
}

func (s *fakeStore) IncrementVote(_ context.Context, id int, field string) error {
	s.voted = append(s.voted, field)
	return nil
}

func TestListingProcessesCommentRows(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(2000, 0) }

	data, err := service.Listing(context.Background(), ListingRequest{PathParams: "61494-1-1"})
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if store.lastOrder != "a.upnum DESC" {
		t.Fatalf("unexpected order %q", store.lastOrder)
	}
	if len(data.Rows) != 1 {
		t.Fatalf("expected one row, got %d", len(data.Rows))
	}
	row := data.Rows[0]
	if row["avatar_url"] != "https://res.example.test/sysavatar/noavatar.png" {
		t.Fatalf("unexpected avatar_url %v", row["avatar_url"])
	}
	if row["isvip"] != 1 || row["gicon"] != "V6" {
		t.Fatalf("unexpected vip/gicon %#v", row)
	}
	if row["addtime"] != "16分钟前 " {
		t.Fatalf("unexpected addtime %v", row["addtime"])
	}
	if len(row["subrows"].([]map[string]interface{})) != 1 {
		t.Fatalf("expected one subrow, got %#v", row["subrows"])
	}
}

func TestListingMissingVOD(t *testing.T) {
	service := NewService(&fakeStore{vod: map[string]interface{}{}}, "https://res.example.test")

	_, err := service.Listing(context.Background(), ListingRequest{PathParams: "999999-0-1"})
	if err != ErrVODNotFound {
		t.Fatalf("expected ErrVODNotFound, got %v", err)
	}
}

func TestVoteWithoutTokenUsesGuestActor(t *testing.T) {
	service := NewService(&fakeStore{}, "https://res.example.test")

	retcode, errmsg, err := service.Vote(context.Background(), "", 1, true)
	if err != nil {
		t.Fatalf("vote: %v", err)
	}
	if retcode != -1 || errmsg != "记录不存在或已被删除" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestVoteMissingComment(t *testing.T) {
	service := NewService(&fakeStore{}, "https://res.example.test")

	retcode, errmsg, err := service.Vote(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1, true)
	if err != nil {
		t.Fatalf("vote: %v", err)
	}
	if retcode != -1 || errmsg != "记录不存在或已被删除" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestVoteSuccessAndDuplicate(t *testing.T) {
	store := &fakeStore{comment: map[string]interface{}{"id": "8"}}
	service := NewService(store, "https://res.example.test")
	token := "3235306637393062613731656332623964333835356634323464623232353965"

	retcode, errmsg, err := service.Vote(context.Background(), token, 8, true)
	if err != nil {
		t.Fatalf("vote: %v", err)
	}
	if retcode != 0 || errmsg != "已赞" || len(store.voted) != 1 || store.voted[0] != "upnum" {
		t.Fatalf("response = %d %q voted=%#v", retcode, errmsg, store.voted)
	}
	retcode, errmsg, err = service.Vote(context.Background(), token, 8, false)
	if err != nil {
		t.Fatalf("vote duplicate: %v", err)
	}
	if retcode != -1 || errmsg != "您已经赞/踩过了" {
		t.Fatalf("duplicate response = %d %q", retcode, errmsg)
	}
}
