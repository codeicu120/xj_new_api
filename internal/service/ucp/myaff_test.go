package ucp

import (
	"context"
	"testing"
	"time"
)

type fakeUserStore struct {
	user map[string]interface{}
}

func (s fakeUserStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s fakeUserStore) Groups(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"gid": "6", "gicon": "V6"}}, nil
}

func (s fakeUserStore) CountRecommended(context.Context, int) (int, error) {
	return 1, nil
}

func (s fakeUserStore) RecommendedUsers(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"uid":             "10",
			"uniqkey":         "12345",
			"username":        "u",
			"nickname":        "",
			"mobi":            "86.1",
			"email":           "~u",
			"sysgid":          "6",
			"sysgid_exptime":  "1000",
			"gid":             "1",
			"regtime":         "100",
			"gender":          "1",
			"avatar":          "",
			"newmsg":          "0",
			"recommend_total": "2",
		},
	}, nil
}

func TestMyAffRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")
	_, retcode, errmsg, err := service.MyAff(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("my aff: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestMyAffFormatsRows(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(2000, 0) }

	data, retcode, errmsg, err := service.MyAff(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatalf("my aff: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if len(data.Rows) != 1 {
		t.Fatalf("expected one row, got %d", len(data.Rows))
	}
	row := data.Rows[0]
	if row["uniqkey"] != "9IX" {
		t.Fatalf("unexpected uniqkey %v", row["uniqkey"])
	}
	if row["gicon"] != "V6" {
		t.Fatalf("unexpected gicon %v", row["gicon"])
	}
	if row["avatar_url"] != "https://res.example.test/sysavatar/noavatar.png" {
		t.Fatalf("unexpected avatar url %v", row["avatar_url"])
	}
	if data.PageInfo["total"] != 1 || data.PageInfo["page"] != 1 {
		t.Fatalf("unexpected pageinfo %#v", data.PageInfo)
	}
}
