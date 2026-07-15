package verification

import (
	"context"
	"testing"
)

type fakeStore struct {
	setting string
	user    map[string]interface{}
}

func (s fakeStore) SettingByUUID(context.Context, string) (map[string]interface{}, error) {
	return map[string]interface{}{"value": s.setting}, nil
}

func (s fakeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

type allowCaptcha struct{}

func (allowCaptcha) VerifyImage(context.Context, string, string) bool { return true }
func (allowCaptcha) VerifyGoogle(context.Context, string) bool        { return true }
func (allowCaptcha) VerifyTencent(context.Context, string, string, string) bool {
	return true
}
func (allowCaptcha) VerifySelf(context.Context, string, string) bool { return true }

type okSMS struct{}

func (okSMS) SendSMS(context.Context, int, string, string, string, int) error { return nil }

type okMail struct{}

func (okMail) SendMail(context.Context, map[string]interface{}, string, string, string) error {
	return nil
}

func TestSendVValidation(t *testing.T) {
	service := NewService(fakeStore{setting: `a:2:{s:10:"smscaptcha";i:1;s:13:"smsphonelimit";i:5;}`}, nil, nil, nil, nil)
	msg, err := service.SendV(context.Background(), SendSMSRequest{Mobi: "bad"})
	if err != nil {
		t.Fatal(err)
	}
	if msg != "手机号码填写不正确" {
		t.Fatalf("msg = %q", msg)
	}
	msg, err = service.SendV(context.Background(), SendSMSRequest{Mobi: "14012340002"})
	if err != nil {
		t.Fatal(err)
	}
	if msg != "未提供图形验证码串" {
		t.Fatalf("msg = %q", msg)
	}
}

func TestSendUSuccessWithFakeSender(t *testing.T) {
	service := NewService(
		fakeStore{
			setting: `a:5:{s:10:"smscaptcha";i:0;s:13:"smsphonelimit";i:5;s:11:"smsplatform";i:0;s:24:"smsplatforminternational";i:0;s:9:"smsconfig";s:2:"{}";}`,
			user:    map[string]interface{}{"uid": "5", "mobi": "86.14012340002"},
		},
		nil,
		allowCaptcha{},
		okSMS{},
		nil,
	)
	msg, err := service.SendU(context.Background(), SendSMSRequest{Token: "token"})
	if err != nil {
		t.Fatal(err)
	}
	if msg != "短信已成功发送" {
		t.Fatalf("msg = %q", msg)
	}
}

func TestEmailValidationAndSuccess(t *testing.T) {
	service := NewService(
		fakeStore{setting: `a:3:{s:10:"smscaptcha";i:0;s:13:"smsphonelimit";i:5;s:8:"mailconf";s:7:"{"a":1}";}`},
		nil,
		allowCaptcha{},
		nil,
		okMail{},
	)
	msg, err := service.SendEmail(context.Background(), SendEmailRequest{Email: "bad"})
	if err != nil {
		t.Fatal(err)
	}
	if msg != "邮箱格式填写不正确" {
		t.Fatalf("msg = %q", msg)
	}
	msg, err = service.SendEmail(context.Background(), SendEmailRequest{Email: "a@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if msg != "验证码已发送至您的邮箱，请10分钟内验证并确认" {
		t.Fatalf("msg = %q", msg)
	}
}
