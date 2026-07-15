package ucp

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type fakeUserStore struct {
	user               map[string]interface{}
	guest              map[string]interface{}
	missingGuest       bool
	coinLogTypes       []int
	coinLogOrderBy     string
	countCoinLogTypes  []int
	countCoinLogResult int
	coinBonusStats     map[string]interface{}
	feedbackRow        map[string]interface{}
	paymentRow         map[string]interface{}
	attachRows         []map[string]interface{}
	posters            []map[string]interface{}
	taskboxes          []map[string]interface{}
	taskboxLog         map[string]interface{}
	taskboxLogs        []map[string]interface{}
	bankcards          []map[string]interface{}
	banks              []map[string]interface{}
}

func (s fakeUserStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s fakeUserStore) Groups(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"gid": "0", "gname": "游客", "gicon": "", "minup": "0", "weight": "0", "scope": "0", "perms": `{"max.vod.play.daynum":"10","max.vod.down.daynum":"8","max.minivod.play.daynum":"20","max.minivod.down.daynum":"18"}`},
		{"gid": "4", "gname": "普通会员", "gicon": "V4", "minup": "100", "weight": "4", "scope": "0", "perms": `{"max.vod.play.daynum":"40","max.vod.down.daynum":"30","max.comment.post.daynum":"4","max.minivod.play.daynum":"150","max.minivod.down.daynum":"150"}`},
		{"gid": "7", "gname": "禁止发言", "gicon": "V7", "minup": "2000000", "weight": "7", "scope": "0", "perms": `{"max.vod.play.daynum":"0","max.vod.down.daynum":"0","max.comment.post.daynum":"0","max.minivod.play.daynum":"0","max.minivod.down.daynum":"0"}`},
		{"gid": "6", "gname": "尊贵VIP", "gicon": "V6", "minup": "1000000", "weight": "6", "scope": "0", "perms": `{"max.vod.play.daynum":"1000","max.vod.down.daynum":"202","max.comment.post.daynum":"50","max.minivod.play.daynum":"999","max.minivod.down.daynum":"200"}`},
	}, nil
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

func (s fakeUserStore) RollTitles(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"id": "1", "message": "这是一条测试消息", "status": "1"}}, nil
}

func (s fakeUserStore) Posters(context.Context) ([]map[string]interface{}, error) {
	return s.posters, nil
}

func (s fakeUserStore) Taskboxes(context.Context) ([]map[string]interface{}, error) {
	return s.taskboxes, nil
}

func (s fakeUserStore) TaskboxLog(context.Context, int, int, int) (map[string]interface{}, error) {
	if s.taskboxLog != nil {
		return s.taskboxLog, nil
	}
	return map[string]interface{}{}, nil
}

func (s fakeUserStore) TaskboxCompletedLogs(context.Context, int) ([]map[string]interface{}, error) {
	return s.taskboxLogs, nil
}

func (s fakeUserStore) CountTaskboxLogs(context.Context, int) (int, error) {
	return len(s.taskboxLogs), nil
}

func (s fakeUserStore) TaskboxLogs(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return s.taskboxLogs, nil
}

func (s fakeUserStore) CountPayments(context.Context, int) (int, error) {
	return 1, nil
}

func (s fakeUserStore) Payments(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"payid":      "2536422300000504",
			"paytype":    "8",
			"payway":     "safepay",
			"paycode":    "safepay",
			"itemname":   "60天",
			"trx_amount": "6800",
			"pay_amount": "6800",
			"uid":        "5",
			"createtime": "1762948948",
			"ispaid":     "1",
			"paidtime":   "1762949999",
			"out_trxid":  "123",
		},
	}, nil
}

func (s fakeUserStore) SafePayLogs(context.Context, int, int64, int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"payid":      "2536422300000504",
			"paytype":    "8",
			"payway":     "safepay",
			"paycode":    "safepay",
			"itemname":   "60天",
			"trx_amount": "6800",
			"pay_amount": "6800",
			"uid":        "5",
			"createtime": "1762948948",
			"ispaid":     "1",
			"paidtime":   "0",
			"out_trxid":  "",
		},
	}, nil
}

func (s fakeUserStore) PaymentsSince(context.Context, int, int64, int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"payid":      "2536422300000504",
			"paytype":    "8",
			"payway":     "safepay",
			"paycode":    "safepay",
			"itemname":   "60天",
			"trx_amount": "6800",
			"pay_amount": "6800",
			"uid":        "5",
			"createtime": "1762948948",
			"ispaid":     "1",
			"paidtime":   "1762949999",
			"out_trxid":  "123",
		},
	}, nil
}

func (s fakeUserStore) Account(context.Context, int) (map[string]interface{}, error) {
	return map[string]interface{}{
		"uid":                    "5",
		"balance":                "1000",
		"frozen":                 "200",
		"deposit":                "300",
		"game_balance":           "400",
		"game_frozen":            "50",
		"available_balance":      "800",
		"game_available_balance": "350",
	}, nil
}

func (s fakeUserStore) Quota(context.Context, int) (map[string]interface{}, error) {
	return map[string]interface{}{"uid": "5", "goldcoin": "625"}, nil
}

func (s fakeUserStore) Goldbean(context.Context, int) (map[string]interface{}, error) {
	return map[string]interface{}{"uid": "5", "gold_bean": "3815"}, nil
}

func (s fakeUserStore) CountVODPlayLogsSince(context.Context, int, int64) (int, error) {
	return 1, nil
}

func (s fakeUserStore) CountVODDownLogsSince(context.Context, int, int64) (int, error) {
	return 2, nil
}

func (s fakeUserStore) GuestBySID(context.Context, string) (map[string]interface{}, error) {
	if s.missingGuest {
		return map[string]interface{}{}, nil
	}
	if s.guest != nil {
		return s.guest, nil
	}
	return map[string]interface{}{"sid": "guestguestguestguestguestguest12", "goldcoin": "9", "signtime": "1770000000"}, nil
}

func (s fakeUserStore) CountGuestVODPlayLogsSince(context.Context, string, int64) (int, error) {
	return 3, nil
}

func (s fakeUserStore) CountGuestVODDownLogsSince(context.Context, string, int64) (int, error) {
	return 4, nil
}

func (s fakeUserStore) CountMiniVODViewLogsSince(_ context.Context, _ int, _ int64, action int) (int, error) {
	if action == 2 {
		return 5, nil
	}
	return 4, nil
}

func (s fakeUserStore) CountGuestMiniVODViewLogsSince(_ context.Context, _ string, _ int64, action int) (int, error) {
	if action == 2 {
		return 7, nil
	}
	return 6, nil
}

func (s fakeUserStore) CountCoinLogsSinceByType(context.Context, int, int, int64) (int, error) {
	return 1, nil
}

func (s fakeUserStore) CountFeedbacks(context.Context, int) (int, error) {
	return 1, nil
}

func (s fakeUserStore) Feedbacks(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"id":         "1917132",
			"cid":        "1",
			"content":    "afasdf",
			"ctimestamp": "1778914934",
			"replytime":  "0",
			"replytext":  "",
			"payid":      "0",
			"payname":    "",
			"payaccount": "",
		},
	}, nil
}

func (s fakeUserStore) CountFeedbacksByType(_ context.Context, _ int, feedbackType int) (int, error) {
	if feedbackType == 2 {
		return 0, nil
	}
	return 1, nil
}

func (s fakeUserStore) FeedbacksByType(_ context.Context, _ int, feedbackType int, _ int, _ int) ([]map[string]interface{}, error) {
	if feedbackType == 2 {
		return []map[string]interface{}{}, nil
	}
	return s.Feedbacks(context.Background(), 0, 1, 20)
}

func (s fakeUserStore) FeedbackByID(_ context.Context, id int) (map[string]interface{}, error) {
	if id == 404 {
		return map[string]interface{}{}, nil
	}
	if s.feedbackRow != nil {
		return s.feedbackRow, nil
	}
	return map[string]interface{}{
		"id":         "1917132",
		"uid":        "5",
		"cid":        "1",
		"content":    "afasdf",
		"ctimestamp": "1778914934",
		"replytime":  "0",
		"replytext":  "",
		"payid":      "0",
		"payname":    "",
		"payaccount": "",
		"aids":       "",
	}, nil
}

func (s fakeUserStore) PaymentByID(context.Context, int) (map[string]interface{}, error) {
	if s.paymentRow != nil {
		return s.paymentRow, nil
	}
	return map[string]interface{}{}, nil
}

func (s fakeUserStore) AttachByIDs(context.Context, []int) ([]map[string]interface{}, error) {
	if s.attachRows != nil {
		return s.attachRows, nil
	}
	return []map[string]interface{}{}, nil
}

func (s fakeUserStore) CountMsgConversations(context.Context, int) (int, error) {
	return 1, nil
}

func (s fakeUserStore) MsgConversations(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"uid":           "5",
			"cid":           "9776805",
			"ruid":          "0",
			"risread":       "1",
			"msgcount":      "1",
			"newmsg":        "1",
			"last_msgid":    "9776805",
			"last_sendtime": "1762921166",
			"senderid":      "0",
			"content":       "你办理的业务已支付成功\n相关业务：60天\n支付金额：￥68.00\n",
			"sendtime":      "1762921166",
			"username":      nil,
			"avatar":        nil,
		},
	}, nil
}

func (s fakeUserStore) MsgConversation(context.Context, int, int) (map[string]interface{}, error) {
	return map[string]interface{}{"uid": "5", "cid": "9", "ruid": "7", "newmsg": "1"}, nil
}

func (s fakeUserStore) UserByID(context.Context, int) (map[string]interface{}, error) {
	if s.user != nil {
		return s.user, nil
	}
	return map[string]interface{}{"uid": "7", "username": "peer"}, nil
}

func (s fakeUserStore) Bankcards(context.Context, int) ([]map[string]interface{}, error) {
	return s.bankcards, nil
}

func (s fakeUserStore) Banks(context.Context) ([]map[string]interface{}, error) {
	return s.banks, nil
}

func (s fakeUserStore) CountMessages(context.Context, int, int) (int, error) {
	return 1, nil
}

func (s fakeUserStore) Messages(context.Context, int, int, int, int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{
		"uid":      "5",
		"cid":      "9",
		"msgid":    "11",
		"senderid": "7",
		"content":  `<a href="foo">foo</a>`,
		"sendtime": "100",
		"username": "peer",
		"avatar":   "",
	}}, nil
}

func (s fakeUserStore) SetMsgRead(context.Context, int, int) error {
	return nil
}

func (s fakeUserStore) CleanMsgRead(context.Context, int) error {
	return nil
}

func (s fakeUserStore) DeleteMsgConversations(context.Context, int, []int) error {
	return nil
}

func (s fakeUserStore) BalanceLogs(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"trxid":   "111",
			"paytype": "8",
			"uid":     "5",
			"trxin":   "0",
			"trxout":  "6800",
			"balance": "0",
			"trxtime": "1762921166",
			"remark":  "60天",
		},
	}, nil
}

func (s fakeUserStore) CoinLogs(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"logid":            "1284862472",
			"uid":              "5",
			"cointype":         "7",
			"coinnum":          "10",
			"balance":          "625",
			"addtime":          "1767830400",
			"remark":           "",
			"invited_uid":      "0",
			"mobi":             nil,
			"invited_username": "",
		},
		{
			"logid":            "1284862400",
			"uid":              "5",
			"cointype":         "999",
			"coinnum":          "-10",
			"balance":          "615",
			"addtime":          "1764720000",
			"remark":           "old",
			"invited_uid":      "9",
			"mobi":             "86.13800138000",
			"invited_username": "好友",
		},
	}, nil
}

func (s fakeUserStore) CountCoinLogsByTypes(_ context.Context, _ int, coinTypes []int) (int, error) {
	if s.countCoinLogTypes != nil && !equalInts(coinTypes, s.countCoinLogTypes) {
		return -1, nil
	}
	if s.countCoinLogResult > 0 {
		return s.countCoinLogResult, nil
	}
	return 2, nil
}

func (s fakeUserStore) CoinLogsByTypes(_ context.Context, _ int, coinTypes []int, _ int, _ int, orderBy string) ([]map[string]interface{}, error) {
	if s.coinLogTypes != nil && !equalInts(coinTypes, s.coinLogTypes) {
		return nil, nil
	}
	if s.coinLogOrderBy != "" && orderBy != s.coinLogOrderBy {
		return nil, nil
	}
	return []map[string]interface{}{
		{
			"logid":            "1284862472",
			"uid":              "5",
			"cointype":         fmt.Sprint(coinTypes[0]),
			"coinnum":          "10",
			"balance":          "625",
			"addtime":          "1764720000",
			"remark":           "invite days",
			"invited_uid":      "9",
			"mobi":             "13800138000",
			"invited_username": "好友",
		},
	}, nil
}

func (s fakeUserStore) CoinBonusStats(context.Context, int) (map[string]interface{}, error) {
	if s.coinBonusStats != nil {
		return s.coinBonusStats, nil
	}
	return map[string]interface{}{"inviteTotal": "3", "activeTotal": "1", "bonusTotal": "6232"}, nil
}

func (s fakeUserStore) CountBalanceLogs(context.Context, int) (int, error) {
	return 1, nil
}

func (s fakeUserStore) SettingExRate(context.Context) (int, error) {
	return 10, nil
}

func TestTaskSharePicEmpty(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	data, err := service.TaskSharePic(context.Background())
	if err != nil {
		t.Fatalf("sharepic: %v", err)
	}
	rows, ok := data["data"].([]interface{})
	if !ok || len(rows) != 0 {
		t.Fatalf("data = %#v", data)
	}
}

func TestTaskSharePicReturnsPoster(t *testing.T) {
	service := NewService(fakeUserStore{posters: []map[string]interface{}{{"id": "1", "pic": "a.png"}}}, "https://res.example.test")

	data, err := service.TaskSharePic(context.Background())
	if err != nil {
		t.Fatalf("sharepic: %v", err)
	}
	row, ok := data["data"].(map[string]interface{})
	if !ok || row["pic"] != "a.png" {
		t.Fatalf("data = %#v", data)
	}
}

func TestTaskboxIndexGuest(t *testing.T) {
	service := NewService(fakeUserStore{
		taskboxes: []map[string]interface{}{
			{"taskid": "1", "taskname": "推广1人", "showtype": "0"},
			{"taskid": "1022", "taskname": "每日神秘", "showtype": "0"},
		},
		taskboxLogs: []map[string]interface{}{
			{"logid": "9", "username": "u", "nickname": "n", "avatar": "", "addtime": "1760000000", "taskid": "1", "addcoin": "3", "prize": "p", "taskstatus": "2"},
		},
	}, "https://res.example.test")
	service.now = func() time.Time { return time.Date(2026, 7, 15, 10, 0, 0, 0, time.Local) }

	data, err := service.TaskboxIndex(context.Background(), "")
	if err != nil {
		t.Fatalf("taskbox: %v", err)
	}
	taskRows := data["taskrows"].([]map[string]interface{})
	if len(taskRows) != 2 || taskRows[0]["taskstatus"] != 0 {
		t.Fatalf("taskrows = %#v", taskRows)
	}
	logRows := data["logrows"].([]map[string]interface{})
	if len(logRows) != 1 || logRows[0]["taskstatus"] != "已发放" {
		t.Fatalf("logrows = %#v", logRows)
	}
}

func TestTaskboxLogListingRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.TaskboxLogListing(context.Background(), "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestTaskboxLogListingFormatsRowsAndPageInfo(t *testing.T) {
	service := NewService(fakeUserStore{
		user: map[string]interface{}{"uid": "5"},
		taskboxLogs: []map[string]interface{}{
			{"logid": "9", "username": "u", "nickname": "n", "avatar": "", "addtime": "1760000000", "taskid": "1", "addcoin": "3", "prize": "p", "taskstatus": "2"},
		},
	}, "https://res.example.test")

	data, retcode, errmsg, err := service.TaskboxLogListing(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	logRows := data["logrows"].([]map[string]interface{})
	if len(logRows) != 1 || logRows[0]["taskstatus"] != "已发放" {
		t.Fatalf("logrows = %#v", logRows)
	}
	pageInfo := data["pageinfo"].(map[string]interface{})
	if pageInfo["page_url"] != "/ucp/taskbox/taskboxlog?page=[?]" || pageInfo["pagesize"] != 20 {
		t.Fatalf("pageinfo = %#v", pageInfo)
	}
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

func TestUserIndexRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.UserIndex(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestUserIndexReturnsProcessedUser(t *testing.T) {
	service := NewService(fakeUserStore{
		user: map[string]interface{}{
			"uid":             "5",
			"uniqkey":         "1904908418",
			"username":        "~1904908418",
			"nickname":        "1400002",
			"mobi":            "86.14012340002",
			"email":           "~1904908418",
			"sysgid":          "6",
			"sysgid_exptime":  "0",
			"gid":             "4",
			"regtime":         "1547688484",
			"gender":          "1",
			"avatar":          "sysavatar/man/5.png",
			"newmsg":          "0",
			"recommend_total": "6",
		},
	}, "https://res.example.test")

	data, retcode, errmsg, err := service.UserIndex(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	user := data["user"].(map[string]interface{})
	if user["uid"] != "5" || user["uniqkey"] != "VI4SQQ" || user["avatar_url"] != "https://res.example.test/sysavatar/man/5.png" {
		t.Fatalf("user = %#v", user)
	}
}

func TestBankcardIndexRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.BankcardIndex(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestBankcardIndexReturnsCardsAndBanks(t *testing.T) {
	service := NewService(fakeUserStore{
		user:      map[string]interface{}{"uid": "5"},
		bankcards: []map[string]interface{}{{"cardid": "1", "uid": "5", "name": "张三", "bankname": "中国银行", "cardnum": "123", "isdef": "1", "type": "2"}},
		banks:     []map[string]interface{}{{"bankid": "8", "bankname": "平安银行", "coverpic": ""}},
	}, "https://res.example.test")

	data, retcode, errmsg, err := service.BankcardIndex(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if data["maxallow"] != 3 || data["allowtype"] != 7 {
		t.Fatalf("limits = %#v", data)
	}
	if len(data["cardrows"].([]map[string]interface{})) != 1 {
		t.Fatalf("cardrows = %#v", data["cardrows"])
	}
	bankRows := data["bankRows"].([]map[string]interface{})
	if bankRows[0]["bankid"] != 8 || bankRows[0]["bankname"] != "平安银行" {
		t.Fatalf("bankRows = %#v", bankRows)
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

func TestRollTitleReturnsMessages(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	data, err := service.RollTitle(context.Background())
	if err != nil {
		t.Fatalf("roll title: %v", err)
	}
	if len(data.Messages) != 1 {
		t.Fatalf("expected one message, got %d", len(data.Messages))
	}
	if data.Messages[0]["message"] != "这是一条测试消息" {
		t.Fatalf("unexpected message %#v", data.Messages[0])
	}
}

func TestAffCenterRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.AffCenter(context.Background(), "")
	if err != nil {
		t.Fatalf("affcenter: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestAffCenterFormatsUserAndUInfo(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{
		"uid":             "5",
		"uniqkey":         "1904908418",
		"username":        "~1904908418",
		"nickname":        "1400002",
		"mobi":            "86.14012340002",
		"email":           "~1904908418",
		"sysgid":          "6",
		"gid":             "4",
		"sysgid_exptime":  "0",
		"regtime":         "1547641684",
		"gender":          "1",
		"avatar":          "sysavatar/man/5.png",
		"newmsg":          "9",
		"recommend_total": "6",
		"perms":           `{"max.vod.play.daynum":1000,"max.vod.down.daynum":202}`,
	}}, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	data, retcode, errmsg, err := service.AffCenter(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatalf("affcenter: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if data.User["uid"] != "5" || data.User["goldcoin"] != 625 || data.User["gold_bean"] != 3815 {
		t.Fatalf("unexpected user %#v", data.User)
	}
	if data.User["gicon"] != "V6" || data.User["avatar_url"] != "https://res.example.test/sysavatar/man/5.png" {
		t.Fatalf("unexpected processed user %#v", data.User)
	}
	if data.UInfo["goldcoin"] != "625" || data.UInfo["gold_bean"] != "3815" {
		t.Fatalf("unexpected uinfo coins %#v", data.UInfo)
	}
	if data.UInfo["play_daily_remainders"] != 999 || data.UInfo["down_daily_remainders"] != 200 {
		t.Fatalf("unexpected remainders %#v", data.UInfo)
	}
	curr, ok := data.UInfo["curr_group"].(map[string]interface{})
	if !ok || curr["gid"] != "6" || curr["gname"] != "尊贵VIP" || curr["minup"] != "1000000" {
		t.Fatalf("unexpected curr_group %#v", data.UInfo["curr_group"])
	}
	next, ok := data.UInfo["next_group"].(map[string]interface{})
	if !ok || next["gid"] != "7" || next["gname"] != "禁止发言" || next["minup"] != "2000000" {
		t.Fatalf("unexpected next_group %#v", data.UInfo["next_group"])
	}
	if data.UInfo["next_upgrade_need"] != 1999994 {
		t.Fatalf("unexpected next_upgrade_need %#v", data.UInfo["next_upgrade_need"])
	}
}

func TestIndexFormatsLoggedInUser(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{
		"uid":             "5",
		"uniqkey":         "1904908418",
		"username":        "~1904908418",
		"nickname":        "1400002",
		"mobi":            "~86.14012340002",
		"email":           "~1904908418",
		"sysgid":          "6",
		"gid":             "4",
		"sysgid_exptime":  "0",
		"regtime":         "1547641684",
		"gender":          "1",
		"avatar":          "sysavatar/man/5.png",
		"newmsg":          "9",
		"recommend_total": "6",
	}}, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	data, retcode, errmsg, err := service.Index(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if data.User["mobi"] != "" || data.User["email"] != "" {
		t.Fatalf("expected tilde contacts to be cleared, got %#v", data.User)
	}
	if data.User["goldcoin"] != 625 || data.User["gold_bean"] != 3815 {
		t.Fatalf("unexpected user money fields %#v", data.User)
	}
	if data.Signed != 1 {
		t.Fatalf("expected signed=1, got %d", data.Signed)
	}
	if data.UInfo["play_daily_remainders"] != 999 ||
		data.UInfo["down_daily_remainders"] != 200 ||
		data.UInfo["minivod_play_daily_remainders"] != 995 ||
		data.UInfo["minivod_down_daily_remainders"] != 195 {
		t.Fatalf("unexpected uinfo remainders %#v", data.UInfo)
	}
	curr, ok := data.UInfo["curr_group"].(map[string]interface{})
	if !ok || curr["gicon"] != "V6" {
		t.Fatalf("unexpected curr_group %#v", data.UInfo["curr_group"])
	}
	if len(data.Groups) != 3 {
		t.Fatalf("expected three visible groups, got %#v", data.Groups)
	}
	if data.Groups[0]["gicon"] != "V4" || data.Groups[1]["gicon"] != "V6" || data.Groups[1]["play_daynum"] != 1000 {
		t.Fatalf("unexpected groups %#v", data.Groups)
	}
}

func TestIndexFormatsGuest(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	data, retcode, errmsg, err := service.Index(context.Background(), "guestguestguestguestguestguest12")
	if err != nil {
		t.Fatalf("index guest: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if data.Signed != 1 {
		t.Fatalf("expected signed=1, got %d", data.Signed)
	}
	if data.UInfo["goldcoin"] != "9" || data.UInfo["curr_group"] != nil || data.UInfo["next_group"] != nil {
		t.Fatalf("unexpected guest uinfo %#v", data.UInfo)
	}
	if data.UInfo["play_daily_remainders"] != 7 ||
		data.UInfo["down_daily_remainders"] != 4 ||
		data.UInfo["minivod_play_daily_remainders"] != 14 ||
		data.UInfo["minivod_down_daily_remainders"] != 11 {
		t.Fatalf("unexpected guest remainders %#v", data.UInfo)
	}
	if data.Groups != nil {
		t.Fatalf("expected guest groups omitted, got %#v", data.Groups)
	}
}

func TestIndexGuestMissingRow(t *testing.T) {
	service := NewService(fakeUserStore{missingGuest: true}, "https://res.example.test")

	_, retcode, errmsg, err := service.Index(context.Background(), "guestguestguestguestguestguest12")
	if err != nil {
		t.Fatalf("index guest missing: %v", err)
	}
	if retcode != -1 || errmsg != "请登录后操作，客户端游客请先携带信息" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestFeedbackListingRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.FeedbackListing(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("feedback listing: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestFeedbackListingFormatsRows(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	data, retcode, errmsg, err := service.FeedbackListing(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatalf("feedback listing: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if len(data.Rows) != 1 {
		t.Fatalf("expected one row, got %d", len(data.Rows))
	}
	row := data.Rows[0]
	if row["id"] != "1917132" || row["cid"] != "1" || row["content"] != "afasdf" {
		t.Fatalf("unexpected row %#v", row)
	}
	if row["ctimestamp"] != "2026-05-16 15:02" || row["replytime"] != "" {
		t.Fatalf("unexpected times %#v", row)
	}
	if row["itemname"] != nil || row["paidtime"] != "" {
		t.Fatalf("unexpected payment compatibility fields %#v", row)
	}
	if data.PageInfo["page_url"] != "/ucp/feedback?page=[?]" || data.PageInfo["pagesize"] != 20 {
		t.Fatalf("unexpected pageinfo %#v", data.PageInfo)
	}
}

func TestFeedbackIndexRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.FeedbackIndex(context.Background(), "")
	if err != nil {
		t.Fatalf("feedback index: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestFeedbackIndexReturnsRecentPayments(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	data, retcode, errmsg, err := service.FeedbackIndex(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatalf("feedback index: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if len(data.PayRows) != 1 {
		t.Fatalf("expected one pay row, got %d", len(data.PayRows))
	}
	row := data.PayRows[0]
	if row["payid"] != "2536422300000504" || row["pay_amount"] != "68.00" || row["payway_name"] != "人工代付" {
		t.Fatalf("unexpected pay row %#v", row)
	}
}

func TestFeedbackNewListingRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.FeedbackNewListing(context.Background(), "", 0, 1)
	if err != nil {
		t.Fatalf("feedback new listing: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestFeedbackNewListingFormatsTypeURL(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	data, retcode, errmsg, err := service.FeedbackNewListing(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 99, 0)
	if err != nil {
		t.Fatalf("feedback new listing: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if len(data.Rows) != 1 {
		t.Fatalf("expected one row, got %d", len(data.Rows))
	}
	if data.PageInfo["page_url"] != "/ucp/feedback/listing?type=0&page=[?]" || data.PageInfo["curr_url"] != "/ucp/feedback/listing?type=0&page=1" {
		t.Fatalf("unexpected pageinfo %#v", data.PageInfo)
	}
}

func TestFeedbackNewListingTypeTwoEmpty(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")

	data, retcode, errmsg, err := service.FeedbackNewListing(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 2, 1)
	if err != nil {
		t.Fatalf("feedback new listing: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if len(data.Rows) != 0 {
		t.Fatalf("expected no rows, got %#v", data.Rows)
	}
	if data.PageInfo["total"] != 0 || data.PageInfo["start"] != 0 || data.PageInfo["page_url"] != "/ucp/feedback/listing?type=2&page=[?]" {
		t.Fatalf("unexpected pageinfo %#v", data.PageInfo)
	}
}

func TestFeedbackDetailRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.FeedbackDetail(context.Background(), "", 1917132)
	if err != nil {
		t.Fatalf("feedback detail: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestFeedbackDetailMissingOrForeignRow(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")

	_, retcode, errmsg, err := service.FeedbackDetail(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 404)
	if err != nil {
		t.Fatalf("feedback detail: %v", err)
	}
	if retcode != -1 || errmsg != "记录不存在或已被删除" {
		t.Fatalf("unexpected missing response %d %q", retcode, errmsg)
	}
}

func TestFeedbackDetailFormatsRowPaymentAndPicURLs(t *testing.T) {
	service := NewService(fakeUserStore{
		user: map[string]interface{}{"uid": "5"},
		feedbackRow: map[string]interface{}{
			"id":         "1917133",
			"uid":        "5",
			"cid":        "5",
			"content":    "支付没到账",
			"ctimestamp": "1778914934",
			"replytime":  "1778918534",
			"replytext":  "已处理",
			"payid":      "2536422300000504",
			"payname":    "支付宝",
			"payaccount": "user@example.test",
			"aids":       "10,20",
		},
		paymentRow: map[string]interface{}{
			"payid":    "2536422300000504",
			"itemname": "60天",
			"paidtime": "1762949999",
		},
		attachRows: []map[string]interface{}{
			{"aid": "10", "uri": "feedback/a.png"},
			{"aid": "20", "uri": "https://cdn.example.test/feedback/b.png"},
		},
	}, "https://res.example.test")

	data, retcode, errmsg, err := service.FeedbackDetail(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1917133)
	if err != nil {
		t.Fatalf("feedback detail: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if data.Row["id"] != "1917133" || data.Row["itemname"] != "60天" || data.Row["paidtime"] != "2025-11-12 20:19" {
		t.Fatalf("unexpected row %#v", data.Row)
	}
	if data.Row["ctimestamp"] != "2026-05-16 15:02" || data.Row["replytime"] != "2026-05-16 16:02" {
		t.Fatalf("unexpected times %#v", data.Row)
	}
	pics, ok := data.PicURLs.([]string)
	if !ok {
		t.Fatalf("expected pic urls []string, got %T", data.PicURLs)
	}
	if len(pics) != 2 || pics[0] != "https://res.example.test/feedback/a.png" || pics[1] != "https://cdn.example.test/feedback/b.png" {
		t.Fatalf("unexpected pic urls %#v", pics)
	}
}

func TestFeedbackDetailEmptyPicURLsAreNil(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")

	data, retcode, errmsg, err := service.FeedbackDetail(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1917132)
	if err != nil {
		t.Fatalf("feedback detail: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if data.PicURLs != nil {
		t.Fatalf("expected nil picurls, got %#v", data.PicURLs)
	}
	if data.Row["itemname"] != nil || data.Row["paidtime"] != "" {
		t.Fatalf("unexpected empty payment fields %#v", data.Row)
	}
}

func TestMsgListingRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.MsgListing(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("msg listing: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestMsgListingFormatsRows(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	data, retcode, errmsg, err := service.MsgListing(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatalf("msg listing: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if len(data.Rows) != 1 {
		t.Fatalf("expected one row, got %d", len(data.Rows))
	}
	row := data.Rows[0]
	if row["cid"] != "9776805" || row["uid"] != "5" || row["username"] != nil || row["avatar"] != nil {
		t.Fatalf("unexpected row %#v", row)
	}
	if row["__url__"] != "/ucp/msg/show?cid=9776805" {
		t.Fatalf("unexpected url %#v", row["__url__"])
	}
	if data.PageInfo["page_url"] != "/ucp/msg?page=[?]" || data.PageInfo["pagesize"] != 20 {
		t.Fatalf("unexpected pageinfo %#v", data.PageInfo)
	}
}

func TestPaymentListingRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.PaymentListing(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("payment listing: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestPaymentListingFormatsRows(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")

	data, retcode, errmsg, err := service.PaymentListing(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatalf("payment listing: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if len(data.Rows) != 1 {
		t.Fatalf("expected one row, got %d", len(data.Rows))
	}
	row := data.Rows[0]
	if row["paytype"] != "购买套餐" {
		t.Fatalf("unexpected paytype %v", row["paytype"])
	}
	if row["payway_name"] != "人工代付" || row["paycode_name"] != "客服支付" {
		t.Fatalf("unexpected pay names %#v", row)
	}
	if row["trx_amount"] != "68.00" || row["pay_amount"] != "68.00" {
		t.Fatalf("unexpected amounts %#v", row)
	}
	if row["ispaid"] != 1 {
		t.Fatalf("unexpected ispaid %v", row["ispaid"])
	}
	if row["paidtime"] == "" {
		t.Fatalf("expected paidtime")
	}
	if data.PageInfo["pagesize"] != 20 || data.PageInfo["page"] != 1 {
		t.Fatalf("unexpected pageinfo %#v", data.PageInfo)
	}
}

func TestSafePayLogFormatsRows(t *testing.T) {
	store := fakeUserStore{user: map[string]interface{}{"uid": "5"}}
	service := NewService(store, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(2000000, 0) }

	data, retcode, errmsg, err := service.SafePayLog(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatalf("safe pay log: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if len(data.PayRows) != 1 {
		t.Fatalf("expected one row, got %d", len(data.PayRows))
	}
	row := data.PayRows[0]
	if row["paidtime"] != "" {
		t.Fatalf("expected empty paidtime, got %v", row["paidtime"])
	}
	if row["payway"] != "safepay" {
		t.Fatalf("unexpected payway %v", row["payway"])
	}
}

func TestProcessPaymentRowsUnknownPaymentNames(t *testing.T) {
	rows := processPaymentRows([]map[string]interface{}{
		{
			"payid":      "1",
			"paytype":    "999",
			"payway":     "unknown",
			"paycode":    "unknown",
			"trx_amount": "1",
			"pay_amount": "2",
		},
	})
	if rows[0]["paytype"] != nil || rows[0]["payway_name"] != nil || rows[0]["paycode_name"] != nil {
		t.Fatalf("expected nil names for unknown payment maps, got %#v", rows[0])
	}
}

func TestAccountIndexRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.AccountIndex(context.Background(), "")
	if err != nil {
		t.Fatalf("account index: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestAccountIndexFormatsRows(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	data, retcode, errmsg, err := service.AccountIndex(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatalf("account index: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if data.GoldCoin != 625 || data.ExRate != 10 {
		t.Fatalf("unexpected coin/exrate %#v", data)
	}
	if data.Coin2RMB != "62.50" || data.Max2RMB != "70.50" {
		t.Fatalf("unexpected rmb fields %#v", data)
	}
	if data.Account["balance"] != "10.00" || data.Account["available_balance"] != "8.00" {
		t.Fatalf("unexpected account %#v", data.Account)
	}
	if data.Account["game_balance"] != 400 || data.Account["game_available_balance"] != 350 {
		t.Fatalf("unexpected game account %#v", data.Account)
	}
	if len(data.LogRows) != 1 {
		t.Fatalf("expected one log row, got %d", len(data.LogRows))
	}
	if data.LogRows[0]["paytype"] != "购买套餐" || data.LogRows[0]["trxout"] != "68.00" {
		t.Fatalf("unexpected log row %#v", data.LogRows[0])
	}
}

func TestBalanceLogRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.BalanceLog(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("balance log: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestBalanceLogFormatsRows(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	data, retcode, errmsg, err := service.BalanceLog(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatalf("balance log: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if len(data.LogRows) != 1 {
		t.Fatalf("expected one log row, got %d", len(data.LogRows))
	}
	if data.LogRows[0]["paytype"] != "购买套餐" || data.LogRows[0]["trxout"] != "68.00" {
		t.Fatalf("unexpected log row %#v", data.LogRows[0])
	}
	if data.PageInfo["pagesize"] != 20 || data.PageInfo["page"] != 1 {
		t.Fatalf("unexpected pageinfo %#v", data.PageInfo)
	}
}

func TestCoinLogIndexRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.CoinLogIndex(context.Background(), "")
	if err != nil {
		t.Fatalf("coin log index: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestCoinLogIndexFormatsRows(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	data, retcode, errmsg, err := service.CoinLogIndex(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatalf("coin log index: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if data.GoldCoin != 625 || data.ExRate != 10 {
		t.Fatalf("unexpected coin/exrate %#v", data)
	}
	if data.Account["balance"] != "10.00" || data.Account["available_balance"] != "8.00" {
		t.Fatalf("unexpected account %#v", data.Account)
	}
	if len(data.LogRows) != 2 {
		t.Fatalf("expected two log rows, got %d", len(data.LogRows))
	}
	first := data.LogRows[0]
	if first["cointype"] != "点击广告送金币" || first["addtime"] != "25天前 " {
		t.Fatalf("unexpected first coin row %#v", first)
	}
	if first["coinnum"] != "10" || first["balance"] != "625" || first["mobi"] != nil {
		t.Fatalf("unexpected first coin values %#v", first)
	}
	second := data.LogRows[1]
	if second["cointype"] != "--" || second["addtime"] != "2025-12-03" || second["invited_username"] != "好友" {
		t.Fatalf("unexpected second coin row %#v", second)
	}
	if second["mobi"] != "86.138****8000" {
		t.Fatalf("unexpected masked mobi %v", second["mobi"])
	}
}

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{name: "nil", input: nil, want: nil},
		{name: "country code", input: "86.13800138000", want: "86.138****8000"},
		{name: "plain phone", input: "13800138000", want: "138****8000"},
		{name: "unmatched", input: "abc", want: "abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maskPhone(tt.input); got != tt.want {
				t.Fatalf("maskPhone(%#v) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCoinLogInviteLogRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.CoinLogInviteLog(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("coin log invite log: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestCoinLogBonusLogRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.CoinLogBonusLog(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("coin log bonus log: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestCoinLogBonusLogFormatsRowsAndStats(t *testing.T) {
	store := fakeUserStore{
		user:               map[string]interface{}{"uid": "5"},
		countCoinLogTypes:  coinLogBonusListTypes,
		coinLogTypes:       coinLogBonusListTypes,
		coinLogOrderBy:     "addtime DESC",
		countCoinLogResult: 1,
		coinBonusStats:     map[string]interface{}{"inviteTotal": "3", "activeTotal": "1", "bonusTotal": "6232"},
	}
	service := NewService(store, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	data, retcode, errmsg, err := service.CoinLogBonusLog(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatalf("coin log bonus log: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if len(data.LogRows) != 1 {
		t.Fatalf("expected one log row, got %d", len(data.LogRows))
	}
	row := data.LogRows[0]
	if row["cointype"] != "未知类型进账" {
		t.Fatalf("unexpected cointype %v", row["cointype"])
	}
	if data.AddInfo["inviteTotal"] != 3 || data.AddInfo["activeTotal"] != 1 || data.AddInfo["bonusTotal"] != 6232 {
		t.Fatalf("unexpected addinfo %#v", data.AddInfo)
	}
	if data.PageInfo["page_url"] != "/ucp/coinlog/bonuslog?page=[?]" || data.PageInfo["pagesize"] != 20 {
		t.Fatalf("unexpected pageinfo %#v", data.PageInfo)
	}
}

func TestCoinLogInviteLogFormatsRows(t *testing.T) {
	store := fakeUserStore{
		user:               map[string]interface{}{"uid": "5"},
		countCoinLogTypes:  []int{201, 32, 11},
		coinLogTypes:       []int{201, 32, 11},
		coinLogOrderBy:     "addtime DESC",
		countCoinLogResult: 1,
	}
	service := NewService(store, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(1770000000, 0) }

	data, retcode, errmsg, err := service.CoinLogInviteLog(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatalf("coin log invite log: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if len(data.LogRows) != 1 {
		t.Fatalf("expected one log row, got %d", len(data.LogRows))
	}
	row := data.LogRows[0]
	if row["cointype"] != "邀请好友赠送vip天数" {
		t.Fatalf("unexpected cointype %v", row["cointype"])
	}
	if row["mobi"] != "138****8000" || row["invited_username"] != "好友" {
		t.Fatalf("unexpected invite row %#v", row)
	}
	if data.PageInfo["page_url"] != "/ucp/coinlog/invitelog?page=[?]" || data.PageInfo["pagesize"] != 20 {
		t.Fatalf("unexpected pageinfo %#v", data.PageInfo)
	}
}

type trackingMsgStore struct {
	fakeUserStore
	readCIDs []int
	cleanUID int
	deleteID int
	deleted  []int
}

func (s *trackingMsgStore) SetMsgRead(_ context.Context, uid int, cid int) error {
	s.readCIDs = append(s.readCIDs, cid)
	return nil
}

func (s *trackingMsgStore) CleanMsgRead(_ context.Context, uid int) error {
	s.cleanUID = uid
	return nil
}

func (s *trackingMsgStore) DeleteMsgConversations(_ context.Context, uid int, cids []int) error {
	s.deleteID = uid
	s.deleted = append([]int{}, cids...)
	return nil
}

func TestMsgSetReadRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	retcode, errmsg, err := service.MsgSetRead(context.Background(), "", []int{1})
	if err != nil {
		t.Fatalf("msg setread: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestMsgSetReadSkipsInvalidIDs(t *testing.T) {
	store := &trackingMsgStore{fakeUserStore: fakeUserStore{user: map[string]interface{}{"uid": "5"}}}
	service := NewService(store, "https://res.example.test")

	retcode, errmsg, err := service.MsgSetRead(context.Background(), "token", []int{9, 0, -1, 12})
	if err != nil {
		t.Fatalf("msg setread: %v", err)
	}
	if retcode != 0 || errmsg != "操作成功" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if !equalInts(store.readCIDs, []int{9, 12}) {
		t.Fatalf("read cids = %#v", store.readCIDs)
	}
}

func TestMsgCleanReadAndDelete(t *testing.T) {
	store := &trackingMsgStore{fakeUserStore: fakeUserStore{user: map[string]interface{}{"uid": "5"}}}
	service := NewService(store, "https://res.example.test")

	retcode, errmsg, err := service.MsgCleanRead(context.Background(), "token")
	if err != nil || retcode != 0 || errmsg != "操作成功" {
		t.Fatalf("cleanread response %d %q err %v", retcode, errmsg, err)
	}
	if store.cleanUID != 5 {
		t.Fatalf("clean uid = %d", store.cleanUID)
	}
	retcode, errmsg, err = service.MsgDelete(context.Background(), "token", []int{3, 4})
	if err != nil || retcode != 0 || errmsg != "操作成功" {
		t.Fatalf("delete response %d %q err %v", retcode, errmsg, err)
	}
	if store.deleteID != 5 || !equalInts(store.deleted, []int{3, 4}) {
		t.Fatalf("delete uid/cids = %d %#v", store.deleteID, store.deleted)
	}
}

func equalInts(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
