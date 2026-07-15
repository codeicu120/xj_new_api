package bought

import (
	"context"
	"testing"
)

type fakeStore struct {
	user    map[string]interface{}
	total   int
	rows    []map[string]interface{}
	deleted []int
	vod     map[string]interface{}
	count   int
	bean    map[string]interface{}
	bought  struct {
		uid   int
		vodid int
		price int
	}
}

func (s *fakeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s *fakeStore) Items(context.Context, int, int, int) (int, []map[string]interface{}, error) {
	return s.total, s.rows, nil
}

func (s *fakeStore) Delete(_ context.Context, _ int, vodid int) error {
	s.deleted = append(s.deleted, vodid)
	return nil
}

func (s *fakeStore) VODByID(context.Context, int) (map[string]interface{}, error) {
	return s.vod, nil
}

func (s *fakeStore) BoughtCount(context.Context, int, int) (int, error) {
	return s.count, nil
}

func (s *fakeStore) Goldbean(context.Context, int) (map[string]interface{}, error) {
	return s.bean, nil
}

func (s *fakeStore) BuyVOD(_ context.Context, uid int, vodid int, price int) error {
	s.bought.uid = uid
	s.bought.vodid = vodid
	s.bought.price = price
	return nil
}

type fakeVODProcessor struct{}

func (fakeVODProcessor) ProcessRows(_ context.Context, rows []map[string]interface{}, _ bool) ([]map[string]interface{}, error) {
	for _, row := range rows {
		row["processed"] = "1"
	}
	return rows, nil
}

func TestListingRequiresLogin(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, store, fakeVODProcessor{})

	_, retcode, errmsg, err := service.Listing(context.Background(), "", 1, false)
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestListingRows(t *testing.T) {
	store := &fakeStore{
		user:  map[string]interface{}{"uid": "5"},
		total: 1,
		rows:  []map[string]interface{}{{"vodid": "9"}},
	}
	service := NewService(store, store, fakeVODProcessor{})

	data, retcode, errmsg, err := service.Listing(context.Background(), "token", 1, true)
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
	if len(data.Rows) != 1 || data.Rows[0]["processed"] != "1" {
		t.Fatalf("rows = %#v", data.Rows)
	}
	if data.PageInfo["total"] != 1 || data.PageInfo["page_url"] != "/bought/listing?page=[?]" {
		t.Fatalf("pageinfo = %#v", data.PageInfo)
	}
}

func TestDeleteRequiresLogin(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, store)

	retcode, errmsg, err := service.Delete(context.Background(), "", []int{1})
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestDeleteVodIDs(t *testing.T) {
	store := &fakeStore{user: map[string]interface{}{"uid": "5"}}
	service := NewService(store, store)

	retcode, errmsg, err := service.Delete(context.Background(), "token", []int{1, 0, 2})
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
	if len(store.deleted) != 3 || store.deleted[0] != 1 || store.deleted[2] != 2 {
		t.Fatalf("deleted = %#v", store.deleted)
	}
}

func TestBuyRequiresLogin(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, store)

	retcode, errmsg, err := service.Buy(context.Background(), "", 9)
	if err != nil {
		t.Fatalf("buy: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestBuyMissingVOD(t *testing.T) {
	store := &fakeStore{user: map[string]interface{}{"uid": "5"}}
	service := NewService(store, store)

	retcode, errmsg, err := service.Buy(context.Background(), "token", 9)
	if err != nil {
		t.Fatalf("buy: %v", err)
	}
	if retcode != -1 || errmsg != "记录不存在或已被删除" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestBuyAlreadyBoughtReturnsOK(t *testing.T) {
	store := &fakeStore{
		user:  map[string]interface{}{"uid": "5"},
		vod:   map[string]interface{}{"vodid": "9", "showtype": "0", "view_price": "30"},
		count: 1,
	}
	service := NewService(store, store)

	retcode, errmsg, err := service.Buy(context.Background(), "token", 9)
	if err != nil {
		t.Fatalf("buy: %v", err)
	}
	if retcode != 0 || errmsg != "" || store.bought.vodid != 0 {
		t.Fatalf("response = %d %q bought=%#v", retcode, errmsg, store.bought)
	}
}

func TestBuyInsufficientGoldbean(t *testing.T) {
	store := &fakeStore{
		user: map[string]interface{}{"uid": "5"},
		vod:  map[string]interface{}{"vodid": "9", "showtype": "0", "view_price": "30"},
		bean: map[string]interface{}{"uid": "5", "gold_bean": "10"},
	}
	service := NewService(store, store)

	retcode, errmsg, err := service.Buy(context.Background(), "token", 9)
	if err != nil {
		t.Fatalf("buy: %v", err)
	}
	if retcode != 4 || errmsg != "金豆余额不足" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestBuyChargesVIPDiscount(t *testing.T) {
	store := &fakeStore{
		user: map[string]interface{}{"uid": "5", "perms": map[string]interface{}{"allow.vod.vip": "1"}},
		vod:  map[string]interface{}{"vodid": "9", "showtype": "0", "view_price": "30"},
		bean: map[string]interface{}{"uid": "5", "gold_bean": "30"},
	}
	service := NewService(store, store).WithVIPDiscount(50)

	retcode, errmsg, err := service.Buy(context.Background(), "token", 9)
	if err != nil {
		t.Fatalf("buy: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
	if store.bought.uid != 5 || store.bought.vodid != 9 || store.bought.price != 15 {
		t.Fatalf("bought = %#v", store.bought)
	}
}

func TestBuyZeroPriceDoesNotWriteBought(t *testing.T) {
	store := &fakeStore{
		user: map[string]interface{}{"uid": "5"},
		vod:  map[string]interface{}{"vodid": "9", "showtype": "0", "view_price": "0"},
		bean: map[string]interface{}{"uid": "5", "gold_bean": "0"},
	}
	service := NewService(store, store)

	retcode, errmsg, err := service.Buy(context.Background(), "token", 9)
	if err != nil {
		t.Fatalf("buy: %v", err)
	}
	if retcode != 0 || errmsg != "" || store.bought.vodid != 0 {
		t.Fatalf("response = %d %q bought=%#v", retcode, errmsg, store.bought)
	}
}
