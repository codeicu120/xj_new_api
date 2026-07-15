package index

import (
	"context"
	"testing"
	"time"
)

type fakeInitStore struct {
	user     map[string]interface{}
	groups   []map[string]interface{}
	quota    map[string]interface{}
	goldbean map[string]interface{}
	settings map[string]map[string]interface{}
	calls    map[string]map[string]interface{}
}

func (s fakeInitStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s fakeInitStore) Groups(context.Context) ([]map[string]interface{}, error) {
	return s.groups, nil
}

func (s fakeInitStore) Quota(context.Context, int) (map[string]interface{}, error) {
	return s.quota, nil
}

func (s fakeInitStore) Goldbean(context.Context, int) (map[string]interface{}, error) {
	return s.goldbean, nil
}

func (s fakeInitStore) SettingByUUID(_ context.Context, uuid string) (map[string]interface{}, error) {
	return s.settings[uuid], nil
}

func (s fakeInitStore) CalldataByUUID(_ context.Context, uuid string) (map[string]interface{}, error) {
	return s.calls[uuid], nil
}

func TestInitGuestShape(t *testing.T) {
	store := fakeInitStore{
		settings: map[string]map[string]interface{}{
			"setting": {"value": `a:4:{s:5:"csurl";s:8:"https://c";s:8:"sitelogo";s:8:"logo.png";s:8:"isclosed";s:1:"0";s:9:"closetips";s:0:"";}`},
			"baseset": {"value": `a:3:{s:13:"inviteCodeUrl";s:8:"https://i";s:15:"inviteCodeAppid";s:3:"app";s:8:"newHosts";s:5:"a.com";}`},
		},
		calls: map[string]map[string]interface{}{
			"global.appver": {"type": "json", "content": `{"AndroidVer":"1.0.0","iOSVer":"1.0.0"}`},
			"playHeaders":   {"type": "json", "content": `{"a":"b"}`},
		},
	}
	global := NewGlobalService(store, "https://res.example")
	service := NewInitService(store, global, "https://res.example")

	data, err := service.Init(context.Background(), InitRequest{Version: "2.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	if data["globalData"] == nil || data["appver"] == nil || data["notification_all"] != nil {
		t.Fatalf("data = %#v", data)
	}
	user := data["user"].(map[string]interface{})
	if user["uid"] != 0 || user["avatar_url"] != "https://res.example/sysavatar/noavatar.png" {
		t.Fatalf("user = %#v", user)
	}
}

func TestInitAuthUserUsesQuotaAndGoldbean(t *testing.T) {
	store := fakeInitStore{
		user: map[string]interface{}{
			"uid": "5", "uniqkey": "1904908418", "username": "~1904908418", "nickname": "1400002",
			"mobi": "86.14012340002", "email": "~1904908418", "sysgid": "6", "gid": "4",
			"sysgid_exptime": "0", "regtime": "1547688484", "gender": "1", "avatar": "sysavatar/man/5.png",
			"newmsg": "0", "recommend_total": "6",
		},
		groups: []map[string]interface{}{{"gid": "6", "gicon": "V6"}, {"gid": "4", "gicon": "V4"}},
		quota:  map[string]interface{}{"goldcoin": "625"},
		goldbean: map[string]interface{}{
			"gold_bean": "3815",
		},
		settings: map[string]map[string]interface{}{
			"setting": {"value": `a:1:{s:5:"csurl";s:8:"https://c";}`},
			"baseset": {"value": `a:0:{}`},
		},
		calls: map[string]map[string]interface{}{
			"global.appver": {"type": "json", "content": `{}`},
		},
	}
	global := NewGlobalService(store, "https://res.example")
	service := NewInitService(store, global, "https://res.example")
	service.now = func() time.Time { return time.Unix(2000, 0) }

	data, err := service.Init(context.Background(), InitRequest{Token: "3235306637393062613731656332623964333835356634323464623232353965"})
	if err != nil {
		t.Fatal(err)
	}
	user := data["user"].(map[string]interface{})
	if user["uid"] != "5" || user["uniqkey"] != "VI4SQQ" || user["goldcoin"] != 625 || user["gold_bean"] != 3815 {
		t.Fatalf("user = %#v", user)
	}
}
