package user

import (
	"context"
	"testing"
	"time"
)

type fakeAuthEdgeStore struct {
	user      map[string]interface{}
	byID      map[string]interface{}
	byMobi    map[string]interface{}
	byEmail   map[string]interface{}
	byUser    map[string]interface{}
	settings  map[string]map[string]interface{}
	keyCounts map[string]int
}

func (s fakeAuthEdgeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s fakeAuthEdgeStore) UserByID(context.Context, int) (map[string]interface{}, error) {
	return s.byID, nil
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

func (s fakeAuthEdgeStore) SettingByUUID(_ context.Context, uuid string) (map[string]interface{}, error) {
	if s.settings != nil {
		return s.settings[uuid], nil
	}
	return map[string]interface{}{}, nil
}

func (s fakeAuthEdgeStore) KeylimitCountSince(_ context.Context, key string, _ int64) (int, error) {
	if s.keyCounts != nil {
		return s.keyCounts[key], nil
	}
	return 0, nil
}

func TestRegisterEdgeBranches(t *testing.T) {
	service := NewAuthEdgeService(fakeAuthEdgeStore{})

	retcode, errmsg, err := service.Register(context.Background(), AuthEdgeRequest{}, false)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if retcode != -1 || errmsg != "请同意用户协议" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}

	logged := NewAuthEdgeService(fakeAuthEdgeStore{user: map[string]interface{}{"uid": "5"}})
	retcode, errmsg, err = logged.Register(context.Background(), AuthEdgeRequest{Token: "250f790ba71ec2b9d3855f424db2259e", AUP: 1}, false)
	if err != nil {
		t.Fatalf("register logged: %v", err)
	}
	if retcode != -1 || errmsg != "用户已登录" {
		t.Fatalf("unexpected logged response %d %q", retcode, errmsg)
	}

	retcode, errmsg, err = service.Register(context.Background(), AuthEdgeRequest{AUP: 1, RegType: 2}, true)
	if err != nil {
		t.Fatalf("v2 register mobile: %v", err)
	}
	if retcode != -1 || errmsg != "手机号码填写不正确" {
		t.Fatalf("unexpected v2 mobile response %d %q", retcode, errmsg)
	}

	retcode, errmsg, err = service.Register(context.Background(), AuthEdgeRequest{AUP: 1, RegType: 3, Email: "bad"}, true)
	if err != nil {
		t.Fatalf("v2 register email: %v", err)
	}
	if retcode != -1 || errmsg != "请输入正确邮箱地址" {
		t.Fatalf("unexpected v2 email response %d %q", retcode, errmsg)
	}

	retcode, errmsg, err = service.Register(context.Background(), AuthEdgeRequest{AUP: 1, RegType: 1, Username: "abcdef", Password: "123"}, true)
	if err != nil {
		t.Fatalf("v2 register password: %v", err)
	}
	if retcode != -1 || errmsg != "密码6-16位" {
		t.Fatalf("unexpected v2 password response %d %q", retcode, errmsg)
	}
}

func TestRegisterReadOnlyValidationBranches(t *testing.T) {
	service := NewAuthEdgeService(fakeAuthEdgeStore{})

	retcode, errmsg, err := service.Register(context.Background(), AuthEdgeRequest{AUP: 1, Mobi: "13800138000"}, false)
	if err != nil {
		t.Fatalf("v1 register mobi: %v", err)
	}
	if retcode != -1 || errmsg != "注册成功分支暂未迁移" {
		t.Fatalf("unexpected v1 mobi response %d %q", retcode, errmsg)
	}

	service = NewAuthEdgeService(fakeAuthEdgeStore{byMobi: map[string]interface{}{"uid": "9"}})
	retcode, errmsg, err = service.Register(context.Background(), AuthEdgeRequest{AUP: 1, Mobi: "13800138000"}, false)
	if err != nil {
		t.Fatalf("v1 register duplicate mobi: %v", err)
	}
	if retcode != -1 || errmsg != "手机号码已被注册" {
		t.Fatalf("unexpected duplicate mobi response %d %q", retcode, errmsg)
	}

	service = NewAuthEdgeService(fakeAuthEdgeStore{})
	retcode, errmsg, err = service.Register(context.Background(), AuthEdgeRequest{AUP: 1, RegType: 1, Username: "123456", Password: "123456"}, true)
	if err != nil {
		t.Fatalf("v2 register numeric username: %v", err)
	}
	if retcode != -1 || errmsg != "用户名不能是纯数字" {
		t.Fatalf("unexpected numeric username response %d %q", retcode, errmsg)
	}

	retcode, errmsg, err = service.Register(context.Background(), AuthEdgeRequest{AUP: 1, RegType: 1, Username: "bad!", Password: "123456"}, true)
	if err != nil {
		t.Fatalf("v2 register invalid username: %v", err)
	}
	if retcode != -1 || errmsg != "用户名2-8个汉字，英文6-16个字符" {
		t.Fatalf("unexpected invalid username response %d %q", retcode, errmsg)
	}

	service = NewAuthEdgeService(fakeAuthEdgeStore{byUser: map[string]interface{}{"uid": "9"}})
	retcode, errmsg, err = service.Register(context.Background(), AuthEdgeRequest{AUP: 1, RegType: 1, Username: "abcdef", Password: "123456"}, true)
	if err != nil {
		t.Fatalf("v2 register duplicate username: %v", err)
	}
	if retcode != -1 || errmsg != "用户名已存在" {
		t.Fatalf("unexpected duplicate username response %d %q", retcode, errmsg)
	}

	service = NewAuthEdgeService(fakeAuthEdgeStore{byEmail: map[string]interface{}{"uid": "9"}})
	retcode, errmsg, err = service.Register(context.Background(), AuthEdgeRequest{AUP: 1, RegType: 3, Email: "used@example.com"}, true)
	if err != nil {
		t.Fatalf("v2 register duplicate email: %v", err)
	}
	if retcode != -1 || errmsg != "该邮箱已经被注册，您可以通过邮箱找回密码" {
		t.Fatalf("unexpected duplicate email response %d %q", retcode, errmsg)
	}
}

func TestRegisterPolicyBranches(t *testing.T) {
	service := NewAuthEdgeService(fakeAuthEdgeStore{settings: map[string]map[string]interface{}{
		"user.regopt": {"value": `a:1:{s:9:"regclosed";i:1;}`},
	}})
	retcode, errmsg, err := service.Register(context.Background(), AuthEdgeRequest{AUP: 1, Mobi: "13800138000"}, false)
	if err != nil {
		t.Fatalf("register closed: %v", err)
	}
	if retcode != -1 || errmsg != "已暂时关闭了注册" {
		t.Fatalf("unexpected register closed response %d %q", retcode, errmsg)
	}

	service = NewAuthEdgeService(fakeAuthEdgeStore{keyCounts: map[string]int{"user.regiser.ip.127.0.0.1": 1}})
	service.now = func() time.Time { return time.Unix(1700000000, 0) }
	retcode, errmsg, err = service.Register(context.Background(), AuthEdgeRequest{AUP: 1, Mobi: "13800138000", ClientIP: "127.0.0.1"}, false)
	if err != nil {
		t.Fatalf("register ip limited: %v", err)
	}
	if retcode != -1 || errmsg != "注册过于频繁，请稍后再试" {
		t.Fatalf("unexpected ip limit response %d %q", retcode, errmsg)
	}
}

func TestLoginPasswordClosed(t *testing.T) {
	service := NewAuthEdgeService(fakeAuthEdgeStore{settings: map[string]map[string]interface{}{
		"setting": {"value": `a:1:{s:15:"pswdLoginStatus";i:0;}`},
	}})
	retcode, errmsg, err := service.Login(context.Background(), AuthEdgeRequest{}, false)
	if err != nil {
		t.Fatalf("login closed: %v", err)
	}
	if retcode != -1 || errmsg != "系统已关闭密码登录" {
		t.Fatalf("unexpected login closed response %d %q", retcode, errmsg)
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

	service = NewAuthEdgeService(fakeAuthEdgeStore{byMobi: map[string]interface{}{"uid": "9"}})
	retcode, errmsg, err = service.Login(context.Background(), AuthEdgeRequest{Mobi: "13800138000"}, true)
	if err != nil {
		t.Fatalf("login password: %v", err)
	}
	if retcode != -1 || errmsg != "密码不能为空" {
		t.Fatalf("unexpected password response %d %q", retcode, errmsg)
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

func TestDeleteGuestAccountBranch(t *testing.T) {
	service := NewAuthEdgeService(fakeAuthEdgeStore{
		user: map[string]interface{}{"uid": "7"},
		byID: map[string]interface{}{"uid": "7", "mobi": "~86.abc", "email": "~abc"},
	})

	retcode, errmsg, err := service.Delete(context.Background(), "250f790ba71ec2b9d3855f424db2259e")
	if err != nil {
		t.Fatalf("delete guest: %v", err)
	}
	if retcode != -1 || errmsg != "游客账号无需注销" {
		t.Fatalf("unexpected guest delete response %d %q", retcode, errmsg)
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
