package user

import (
	"context"
	"testing"
)

type fakeAuthEdgeStore struct {
	user    map[string]interface{}
	byMobi  map[string]interface{}
	byEmail map[string]interface{}
	byUser  map[string]interface{}
}

func (s fakeAuthEdgeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s fakeAuthEdgeStore) UserByMobi(context.Context, string) (map[string]interface{}, error) {
	return s.byMobi, nil
}

func (s fakeAuthEdgeStore) UserByEmail(context.Context, string) (map[string]interface{}, error) {
	return s.byEmail, nil
}

func (s fakeAuthEdgeStore) UserByUsername(context.Context, string) (map[string]interface{}, error) {
	return s.byUser, nil
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

	retcode, errmsg, err = service.Login(context.Background(), AuthEdgeRequest{Mobi: "13800138000"}, true)
	if err != nil {
		t.Fatalf("login mobi: %v", err)
	}
	if retcode != -1 || errmsg != "手机号码未注册" {
		t.Fatalf("unexpected mobi response %d %q", retcode, errmsg)
	}

	retcode, errmsg, err = service.Login(context.Background(), AuthEdgeRequest{Email: "nobody@example.com"}, true)
	if err != nil {
		t.Fatalf("login email: %v", err)
	}
	if retcode != -1 || errmsg != "邮箱未注册" {
		t.Fatalf("unexpected email response %d %q", retcode, errmsg)
	}
}

func TestForgotEdgeBranches(t *testing.T) {
	service := NewAuthEdgeService(nil)

	if retcode, errmsg, err := service.Forgot(context.Background(), AuthEdgeRequest{}, true); err != nil || retcode != -1 || errmsg != "请填写手机号码或者邮箱" {
		t.Fatalf("unexpected v2 response %d %q", retcode, errmsg)
	}
	if retcode, errmsg, err := service.Forgot(context.Background(), AuthEdgeRequest{}, false); err != nil || retcode != -1 || errmsg != "手机号码填写不正确" {
		t.Fatalf("unexpected v1 response %d %q", retcode, errmsg)
	}
	if retcode, errmsg, err := service.Forgot(context.Background(), AuthEdgeRequest{Mobi: "13800138000"}, false); err != nil || retcode != -1 || errmsg != "无效的操作" {
		t.Fatalf("unexpected step response %d %q", retcode, errmsg)
	}

	service = NewAuthEdgeService(fakeAuthEdgeStore{})
	if retcode, errmsg, err := service.Forgot(context.Background(), AuthEdgeRequest{Mobi: "13800138000", Step: "step1"}, false); err != nil || retcode != -1 || errmsg != "输入的手机号码不存在" {
		t.Fatalf("unexpected missing mobile response %d %q err=%v", retcode, errmsg, err)
	}

	service = NewAuthEdgeService(fakeAuthEdgeStore{byEmail: map[string]interface{}{"uid": "9"}})
	if retcode, errmsg, err := service.Forgot(context.Background(), AuthEdgeRequest{Email: "ok@example.com", Step: "step1"}, true); err != nil || retcode != 0 || errmsg != "step1->step2" {
		t.Fatalf("unexpected email step1 response %d %q err=%v", retcode, errmsg, err)
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

func TestChangePhoneStep1ReadOnlyBranches(t *testing.T) {
	service := NewAuthEdgeService(fakeAuthEdgeStore{user: map[string]interface{}{"uid": "7", "mobi": "86.13800138000"}})
	retcode, errmsg, err := service.ChangePhone(context.Background(), AuthEdgeRequest{Token: "250f790ba71ec2b9d3855f424db2259e", Mobi: "13800138000", Step: "step1"})
	if err != nil {
		t.Fatalf("same mobi: %v", err)
	}
	if retcode != -1 || errmsg != "更换的手机号和当前手机号相同！" {
		t.Fatalf("unexpected same mobi response %d %q", retcode, errmsg)
	}

	service = NewAuthEdgeService(fakeAuthEdgeStore{user: map[string]interface{}{"uid": "7", "mobi": "86.13800138000"}, byMobi: map[string]interface{}{"uid": "8"}})
	retcode, errmsg, err = service.ChangePhone(context.Background(), AuthEdgeRequest{Token: "250f790ba71ec2b9d3855f424db2259e", Mobi: "13900139000", Step: "step1"})
	if err != nil {
		t.Fatalf("existing mobi: %v", err)
	}
	if retcode != -1 || errmsg != "手机号已经存在" {
		t.Fatalf("unexpected existing mobi response %d %q", retcode, errmsg)
	}

	service = NewAuthEdgeService(fakeAuthEdgeStore{user: map[string]interface{}{"uid": "7", "mobi": "86.13800138000"}})
	retcode, errmsg, err = service.ChangePhone(context.Background(), AuthEdgeRequest{Token: "250f790ba71ec2b9d3855f424db2259e", Mobi: "13900139000", Step: "step1"})
	if err != nil {
		t.Fatalf("step1: %v", err)
	}
	if retcode != 0 || errmsg != "step1->step2" {
		t.Fatalf("unexpected step1 response %d %q", retcode, errmsg)
	}
}
