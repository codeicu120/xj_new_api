package starlive

import (
	"context"
	"testing"

	"xj_comp/internal/domain"
)

type fakeStore struct {
	info  domain.StarLiveInfo
	user  map[string]interface{}
	guest map[string]interface{}
	quota map[string]interface{}
}

func (s fakeStore) Info(context.Context) (domain.StarLiveInfo, error) {
	return s.info, nil
}

func (s fakeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s fakeStore) GuestBySID(context.Context, string) (map[string]interface{}, error) {
	return s.guest, nil
}

func (s fakeStore) Quota(context.Context, int) (map[string]interface{}, error) {
	return s.quota, nil
}

func TestIndexRequiresGuestToken(t *testing.T) {
	service := NewService(fakeStore{}, nil, fakeStore{}, nil)

	_, retcode, errmsg, err := service.Index(context.Background(), "")
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	if retcode != -9999 || errmsg != "客户端游客请先携带信息" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestIndexReturnsNotOpenWhenInfoMissing(t *testing.T) {
	store := fakeStore{guest: map[string]interface{}{"sid": "250f790ba71ec2b9d3855f424db2259e"}}
	service := NewService(store, nil, store, nil)

	_, retcode, errmsg, err := service.Index(context.Background(), "250f790ba71ec2b9d3855f424db2259e")
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	if retcode != -1 || errmsg != "暂未开放" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestIndexReturnsLiveConfigForUser(t *testing.T) {
	store := fakeStore{
		info: domain.StarLiveInfo{
			AppID:    "app",
			SecKey:   "1234567890abcdef",
			APIHost:  "https://api.example",
			Env:      "2",
			Src:      "https://src.example",
			LiveHost: "https://live.example",
		},
		user: map[string]interface{}{"uid": "5"},
	}
	service := NewService(store, store, nil, nil)

	data, retcode, errmsg, err := service.Index(context.Background(), "250f790ba71ec2b9d3855f424db2259e")
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if data.Data["encryptUid"] != "x8oY26MqLNZyOfLCkspXeg==" || data.Data["token"] != "22126bb2136f0718741a54d319a7b999" {
		t.Fatalf("unexpected auth fields %#v", data.Data)
	}
	if data.Data["appId"] != "app" || data.Data["apiHost"] != "https://api.example" || data.Data["liveHost"] != "https://live.example" {
		t.Fatalf("unexpected data %#v", data.Data)
	}
}

func TestQueryCoinBalanceBranches(t *testing.T) {
	service := NewService(nil, nil, nil, fakeStore{quota: map[string]interface{}{"goldcoin": "7"}})

	longID, err := service.QueryCoinBalance(context.Background(), "tourist-member-123456")
	if err != nil {
		t.Fatalf("long member id: %v", err)
	}
	if longID.Code != 0 || longID.Data["balance"] != 0 {
		t.Fatalf("long member id response %#v", longID)
	}

	found, err := service.QueryCoinBalance(context.Background(), "5")
	if err != nil {
		t.Fatalf("found member: %v", err)
	}
	if found.Code != 0 || found.Data["balance"] != 70 {
		t.Fatalf("found response %#v", found)
	}

	missingService := NewService(nil, nil, nil, fakeStore{})
	missing, err := missingService.QueryCoinBalance(context.Background(), "6")
	if err != nil {
		t.Fatalf("missing member: %v", err)
	}
	if missing.Code != -1 || missing.Data["msg"] != "未知用户" {
		t.Fatalf("missing response %#v", missing)
	}
}

func TestAssetEdgesReturnSafeFailures(t *testing.T) {
	service := NewService(nil, nil, nil, nil)

	guest := service.GameBetEdge(map[string]interface{}{"memberId": "tourist-member-123456"})
	if guest.Code != -1 || guest.Data["msg"] != "游客用户请先登录" {
		t.Fatalf("guest response %#v", guest)
	}

	unknown := service.GameWinEdge(map[string]interface{}{"memberId": ""})
	if unknown.Code != -1 || unknown.Data["msg"] != "未知用户" {
		t.Fatalf("unknown response %#v", unknown)
	}

	translated := service.TranslateEdge(map[string]interface{}{"memberId": "tourist-member-123456"})
	if translated.Code != -1 || translated.Data["msg"] != "游客用户请先登录" {
		t.Fatalf("translate response %#v", translated)
	}

	retry := service.TryAgainEdge(map[string]interface{}{"busiType": 9})
	if retry.Code != -1 || retry.Data["msg"] != "未知业务类型" {
		t.Fatalf("tryAgain response %#v", retry)
	}
}
