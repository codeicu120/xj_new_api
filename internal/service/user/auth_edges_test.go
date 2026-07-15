package user

import (
	"context"
	"testing"
)

type fakeAuthEdgeStore struct {
	user map[string]interface{}
}

func (s fakeAuthEdgeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func TestRegisterEdgeBranches(t *testing.T) {
	service := NewAuthEdgeService(fakeAuthEdgeStore{})

	retcode, errmsg, err := service.Register(context.Background(), AuthEdgeRequest{})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if retcode != -1 || errmsg != "请同意用户协议" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}

	logged := NewAuthEdgeService(fakeAuthEdgeStore{user: map[string]interface{}{"uid": "5"}})
	retcode, errmsg, err = logged.Register(context.Background(), AuthEdgeRequest{Token: "250f790ba71ec2b9d3855f424db2259e", AUP: 1})
	if err != nil {
		t.Fatalf("register logged: %v", err)
	}
	if retcode != -1 || errmsg != "用户已登录" {
		t.Fatalf("unexpected logged response %d %q", retcode, errmsg)
	}
}

func TestV2LoginEmptyUsernameBranch(t *testing.T) {
	service := NewAuthEdgeService(fakeAuthEdgeStore{})

	retcode, errmsg, err := service.Login(context.Background(), AuthEdgeRequest{}, true)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if retcode != -1 || errmsg != "用户名未注册" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestForgotEdgeBranches(t *testing.T) {
	service := NewAuthEdgeService(nil)

	if retcode, errmsg := service.Forgot(AuthEdgeRequest{}, true); retcode != -1 || errmsg != "请填写手机号码或者邮箱" {
		t.Fatalf("unexpected v2 response %d %q", retcode, errmsg)
	}
	if retcode, errmsg := service.Forgot(AuthEdgeRequest{}, false); retcode != -1 || errmsg != "手机号码填写不正确" {
		t.Fatalf("unexpected v1 response %d %q", retcode, errmsg)
	}
	if retcode, errmsg := service.Forgot(AuthEdgeRequest{Mobi: "13800138000"}, false); retcode != -1 || errmsg != "无效的操作" {
		t.Fatalf("unexpected step response %d %q", retcode, errmsg)
	}
}

func TestDeleteAndChangePhoneRequireLogin(t *testing.T) {
	service := NewAuthEdgeService(fakeAuthEdgeStore{})

	retcode, errmsg, err := service.Delete(context.Background(), "")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected delete response %d %q", retcode, errmsg)
	}

	retcode, errmsg, err = service.ChangePhone(context.Background(), AuthEdgeRequest{})
	if err != nil {
		t.Fatalf("change phone: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作" {
		t.Fatalf("unexpected change phone response %d %q", retcode, errmsg)
	}
}
