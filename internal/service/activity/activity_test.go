package activity

import (
	"context"
	"testing"
	"time"
)

type fakeStore struct {
	activities []map[string]interface{}
	activity   map[string]interface{}
	prizes     []map[string]interface{}
	logs       []map[string]interface{}
	user       map[string]interface{}
	records    []map[string]interface{}
	ranking    map[string]interface{}
	users      []map[string]interface{}
	groups     []map[string]interface{}
}

func (s fakeStore) CurrentActivities(context.Context, int64, int) ([]map[string]interface{}, error) {
	return s.activities, nil
}

func (s fakeStore) ActivityByID(context.Context, int) (map[string]interface{}, error) {
	return s.activity, nil
}

func (s fakeStore) PrizesByActivityID(context.Context, int) ([]map[string]interface{}, error) {
	return s.prizes, nil
}

func (s fakeStore) PrizeLogs(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return s.logs, nil
}

func (s fakeStore) CountActivityRecords(context.Context, int) (int, error) {
	return len(s.records), nil
}

func (s fakeStore) ActivityRecords(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return s.records, nil
}

func (s fakeStore) ActivityRanking(context.Context, int, int) (map[string]interface{}, error) {
	return s.ranking, nil
}

func (s fakeStore) BotByID(context.Context, int) (map[string]interface{}, error) {
	return map[string]interface{}{"username": "bot", "avatar": "bot.jpg"}, nil
}

func (s fakeStore) CountRecommendedUsers(context.Context, int, int64, int64) (int, error) {
	return len(s.users), nil
}

func (s fakeStore) RecommendedUsers(context.Context, int, int64, int64, int, int) ([]map[string]interface{}, error) {
	return s.users, nil
}

func (s fakeStore) UserGroups(context.Context) ([]map[string]interface{}, error) {
	return s.groups, nil
}

func (s fakeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func TestLuckyPrizes(t *testing.T) {
	service := NewService(fakeStore{}, fakeStore{}, "https://res.example.test")
	data := service.LuckyPrizes()
	rows, ok := data["data"].([]map[string]interface{})
	if !ok {
		t.Fatalf("data rows type = %T", data["data"])
	}
	if len(rows) != 5 {
		t.Fatalf("len rows = %d", len(rows))
	}
	if rows[0]["keyid"] != "prize.vip.365" || rows[4]["prizename"] != "30天VIP" {
		t.Fatalf("unexpected prizes %#v", rows)
	}
}

func TestExpiredActivityResponses(t *testing.T) {
	service := NewService(fakeStore{}, fakeStore{}, "https://res.example.test")
	service.now = func() time.Time { return time.Date(2026, 7, 15, 0, 0, 0, 0, time.Local) }

	for name, fn := range map[string]func() (int, string){
		"newyear2020": service.NewYear2020,
		"luckydraw":   service.LuckyDraw,
	} {
		retcode, errmsg := fn()
		if retcode != -1 || errmsg != "抽奖活动已结束，谢谢支持" {
			t.Fatalf("%s response = %d %q", name, retcode, errmsg)
		}
	}
}

func TestIndexNoCurrentActivity(t *testing.T) {
	service := NewService(fakeStore{}, fakeStore{}, "https://res.example.test")

	data, retcode, errmsg, err := service.Index(context.Background(), 1)
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	if data != nil || retcode != -9999 || errmsg != "当前没有进行中的活动" {
		t.Fatalf("index response = %#v %d %q", data, retcode, errmsg)
	}
}

func TestDetailsProcessesPrizes(t *testing.T) {
	store := fakeStore{
		activity: map[string]interface{}{"id": "9", "title": "活动"},
		prizes: []map[string]interface{}{
			{"id": "1", "aid": "9", "level": "一", "ranking": "1", "prize": "100"},
			{"id": "2", "aid": "9", "level": "二", "ranking": "3", "prize": "50"},
		},
	}
	service := NewService(store, store, "https://res.example.test")

	data, retcode, errmsg, err := service.Details(context.Background(), 9)
	if err != nil {
		t.Fatalf("details: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("details response = %d %q", retcode, errmsg)
	}
	payload := data["data"].(map[string]interface{})
	prizes := payload["activity_prize"].([]map[string]interface{})
	if prizes[0]["ranking"] != "1" || prizes[0]["prize_users"] != "1" {
		t.Fatalf("first prize = %#v", prizes[0])
	}
	if prizes[1]["ranking"] != "2-3" || prizes[1]["prize_users"] != "2" {
		t.Fatalf("second prize = %#v", prizes[1])
	}
}

func TestLuckyDrawHistoryRequiresLogin(t *testing.T) {
	service := NewService(fakeStore{}, fakeStore{}, "https://res.example.test")

	data, retcode, errmsg, err := service.LuckyDrawHistory(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if data != nil || retcode != -9999 || errmsg != "请登录后操作" {
		t.Fatalf("history response = %#v %d %q", data, retcode, errmsg)
	}
}

func TestLuckyDrawHistoryAddsPrizeName(t *testing.T) {
	store := fakeStore{
		user: map[string]interface{}{"uid": "5"},
		logs: []map[string]interface{}{
			{"id": "1", "uid": "5", "keyid": "prize.vip.90", "createtime": "100"},
		},
	}
	service := NewService(store, store, "https://res.example.test")

	data, retcode, errmsg, err := service.LuckyDrawHistory(context.Background(), "token", 1)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("history response = %d %q", retcode, errmsg)
	}
	rows := data["data"].([]map[string]interface{})
	if rows[0]["prizename"] != "90天VIP" {
		t.Fatalf("history row = %#v", rows[0])
	}
}

func TestRankingRequiresLogin(t *testing.T) {
	service := NewService(fakeStore{}, fakeStore{}, "https://res.example.test")

	data, retcode, errmsg, err := service.Ranking(context.Background(), "", 1, 1)
	if err != nil {
		t.Fatalf("ranking: %v", err)
	}
	if data != nil || retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("ranking response = %#v %d %q", data, retcode, errmsg)
	}
}

func TestRankingProcessesRecords(t *testing.T) {
	store := fakeStore{
		user:     map[string]interface{}{"uid": "5"},
		activity: map[string]interface{}{"id": "9", "prize_users": "10"},
		prizes: []map[string]interface{}{
			{"ranking": "1", "level": "一", "prize": "100"},
			{"ranking": "3", "level": "二", "prize": "50"},
		},
		records: []map[string]interface{}{
			{"id": "1", "aid": "9", "uid": "5", "username": "u", "avatar": "a.jpg", "score": "8", "received": "0", "create_time": "1", "update_time": "2"},
			{"id": "2", "aid": "9", "uid": "-1", "username": nil, "avatar": nil, "score": "7", "received": "0", "create_time": "1", "update_time": "2"},
		},
	}
	service := NewService(store, store, "https://res.example.test")

	data, retcode, errmsg, err := service.Ranking(context.Background(), "token", 9, 1)
	if err != nil {
		t.Fatalf("ranking: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("ranking response = %d %q", retcode, errmsg)
	}
	rows := data["data"].([]map[string]interface{})
	if rows[0]["prize_level"] != "一" || rows[1]["prize_level"] != "二" || rows[1]["username"] != "bot" {
		t.Fatalf("ranking rows = %#v", rows)
	}
}

func TestReceiveExpired(t *testing.T) {
	store := fakeStore{
		user:     map[string]interface{}{"uid": "5"},
		activity: map[string]interface{}{"id": "9", "prize_users": "10", "reward_expire_time": "1"},
		prizes:   []map[string]interface{}{{"ranking": "1", "level": "一", "prize": "100"}},
		ranking:  map[string]interface{}{"uid": "5", "ranking": "1"},
	}
	service := NewService(store, store, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(2, 0) }

	data, retcode, errmsg, err := service.Receive(context.Background(), "token", 9)
	if err != nil {
		t.Fatalf("receive: %v", err)
	}
	if data != nil || retcode != -9999 || errmsg != "超过该活动领奖截止日期" {
		t.Fatalf("receive response = %#v %d %q", data, retcode, errmsg)
	}
}

func TestRecommendsProcessesUsers(t *testing.T) {
	store := fakeStore{
		user:     map[string]interface{}{"uid": "5"},
		activity: map[string]interface{}{"id": "9", "effect_time": "1", "expire_time": "999"},
		groups:   []map[string]interface{}{{"gid": "6", "gicon": "V6"}},
		users: []map[string]interface{}{{
			"uid":             "10",
			"uniqkey":         "12345",
			"username":        "child",
			"nickname":        "",
			"mobi":            "",
			"email":           "",
			"sysgid":          "6",
			"sysgid_exptime":  "200",
			"gid":             "1",
			"regtime":         "100",
			"gender":          "1",
			"avatar":          "avatar.jpg",
			"newmsg":          "0",
			"goldcoin":        "7",
			"gold_bean":       "8",
			"recommend_total": "2",
		}},
	}
	service := NewService(store, store, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(150, 0) }

	data, retcode, errmsg, err := service.Recommends(context.Background(), "token", 9, 1)
	if err != nil {
		t.Fatalf("recommends: %v", err)
	}
	if retcode != 0 || errmsg != "" || data["total"] != 1 {
		t.Fatalf("recommends response = %#v %d %q", data, retcode, errmsg)
	}
	rows := data["data"].([]map[string]interface{})
	row := rows[0]
	if row["uniqkey"] != "9IX" || row["gicon"] != "V6" || row["isvip"] != 1 {
		t.Fatalf("unexpected user row %#v", row)
	}
	if row["avatar_url"] != "https://res.example.test/C1/avatar/avatar.jpg" {
		t.Fatalf("unexpected avatar_url %v", row["avatar_url"])
	}
}
