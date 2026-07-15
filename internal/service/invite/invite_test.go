package invite

import (
	"context"
	"testing"
)

type fakeStore struct {
	user map[string]interface{}
	row  map[string]interface{}
}

func (s fakeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s fakeStore) RecordRecommend(context.Context, int) (map[string]interface{}, error) {
	return s.row, nil
}

func TestInfoRequiresLogin(t *testing.T) {
	service := NewService(fakeStore{}, fakeStore{})

	data, retcode, errmsg, err := service.Info(context.Background(), "")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if data != nil || retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("response = %#v %d %q", data, retcode, errmsg)
	}
}

func TestInfoNotBound(t *testing.T) {
	store := fakeStore{user: map[string]interface{}{"uid": "5"}, row: map[string]interface{}{}}
	service := NewService(store, store)

	data, retcode, errmsg, err := service.Info(context.Background(), "token")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if retcode != 0 || errmsg != "" || data["data"] != nil {
		t.Fatalf("response = %#v %d %q", data, retcode, errmsg)
	}
}

func TestInfoReturnsBase36RecommendKey(t *testing.T) {
	store := fakeStore{user: map[string]interface{}{"uid": "5"}, row: map[string]interface{}{"uniqkey": "12345"}}
	service := NewService(store, store)

	data, retcode, errmsg, err := service.Info(context.Background(), "token")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if retcode != 0 || errmsg != "" || data["data"] != "9ix" {
		t.Fatalf("response = %#v %d %q", data, retcode, errmsg)
	}
}
