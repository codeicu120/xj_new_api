package sendfile

import (
	"context"
	"testing"
)

type fakeAuthStore struct {
	user map[string]interface{}
}

func (s fakeAuthStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

type fakeVODStore struct {
	row map[string]interface{}
}

func (s fakeVODStore) VODByID(context.Context, int) (map[string]interface{}, error) {
	return s.row, nil
}

func TestPlayRequiresLogin(t *testing.T) {
	service := NewService(fakeAuthStore{}, fakeVODStore{})

	retcode, errmsg, err := service.Play(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("play: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestPlayMissingVOD(t *testing.T) {
	service := NewService(fakeAuthStore{user: map[string]interface{}{"uid": "5"}}, fakeVODStore{row: map[string]interface{}{}})

	retcode, errmsg, err := service.Play(context.Background(), "token", 0)
	if err != ErrVODNotFound {
		t.Fatalf("expected ErrVODNotFound, got %v", err)
	}
	if retcode != -1 || errmsg != "记录不存在或已被删除" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestPlayValidVODReturnsEmptySuccess(t *testing.T) {
	service := NewService(
		fakeAuthStore{user: map[string]interface{}{"uid": "5"}},
		fakeVODStore{row: map[string]interface{}{"vodid": "1", "showtype": "0"}},
	)

	retcode, errmsg, err := service.Play(context.Background(), "token", 1)
	if err != nil {
		t.Fatalf("play: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}
