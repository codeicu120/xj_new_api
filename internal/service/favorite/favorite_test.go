package favorite

import (
	"context"
	"testing"

	favoriteRepo "xj_comp/internal/repository/favorite"
)

type fakeStore struct {
	user    map[string]interface{}
	kind    favoriteRepo.Kind
	uid     int
	keyword string
	total   int
	rows    []map[string]interface{}
	removed []int
}

func (s *fakeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s *fakeStore) Items(_ context.Context, kind favoriteRepo.Kind, uid int, _ int, _ int, keyword string) (int, []map[string]interface{}, error) {
	s.kind = kind
	s.uid = uid
	s.keyword = keyword
	return s.total, s.rows, nil
}

func (s *fakeStore) Remove(_ context.Context, kind favoriteRepo.Kind, uid int, vodid int) (int, error) {
	s.kind = kind
	s.uid = uid
	s.removed = append(s.removed, vodid)
	if vodid <= 0 {
		return 0, nil
	}
	return 1, nil
}

type fakeVODProcessor struct{}

func (fakeVODProcessor) ProcessRows(_ context.Context, rows []map[string]interface{}, _ bool) ([]map[string]interface{}, error) {
	for _, row := range rows {
		row["processed"] = "vod"
	}
	return rows, nil
}

func (fakeVODProcessor) ProcessRowsPlain(_ context.Context, rows []map[string]interface{}, _ bool) ([]map[string]interface{}, error) {
	for _, row := range rows {
		row["processed"] = "plain"
	}
	return rows, nil
}

func (fakeVODProcessor) ProcessMiniRows(_ context.Context, rows []map[string]interface{}, _ bool) ([]map[string]interface{}, error) {
	for _, row := range rows {
		row["processed"] = "mini"
	}
	return rows, nil
}

func (fakeVODProcessor) ProcessMiniRowsFullPrice(_ context.Context, rows []map[string]interface{}, _ bool) ([]map[string]interface{}, error) {
	for _, row := range rows {
		row["processed"] = "mini-full"
	}
	return rows, nil
}

func TestListingRequiresLogin(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, store, fakeVODProcessor{})

	_, retcode, errmsg, err := service.Listing(context.Background(), "", favoriteRepo.KindVOD, 1, "", false)
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestListingProcessesVODRowsAndKeywordPageURL(t *testing.T) {
	store := &fakeStore{
		user:  map[string]interface{}{"uid": "5"},
		total: 1,
		rows:  []map[string]interface{}{{"vodid": "9"}},
	}
	service := NewService(store, store, fakeVODProcessor{})

	data, retcode, errmsg, err := service.Listing(context.Background(), "token", favoriteRepo.KindVOD, 1, "abc", true)
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
	if store.kind != favoriteRepo.KindVOD || store.uid != 5 || store.keyword != "abc" {
		t.Fatalf("lookup = kind:%s uid:%d keyword:%q", store.kind, store.uid, store.keyword)
	}
	if len(data.Rows) != 1 || data.Rows[0]["processed"] != "plain" {
		t.Fatalf("rows = %#v", data.Rows)
	}
	if data.PageInfo["page_url"] != "/favorite/listing?page=[?]&wd=abc" {
		t.Fatalf("pageinfo = %#v", data.PageInfo)
	}
}

func TestMiniListingMarksFavorite(t *testing.T) {
	store := &fakeStore{
		user:  map[string]interface{}{"uid": "5"},
		total: 1,
		rows:  []map[string]interface{}{{"vodid": "9"}},
	}
	service := NewService(store, store, fakeVODProcessor{})

	data, _, _, err := service.Listing(context.Background(), "token", favoriteRepo.KindMini, 1, "", false)
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if data.PageInfo["page_url"] != "/minifavorite/listing?page=[?]" {
		t.Fatalf("pageinfo = %#v", data.PageInfo)
	}
	if data.Rows[0]["processed"] != "mini-full" || data.Rows[0]["isfavorite"] != 1 {
		t.Fatalf("rows = %#v", data.Rows)
	}
}

func TestRemoveRequiresLogin(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, store, nil)

	retcode, errmsg, err := service.Remove(context.Background(), "", favoriteRepo.KindVOD, []int{1})
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestRemoveCountsRows(t *testing.T) {
	store := &fakeStore{user: map[string]interface{}{"uid": "5"}}
	service := NewService(store, store, nil)

	retcode, errmsg, err := service.Remove(context.Background(), "token", favoriteRepo.KindMini, []int{1, 0, 2})
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if retcode != 0 || errmsg != "已删除2项" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
	if store.kind != favoriteRepo.KindMini || len(store.removed) != 3 {
		t.Fatalf("store = kind:%s removed:%#v", store.kind, store.removed)
	}
}
