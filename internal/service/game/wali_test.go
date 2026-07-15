package game

import (
	"context"
	"encoding/base64"
	"net/url"
	"strings"
	"testing"
	"time"
)

type fakeWaliStore struct {
	row      map[string]interface{}
	settings map[string]string
	quota    map[string]interface{}
}

func (s fakeWaliStore) PlatformByID(context.Context, int) (map[string]interface{}, error) {
	return s.row, nil
}

func (s fakeWaliStore) Setting(_ context.Context, key string) (string, error) {
	return s.settings[key], nil
}

func (s fakeWaliStore) Quota(context.Context, int) (map[string]interface{}, error) {
	return s.quota, nil
}

type fakeWaliAuthStore struct {
	user map[string]interface{}
}

func (s fakeWaliAuthStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

type fakeWaliClient struct {
	rawURL string
	body   []byte
}

func (c *fakeWaliClient) Get(_ context.Context, rawURL string) ([]byte, error) {
	c.rawURL = rawURL
	return c.body, nil
}

func TestWaliServicePingBuildsSignedEncryptedRequest(t *testing.T) {
	client := &fakeWaliClient{body: []byte(`{"code":0,"data":{"text":"helloThere"}}`)}
	service := NewWaliService(fakeWaliStore{row: map[string]interface{}{
		"status":      "1",
		"config_json": `{"url":"https://wali.example/api","account":"acct","aesKey":"1234567890abcdef","signKey":"sign","agentId":"agent"}`,
	}}, nil, client)
	service.now = func() time.Time { return time.Unix(100, 0) }

	data, err := service.Ping(context.Background())
	if err != nil {
		t.Fatalf("ping: %v", err)
	}
	if data.Data["text"] != "helloThere" || data.Data["code"] != 0 || data.Data["msg"] != "" {
		t.Fatalf("unexpected data %#v", data.Data)
	}
	if !strings.HasPrefix(client.rawURL, "https://wali.example/api/ping?a=acct&t=100&p=") {
		t.Fatalf("unexpected url %q", client.rawURL)
	}
	parsed, err := url.Parse(client.rawURL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	param := parsed.Query().Get("p")
	if param == "" {
		t.Fatalf("expected encrypted param in %q", client.rawURL)
	}
	if _, err := base64.StdEncoding.DecodeString(param); err != nil {
		t.Fatalf("param is not base64: %v", err)
	}
	if got, want := parsed.Query().Get("k"), waliSign(param, 100, "sign"); got != want {
		t.Fatalf("unexpected sign %q want %q", got, want)
	}
}

func TestEncryptWaliParamsMatchesPHP(t *testing.T) {
	param, err := encryptWaliParams(map[string]string{"text": "helloThere"}, ")s)!mv)5_pn)c^l_")
	if err != nil {
		t.Fatalf("encrypt params: %v", err)
	}
	if param != "DvcLGRiX2JwaZCLpGjLwIA==" {
		t.Fatalf("unexpected encrypted param %q", param)
	}
}

func TestWaliServicePingFailure(t *testing.T) {
	client := &fakeWaliClient{body: []byte(`{"code":1,"msg":"bad","data":{}}`)}
	service := NewWaliService(fakeWaliStore{row: map[string]interface{}{
		"status":      "1",
		"config_json": `{"url":"https://wali.example/api","account":"acct","aesKey":"1234567890abcdef","signKey":"sign","agentId":"agent"}`,
	}}, nil, client)

	_, err := service.Ping(context.Background())
	if err == nil {
		t.Fatal("expected ping failure")
	}
}

func TestWaliServiceBalanceRequiresLogin(t *testing.T) {
	service := NewWaliService(fakeWaliStore{}, fakeWaliAuthStore{}, &fakeWaliClient{})

	_, retcode, errmsg, err := service.Balance(context.Background(), "")
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth result %d %q", retcode, errmsg)
	}
}

func TestWaliTransferEdgePrechecks(t *testing.T) {
	service := NewWaliService(fakeWaliStore{settings: map[string]string{"gamecoinlimit": "100"}}, fakeWaliAuthStore{}, &fakeWaliClient{})

	retcode, errmsg, err := service.TopupEdge(context.Background(), "", "0", "")
	if err != nil {
		t.Fatalf("topup unauth: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected unauth response %d %q", retcode, errmsg)
	}

	service = NewWaliService(fakeWaliStore{settings: map[string]string{"gamecoinlimit": "100"}}, fakeWaliAuthStore{user: map[string]interface{}{"uid": "5"}}, &fakeWaliClient{})
	retcode, errmsg, err = service.TopupEdge(context.Background(), "token", "99", "上分成功分支暂未迁移")
	if err != nil {
		t.Fatalf("topup low: %v", err)
	}
	if retcode != -1 || errmsg != "转入金币不能低于100" {
		t.Fatalf("unexpected low response %d %q", retcode, errmsg)
	}

	service = NewWaliService(fakeWaliStore{settings: map[string]string{"gamecoinlimit": "100"}, quota: map[string]interface{}{"goldcoin": "120"}}, fakeWaliAuthStore{user: map[string]interface{}{"uid": "5"}}, &fakeWaliClient{})
	retcode, errmsg, err = service.TopupEdge(context.Background(), "token", "200", "上分成功分支暂未迁移")
	if err != nil {
		t.Fatalf("topup balance: %v", err)
	}
	if retcode != -1 || errmsg != "余额不足:120" {
		t.Fatalf("unexpected balance response %d %q", retcode, errmsg)
	}

	service = NewWaliService(fakeWaliStore{settings: map[string]string{"gamecoinlimit": "100"}, quota: map[string]interface{}{"goldcoin": "220"}}, fakeWaliAuthStore{user: map[string]interface{}{"uid": "5"}}, &fakeWaliClient{})
	retcode, errmsg, err = service.TopupEdge(context.Background(), "token", "200", "上分成功分支暂未迁移")
	if err != nil {
		t.Fatalf("topup pending: %v", err)
	}
	if retcode != -1 || errmsg != "上分成功分支暂未迁移" {
		t.Fatalf("unexpected pending response %d %q", retcode, errmsg)
	}

	retcode, errmsg, err = service.WithdrawEdge(context.Background(), "token", "0", "下分成功分支暂未迁移")
	if err != nil {
		t.Fatalf("withdraw invalid: %v", err)
	}
	if retcode != -1 || errmsg != "金额输入不正确" {
		t.Fatalf("unexpected withdraw response %d %q", retcode, errmsg)
	}
}

func TestWaliServiceBalanceSuccess(t *testing.T) {
	client := &fakeWaliClient{body: []byte(`{"code":0,"data":{"status":0,"balance":"0","transferable":"0"}}`)}
	service := NewWaliService(fakeWaliStore{row: map[string]interface{}{
		"status":      "1",
		"config_json": `{"url":"https://wali.example/api","account":"acct","aesKey":"1234567890abcdef","signKey":"sign","agentId":"agent"}`,
	}}, fakeWaliAuthStore{user: map[string]interface{}{"uid": "5"}}, client)
	service.now = func() time.Time { return time.Unix(100, 0) }

	data, retcode, errmsg, err := service.Balance(context.Background(), "token")
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected result %d %q", retcode, errmsg)
	}
	if data.Data["status"] != float64(0) || data.Data["balance"] != "0" || data.Data["transferable"] != "0" {
		t.Fatalf("unexpected balance data %#v", data.Data)
	}
	if !strings.Contains(client.rawURL, "/getBalance?") || !strings.Contains(client.rawURL, "a=acct") {
		t.Fatalf("unexpected url %q", client.rawURL)
	}
}
