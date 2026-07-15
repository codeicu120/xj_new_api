package vod

import (
	"context"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"xj_comp/internal/domain"
	vodRepo "xj_comp/internal/repository/vod"
)

type fakeListingStore struct {
	lastFilter          vodRepo.ListingFilter
	lastTotal           int
	lastPage            int
	lastSize            int
	lastOrder           string
	randomUsed          bool
	vodByID             map[string]interface{}
	searchLog           map[string]interface{}
	miniSearchCalls     int
	miniCachedListCalls int
	miniUpsertCalls     int
	miniIncrementCalls  int
	updown              map[string]interface{}
	breaking            map[string]interface{}
	errorReport         map[string]interface{}
	savedErrorReport    *vodRepo.ErrorReportInput
	savedUpdown         int
	counters            []string
}

type fakeM3U8Fetcher map[string]string

func (f fakeM3U8Fetcher) Get(_ context.Context, url string) (string, error) {
	return f[url], nil
}

func (s *fakeListingStore) Categories(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"cateid": "1", "parentid": "0", "uuid": "a", "catename": "主类"},
		{"cateid": "2", "parentid": "1", "uuid": "b", "catename": "子类"},
	}, nil
}

func (s *fakeListingStore) Areas(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"areaid": "3", "areaname": "日本", "sortnum": "1"}}, nil
}

func (s *fakeListingStore) Years(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"yearid": "4", "yearname": "2025", "sortnum": "1"}}, nil
}

func (s *fakeListingStore) Servers(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"srvid": "9", "srvtype": "cover", "srvhost": "https://cover.example.test", "sortnum": "1"},
		{"srvid": "9", "srvtype": "play", "srvhost": "https://cover.example.test", "sortnum": "1"},
		{"srvid": "10", "srvtype": "preview", "srvhost": "https://preview.example.test", "sortnum": "1"},
	}, nil
}

func (s *fakeListingStore) CountVODs(_ context.Context, filter vodRepo.ListingFilter) (int, error) {
	s.lastFilter = filter
	return 20, nil
}

func (s *fakeListingStore) ListVODs(_ context.Context, filter vodRepo.ListingFilter, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error) {
	s.lastFilter = filter
	s.lastTotal = total
	s.lastPage = page
	s.lastSize = pageSize
	s.lastOrder = orderBy
	return fixtureVODRows(), nil
}

func (s *fakeListingStore) RandomVODs(_ context.Context, pageSize int) ([]map[string]interface{}, error) {
	s.randomUsed = true
	s.lastSize = pageSize
	return fixtureVODRows(), nil
}

func (s *fakeListingStore) RandomVODsExcept(_ context.Context, pageSize int, _ int, _ int) ([]map[string]interface{}, error) {
	s.randomUsed = true
	s.lastSize = pageSize
	rows := []map[string]interface{}{}
	for len(rows) < pageSize {
		rows = append(rows, fixtureVODRows()[0])
	}
	return rows, nil
}

func (s *fakeListingStore) VODByID(context.Context, int) (map[string]interface{}, error) {
	if s.vodByID != nil {
		return s.vodByID, nil
	}
	return fixtureVODRows()[0], nil
}

func (s *fakeListingStore) VODsByIDs(context.Context, []int, string) ([]map[string]interface{}, error) {
	return fixtureVODRows(), nil
}

func (s *fakeListingStore) SimilarVODsByTagIDs(context.Context, []int, int, int64, int) ([]map[string]interface{}, error) {
	return fixtureVODRows(), nil
}

func (s *fakeListingStore) TagsByNames(context.Context, []string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"tagid": "7", "tagtype": "1", "tagname": "剧情", "itemcount": "12"},
		{"tagid": "8", "tagtype": "0", "tagname": "演员", "itemcount": "3"},
	}, nil
}

func (s *fakeListingStore) CalldataByUUID(_ context.Context, uuid string) (map[string]interface{}, error) {
	switch uuid {
	case "search.hotwords":
		return map[string]interface{}{"type": "json", "content": `["剧情","演员"]`}, nil
	case "search.hotvods":
		return map[string]interface{}{"type": "code", "content": "100"}, nil
	case "search.minihotwords":
		return map[string]interface{}{"type": "json", "content": `["短片","作者"]`}, nil
	case "search.minihotvods":
		return map[string]interface{}{"type": "code", "content": "100"}, nil
	default:
		return map[string]interface{}{}, nil
	}
}

func (s *fakeListingStore) VODsByIDsLimited(_ context.Context, _ []int, _ bool, _ int, _ bool) ([]map[string]interface{}, error) {
	return fixtureVODRows(), nil
}

func (s *fakeListingStore) SearchVODs(_ context.Context, _ string, _ bool, _ int) ([]map[string]interface{}, error) {
	return fixtureVODRows(), nil
}

func (s *fakeListingStore) SearchLog(context.Context, string) (map[string]interface{}, error) {
	if s.searchLog != nil {
		return s.searchLog, nil
	}
	return map[string]interface{}{}, nil
}

func (s *fakeListingStore) UpsertSearchLog(context.Context, string, int64, int, []int) error {
	return nil
}

func (s *fakeListingStore) IncrementSearchLog(context.Context, string, int64, int64) error {
	return nil
}

func (s *fakeListingStore) TopSearchVODIDs(context.Context) (string, error) {
	return "100", nil
}

func (s *fakeListingStore) MiniVODsByIDsLimited(_ context.Context, _ []int, _ int, _ bool) ([]map[string]interface{}, error) {
	s.miniCachedListCalls++
	return fixtureVODRows(), nil
}

func (s *fakeListingStore) MiniVODsByIDs(_ context.Context, _ []int, _ string) ([]map[string]interface{}, error) {
	return fixtureVODRows(), nil
}

func (s *fakeListingStore) MiniSearchVODs(_ context.Context, _ string, _ int) ([]map[string]interface{}, error) {
	s.miniSearchCalls++
	return fixtureVODRows(), nil
}

func (s *fakeListingStore) MiniSearchLog(context.Context, string) (map[string]interface{}, error) {
	if s.searchLog != nil {
		return s.searchLog, nil
	}
	return map[string]interface{}{}, nil
}

func (s *fakeListingStore) UpsertMiniSearchLog(context.Context, string, int64, int, []int) error {
	s.miniUpsertCalls++
	return nil
}

func (s *fakeListingStore) IncrementMiniSearchLog(context.Context, string, int64, int64) error {
	s.miniIncrementCalls++
	return nil
}

func (s *fakeListingStore) TopMiniSearchVODIDs(context.Context) (string, error) {
	return "100", nil
}

func (s *fakeListingStore) BreakingVOD(context.Context, int, int64) (map[string]interface{}, error) {
	if s.breaking != nil {
		return s.breaking, nil
	}
	return map[string]interface{}{"vodid": "99", "title": "爆料", "showtype": "0"}, nil
}

func (s *fakeListingStore) VODErrorByUID(context.Context, string, int) (map[string]interface{}, error) {
	if s.errorReport == nil {
		return map[string]interface{}{}, nil
	}
	return s.errorReport, nil
}

func (s *fakeListingStore) SaveVODError(_ context.Context, input vodRepo.ErrorReportInput) (int, error) {
	if s.savedErrorReport != nil {
		*s.savedErrorReport = input
	}
	return 1, nil
}

func (s *fakeListingStore) UpDownByUser(context.Context, int, int) (map[string]interface{}, error) {
	if s.updown != nil {
		return s.updown, nil
	}
	return map[string]interface{}{}, nil
}

func (s *fakeListingStore) DeleteUpDown(context.Context, int, int) error {
	s.savedUpdown = 0
	return nil
}

func (s *fakeListingStore) SaveUpDown(_ context.Context, _ int, _ int, updown int, _ int64) (int, error) {
	s.savedUpdown = updown
	return 1, nil
}

func (s *fakeListingStore) IncrementVODCounter(_ context.Context, _ int, field string, delta int) error {
	s.counters = append(s.counters, field+":"+strconv.Itoa(delta))
	return nil
}

func (s *fakeListingStore) RecountUpDown(context.Context, int) error {
	return nil
}

func TestListingServiceSearchIndex(t *testing.T) {
	store := &fakeListingStore{}
	service := NewListingService(store, "https://res.example.test", 50)
	service.now = func() time.Time { return time.Unix(2000, 0) }

	data, err := service.Search(context.Background(), "", false, 0, false)
	if err != nil {
		t.Fatalf("search index: %v", err)
	}
	index, ok := data.(domain.SearchIndexData)
	if !ok {
		t.Fatalf("expected SearchIndexData, got %T", data)
	}
	hotwords := index.HotWords.([]interface{})
	if len(hotwords) != 2 || hotwords[0] != "剧情" {
		t.Fatalf("unexpected hotwords %#v", index.HotWords)
	}
	if len(index.HotRows) != 1 || index.HotRows[0]["coverpic"] != "https://cover.example.test/202501/a.jpg" {
		t.Fatalf("unexpected hot rows %#v", index.HotRows)
	}
}

func TestListingServiceSearchList(t *testing.T) {
	store := &fakeListingStore{}
	service := NewListingService(store, "https://res.example.test", 50)
	service.now = func() time.Time { return time.Unix(2000, 0) }

	data, err := service.Search(context.Background(), "标题", false, 1, false)
	if err != nil {
		t.Fatalf("search list: %v", err)
	}
	list, ok := data.(domain.SearchListData)
	if !ok {
		t.Fatalf("expected SearchListData, got %T", data)
	}
	if len(list.VODRows) != 1 {
		t.Fatalf("expected one search row, got %d", len(list.VODRows))
	}
	if list.PageInfo["total"] != 1 {
		t.Fatalf("expected total 1, got %v", list.PageInfo["total"])
	}
}

func TestListingServiceMiniSearchIndex(t *testing.T) {
	store := &fakeListingStore{}
	service := NewListingService(store, "https://res.example.test", 50)
	service.now = func() time.Time { return time.Unix(2000, 0) }

	data, err := service.MiniSearch(context.Background(), "", 0, false)
	if err != nil {
		t.Fatalf("mini search index: %v", err)
	}
	index, ok := data.(domain.SearchIndexData)
	if !ok {
		t.Fatalf("expected SearchIndexData, got %T", data)
	}
	if len(index.HotRows) != 1 {
		t.Fatalf("expected one hot row, got %d", len(index.HotRows))
	}
	vodrow := index.HotRows[0]["vodrow"].(map[string]interface{})
	if vodrow["play_url"] != "/minivod/reqplay/100" || !strings.Contains(vodrow["preview_url"].(string), "/minivod/preView/100/index.m3u8") {
		t.Fatalf("unexpected mini vod urls %#v", vodrow)
	}
}

func TestListingServiceMiniSearchList(t *testing.T) {
	store := &fakeListingStore{}
	service := NewListingService(store, "https://res.example.test", 50)
	service.now = func() time.Time { return time.Unix(2000, 0) }

	data, err := service.MiniSearch(context.Background(), "标题", 1, false)
	if err != nil {
		t.Fatalf("mini search list: %v", err)
	}
	list, ok := data.(domain.MiniSearchListData)
	if !ok {
		t.Fatalf("expected MiniSearchListData, got %T", data)
	}
	if len(list.Rows) != 1 || list.PageInfo["page_url"] != "/search?wd=%E6%A0%87%E9%A2%98&page=[?]" {
		t.Fatalf("unexpected mini search list %#v", list)
	}
	vodrow := list.Rows[0]["vodrow"].(map[string]interface{})
	if vodrow["down_url"] != "/minivod/reqdown/100" {
		t.Fatalf("unexpected mini vod row %#v", vodrow)
	}
}

func TestListingServiceMiniSearchRefreshesExpiredLogAndIncrementsFirstPage(t *testing.T) {
	store := &fakeListingStore{searchLog: map[string]interface{}{
		"schwd":        "标题",
		"schtime":      "1",
		"sch_lasttime": "100",
		"total":        "1",
		"vodids":       "100",
	}}
	service := NewListingService(store, "https://res.example.test", 50)
	service.now = func() time.Time { return time.Unix(4000, 0) }

	if _, err := service.MiniSearch(context.Background(), "标题", 1, false); err != nil {
		t.Fatalf("mini search list: %v", err)
	}
	if store.miniSearchCalls != 1 || store.miniUpsertCalls != 1 || store.miniIncrementCalls != 1 {
		t.Fatalf("expected expired log refresh and increment, got search=%d upsert=%d increment=%d", store.miniSearchCalls, store.miniUpsertCalls, store.miniIncrementCalls)
	}
}

func TestListingServiceMiniSearchUsesFreshCachedLogWithoutRebuild(t *testing.T) {
	store := &fakeListingStore{searchLog: map[string]interface{}{
		"schwd":        "标题",
		"schtime":      "3900",
		"sch_lasttime": "3900",
		"total":        "20",
		"vodids":       "100,101,102,103,104,105,106,107,108,109,110,111,112,113,114,115,116,117,118,119",
	}}
	service := NewListingService(store, "https://res.example.test", 50)
	service.now = func() time.Time { return time.Unix(4000, 0) }

	if _, err := service.MiniSearch(context.Background(), "标题", 2, false); err != nil {
		t.Fatalf("mini search list: %v", err)
	}
	if store.miniSearchCalls != 0 || store.miniUpsertCalls != 0 || store.miniIncrementCalls != 0 {
		t.Fatalf("expected cached log without rebuild/increment, got search=%d upsert=%d increment=%d", store.miniSearchCalls, store.miniUpsertCalls, store.miniIncrementCalls)
	}
	if store.miniCachedListCalls != 1 {
		t.Fatalf("expected cached vod lookup, got %d", store.miniCachedListCalls)
	}
}

func TestListingServiceShowProcessesDetailRows(t *testing.T) {
	store := &fakeListingStore{}
	service := NewListingService(store, "https://res.example.test", 50)
	service.now = func() time.Time { return time.Unix(2000, 0) }

	data, err := service.Show(context.Background(), 100, false)
	if err != nil {
		t.Fatalf("show vod: %v", err)
	}
	if data.VODRow["vodid"] != "100" {
		t.Fatalf("unexpected vodrow %#v", data.VODRow)
	}
	if len(data.Categories) != 1 || data.Categories[0]["cateid"] != "1" {
		t.Fatalf("unexpected categories %#v", data.Categories)
	}
	if len(data.SimilarRows) != 10 {
		t.Fatalf("expected 10 similar rows, got %d", len(data.SimilarRows))
	}
	if len(data.LikeRows) != 5 {
		t.Fatalf("expected 5 like rows, got %d", len(data.LikeRows))
	}
}

func TestListingServiceShowMissingVOD(t *testing.T) {
	store := &fakeListingStore{vodByID: map[string]interface{}{}}
	service := NewListingService(store, "https://res.example.test", 50)

	_, err := service.Show(context.Background(), 404, false)
	if err != ErrVODNotFound {
		t.Fatalf("expected ErrVODNotFound, got %v", err)
	}
}

func TestListingServicePreviewProcessesM3U8(t *testing.T) {
	store := &fakeListingStore{}
	service := NewListingService(store, "https://res.example.test", 50)
	service.fetcher = fakeM3U8Fetcher{
		"https://cover.example.test/202501/a.jpg":  "",
		"https://cover.example.test/202501/a.m3u8": "#EXTM3U\nlow/index.m3u8\n",
		"https://cover.example.test/low/index.m3u8": strings.Join([]string{
			"#EXTM3U",
			"#EXT-X-VERSION:3",
			`#EXT-X-KEY:METHOD=AES-128,URI="key.key"`,
			"#EXTINF:10,",
			"seg1.ts",
			"#EXTINF:10,",
			"http://cdn.example.test/seg2.ts",
		}, "\n"),
	}
	store.vodByID = fixtureVODRows()[0]
	store.vodByID["play_url"] = "202501/a.m3u8"
	store.vodByID["play_srvid"] = "9"

	body, err := service.Preview(context.Background(), 100)
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	for _, want := range []string{
		"#EXTM3U",
		`URI="https://cover.example.test/key.key"`,
		"https://cover.example.test/seg1.ts",
		"http://cdn.example.test/seg2.ts",
		"#EXT-X-ENDLIST",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected preview body to contain %q, got:\n%s", want, body)
		}
	}
}

func TestListingServicePreviewMissingVODReturnsEmpty(t *testing.T) {
	store := &fakeListingStore{vodByID: map[string]interface{}{}}
	service := NewListingService(store, "https://res.example.test", 50)

	body, err := service.Preview(context.Background(), 404)
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if body != "" {
		t.Fatalf("expected empty body, got %q", body)
	}
}

func TestListingServiceParsesParamsAndProcessesRows(t *testing.T) {
	store := &fakeListingStore{}
	service := NewListingService(store, "https://res.example.test", 50)
	service.now = func() time.Time { return time.Unix(2000, 0) }

	data, err := service.List(context.Background(), ListingRequest{
		Action:     "listing",
		PathParams: "1-3-4-2-1-1-2-3-2-0",
		QueryPage:  "2",
	})
	if err != nil {
		t.Fatalf("list vods: %v", err)
	}

	if data.Params["page"] != "2" {
		t.Fatalf("expected query page fallback, got %q", data.Params["page"])
	}
	if store.lastOrder != "playcount_total DESC" {
		t.Fatalf("unexpected order %q", store.lastOrder)
	}
	if !reflect.DeepEqual(store.lastFilter.CateIDs, []int{1, 2}) {
		t.Fatalf("unexpected cate ids %#v", store.lastFilter.CateIDs)
	}
	if store.lastFilter.AreaID != 3 || store.lastFilter.YearID != 4 || store.lastFilter.Definition != 2 {
		t.Fatalf("unexpected filter %#v", store.lastFilter)
	}
	if len(data.VODRows) != 1 {
		t.Fatalf("expected one vod row, got %d", len(data.VODRows))
	}

	row := data.VODRows[0]
	if row["coverpic"] != "https://cover.example.test/202501/a.jpg" {
		t.Fatalf("unexpected coverpic %v", row["coverpic"])
	}
	if row["duration"] != "30:01" {
		t.Fatalf("unexpected duration %v", row["duration"])
	}
	if row["vip_price"] != float64(50) {
		t.Fatalf("unexpected vip price %v", row["vip_price"])
	}
	if row["limit_free"] != 1 {
		t.Fatalf("unexpected limit_free %v", row["limit_free"])
	}
	if row["playcount_total"] != 2 {
		t.Fatalf("unexpected playcount_total %v", row["playcount_total"])
	}
	tags := row["tags"].([]map[string]interface{})
	if len(tags) != 1 || tags[0]["tagname"] != "剧情" {
		t.Fatalf("unexpected tags %#v", tags)
	}
}

func TestListingServiceActionOrdering(t *testing.T) {
	tests := []struct {
		action string
		order  string
	}{
		{action: "hot", order: "playcount_week DESC"},
		{action: "latest", order: "vodid DESC"},
	}
	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			store := &fakeListingStore{}
			service := NewListingService(store, "https://res.example.test", 50)
			_, err := service.List(context.Background(), ListingRequest{Action: tt.action})
			if err != nil {
				t.Fatalf("list vods: %v", err)
			}
			if store.lastOrder != tt.order {
				t.Fatalf("expected order %q, got %q", tt.order, store.lastOrder)
			}
		})
	}
}

func TestListingServiceRecommendUsesRandomRows(t *testing.T) {
	store := &fakeListingStore{}
	service := NewListingService(store, "https://res.example.test", 50)

	data, err := service.List(context.Background(), ListingRequest{Action: "recommend"})
	if err != nil {
		t.Fatalf("list vods: %v", err)
	}
	if !store.randomUsed {
		t.Fatal("expected recommend to use random rows")
	}
	if data.PageInfo["total"] != 0 {
		t.Fatalf("expected recommend page total 0, got %v", data.PageInfo["total"])
	}
}

func TestListingServiceLikeRowsUsesSixRandomRows(t *testing.T) {
	store := &fakeListingStore{}
	service := NewListingService(store, "https://res.example.test", 50)
	service.now = func() time.Time { return time.Unix(2000, 0) }

	data, err := service.LikeRows(context.Background(), false)
	if err != nil {
		t.Fatalf("list like rows: %v", err)
	}
	if !store.randomUsed {
		t.Fatal("expected like rows to use random rows")
	}
	if store.lastSize != 6 {
		t.Fatalf("expected PHP-compatible page size 6, got %d", store.lastSize)
	}
	if len(data.LikeRows) != 1 {
		t.Fatalf("expected one fixture row, got %d", len(data.LikeRows))
	}
	if data.LikeRows[0]["coverpic"] != "https://cover.example.test/202501/a.jpg" {
		t.Fatalf("unexpected coverpic %v", data.LikeRows[0]["coverpic"])
	}
}

func TestVoteMissingVOD(t *testing.T) {
	store := &fakeListingStore{vodByID: map[string]interface{}{}}
	service := NewListingService(store, "https://res.example.test", 50)

	retcode, errmsg, err := service.Vote(context.Background(), "", 0, true)
	if err != nil {
		t.Fatalf("vote: %v", err)
	}
	if retcode != -1 || errmsg != "记录不存在或已被删除" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestVoteGuestUpToggle(t *testing.T) {
	store := &fakeListingStore{}
	service := NewListingService(store, "https://res.example.test", 50)

	retcode, errmsg, err := service.Vote(context.Background(), "", 100, true)
	if err != nil {
		t.Fatalf("vote: %v", err)
	}
	if retcode != 0 || errmsg != "已赞" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
	retcode, errmsg, err = service.Vote(context.Background(), "", 100, true)
	if err != nil {
		t.Fatalf("vote toggle: %v", err)
	}
	if retcode != 0 || errmsg != "已取消赞" {
		t.Fatalf("toggle response = %d %q", retcode, errmsg)
	}
}

func TestVoteUserSwitchesState(t *testing.T) {
	store := &fakeListingStore{updown: map[string]interface{}{"updown": "2"}}
	service := NewListingService(store, "https://res.example.test", 50).WithAuth(fakeVODAuth{user: map[string]interface{}{"uid": "5"}})

	retcode, errmsg, err := service.Vote(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 100, true)
	if err != nil {
		t.Fatalf("vote: %v", err)
	}
	if retcode != 0 || errmsg != "已赞" || store.savedUpdown != 1 {
		t.Fatalf("response = %d %q saved=%d", retcode, errmsg, store.savedUpdown)
	}
}

func TestBreakingReturnsVODIDAndTitle(t *testing.T) {
	service := NewListingService(&fakeListingStore{}, "https://res.example.test", 100)
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	data, retcode, errmsg, err := service.Breaking(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "ok" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if data["vodid"] != "99" || data["title"] != "爆料" {
		t.Fatalf("data = %#v", data)
	}
}

func TestBreakingMissing(t *testing.T) {
	service := NewListingService(&fakeListingStore{breaking: map[string]interface{}{}}, "https://res.example.test", 100)

	_, retcode, errmsg, err := service.Breaking(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "记录不存在或已被删除" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestErrorReportValidatesRequiredFields(t *testing.T) {
	service := NewListingService(&fakeListingStore{}, "https://res.example.test", 100)

	retcode, errmsg, err := service.ErrorReport(context.Background(), ErrorReportRequest{VODID: 100})
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "缺少参数" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestErrorReportMissingVOD(t *testing.T) {
	service := NewListingService(&fakeListingStore{vodByID: map[string]interface{}{}}, "https://res.example.test", 100)

	retcode, errmsg, err := service.ErrorReport(context.Background(), validErrorReportRequest())
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "该视频不存在或者已删除" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestErrorReportDuplicate(t *testing.T) {
	service := NewListingService(&fakeListingStore{errorReport: map[string]interface{}{"id": "1"}}, "https://res.example.test", 100)

	retcode, errmsg, err := service.ErrorReport(context.Background(), validErrorReportRequest())
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您已提交过该视频报错反馈" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestErrorReportSavesInput(t *testing.T) {
	saved := vodRepo.ErrorReportInput{}
	service := NewListingService(&fakeListingStore{savedErrorReport: &saved}, "https://res.example.test", 100)
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	retcode, errmsg, err := service.ErrorReport(context.Background(), validErrorReportRequest())
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if saved.UID != "0" || saved.VODID != 100 || saved.PlayURL != "https://play.example.test/a.m3u8" {
		t.Fatalf("saved = %#v", saved)
	}
	if saved.Now != 1770000000 || saved.ClientIP != "127.0.0.1" {
		t.Fatalf("saved time/ip = %#v", saved)
	}
}

func validErrorReportRequest() ErrorReportRequest {
	return ErrorReportRequest{
		VODID:      100,
		PlayURL:    "https://play.example.test/a.m3u8",
		AppVersion: "1.0.0",
		SysVersion: "iOS 18",
		Model:      "iPhone",
		Channel:    "appstore",
		Network:    "wifi",
		Details:    "broken",
		ClientIP:   "127.0.0.1",
	}
}

type fakeVODAuth struct {
	user map[string]interface{}
}

func (a fakeVODAuth) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return a.user, nil
}

func fixtureVODRows() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"vodid":           "100",
			"title":           "标题",
			"intro":           "",
			"coverpic":        "202501/a.jpg",
			"cover_srvid":     "9",
			"ctimestamp":      "1735689600",
			"utimestamp":      "1735689600",
			"vodkey":          "ABC",
			"scorenum":        "9.0",
			"upnum":           "1",
			"downnum":         "2",
			"authorid":        "0",
			"author":          "",
			"play_url":        "a.m3u8",
			"down_url":        "a.mp4",
			"definition":      "2",
			"duration":        "1801",
			"yearid":          "4",
			"mosaic":          "1",
			"portrait":        "0",
			"view_price":      "100",
			"free_sdate":      "1000",
			"free_edate":      "3000",
			"isvip":           "2",
			"islimit":         "0",
			"islimitv3":       "0",
			"prop4":           "1",
			"commentcount":    "5",
			"playcount_total": "20000",
			"downcount_total": "6",
			"tags":            "剧情",
			"actor_tags":      "演员",
			"areaid":          "3",
			"cateid":          "1",
			"playlist":        "第一集$http://example.test$10",
			"downlist":        "",
			"episode_total":   "1",
			"episode_status":  "0",
		},
	}
}
