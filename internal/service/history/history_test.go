package history

import (
	"context"
	"testing"
	"time"

	historyRepo "xj_comp/internal/repository/history"
)

type fakeStore struct {
	user     map[string]interface{}
	kind     historyRepo.Kind
	uid      int
	sid      string
	timeline int
	total    int
	rows     []map[string]interface{}
	removed  []int
}

func (s *fakeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s *fakeStore) Items(_ context.Context, kind historyRepo.Kind, uid int, sid string, _ int, _ int, timeline int, _ int64) (int, []map[string]interface{}, error) {
	s.kind = kind
	s.uid = uid
	s.sid = sid
	s.timeline = timeline
	return s.total, s.rows, nil
}

func (s *fakeStore) Remove(_ context.Context, kind historyRepo.Kind, uid int, sid string, vodid int) (int, error) {
	s.kind = kind
	s.uid = uid
	s.sid = sid
	s.removed = append(s.removed, vodid)
	if vodid <= 0 {
		return 0, nil
	}
	return 1, nil
}

type fakeVODProcessor struct{}

func (fakeVODProcessor) ProcessRows(_ context.Context, rows []map[string]interface{}, _ bool) ([]map[string]interface{}, error) {
	for _, row := range rows {
		row["processed"] = "1"
	}
	return rows, nil
}

func (fakeVODProcessor) ProcessMiniRowsFullPrice(_ context.Context, rows []map[string]interface{}, _ bool) ([]map[string]interface{}, error) {
	for _, row := range rows {
		row["mini_processed"] = "1"
	}
	return rows, nil
}

func TestListingUsesGuestSIDWhenNotLoggedIn(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, store, fakeVODProcessor{})
	service.now = func() time.Time { return time.Unix(1700000000, 0) }

	data, err := service.Listing(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", historyRepo.KindPlay, 1, 2, false)
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if store.uid != 0 || store.sid != "250f790ba71ec2b9d3855f424db2259e" || store.timeline != 2 {
		t.Fatalf("lookup = uid:%d sid:%q timeline:%d", store.uid, store.sid, store.timeline)
	}
	if data.PageInfo["page_url"] != "/playlog/listing?timeline=2&page=[?]" {
		t.Fatalf("pageinfo = %#v", data.PageInfo)
	}
}

func TestListingFormatsPlaytimeAndProcessesRows(t *testing.T) {
	store := &fakeStore{
		user:  map[string]interface{}{"uid": "7", "sid": "sid"},
		total: 1,
		rows:  []map[string]interface{}{{"vodid": "9", "logid": "3", "playtime": "1699996400"}},
	}
	service := NewService(store, store, fakeVODProcessor{})
	service.now = func() time.Time { return time.Unix(1700000000, 0) }

	data, err := service.Listing(context.Background(), "token", historyRepo.KindPlay, 1, 0, true)
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if store.kind != historyRepo.KindPlay || store.uid != 7 {
		t.Fatalf("lookup = kind:%s uid:%d", store.kind, store.uid)
	}
	if len(data.Rows) != 1 || data.Rows[0]["processed"] != "1" || data.Rows[0]["playtime"] != "1小时前" {
		t.Fatalf("rows = %#v", data.Rows)
	}
}

func TestListingFormatsOldDowntimeAsDate(t *testing.T) {
	store := &fakeStore{
		user:  map[string]interface{}{"uid": "7", "sid": "sid"},
		total: 1,
		rows:  []map[string]interface{}{{"vodid": "9", "logid": "3", "downtime": "1600000000"}},
	}
	service := NewService(store, store, fakeVODProcessor{})
	service.now = func() time.Time { return time.Unix(1700000000, 0) }

	data, err := service.Listing(context.Background(), "token", historyRepo.KindDown, 1, 3, false)
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if data.PageInfo["page_url"] != "/downlog/listing?timeline=3&page=[?]" {
		t.Fatalf("pageinfo = %#v", data.PageInfo)
	}
	if data.Rows[0]["downtime"] != "2020-09-13" {
		t.Fatalf("rows = %#v", data.Rows)
	}
}

func TestRemoveUsesGuestSIDAndCountsRows(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, store, nil)

	msg, err := service.Remove(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", historyRepo.KindDown, []int{1, 0, 2})
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if msg != "已删除2项" {
		t.Fatalf("msg = %q", msg)
	}
	if store.kind != historyRepo.KindDown || store.uid != 0 || store.sid != "250f790ba71ec2b9d3855f424db2259e" {
		t.Fatalf("lookup = kind:%s uid:%d sid:%q", store.kind, store.uid, store.sid)
	}
	if len(store.removed) != 3 || store.removed[0] != 1 || store.removed[2] != 2 {
		t.Fatalf("removed = %#v", store.removed)
	}
}

func TestRemoveWithNoIDsMatchesPHPSuccess(t *testing.T) {
	store := &fakeStore{user: map[string]interface{}{"uid": "8", "sid": "sid"}}
	service := NewService(store, store, nil)

	msg, err := service.Remove(context.Background(), "token", historyRepo.KindPlay, nil)
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if msg != "已删除0项" || len(store.removed) != 0 {
		t.Fatalf("msg=%q removed=%#v", msg, store.removed)
	}
}

func TestMiniPlayListingUsesMiniProcessorAndPageURL(t *testing.T) {
	store := &fakeStore{
		user:  map[string]interface{}{"uid": "7", "sid": "sid"},
		total: 1,
		rows:  []map[string]interface{}{{"vodid": "9", "logid": "3", "playtime": "1699999940"}},
	}
	service := NewService(store, store, fakeVODProcessor{})
	service.now = func() time.Time { return time.Unix(1700000000, 0) }

	data, err := service.Listing(context.Background(), "token", historyRepo.KindMiniPlay, 1, 1, false)
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if store.kind != historyRepo.KindMiniPlay || store.uid != 7 {
		t.Fatalf("lookup = kind:%s uid:%d", store.kind, store.uid)
	}
	if data.PageInfo["page_url"] != "/miniplaylog/listing?timeline=1&page=[?]" {
		t.Fatalf("pageinfo = %#v", data.PageInfo)
	}
	if data.Rows[0]["mini_processed"] != "1" || data.Rows[0]["playtime"] != "1分钟前" {
		t.Fatalf("rows = %#v", data.Rows)
	}
}
