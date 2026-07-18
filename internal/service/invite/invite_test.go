package invite

import (
	"context"
	"testing"

	"xj_comp/internal/domain"
)

type fakeStore struct {
	user       map[string]interface{}
	row        map[string]interface{}
	inviter    map[string]interface{}
	groups     []map[string]interface{}
	setting    map[string]interface{}
	deletedTag bool
	bindOK     bool
	bindInput  domain.InviteBindInput
}

func (s *fakeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s *fakeStore) RecordRecommend(context.Context, int) (map[string]interface{}, error) {
	return s.row, nil
}

func (s *fakeStore) UserByInviteKey(context.Context, string) (map[string]interface{}, error) {
	return s.inviter, nil
}

func (s *fakeStore) Groups(context.Context) ([]map[string]interface{}, error) {
	return s.groups, nil
}

func (s *fakeStore) SettingByUUID(context.Context, string) (map[string]interface{}, error) {
	return s.setting, nil
}

func (s *fakeStore) DeletedUserTag(context.Context, string) (bool, error) {
	return s.deletedTag, nil
}

func (s *fakeStore) BindInvite(_ context.Context, input domain.InviteBindInput) (bool, error) {
	s.bindInput = input
	return s.bindOK, nil
}

func TestInfoRequiresLogin(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, store)

	data, retcode, errmsg, err := service.Info(context.Background(), "")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if data != nil || retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("response = %#v %d %q", data, retcode, errmsg)
	}
}

func TestInfoNotBound(t *testing.T) {
	store := &fakeStore{user: map[string]interface{}{"uid": "5"}, row: map[string]interface{}{}}
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
	store := &fakeStore{user: map[string]interface{}{"uid": "5"}, row: map[string]interface{}{"uniqkey": "12345"}}
	service := NewService(store, store)

	data, retcode, errmsg, err := service.Info(context.Background(), "token")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if retcode != 0 || errmsg != "" || data["data"] != "9ix" {
		t.Fatalf("response = %#v %d %q", data, retcode, errmsg)
	}
}

func TestBindEdgePrechecks(t *testing.T) {
	tests := []struct {
		name    string
		user    map[string]interface{}
		row     map[string]interface{}
		inviter map[string]interface{}
		code    string
		retcode int
		errmsg  string
	}{
		{name: "guest", retcode: -9999, errmsg: "您还没有登录"},
		{name: "alreadyBoundBeforeMissingCode", user: map[string]interface{}{"uid": "5"}, row: map[string]interface{}{"uniqkey": "12345"}, retcode: -1, errmsg: "您已经绑定了邀请码:9ix"},
		{name: "missingCode", user: map[string]interface{}{"uid": "5"}, retcode: -1, errmsg: "请输入邀请码"},
		{name: "invalidCode", user: map[string]interface{}{"uid": "5"}, code: "abc", retcode: -1, errmsg: "无效邀请码"},
		{name: "self", user: map[string]interface{}{"uid": "5"}, inviter: map[string]interface{}{"uid": "5"}, code: "5", retcode: -1, errmsg: "无法绑定自己"},
		{name: "bindFailed", user: map[string]interface{}{"uid": "5"}, inviter: map[string]interface{}{"uid": "8"}, code: "abc", retcode: -1, errmsg: "绑定失败，请重试"},
	}

	for _, tt := range tests {
		auth := &fakeStore{user: tt.user}
		store := &fakeStore{row: tt.row, inviter: tt.inviter}
		service := NewService(auth, store)
		retcode, errmsg, err := service.BindEdge(context.Background(), "token", tt.code)
		if err != nil {
			t.Fatalf("%s bind edge: %v", tt.name, err)
		}
		if retcode != tt.retcode || errmsg != tt.errmsg {
			t.Fatalf("%s response = %d %q", tt.name, retcode, errmsg)
		}
	}
}

func TestBindReturnsInviteCodeOnSuccess(t *testing.T) {
	auth := &fakeStore{user: map[string]interface{}{"uid": "5", "mobi": "86.13800000000"}}
	store := &fakeStore{
		inviter: map[string]interface{}{"uid": "8"},
		groups: []map[string]interface{}{
			{"gid": "0", "minup": "0"},
			{"gid": "1", "minup": "1"},
		},
		setting: map[string]interface{}{"value": `a:5:{s:4:"reg1";i:10;s:4:"reg2";i:5;s:4:"reg3";i:1;s:7:"invite1";i:3;s:7:"invite2";i:2;}`},
		bindOK:  true,
	}
	service := NewService(auth, store)

	data, retcode, errmsg, err := service.Bind(context.Background(), "token", " abc ")
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if retcode != 0 || errmsg != "" || data["data"] != "abc" {
		t.Fatalf("response = %#v %d %q", data, retcode, errmsg)
	}
	if store.bindInput.UID != 5 || store.bindInput.InviterUID != 8 || store.bindInput.InviteCode != "abc" {
		t.Fatalf("bind input = %#v", store.bindInput)
	}
	if store.bindInput.Bonus["reg1"] != 10 || store.bindInput.Bonus["invite1"] != 3 || len(store.bindInput.Groups) != 2 {
		t.Fatalf("bind input bonus/groups = %#v", store.bindInput)
	}
}
