package comment

import (
	"context"
	"testing"
	"time"
)

type fakeStore struct {
	lastOrder string
	vod       map[string]interface{}
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
