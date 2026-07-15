package index

import (
	"context"
	"testing"
)

type fakeGlobalStore struct {
	calldata map[string]map[string]interface{}
	settings map[string]map[string]interface{}
}

func (s fakeGlobalStore) CalldataByUUID(_ context.Context, uuid string) (map[string]interface{}, error) {
	return s.calldata[uuid], nil
}

func (s fakeGlobalStore) SettingByUUID(_ context.Context, uuid string) (map[string]interface{}, error) {
	return s.settings[uuid], nil
}

func TestGlobalDataBuildsCoreFields(t *testing.T) {
	store := fakeGlobalStore{
		calldata: map[string]map[string]interface{}{
			"global.appver":                  {"type": "json", "content": `{"AndroidVer":"1","iOSVer":"1"}`},
			"search.hotwords":                {"type": "json", "content": `[{"schwd":"a"}]`},
			"global.hottags":                 {"type": "json", "content": `["tag"]`},
			"global.hotcategories":           {"type": "json", "content": `[{"catename":"c"}]`},
			"global.appdownurl":              {"type": "code", "content": `https://down.test`},
			"global.qrcode.link":             {"type": "code", "content": `https://{inviteUrl}?c={inviteCode}`},
			"global.share.text":              {"type": "html", "content": `code {inviteCode}`},
			"global.app.interval_time":       {"type": "code", "content": `120`},
			"global.app.launch.times.adshow": {"type": "code", "content": `3`},
			"global.app.launch.type.adshow":  {"type": "json", "content": `{"type":1}`},
			"promotion.earn.dscr":            {"type": "json", "content": `{"title":"p"}`},
			"global.ads":                     {"type": "rows", "content": `[{"title0":"url","url0":"u","pic":"p.jpg","showweight":"1"}]`},
		},
		settings: map[string]map[string]interface{}{
			"setting":     {"value": `a:4:{s:12:"gameDisabled";i:1;s:6:"csurl";s:8:"https://";s:8:"sitelogo";s:8:"logo.png";s:10:"smscaptcha";i:2;}`},
			"baseset":     {"value": `a:2:{s:10:"inviteUrls";s:9:"a.test\nb";s:7:"newUrls";s:8:"new.test";}`},
			"user.regopt": {"value": `a:1:{s:6:"webreg";i:1;}`},
		},
	}
	service := NewGlobalService(store, "https://res.test")

	data, err := service.GetGlobalData(context.Background(), GlobalRequest{Version: "2"})
	if err != nil {
		t.Fatalf("global data: %v", err)
	}
	appver := data["appver"].(map[string]interface{})
	if appver["AndroidVer"] != "2" || appver["iOSVer"] != "2" {
		t.Fatalf("appver=%#v", appver)
	}
	if data["webreg"] != 1 || data["gameDisabled"] != 1 || data["smscaptcha"] != 2 {
		t.Fatalf("switches=%#v", data)
	}
	if data["sitelogo"] != "https://res.test/logo.png" || data["appintervaltime"] != 120 {
		t.Fatalf("resource/timing=%#v", data)
	}
}
