package open

import (
	"context"
	"testing"
	"time"
)

type fakeAuthStore struct {
	user map[string]interface{}
	err  error
}

func (s fakeAuthStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, s.err
}

type fakeGuestStore struct {
	exists  bool
	created bool
	sid     string
	now     int64
	err     error
}

func (s *fakeGuestStore) GuestExists(context.Context, string) (bool, error) {
	return s.exists, s.err
}

func (s *fakeGuestStore) CreateGuest(_ context.Context, sid string, now int64) error {
	s.created = true
	s.sid = sid
	s.now = now
	return s.err
}

func TestReqAuthInvalidAppID(t *testing.T) {
	service := NewService(fakeAuthStore{}, &fakeGuestStore{}, "")

	data, retcode, errmsg, err := service.ReqAuth(context.Background(), "", "127.0.0.1", "bad")
	if err != nil {
		t.Fatalf("ReqAuth error = %v", err)
	}
	if data != nil || retcode != -1 || errmsg != "请输入正确的appid" {
		t.Fatalf("ReqAuth invalid app = data:%v retcode:%d errmsg:%q", data, retcode, errmsg)
	}
}

func TestReqAuthGuestCreatesGuestAndSignsPayload(t *testing.T) {
	guestStore := &fakeGuestStore{}
	service := NewService(fakeAuthStore{}, guestStore, "https://res.example")
	service.now = func() time.Time { return time.Unix(1700000000, 0) }

	data, retcode, errmsg, err := service.ReqAuth(context.Background(), "", "127.0.0.1", "4b4131e49")
	if err != nil {
		t.Fatalf("ReqAuth error = %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("ReqAuth retcode = %d errmsg = %q", retcode, errmsg)
	}
	wantSID := guestSID("127.0.0.1")
	if !guestStore.created || guestStore.sid != wantSID || guestStore.now != 1700000000 {
		t.Fatalf("guest creation = created:%v sid:%q now:%d", guestStore.created, guestStore.sid, guestStore.now)
	}
	authrow := data["authrow"].(map[string]interface{})
	if authrow["deviceString"] != wantSID {
		t.Fatalf("deviceString = %v, want %s", authrow["deviceString"], wantSID)
	}
	if _, ok := authrow["phoneNumber"]; ok {
		t.Fatalf("guest authrow should omit phoneNumber: %#v", authrow)
	}
	if data["time"] != int64(1700000000) {
		t.Fatalf("time = %v", data["time"])
	}
	if data["sign"] == "" || data["openid"] == "" {
		t.Fatalf("missing sign/openid: %#v", data)
	}
}

func TestReqAuthLoggedUserUsesUserProfile(t *testing.T) {
	service := NewService(fakeAuthStore{user: map[string]interface{}{
		"uid":      "42",
		"mobi":     "13800138000",
		"avatar":   "avatar.jpg",
		"gender":   "1",
		"nickname": "",
		"username": "tester",
	}}, &fakeGuestStore{exists: true}, "https://res.example")
	service.now = func() time.Time { return time.Unix(1700000000, 0) }

	data, retcode, errmsg, err := service.ReqAuth(context.Background(), "0123456789abcdef0123456789abcdef", "127.0.0.1", "4b4131e49")
	if err != nil {
		t.Fatalf("ReqAuth error = %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("ReqAuth retcode = %d errmsg = %q", retcode, errmsg)
	}
	authrow := data["authrow"].(map[string]interface{})
	if authrow["phoneNumber"] != "13800138000" || authrow["nickName"] != "tester" || authrow["gender"] != 1 {
		t.Fatalf("authrow mismatch: %#v", authrow)
	}
	if authrow["headUrl"] != "https://res.example/C1/avatar.jpg" {
		t.Fatalf("headUrl = %v", authrow["headUrl"])
	}
	if _, ok := authrow["deviceString"]; ok {
		t.Fatalf("logged authrow should omit deviceString: %#v", authrow)
	}
}
