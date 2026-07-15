package explore

import (
	"context"
	"testing"
	"time"
)

type fakeStore struct {
	user          map[string]interface{}
	groups        []map[string]interface{}
	tabs          []map[string]interface{}
	userNotified  string
	guestNotified string
}

func (s fakeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s fakeStore) Groups(context.Context) ([]map[string]interface{}, error) {
	return s.groups, nil
}

func (s fakeStore) Tabs(context.Context) ([]map[string]interface{}, error) {
	return s.tabs, nil
}

func (s *fakeStore) UpdateUserNotificationAll(_ context.Context, _ int, value string) error {
	s.userNotified = value
	return nil
}

func (s *fakeStore) UpdateGuestNotificationAll(_ context.Context, _ string, value string) error {
	s.guestNotified = value
	return nil
}

func TestIndexGuest(t *testing.T) {
	store := &fakeStore{
		groups: []map[string]interface{}{
			{"gid": "0", "weight": "0", "perms": `{"max.signtask.coinnum1":1,"max.signtask.coinnum2":2,"max.signtask.coinnum3":3,"max.signtask.coinnum4":4,"max.signtask.coinnum5":5,"max.signtask.coinnum6":6,"max.signtask.coinnum7":7}`},
		},
		tabs: []map[string]interface{}{
			{"tabkey": "sign", "tabname": "签到", "intro": "i", "coverpic": "a.png", "coverpic2": "", "extjson": `{"x":1}`},
		},
	}
	service := NewService(store, store, "https://img.test")
	service.now = func() time.Time { return time.Date(2026, 7, 14, 12, 0, 0, 0, chinaLocation()) }

	data, err := service.Index(context.Background(), "")
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	if len(data.TabRows) != 1 || data.TabRows[0]["coverpic"] != "https://img.test/a.png" {
		t.Fatalf("tabrows = %#v", data.TabRows)
	}
	if len(data.DayRows) != 7 || data.DayRows[0]["day"] != "07-14" || data.DayRows[0]["coinnum"] != 1 {
		t.Fatalf("dayrows = %#v", data.DayRows)
	}
	if data.SignData["signed_today"] != 0 {
		t.Fatalf("signdata = %#v", data.SignData)
	}
}

func TestIndexSignedToday(t *testing.T) {
	store := &fakeStore{
		user: map[string]interface{}{
			"uid":             "5",
			"gid":             "0",
			"signed_lasttime": "1784019600",
			"signed_unitdays": "3",
			"signed_peakdays": "9",
			"signed_contdays": "4",
			"sysgid":          "0",
			"sysgid_exptime":  "0",
			"gids":            "",
		},
		groups: []map[string]interface{}{
			{"gid": "0", "weight": "0", "perms": `{"max.signtask.coinnum1":1,"max.signtask.coinnum2":2,"max.signtask.coinnum3":3,"max.signtask.coinnum4":4}`},
		},
	}
	service := NewService(store, store, "")
	service.now = func() time.Time { return time.Date(2026, 7, 14, 12, 0, 0, 0, chinaLocation()) }

	data, err := service.Index(context.Background(), "token")
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	if data.SignData["signed_today"] != 1 || data.SignData["signed_unitdays"] != 3 {
		t.Fatalf("signdata = %#v", data.SignData)
	}
	if data.DayRows[0]["coinnum"] != 3 {
		t.Fatalf("dayrows = %#v", data.DayRows)
	}
}

func TestCleanNotificationRequiresTabKey(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, store, "")

	_, retcode, errmsg, err := service.CleanNotification(context.Background(), "", "")
	if err != nil {
		t.Fatalf("clean: %v", err)
	}
	if retcode != -1 || errmsg != "请提供频道键名" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestCleanNotificationAll(t *testing.T) {
	store := &fakeStore{
		user: map[string]interface{}{"uid": "5", "notification_all": `{"sign":2}`},
	}
	service := NewService(store, store, "")

	data, retcode, errmsg, err := service.CleanNotification(context.Background(), "token", "all")
	if err != nil {
		t.Fatalf("clean: %v", err)
	}
	if retcode != 0 || errmsg != "" || store.userNotified != "null" {
		t.Fatalf("response = %d %q notified=%q", retcode, errmsg, store.userNotified)
	}
	if data["notification_all"] != nil {
		t.Fatalf("data = %#v", data)
	}
}

func TestCleanNotificationTab(t *testing.T) {
	store := &fakeStore{
		user: map[string]interface{}{"uid": "0", "sid": "guest", "notification_all": `{"sign":2}`},
	}
	service := NewService(store, store, "")

	data, retcode, errmsg, err := service.CleanNotification(context.Background(), "", "sign")
	if err != nil {
		t.Fatalf("clean: %v", err)
	}
	if retcode != 0 || errmsg != "" || store.guestNotified != `{"sign":0}` {
		t.Fatalf("response = %d %q notified=%q", retcode, errmsg, store.guestNotified)
	}
	values := data["notification_all"].(map[string]interface{})
	if values["sign"] != float64(0) {
		t.Fatalf("data = %#v", data)
	}
}
