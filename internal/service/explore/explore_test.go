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
	vodTask       map[string]interface{}
	userLog       map[string]interface{}
	guestLog      map[string]interface{}
	createdUser   map[string]interface{}
	createdGuest  map[string]interface{}
	reqCoinCall   map[string]interface{}
	reqCoinRet    int
	reqCoinMsg    string
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

func (s *fakeStore) VodTaskByID(context.Context, int) (map[string]interface{}, error) {
	if s.vodTask != nil {
		return s.vodTask, nil
	}
	return map[string]interface{}{}, nil
}

func (s *fakeStore) UserVodTaskLog(context.Context, int, int64, int) (map[string]interface{}, error) {
	if s.userLog != nil {
		return s.userLog, nil
	}
	return map[string]interface{}{}, nil
}

func (s *fakeStore) GuestVodTaskLog(context.Context, string, int64, int) (map[string]interface{}, error) {
	if s.guestLog != nil {
		return s.guestLog, nil
	}
	return map[string]interface{}{}, nil
}

func (s *fakeStore) CreateUserVodTaskLog(_ context.Context, uid int, vid int, addtime int64, reqcoin int) (int, error) {
	s.createdUser = map[string]interface{}{"uid": uid, "vid": vid, "addtime": addtime, "reqcoin": reqcoin}
	return 77, nil
}

func (s *fakeStore) CreateGuestVodTaskLog(_ context.Context, sid string, vid int, addtime int64, reqcoin int) (int, error) {
	s.createdGuest = map[string]interface{}{"sid": sid, "vid": vid, "addtime": addtime, "reqcoin": reqcoin}
	return 88, nil
}

func (s *fakeStore) ReqVodTaskCoin(_ context.Context, uid int, sid string, logid int, now int64) (int, string, error) {
	s.reqCoinCall = map[string]interface{}{"uid": uid, "sid": sid, "logid": logid, "now": now}
	return s.reqCoinRet, s.reqCoinMsg, nil
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

func TestVodTaskShowMissing(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, store, "")

	_, retcode, errmsg, err := service.VodTaskShow(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if retcode != -1 || errmsg != "记录不存在或已被删除" {
		t.Fatalf("response=%d %q", retcode, errmsg)
	}
}

func TestVodTaskShowExistingUserLog(t *testing.T) {
	store := &fakeStore{
		user:    map[string]interface{}{"uid": "5", "gid": "0"},
		groups:  []map[string]interface{}{},
		vodTask: vodTaskRow(),
		userLog: map[string]interface{}{"logid": "9", "reqcoin": "6", "reqtime": "123"},
	}
	service := NewService(store, store, "https://res.test")
	service.now = func() time.Time { return time.Date(2026, 7, 14, 12, 0, 0, 0, chinaLocation()) }

	data, retcode, errmsg, err := service.VodTaskShow(context.Background(), "token", 3)
	if err != nil || retcode != 0 || errmsg != "" {
		t.Fatalf("data=%#v response=%d %q err=%v", data, retcode, errmsg, err)
	}
	if data["logid"] != 9 || data["reqcoin"] != 6 || data["reqtime"] != 123 {
		t.Fatalf("data=%#v", data)
	}
	vodrow := data["vodrow"].(map[string]interface{})
	if vodrow["coverpic"] != "https://res.test/c.png" || vodrow["picon"] != "https://res.test/p.png" {
		t.Fatalf("vodrow=%#v", vodrow)
	}
}

func TestVodTaskShowCreatesGuestLog(t *testing.T) {
	store := &fakeStore{
		user:    map[string]interface{}{"uid": "0", "sid": "guest"},
		groups:  []map[string]interface{}{},
		vodTask: vodTaskRow(),
	}
	service := NewService(store, store, "")
	service.now = func() time.Time { return time.Unix(1700000000, 0) }
	service.randIntn = func(n int) int { return n - 1 }

	data, retcode, errmsg, err := service.VodTaskShow(context.Background(), "", 3)
	if err != nil || retcode != 0 || errmsg != "" {
		t.Fatalf("data=%#v response=%d %q err=%v", data, retcode, errmsg, err)
	}
	if data["logid"] != 88 || data["reqcoin"] != 5 || data["reqtime"] != 0 {
		t.Fatalf("data=%#v", data)
	}
	if store.createdGuest["sid"] != "guest" || store.createdGuest["reqcoin"] != 5 {
		t.Fatalf("created=%#v", store.createdGuest)
	}
}

func TestVodTaskReqCoinUser(t *testing.T) {
	store := &fakeStore{
		user:       map[string]interface{}{"uid": "5", "sid": "s", "gid": "0"},
		groups:     []map[string]interface{}{},
		reqCoinMsg: "领取成功",
	}
	service := NewService(store, store, "")
	service.now = func() time.Time { return time.Unix(1700000000, 0) }

	retcode, errmsg, err := service.VodTaskReqCoin(context.Background(), "token", 9)
	if err != nil {
		t.Fatalf("reqcoin: %v", err)
	}
	if retcode != 0 || errmsg != "领取成功" {
		t.Fatalf("response=%d %q", retcode, errmsg)
	}
	if store.reqCoinCall["uid"] != 5 || store.reqCoinCall["logid"] != 9 || store.reqCoinCall["now"] != int64(1700000000) {
		t.Fatalf("call=%#v", store.reqCoinCall)
	}
}

func TestVodTaskReqCoinPassesStoreError(t *testing.T) {
	store := &fakeStore{
		user:       map[string]interface{}{"uid": "0", "sid": "guest"},
		reqCoinRet: -1,
		reqCoinMsg: "您已经领取过金币了",
	}
	service := NewService(store, store, "")

	retcode, errmsg, err := service.VodTaskReqCoin(context.Background(), "", 9)
	if err != nil {
		t.Fatalf("reqcoin: %v", err)
	}
	if retcode != -1 || errmsg != "您已经领取过金币了" || store.reqCoinCall["sid"] != "guest" {
		t.Fatalf("response=%d %q call=%#v", retcode, errmsg, store.reqCoinCall)
	}
}

func vodTaskRow() map[string]interface{} {
	return map[string]interface{}{
		"vid":       "3",
		"showtype":  "0",
		"title":     "视频任务",
		"intro":     "intro",
		"coverpic":  "c.png",
		"playurl":   "https://video.test/a.mp4",
		"portrait":  "1",
		"countdown": "15",
		"pname":     "产品",
		"pdscr":     "说明",
		"picon":     "p.png",
		"purl":      "https://p.test",
		"mincoin":   "3",
		"maxcoin":   "5",
	}
}
