package minivod

import (
	"context"
	"strconv"
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
	updown     map[string]interface{}
	viewlog    map[string]interface{}
	saved      int
	deleted    bool
	counters   []string
	daycount   int
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
	return []map[string]interface{}{{"srvid": "3", "srvhost": "https://cdn.test", "cdnkey": "", "cdnparam": ""}}, nil
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

func (s *fakeStore) UpDownByUser(context.Context, int, int) (map[string]interface{}, error) {
	if s.updown != nil {
		return s.updown, nil
	}
	return map[string]interface{}{}, nil
}

func (s *fakeStore) DeleteUpDown(context.Context, int, int) error {
	s.deleted = true
	s.updown = map[string]interface{}{}
	return nil
}

func (s *fakeStore) SaveUpDown(_ context.Context, _ int, _ int, updown int, _ int64) (int, error) {
	s.saved = updown
	return 1, nil
}

func (s *fakeStore) IncrementVODCounter(_ context.Context, _ int, field string, delta int) error {
	s.counters = append(s.counters, field+":"+strconv.Itoa(delta))
	return nil
}

func (s *fakeStore) RecountUpDown(context.Context, int) error {
	return nil
}

func (s *fakeStore) FavoriteCount(context.Context, int, int) (int, error) {
	return 1, nil
}

func (s *fakeStore) MiniViewLog(context.Context, int, string, int) (map[string]interface{}, error) {
	if s.viewlog != nil {
		return s.viewlog, nil
	}
	return map[string]interface{}{}, nil
}

func (s *fakeStore) CountMiniViewLogsSince(context.Context, int, string, int64, int) (int, error) {
	return s.daycount, nil
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

func TestVoteRejectsNonMiniVOD(t *testing.T) {
	store := &fakeStore{vod: map[string]interface{}{"vodid": "9", "showtype": "0"}}
	service := NewService(store, fakeProcessor{}, "https://res.test")

	retcode, errmsg, err := service.Vote(context.Background(), "", 9, true)
	if err != nil {
		t.Fatalf("vote: %v", err)
	}
	if retcode != -1 || errmsg != "记录不存在或已被删除" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestVoteGuestUpToggle(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, fakeProcessor{}, "https://res.test")

	retcode, errmsg, err := service.Vote(context.Background(), "", 9, true)
	if err != nil || retcode != 0 || errmsg != "已赞" {
		t.Fatalf("first retcode=%d errmsg=%q err=%v", retcode, errmsg, err)
	}
	retcode, errmsg, err = service.Vote(context.Background(), "", 9, true)
	if err != nil || retcode != 0 || errmsg != "已取消赞" {
		t.Fatalf("second retcode=%d errmsg=%q err=%v", retcode, errmsg, err)
	}
	if len(store.counters) < 2 || store.counters[0] != "upnum:1" || store.counters[1] != "upnum:-1" {
		t.Fatalf("counters=%#v", store.counters)
	}
}

type fakeAuth struct {
	user map[string]interface{}
}

func (a fakeAuth) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return a.user, nil
}

func TestVoteUserSwitchesState(t *testing.T) {
	store := &fakeStore{updown: map[string]interface{}{"updown": "1"}}
	service := NewService(store, fakeProcessor{}, "https://res.test").WithAuth(fakeAuth{user: map[string]interface{}{"uid": "7"}})

	retcode, errmsg, err := service.Vote(context.Background(), "token", 9, false)
	if err != nil || retcode != 0 || errmsg != "已踩" {
		t.Fatalf("retcode=%d errmsg=%q err=%v", retcode, errmsg, err)
	}
	if !store.deleted || store.saved != 2 {
		t.Fatalf("deleted=%v saved=%d", store.deleted, store.saved)
	}
	if len(store.counters) < 2 || store.counters[0] != "downnum:1" || store.counters[1] != "upnum:-1" {
		t.Fatalf("counters=%#v", store.counters)
	}
}

func TestReqLongRejectsMiniVOD(t *testing.T) {
	store := &fakeStore{vod: map[string]interface{}{"vodid": "9", "showtype": "1", "play_url": "x.m3u8"}}
	service := NewService(store, fakeProcessor{}, "https://res.test")

	_, retcode, errmsg, err := service.ReqLong(context.Background(), "", 9)
	if err != nil {
		t.Fatalf("reqlong: %v", err)
	}
	if retcode != 1 || errmsg != "记录不存在或已被删除" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestReqPlayFreeMiniVOD(t *testing.T) {
	store := &fakeStore{vod: map[string]interface{}{
		"vodid":      "9",
		"showtype":   "1",
		"play_url":   "p/index.m3u8",
		"play_srvid": "3",
		"view_price": "0",
		"isvip":      "0",
		"islimit":    "0",
		"islimitv3":  "0",
		"free_sdate": "0",
		"free_edate": "0",
	}}
	service := NewService(store, fakeProcessor{}, "https://res.test")
	service.auth = fakeAuth{user: map[string]interface{}{"uid": "7", "sid": "s", "perms": map[string]interface{}{}, "uniqkey": "9"}}

	data, retcode, errmsg, err := service.ReqPlay(context.Background(), "token", 9, 0)
	if err != nil {
		t.Fatalf("reqplay: %v", err)
	}
	if retcode != 0 || errmsg != "免费观看" || data["httpurl"] != "https://cdn.test/p/index.m3u8" {
		t.Fatalf("retcode=%d errmsg=%q data=%#v", retcode, errmsg, data)
	}
	if data["isfavorite"] != 1 || data["iszan"] != 0 {
		t.Fatalf("flags=%#v", data)
	}
}

func TestReqPlayRejectsVIPWithoutPerm(t *testing.T) {
	store := &fakeStore{vod: map[string]interface{}{"vodid": "9", "showtype": "1", "isvip": "1"}}
	service := NewService(store, fakeProcessor{}, "https://res.test")
	service.auth = fakeAuth{user: map[string]interface{}{"uid": "7", "sid": "s", "perms": map[string]interface{}{}}}

	_, retcode, errmsg, err := service.ReqPlay(context.Background(), "token", 9, 0)
	if err != nil {
		t.Fatalf("reqplay: %v", err)
	}
	if retcode != 5 || errmsg != "VIP独享内容，请升级" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestReqDownFreeMiniVOD(t *testing.T) {
	store := &fakeStore{vod: map[string]interface{}{
		"vodid":      "9",
		"showtype":   "1",
		"down_url":   "d/file.mp4",
		"down_srvid": "3",
		"view_price": "0",
	}}
	service := NewService(store, fakeProcessor{}, "https://res.test")
	service.auth = fakeAuth{user: map[string]interface{}{"uid": "7", "sid": "s", "perms": map[string]interface{}{}}}

	data, retcode, errmsg, err := service.ReqDown(context.Background(), "token", 9, 0)
	if err != nil {
		t.Fatalf("reqdown: %v", err)
	}
	if retcode != 0 || errmsg != "免费观看提供下载" || data["httpurl"] != "https://cdn.test/d/file.mp4" {
		t.Fatalf("retcode=%d errmsg=%q data=%#v", retcode, errmsg, data)
	}
}

func TestReqLongReturnsAbsoluteURL(t *testing.T) {
	store := &fakeStore{vod: map[string]interface{}{"vodid": "9", "showtype": "0", "play_url": "https://cdn.test/a/index.m3u8", "play_srvid": "3"}}
	service := NewService(store, fakeProcessor{}, "https://res.test")

	body, retcode, errmsg, err := service.ReqLong(context.Background(), "", 9)
	if err != nil || retcode != 0 || errmsg != "" {
		t.Fatalf("body=%q retcode=%d errmsg=%q err=%v", body, retcode, errmsg, err)
	}
	if body != "https://cdn.test/a/index.m3u8" {
		t.Fatalf("body=%q", body)
	}
}

func TestReqLongAddsServerHostForRelativeURL(t *testing.T) {
	store := &fakeStore{vod: map[string]interface{}{"vodid": "9", "showtype": "0", "play_url": "a/index.m3u8", "play_srvid": "3"}}
	service := NewService(store, fakeProcessor{}, "https://res.test")

	body, retcode, _, err := service.ReqLong(context.Background(), "", 9)
	if err != nil || retcode != 0 {
		t.Fatalf("body=%q retcode=%d err=%v", body, retcode, err)
	}
	if body != "https://cdn.test/a/index.m3u8" {
		t.Fatalf("body=%q", body)
	}
}
