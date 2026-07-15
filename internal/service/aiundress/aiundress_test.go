package aiundress

import (
	"context"
	"testing"
	"time"
)

type fakeAuth struct {
	user map[string]interface{}
}

func (f fakeAuth) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return f.user, nil
}

type fakeStore struct {
	setting string
	rows    []map[string]interface{}
}

func (f fakeStore) Count(context.Context, int, int) (int, error) {
	return len(f.rows), nil
}

func (f fakeStore) List(context.Context, int, int, int, int, int) ([]map[string]interface{}, error) {
	return f.rows, nil
}

func (f fakeStore) SettingByUUID(context.Context, string) (string, error) {
	return f.setting, nil
}

func TestListingRequiresLoginWithPHPErrorCode(t *testing.T) {
	service := NewService(fakeAuth{}, fakeStore{}, "https://res.example")

	_, retcode, errmsg, err := service.Listing(context.Background(), "", 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "请先登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestListingBuildsResourceURLsAndPageInfo(t *testing.T) {
	service := NewService(
		fakeAuth{user: map[string]interface{}{"uid": "7"}},
		fakeStore{
			setting: `a:1:{s:12:"resurl_h5_ai";s:36:"https://ai-{rand}.example.com/assets";}`,
			rows: []map[string]interface{}{
				{"id": "1", "uid": "7", "image": "ai_undress/a.jpg", "output": "out/b.jpg"},
			},
		},
		"https://res.example",
	)
	service.now = func() time.Time { return time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC) }

	data, retcode, errmsg, err := service.Listing(context.Background(), "250f790ba71ec2b9d3855f424db2259e", 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if data.Rows[0]["image"] != "https://ai-2026071517.example.com/assets/ai_undress/a.jpg" {
		t.Fatalf("image url = %#v", data.Rows[0]["image"])
	}
	if data.Rows[0]["output"] != "https://ai-2026071517.example.com/assets/out/b.jpg" {
		t.Fatalf("output url = %#v", data.Rows[0]["output"])
	}
	if data.PageInfo["page_url"] != "/aiundress/listing?page=[?]" || data.PageInfo["pagesize"] != 10 {
		t.Fatalf("pageinfo = %#v", data.PageInfo)
	}
}
