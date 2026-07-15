package attach

import (
	"context"
	"testing"
)

type fakeStore struct {
	user    map[string]interface{}
	updated bool
	uid     int
	avatar  string
}

func (s *fakeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s *fakeStore) UpdateAvatar(_ context.Context, uid int, avatarID string) error {
	s.updated = true
	s.uid = uid
	s.avatar = avatarID
	return nil
}

func TestUpAvatarRequiresLogin(t *testing.T) {
	service := NewService(&fakeStore{})

	retcode, errmsg, err := service.UpAvatar(context.Background(), "", "1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestUpAvatarRejectsInvalidAvatarID(t *testing.T) {
	store := &fakeStore{user: map[string]interface{}{"uid": "7"}}
	service := NewService(store)

	retcode, errmsg, err := service.UpAvatar(context.Background(), "sid", "abc")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if retcode != -1 || errmsg != "请选择系统头像" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if store.updated {
		t.Fatal("expected update not to be called")
	}
}

func TestUpAvatarUpdatesAvatar(t *testing.T) {
	store := &fakeStore{user: map[string]interface{}{"uid": "7"}}
	service := NewService(store)

	retcode, errmsg, err := service.UpAvatar(context.Background(), "sid", "12")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if !store.updated || store.uid != 7 || store.avatar != "12" {
		t.Fatalf("unexpected update uid=%d avatar=%q updated=%v", store.uid, store.avatar, store.updated)
	}
}
