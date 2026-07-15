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
	users   []map[string]interface{}
	removed []int
	vodrow  map[string]interface{}
	count   int
	added   []int
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

func (s *fakeStore) VODByID(_ context.Context, vodid int) (map[string]interface{}, error) {
	if vodid <= 0 || s.vodrow == nil {
		return map[string]interface{}{}, nil
	}
	return s.vodrow, nil
}

func (s *fakeStore) Count(_ context.Context, kind favoriteRepo.Kind, uid int, vodid int, _ int64) (int, error) {
	s.kind = kind
	s.uid = uid
	if vodid > 0 {
		return s.count, nil
	}
	return 0, nil
}

func (s *fakeStore) Add(_ context.Context, kind favoriteRepo.Kind, uid int, vodid int, _ int64) error {
	s.kind = kind
	s.uid = uid
	s.added = append(s.added, vodid)
	return nil
}

func (s *fakeStore) UsersByIDs(_ context.Context, ids []int) ([]map[string]interface{}, error) {
	return s.users, nil
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

func TestMiniListingIgnoresKeywordForLegacyAPI(t *testing.T) {
	store := &fakeStore{
		user:  map[string]interface{}{"uid": "5"},
		total: 1,
		rows:  []map[string]interface{}{{"vodid": "9"}},
	}
	service := NewService(store, store, fakeVODProcessor{})

	_, _, _, err := service.Listing(context.Background(), "token", favoriteRepo.KindMini, 1, "abc", false)
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if store.keyword != "" {
		t.Fatalf("legacy mini keyword = %q", store.keyword)
	}
}

func TestMiniV2ListingWrapsRowsWithUsersAndKeywordPageURL(t *testing.T) {
	store := &fakeStore{
		user:  map[string]interface{}{"uid": "5"},
		total: 1,
		rows:  []map[string]interface{}{{"vodid": "9", "authorid": "7"}},
		users: []map[string]interface{}{{"uid": "7", "username": "u7", "nickname": "n7", "avatar": "avatar/a.jpg", "gender": "1"}},
	}
	service := NewService(store, store, fakeVODProcessor{}).WithResourceBaseURL("https://res.test")

	data, retcode, errmsg, err := service.MiniV2Listing(context.Background(), "token", 1, "abc", false)
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
	if store.kind != favoriteRepo.KindMini || store.keyword != "abc" {
		t.Fatalf("lookup = kind:%s keyword:%q", store.kind, store.keyword)
	}
	if data.PageInfo["page_url"] != "/minifavorite/listing?page=[?]&wd=abc" {
		t.Fatalf("pageinfo = %#v", data.PageInfo)
	}
	row := data.Rows[0]
	vodrow, ok := row["vodrow"].(map[string]interface{})
	if !ok || vodrow["vodid"] != "9" || vodrow["isfavorite"] != 1 {
		t.Fatalf("vodrow = %#v", row["vodrow"])
	}
	user, ok := row["user"].(map[string]interface{})
	if !ok || user["uid"] != "7" || user["avatar_url"] != "https://res.test/C1/avatar/a.jpg" {
		t.Fatalf("user = %#v", row["user"])
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

func TestAddRequiresLogin(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, store, nil)

	_, retcode, errmsg, err := service.Add(context.Background(), "", favoriteRepo.KindVOD, 9)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestAddValidatesVOD(t *testing.T) {
	store := &fakeStore{user: map[string]interface{}{"uid": "5"}, vodrow: map[string]interface{}{"vodid": "9", "showtype": "1"}}
	service := NewService(store, store, nil)

	_, retcode, errmsg, err := service.Add(context.Background(), "token", favoriteRepo.KindVOD, 9)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if retcode != -1 || errmsg != "记录不存在或已被删除" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestMiniAddRequiresMiniVOD(t *testing.T) {
	store := &fakeStore{user: map[string]interface{}{"uid": "5"}, vodrow: map[string]interface{}{"vodid": "9", "showtype": "0"}}
	service := NewService(store, store, nil)

	_, retcode, errmsg, err := service.Add(context.Background(), "token", favoriteRepo.KindMini, 9)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if retcode != -1 || errmsg != "记录不存在或已被删除" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestAddDuplicate(t *testing.T) {
	store := &fakeStore{user: map[string]interface{}{"uid": "5"}, vodrow: map[string]interface{}{"vodid": "9", "showtype": "0"}, count: 1}
	service := NewService(store, store, nil)

	_, retcode, errmsg, err := service.Add(context.Background(), "token", favoriteRepo.KindVOD, 9)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if retcode != -1 || errmsg != "您已经收藏过了" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestAddSuccess(t *testing.T) {
	store := &fakeStore{user: map[string]interface{}{"uid": "5"}, vodrow: map[string]interface{}{"vodid": "9", "showtype": "0"}}
	service := NewService(store, store, nil)

	data, retcode, errmsg, err := service.Add(context.Background(), "token", favoriteRepo.KindVOD, 9)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if retcode != 0 || errmsg != "已收藏" || data == nil {
		t.Fatalf("response = %d %q %#v", retcode, errmsg, data)
	}
	if len(store.added) != 1 || store.added[0] != 9 {
		t.Fatalf("added = %#v", store.added)
	}
}
