package game

import (
	"context"
	"errors"
	"testing"
)

type fakeLotteryStore struct {
	row     map[string]interface{}
	history map[string]interface{}
	deleted []int
	saved   []struct {
		uid        int
		platformID int
		gameID     int
	}
}

func (s *fakeLotteryStore) PlatformBySlug(_ context.Context, slug string) (map[string]interface{}, error) {
	if slug != "lottery" {
		return map[string]interface{}{}, nil
	}
	return s.row, nil
}

func (s *fakeLotteryStore) GameHistoryByUniqueKey(_ context.Context, uid int, platformID int, gameID int) (map[string]interface{}, error) {
	return s.history, nil
}

func (s *fakeLotteryStore) DeleteGameHistory(_ context.Context, id int) error {
	s.deleted = append(s.deleted, id)
	return nil
}

func (s *fakeLotteryStore) SaveGameHistory(_ context.Context, uid int, platformID int, gameID int) (int, error) {
	s.saved = append(s.saved, struct {
		uid        int
		platformID int
		gameID     int
	}{uid: uid, platformID: platformID, gameID: gameID})
	return 99, nil
}

type fakeLotteryClient struct {
	gameURL string
	balance LotteryBalance
	err     error
	cfg     LotteryConfig
	req     LotteryEnterRequest
	uid     int
}

func (c *fakeLotteryClient) EnterGame(_ context.Context, cfg LotteryConfig, req LotteryEnterRequest) (string, error) {
	c.cfg = cfg
	c.req = req
	if c.err != nil {
		return "", c.err
	}
	return c.gameURL, nil
}

func (c *fakeLotteryClient) Balance(_ context.Context, cfg LotteryConfig, uid int) (LotteryBalance, error) {
	c.cfg = cfg
	c.uid = uid
	if c.err != nil {
		return LotteryBalance{}, c.err
	}
	return c.balance, nil
}

func TestLotteryServiceEnterRequiresLogin(t *testing.T) {
	service := NewLotteryService(&fakeLotteryStore{}, fakeWaliAuthStore{}, &fakeLotteryClient{})

	_, retcode, errmsg, err := service.Enter(context.Background(), "", "")
	if err != nil {
		t.Fatalf("enter: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth result %d %q", retcode, errmsg)
	}
}

func TestLotteryServiceEnterSuccessSavesHistoryAndReturnsURL(t *testing.T) {
	client := &fakeLotteryClient{gameURL: "https://lottery.example/play"}
	store := &fakeLotteryStore{
		row: map[string]interface{}{
			"id":          "8",
			"status":      "1",
			"config_json": `{"apiUrl":"https://lottery.example/api","agent":"agent","encryptKey":"1234567890abcdef","signKey":"sign","platform":"xjlottery"}`,
		},
		history: map[string]interface{}{"id": "13"},
	}
	service := NewLotteryService(store, fakeWaliAuthStore{user: map[string]interface{}{"uid": "5"}}, client)

	gameURL, retcode, errmsg, err := service.Enter(context.Background(), "token", "12")
	if err != nil {
		t.Fatalf("enter: %v", err)
	}
	if retcode != 0 || errmsg != "" || gameURL != "https://lottery.example/play" {
		t.Fatalf("unexpected enter result url=%q retcode=%d errmsg=%q", gameURL, retcode, errmsg)
	}
	if client.cfg.PlatformID != 8 || client.cfg.Platform != "xjlottery" || client.cfg.Lang != "zh-CN" {
		t.Fatalf("unexpected client config %#v", client.cfg)
	}
	if client.req.UID != 5 || client.req.LotID != 12 {
		t.Fatalf("unexpected client request %#v", client.req)
	}
	if len(store.deleted) != 1 || store.deleted[0] != 13 {
		t.Fatalf("unexpected deleted history %#v", store.deleted)
	}
	if len(store.saved) != 1 || store.saved[0].uid != 5 || store.saved[0].platformID != 8 || store.saved[0].gameID != 12 {
		t.Fatalf("unexpected saved history %#v", store.saved)
	}
}

func TestLotteryServiceEnterExternalFailureIsSafe(t *testing.T) {
	client := &fakeLotteryClient{err: errors.New("network down")}
	store := &fakeLotteryStore{row: map[string]interface{}{
		"id":          "8",
		"status":      "1",
		"config_json": `{"apiUrl":"https://lottery.example/api","agent":"agent","encryptKey":"1234567890abcdef","signKey":"sign","platform":"xjlottery"}`,
	}}
	service := NewLotteryService(store, fakeWaliAuthStore{user: map[string]interface{}{"uid": "5"}}, client)

	_, retcode, errmsg, err := service.Enter(context.Background(), "token", "12")
	if err != nil {
		t.Fatalf("enter failure should be safe: %v", err)
	}
	if retcode != -1 || errmsg != "进入游戏失败" {
		t.Fatalf("unexpected failure response %d %q", retcode, errmsg)
	}
	if len(store.saved) != 0 {
		t.Fatalf("history should not be saved on external failure: %#v", store.saved)
	}
}

func TestLotteryServiceEnterMissingConfigIsSafe(t *testing.T) {
	service := NewLotteryService(&fakeLotteryStore{}, fakeWaliAuthStore{user: map[string]interface{}{"uid": "5"}}, &fakeLotteryClient{})

	_, retcode, errmsg, err := service.Enter(context.Background(), "token", "12")
	if err != nil {
		t.Fatalf("enter missing config should be safe: %v", err)
	}
	if retcode != -1 || errmsg != "进入游戏失败" {
		t.Fatalf("unexpected missing config response %d %q", retcode, errmsg)
	}
}

func TestLotteryServiceBalanceRequiresLogin(t *testing.T) {
	service := NewLotteryService(&fakeLotteryStore{}, fakeWaliAuthStore{}, &fakeLotteryClient{})

	_, retcode, errmsg, err := service.Balance(context.Background(), "")
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth result %d %q", retcode, errmsg)
	}
}

func TestLotteryServiceBalanceSuccess(t *testing.T) {
	client := &fakeLotteryClient{balance: LotteryBalance{
		Status:       float64(10),
		Balance:      "12.34",
		Transferable: "5.60",
		Currency:     "CNY",
	}}
	store := &fakeLotteryStore{row: map[string]interface{}{
		"id":          "8",
		"status":      "1",
		"config_json": `{"apiUrl":"https://lottery.example/api","agent":"agent","encryptKey":"1234567890abcdef","signKey":"sign","platform":"xjlottery"}`,
	}}
	service := NewLotteryService(store, fakeWaliAuthStore{user: map[string]interface{}{"uid": "5"}}, client)

	data, retcode, errmsg, err := service.Balance(context.Background(), "token")
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected balance result retcode=%d errmsg=%q", retcode, errmsg)
	}
	if client.uid != 5 || client.cfg.Platform != "xjlottery" {
		t.Fatalf("unexpected client call uid=%d cfg=%#v", client.uid, client.cfg)
	}
	if data["balance"] != "12.34" || data["transferable"] != "5.60" || data["currency"] != "CNY" || data["status"] != float64(10) {
		t.Fatalf("unexpected balance data %#v", data)
	}
}

func TestLotteryServiceBalanceExternalFailureIsSafe(t *testing.T) {
	client := &fakeLotteryClient{err: errors.New("network down")}
	store := &fakeLotteryStore{row: map[string]interface{}{
		"id":          "8",
		"status":      "1",
		"config_json": `{"apiUrl":"https://lottery.example/api","agent":"agent","encryptKey":"1234567890abcdef","signKey":"sign","platform":"xjlottery"}`,
	}}
	service := NewLotteryService(store, fakeWaliAuthStore{user: map[string]interface{}{"uid": "5"}}, client)

	_, retcode, errmsg, err := service.Balance(context.Background(), "token")
	if err != nil {
		t.Fatalf("balance failure should be safe: %v", err)
	}
	if retcode != -1 || errmsg != "查询余额失败" {
		t.Fatalf("unexpected failure response %d %q", retcode, errmsg)
	}
}

func TestLotteryBalanceFromResultFormatsCents(t *testing.T) {
	balance := lotteryBalanceFromResult(map[string]interface{}{
		"status":     "10",
		"totalMoney": "1234",
		"freeMoney":  float64(560),
		"currency":   "CNY",
	})
	if balance.Balance != "12.34" || balance.Transferable != "5.60" || balance.Currency != "CNY" || balance.Status != "10" {
		t.Fatalf("unexpected balance %#v", balance)
	}
}
