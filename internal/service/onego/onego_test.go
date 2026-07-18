package onego

import (
	"context"
	"errors"
	"testing"
	"time"

	"xj_comp/internal/domain"
)

type fakeStore struct {
	rules          map[string]interface{}
	rooms          []map[string]interface{}
	room           map[string]interface{}
	currentRecords []map[string]interface{}
	latestRecord   map[string]interface{}
	periodRecords  []map[string]interface{}
	roomRecords    []map[string]interface{}
	rankWinCoins   []map[string]interface{}
	userWins       []map[string]interface{}
	userOrders     []map[string]interface{}
	periodOrders   []map[string]interface{}
	record         map[string]interface{}
	rankBetCoins   []map[string]interface{}
	user           map[string]interface{}
	bot            map[string]interface{}
	quota          map[string]interface{}
	betResult      domain.OneGoBetResult
	betRet         int
	betMsg         string
	err            error
}

type fakeAuth struct {
	user map[string]interface{}
}

func (a fakeAuth) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return a.user, nil
}

func (s fakeStore) Rules(context.Context) (map[string]interface{}, error) {
	return s.rules, s.err
}

func (s fakeStore) Rooms(context.Context) ([]map[string]interface{}, error) {
	return s.rooms, s.err
}

func (s fakeStore) RoomByID(context.Context, int) (map[string]interface{}, error) {
	return s.room, s.err
}

func (s fakeStore) CurrentRecords(context.Context, int, int64) ([]map[string]interface{}, error) {
	return s.currentRecords, s.err
}

func (s fakeStore) LatestRecord(context.Context) (map[string]interface{}, error) {
	return s.latestRecord, s.err
}

func (s fakeStore) RecordsByRoom(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return s.roomRecords, s.err
}

func (s fakeStore) RecordsByPeriod(context.Context, string, int, int) ([]map[string]interface{}, error) {
	return s.periodRecords, s.err
}

func (s fakeStore) RankWinCoins(context.Context) ([]map[string]interface{}, error) {
	return s.rankWinCoins, s.err
}

func (s fakeStore) UserWins(context.Context, int) ([]map[string]interface{}, error) {
	return s.userWins, s.err
}

func (s fakeStore) UserOrdersGrouped(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return s.userOrders, s.err
}

func (s fakeStore) UserOrdersByPeriod(context.Context, string, int, int) ([]map[string]interface{}, error) {
	return s.periodOrders, s.err
}

func (s fakeStore) RecordByPeriod(context.Context, string, int) (map[string]interface{}, error) {
	return s.record, s.err
}

func (s fakeStore) RankBetCoins(context.Context, string, int, int, int) ([]map[string]interface{}, error) {
	return s.rankBetCoins, s.err
}

func (s fakeStore) UserByID(context.Context, int) (map[string]interface{}, error) {
	return s.user, s.err
}

func (s fakeStore) BotByID(context.Context, int) (map[string]interface{}, error) {
	return s.bot, s.err
}

func (s fakeStore) Quota(context.Context, int) (map[string]interface{}, error) {
	return s.quota, s.err
}

func (s fakeStore) Bet(context.Context, domain.OneGoBetInput) (domain.OneGoBetResult, int, string, error) {
	if s.betRet != 0 || s.betMsg != "" || len(s.betResult.BetNo) > 0 || len(s.betResult.TotalBetNo) > 0 {
		return s.betResult, s.betRet, s.betMsg, s.err
	}
	return domain.OneGoBetResult{BetNo: []int{0}, TotalBetNo: []int{0}}, 0, "", s.err
}

func TestRulesReturnsData(t *testing.T) {
	service := NewService(fakeStore{rules: map[string]interface{}{"id": "1", "title": "一元购规则"}})

	data, err := service.Rules(context.Background())
	if err != nil {
		t.Fatalf("rules: %v", err)
	}
	row, ok := data.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected rules map, got %T", data.Data)
	}
	if row["title"] != "一元购规则" {
		t.Fatalf("unexpected rules %#v", row)
	}
}

func TestRulesNotOpen(t *testing.T) {
	service := NewService(fakeStore{})

	_, err := service.Rules(context.Background())
	if !errors.Is(err, ErrNotOpen) {
		t.Fatalf("expected ErrNotOpen, got %v", err)
	}
}

func TestRoomsReturnsData(t *testing.T) {
	service := NewService(fakeStore{rooms: []map[string]interface{}{{"id": "1", "name": "10金币场"}}})

	data, err := service.Rooms(context.Background())
	if err != nil {
		t.Fatalf("rooms: %v", err)
	}
	rows, ok := data.Data.([]map[string]interface{})
	if !ok {
		t.Fatalf("expected room rows, got %T", data.Data)
	}
	if len(rows) != 1 || rows[0]["name"] != "10金币场" {
		t.Fatalf("unexpected rooms %#v", rows)
	}
}

func TestRoomsNotOpen(t *testing.T) {
	service := NewService(fakeStore{})

	_, err := service.Rooms(context.Background())
	if !errors.Is(err, ErrNotOpen) {
		t.Fatalf("expected ErrNotOpen, got %v", err)
	}
}

func TestCurrentReturnsRulesAndRecord(t *testing.T) {
	service := NewService(fakeStore{
		rules:          map[string]interface{}{"id": "1"},
		room:           map[string]interface{}{"id": "2"},
		currentRecords: []map[string]interface{}{sampleRecord("5")},
		user:           map[string]interface{}{"uid": "5", "username": "winner", "avatar": ""},
	})

	data, err := service.Current(context.Background(), 2)
	if err != nil {
		t.Fatalf("current: %v", err)
	}
	payload, ok := data.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected current payload, got %T", data.Data)
	}
	current := payload["current"].(map[string]interface{})
	if current["id"] != 10 || current["winner"].(map[string]interface{})["username"] != "winner" {
		t.Fatalf("unexpected current %#v", current)
	}
}

func TestCurrentRequiresRulesAndRoom(t *testing.T) {
	service := NewService(fakeStore{})
	if _, err := service.Current(context.Background(), 1); !errors.Is(err, ErrNotOpen) {
		t.Fatalf("expected ErrNotOpen, got %v", err)
	}

	service = NewService(fakeStore{rules: map[string]interface{}{"id": "1"}})
	if _, err := service.Current(context.Background(), 1); !errors.Is(err, ErrSelectRoom) {
		t.Fatalf("expected ErrSelectRoom, got %v", err)
	}
}

func TestLastReturnsLatestPeriodRecords(t *testing.T) {
	service := NewService(fakeStore{
		latestRecord:  map[string]interface{}{"period": "2026071401"},
		periodRecords: []map[string]interface{}{sampleRecord("0")},
	})

	data, err := service.Last(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("last: %v", err)
	}
	rows, ok := data.Data.([]map[string]interface{})
	if !ok {
		t.Fatalf("expected rows, got %T", data.Data)
	}
	if len(rows) != 1 || rows[0]["winner"] != 0 || rows[0]["room_id"] != 1 {
		t.Fatalf("unexpected rows %#v", rows)
	}
}

func TestLastNoData(t *testing.T) {
	service := NewService(fakeStore{})

	_, err := service.Last(context.Background(), 0, 1)
	if !errors.Is(err, ErrNoData) {
		t.Fatalf("expected ErrNoData, got %v", err)
	}
}

func TestHashReturnsSHA256AndNumber(t *testing.T) {
	service := NewService(fakeStore{})

	data, err := service.Hash(" abc ")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	payload, ok := data.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected hash payload, got %T", data.Data)
	}
	if payload["hash_code"] != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("unexpected hash_code %v", payload["hash_code"])
	}
	if payload["hash_number"] != "120015" {
		t.Fatalf("unexpected hash_number %v", payload["hash_number"])
	}
}

func TestHashRequiresPlaintext(t *testing.T) {
	service := NewService(fakeStore{})

	_, err := service.Hash(" \t ")
	if !errors.Is(err, ErrMissingPlaintext) {
		t.Fatalf("expected ErrMissingPlaintext, got %v", err)
	}
}

func TestLuckyReturnsRanksWithWinsAndWinner(t *testing.T) {
	service := NewService(fakeStore{
		rankWinCoins: []map[string]interface{}{
			{"total_awards": "300", "winner": "5"},
		},
		userWins: []map[string]interface{}{
			{"wins": "2", "room_id": "1"},
		},
		user: map[string]interface{}{"uid": "5", "username": "winner", "avatar": ""},
	})

	data, err := service.Lucky(context.Background())
	if err != nil {
		t.Fatalf("lucky: %v", err)
	}
	rows, ok := data.Data.([]map[string]interface{})
	if !ok {
		t.Fatalf("expected ranks, got %T", data.Data)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one rank, got %d", len(rows))
	}
	row := rows[0]
	if row["total_awards"] != 300 {
		t.Fatalf("unexpected total_awards %#v", row["total_awards"])
	}
	winner, ok := row["winner"].(map[string]interface{})
	if !ok || winner["username"] != "winner" {
		t.Fatalf("unexpected winner %#v", row["winner"])
	}
	wins := row["wins"].([]map[string]interface{})
	if wins[0]["wins"] != 2 || wins[0]["room_id"] != 1 {
		t.Fatalf("unexpected wins %#v", wins)
	}
	if row["id"] != 0 || row["awards"] != 0 || row["open_time"] != 0 {
		t.Fatalf("expected PHP procRow zero-fill fields, got %#v", row)
	}
}

func TestLuckyEmptyRanks(t *testing.T) {
	service := NewService(fakeStore{})

	data, err := service.Lucky(context.Background())
	if err != nil {
		t.Fatalf("lucky: %v", err)
	}
	rows, ok := data.Data.([]map[string]interface{})
	if !ok {
		t.Fatalf("expected ranks, got %T", data.Data)
	}
	if len(rows) != 0 {
		t.Fatalf("expected empty ranks, got %#v", rows)
	}
}

func TestHistoryRequiresLogin(t *testing.T) {
	service := NewService(fakeStore{})

	_, retcode, errmsg, err := service.History(context.Background(), "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestBetEdgePrechecks(t *testing.T) {
	service := NewService(fakeStore{})

	_, retcode, errmsg, err := service.BetEdge(context.Background(), "", "", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("guest retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeStore{}, fakeAuth{user: map[string]interface{}{"uid": "5"}})
	_, retcode, errmsg, err = service.BetEdge(context.Background(), "token", "", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "押注数量不能为零" {
		t.Fatalf("quantity retcode=%d errmsg=%q", retcode, errmsg)
	}

	now := time.Unix(1000, 0)
	tests := []struct {
		name    string
		store   fakeStore
		retcode int
		errmsg  string
	}{
		{name: "room", store: fakeStore{}, retcode: -1, errmsg: "无效场次"},
		{name: "period", store: fakeStore{room: map[string]interface{}{"id": "1", "coins": "10"}}, retcode: -1, errmsg: "无效的活动期号"},
		{name: "notStarted", store: fakeStore{room: map[string]interface{}{"id": "1", "coins": "10"}, record: map[string]interface{}{"start_time": "1001", "end_time": "2000"}}, retcode: -1, errmsg: "活动尚未开始"},
		{name: "ended", store: fakeStore{room: map[string]interface{}{"id": "1", "coins": "10"}, record: map[string]interface{}{"start_time": "1", "end_time": "999"}}, retcode: -1, errmsg: "活动已结束"},
		{name: "unknownUser", store: fakeStore{room: map[string]interface{}{"id": "1", "coins": "10"}, record: map[string]interface{}{"start_time": "1", "end_time": "2000"}}, retcode: -1, errmsg: "未知用户"},
		{name: "balance", store: fakeStore{room: map[string]interface{}{"id": "1", "coins": "10"}, record: map[string]interface{}{"start_time": "1", "end_time": "2000"}, quota: map[string]interface{}{"goldcoin": "9"}}, retcode: -1, errmsg: "余额不足"},
		{name: "success", store: fakeStore{room: map[string]interface{}{"id": "1", "coins": "10"}, record: map[string]interface{}{"start_time": "1", "end_time": "2000"}, quota: map[string]interface{}{"goldcoin": "20"}}, retcode: 0, errmsg: ""},
	}
	for _, tt := range tests {
		service = NewService(tt.store, fakeAuth{user: map[string]interface{}{"uid": "5"}})
		service.now = func() time.Time { return now }
		_, retcode, errmsg, err = service.BetEdge(context.Background(), "token", "2026071501", 1, 1)
		if err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		if retcode != tt.retcode || errmsg != tt.errmsg {
			t.Fatalf("%s retcode=%d errmsg=%q", tt.name, retcode, errmsg)
		}
	}
}

func TestBetEdgeReturnsBetData(t *testing.T) {
	store := fakeStore{
		room:      map[string]interface{}{"id": "1", "coins": "10"},
		record:    map[string]interface{}{"start_time": "1", "end_time": "2000"},
		quota:     map[string]interface{}{"goldcoin": "20"},
		betResult: domain.OneGoBetResult{BetNo: []int{3}, TotalBetNo: []int{1, 2, 3}},
	}
	service := NewService(store, fakeAuth{user: map[string]interface{}{"uid": "5"}})
	service.now = func() time.Time { return time.Unix(1000, 0) }

	data, retcode, errmsg, err := service.BetEdge(context.Background(), "token", "2026071501", 1, 1)
	if err != nil {
		t.Fatalf("bet: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	row, ok := data["data"].(domain.OneGoBetResult)
	if !ok || len(row.BetNo) != 1 || row.BetNo[0] != 3 || len(row.TotalBetNo) != 3 {
		t.Fatalf("data=%#v", data)
	}
}

func TestHistoryBuildsUserBetRows(t *testing.T) {
	service := NewService(fakeStore{
		userOrders: []map[string]interface{}{
			{"id": "1", "uid": "5", "period": "2026071401", "room_id": "2", "bet_coins": "20"},
		},
		periodOrders: []map[string]interface{}{
			{"bet_no": "1,2"},
			{"bet_no": "3"},
		},
		record: map[string]interface{}{"open_no": "2", "awards": "100"},
	}, fakeAuth{user: map[string]interface{}{"uid": "5"}})

	data, retcode, errmsg, err := service.History(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	rows := data.Data.([]map[string]interface{})
	if rows[0]["id"] != 1 || rows[0]["win_no"] != 2 || rows[0]["win_coins"] != 100 {
		t.Fatalf("row = %#v", rows[0])
	}
	betNos := rows[0]["bet_no"].([]string)
	if len(betNos) != 3 || betNos[2] != "3" {
		t.Fatalf("bet_no = %#v", betNos)
	}
}

func TestBetRanksValidatesRoomAndPeriod(t *testing.T) {
	service := NewService(fakeStore{})
	if _, err := service.BetRanks(context.Background(), "", 0, 1); !errors.Is(err, ErrInvalidRoom) {
		t.Fatalf("expected ErrInvalidRoom, got %v", err)
	}

	service = NewService(fakeStore{room: map[string]interface{}{"id": "2"}})
	if _, err := service.BetRanks(context.Background(), "2026071401", 2, 1); !errors.Is(err, ErrInvalidPeriod) {
		t.Fatalf("expected ErrInvalidPeriod, got %v", err)
	}
}

func TestBetRanksFormatsOrderRows(t *testing.T) {
	service := NewService(fakeStore{
		room:   map[string]interface{}{"id": "2", "name": "初级场", "coins": "10"},
		record: map[string]interface{}{"id": "10", "period": "2026071401"},
		rankBetCoins: []map[string]interface{}{
			{"uid": "5", "room_id": "2", "total_coins": "30", "total_bets": "2"},
		},
		user: map[string]interface{}{"uid": "5", "username": "bettor", "avatar": ""},
	})

	data, err := service.BetRanks(context.Background(), "2026071401", 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	rows := data.Data.([]map[string]interface{})
	if rows[0]["uid"] != 5 || rows[0]["total_coins"] != 30 || rows[0]["total_bets"] != 3 || rows[0]["room_name"] != "初级场" {
		t.Fatalf("rank row = %#v", rows[0])
	}
	if rows[0]["user"].(map[string]interface{})["username"] != "bettor" {
		t.Fatalf("user = %#v", rows[0]["user"])
	}
}

func TestMarqueeReturnsMessages(t *testing.T) {
	service := NewService(fakeStore{
		latestRecord: map[string]interface{}{"period": "2026071401"},
		rules:        map[string]interface{}{"marquee": "{user} 在 {room} 第 {period} 期赢得 {awards} 金币，胜率 {win_rate}%"},
		periodRecords: []map[string]interface{}{
			sampleRecord("5"),
			func() map[string]interface{} {
				row := sampleRecord("5")
				row["id"] = "11"
				row["awards"] = "0"
				return row
			}(),
		},
		room: map[string]interface{}{"id": "1", "name": "初级场"},
		user: map[string]interface{}{"uid": "5", "username": "winner", "avatar": ""},
	})

	data, err := service.Marquee(context.Background())
	if err != nil {
		t.Fatalf("marquee: %v", err)
	}
	messages, ok := data.Data.([]string)
	if !ok {
		t.Fatalf("expected messages, got %T", data.Data)
	}
	if len(messages) != 1 {
		t.Fatalf("expected one message, got %#v", messages)
	}
	want := "winner 在 初级场 第 2026071401 期赢得 100 金币，胜率 65%"
	if messages[0] != want {
		t.Fatalf("unexpected message %q", messages[0])
	}
}

func TestMarqueeNoDataAndNotOpen(t *testing.T) {
	service := NewService(fakeStore{})
	if _, err := service.Marquee(context.Background()); !errors.Is(err, ErrNoData) {
		t.Fatalf("expected ErrNoData, got %v", err)
	}

	service = NewService(fakeStore{latestRecord: map[string]interface{}{"period": "2026071401"}})
	if _, err := service.Marquee(context.Background()); !errors.Is(err, ErrNotOpen) {
		t.Fatalf("expected ErrNotOpen, got %v", err)
	}
}

func sampleRecord(winner string) map[string]interface{} {
	return map[string]interface{}{
		"id":          "10",
		"start_time":  "1770000000",
		"end_time":    "1770000600",
		"period":      "2026071401",
		"hash_code":   "abc",
		"hash_period": "123456",
		"room_id":     "1",
		"total_bets":  "3",
		"total_coins": "30",
		"open_no":     "-1",
		"winner":      winner,
		"awards":      "100",
		"win_rate":    "6500",
		"bot":         "0",
		"paid":        "0",
		"open_time":   "0",
	}
}
