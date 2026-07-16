package ucp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"xj_comp/internal/domain"
)

type fakeUserStore struct {
	user                map[string]interface{}
	guest               map[string]interface{}
	account             map[string]interface{}
	quota               map[string]interface{}
	missingGuest        bool
	coinLogTypes        []int
	coinLogOrderBy      string
	countCoinLogSince   *int
	countCoinLogTypes   []int
	countCoinLogResult  int
	sumCoinLogByType    map[int]int
	awardedCoins        *map[string]interface{}
	awardCoinsErr       error
	signedGuest         *map[string]interface{}
	signGuestErr        error
	upgradedVIP         *map[string]interface{}
	upgradeVIPErr       error
	boughtBeans         *map[string]interface{}
	buyBeansErr         error
	exchangedCoins      *map[string]interface{}
	exchangeCoinsErr    error
	coinBonusStats      map[string]interface{}
	feedbackRow         map[string]interface{}
	feedbackSinceCount  int
	createdFeedback     *domain.FeedbackCreateInput
	paymentRow          map[string]interface{}
	paymentStatusCounts map[string]int
	attachRows          []map[string]interface{}
	posters             []map[string]interface{}
	nicknames           []map[string]interface{}
	taskboxes           []map[string]interface{}
	taskboxRow          map[string]interface{}
	taskboxLog          map[string]interface{}
	taskboxLogs         []map[string]interface{}
	openedTaskbox       *map[string]interface{}
	openTaskboxMessage  string
	openTaskboxErr      error
	bankcards           []map[string]interface{}
	bankcardRow         map[string]interface{}
	createdBankcard     map[string]interface{}
	updatedBankcard     map[string]interface{}
	deletedBankcard     map[string]interface{}
	defaultBankcard     map[string]interface{}
	banks               []map[string]interface{}
	settings            map[string]map[string]interface{}
	calldata            map[string]map[string]interface{}
	packages            []map[string]interface{}
	packageRow          map[string]interface{}
	payments            []map[string]interface{}
	vodPlayCount        *int
	withdraws           []map[string]interface{}
	withdrawTotal       int
	withdrawSinceCount  *int
	exrate              *int
	vodOrders           []map[string]interface{}
	vodOrderRow         map[string]interface{}
	vodSupports         []map[string]interface{}
	latestVODIssue      map[string]interface{}
	maxVODSupport       map[string]interface{}
	myVODSupportCoins   int
	userByID            map[string]interface{}
	updatedProfile      *map[string]interface{}
	changedPassword     *map[string]interface{}
	changePasswordErr   error
	verifiedEmail       *map[string]interface{}
	boundMobi           *map[string]interface{}
	userByEmail         map[string]interface{}
	userByMobi          map[string]interface{}
	keylimitCounts      map[string]int
	keylimitTotalCounts map[string]int
	keylimitData        map[string]string
	setKeylimit         *map[string]interface{}
	botByID             map[string]interface{}
	sentMessage         map[string]interface{}
}

type fakeQRCodeRenderer struct {
	content string
	body    []byte
	err     error
}

func (r *fakeQRCodeRenderer) PNG(content string) ([]byte, error) {
	r.content = content
	if r.err != nil {
		return nil, r.err
	}
	if r.body != nil {
		return r.body, nil
	}
	return []byte("png-body"), nil
}

func (s fakeUserStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s fakeUserStore) Groups(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"gid": "0", "gname": "游客", "gicon": "", "minup": "0", "weight": "0", "scope": "0", "perms": `{"max.vod.play.daynum":"10","max.vod.down.daynum":"8","max.minivod.play.daynum":"20","max.minivod.down.daynum":"18","max.goldcoin.sign.num":"3"}`},
		{"gid": "4", "gname": "普通会员", "gicon": "V4", "minup": "100", "weight": "4", "scope": "0", "perms": `{"max.vod.play.daynum":"40","max.vod.down.daynum":"30","max.comment.post.daynum":"4","max.minivod.play.daynum":"150","max.minivod.down.daynum":"150","max.goldcoin.sign.num":"3","max.goldcoin.email.num":"2","max.goldcoin.mobi.num":"4","max.goldcoin.qrcode.num":"5","max.goldcoin.share.num":"1","max.goldcoin.share.limit":"3","max.goldcoin.comment.num":"2","max.goldcoin.comment.limit":"6","max.goldcoin.favorite.num":"3","max.goldcoin.favorite.limit":"9","max.goldcoin.play10.num":"4","max.goldcoin.play10.limit":"12","max.goldcoin.saveqrcode.num":"5","max.goldcoin.adviewclick.num":"6","max.goldcoin.minivod.down.coinnum":"7","max.goldcoin.minivod.down.limit":"21"}`},
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

func (s fakeUserStore) Nicknames(context.Context) ([]map[string]interface{}, error) {
	return s.nicknames, nil
}

func (s fakeUserStore) Taskboxes(context.Context) ([]map[string]interface{}, error) {
	return s.taskboxes, nil
}

func (s fakeUserStore) TaskboxByID(context.Context, int) (map[string]interface{}, error) {
	if s.taskboxRow != nil {
		return s.taskboxRow, nil
	}
	return map[string]interface{}{}, nil
}

func (s fakeUserStore) TaskboxLog(context.Context, int, int, int) (map[string]interface{}, error) {
	if s.taskboxLog != nil {
		return s.taskboxLog, nil
	}
	return map[string]interface{}{}, nil
}

func (s fakeUserStore) OpenTaskbox(_ context.Context, uid int, task map[string]interface{}, dayKey int, addCoin int, now int64, duplicateMessage string) (string, error) {
	if s.openedTaskbox != nil {
		*s.openedTaskbox = map[string]interface{}{
			"uid":              uid,
			"taskid":           task["taskid"],
			"daykey":           dayKey,
			"addcoin":          addCoin,
			"now":              now,
			"duplicateMessage": duplicateMessage,
		}
	}
	return s.openTaskboxMessage, s.openTaskboxErr
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

func (s fakeUserStore) CountPaymentsByStatusSince(_ context.Context, _ int, isPaid int, since int64) (int, error) {
	if s.paymentStatusCounts != nil {
		if count, ok := s.paymentStatusCounts[fmt.Sprintf("%d:%d", isPaid, since)]; ok {
			return count, nil
		}
		if count, ok := s.paymentStatusCounts[fmt.Sprintf("%d:*", isPaid)]; ok {
			return count, nil
		}
	}
	return 0, nil
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
	if s.account != nil {
		return s.account, nil
	}
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
	if s.quota != nil {
		return s.quota, nil
	}
	return map[string]interface{}{"uid": "5", "goldcoin": "625"}, nil
}

func (s fakeUserStore) Goldbean(context.Context, int) (map[string]interface{}, error) {
	return map[string]interface{}{"uid": "5", "gold_bean": "3815"}, nil
}

func (s fakeUserStore) CountVODPlayLogsSince(context.Context, int, int64) (int, error) {
	if s.vodPlayCount != nil {
		return *s.vodPlayCount, nil
	}
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
	if s.countCoinLogSince != nil {
		return *s.countCoinLogSince, nil
	}
	return 1, nil
}

func (s fakeUserStore) SumCoinLogsSinceByType(_ context.Context, _ int, coinType int, _ int64) (int, error) {
	if s.sumCoinLogByType != nil {
		return s.sumCoinLogByType[coinType], nil
	}
	return coinType, nil
}

func (s fakeUserStore) AwardCoins(_ context.Context, uid int, coinType int, addCoin int, now int64, remark string) error {
	if s.awardedCoins != nil {
		*s.awardedCoins = map[string]interface{}{
			"uid":      uid,
			"cointype": coinType,
			"addcoin":  addCoin,
			"now":      now,
			"remark":   remark,
		}
	}
	return s.awardCoinsErr
}

func (s fakeUserStore) SignGuest(_ context.Context, sid string, addCoin int, now int64) error {
	if s.signedGuest != nil {
		*s.signedGuest = map[string]interface{}{
			"sid":     sid,
			"addcoin": addCoin,
			"now":     now,
		}
	}
	return s.signGuestErr
}

func (s fakeUserStore) UpgradeVIP(_ context.Context, uid int, deductCoin int, vipGID int, expiry int64, now int64) error {
	if s.upgradedVIP != nil {
		*s.upgradedVIP = map[string]interface{}{
			"uid":        uid,
			"deductCoin": deductCoin,
			"vipGID":     vipGID,
			"expiry":     expiry,
			"now":        now,
		}
	}
	return s.upgradeVIPErr
}

func (s fakeUserStore) BuyBeansWithCoins(_ context.Context, uid int, deductCoin int, addBeans int, now int64) error {
	if s.boughtBeans != nil {
		*s.boughtBeans = map[string]interface{}{
			"uid":        uid,
			"deductCoin": deductCoin,
			"addBeans":   addBeans,
			"now":        now,
		}
	}
	return s.buyBeansErr
}

func (s fakeUserStore) ExchangeCoinsAndBalance(_ context.Context, uid int, extype int, coinnum int, amount int, now int64) error {
	if s.exchangedCoins != nil {
		*s.exchangedCoins = map[string]interface{}{
			"uid":     uid,
			"extype":  extype,
			"coinnum": coinnum,
			"amount":  amount,
			"now":     now,
		}
	}
	return s.exchangeCoinsErr
}

func (s fakeUserStore) CountVODCommentsSince(context.Context, int, int64, bool) (int, error) {
	return 3, nil
}

func (s fakeUserStore) CountVODFavoritesSince(context.Context, int, int64) (int, error) {
	return 4, nil
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

func (s fakeUserStore) CountFeedbacksSince(context.Context, int, int64) (int, error) {
	return s.feedbackSinceCount, nil
}

func (s fakeUserStore) CreateFeedback(_ context.Context, input domain.FeedbackCreateInput) (int, error) {
	if s.createdFeedback != nil {
		*s.createdFeedback = input
	}
	return 123, nil
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
	if s.userByID != nil {
		return s.userByID, nil
	}
	if s.user != nil {
		return s.user, nil
	}
	return map[string]interface{}{"uid": "7", "username": "peer"}, nil
}

func (s fakeUserStore) UpdateUserProfile(_ context.Context, uid int, gender int, nickname *string) error {
	if s.updatedProfile != nil {
		row := map[string]interface{}{"uid": uid, "gender": gender}
		if nickname != nil {
			row["nickname"] = *nickname
		}
		*s.updatedProfile = row
	}
	return nil
}

func (s fakeUserStore) ChangePasswordAndLogin(_ context.Context, uid int, passwordHash string, salt string, sid string, token string, now int64) (map[string]interface{}, error) {
	if s.changedPassword != nil {
		*s.changedPassword = map[string]interface{}{
			"uid":      uid,
			"password": passwordHash,
			"salt":     salt,
			"sid":      sid,
			"token":    token,
			"now":      now,
		}
	}
	if s.changePasswordErr != nil {
		return nil, s.changePasswordErr
	}
	row := map[string]interface{}{}
	for k, v := range s.user {
		row[k] = v
	}
	for k, v := range s.userByID {
		row[k] = v
	}
	row["uid"] = fmt.Sprint(uid)
	row["password"] = passwordHash
	row["salt"] = salt
	return row, nil
}

func (s fakeUserStore) VerifyEmail(_ context.Context, uid int, email string, key string) error {
	if s.verifiedEmail != nil {
		*s.verifiedEmail = map[string]interface{}{
			"uid":   uid,
			"email": email,
			"key":   key,
		}
	}
	return nil
}

func (s fakeUserStore) BindMobi(_ context.Context, uid int, mobi string) error {
	if s.boundMobi != nil {
		*s.boundMobi = map[string]interface{}{
			"uid":  uid,
			"mobi": mobi,
		}
	}
	return nil
}

func (s fakeUserStore) UserByEmail(context.Context, string) (map[string]interface{}, error) {
	if s.userByEmail != nil {
		return s.userByEmail, nil
	}
	return map[string]interface{}{}, nil
}

func (s fakeUserStore) UserByMobi(context.Context, string) (map[string]interface{}, error) {
	if s.userByMobi != nil {
		return s.userByMobi, nil
	}
	return map[string]interface{}{}, nil
}

func (s fakeUserStore) BotByID(context.Context, int) (map[string]interface{}, error) {
	if s.botByID != nil {
		return s.botByID, nil
	}
	return map[string]interface{}{}, nil
}

func (s fakeUserStore) Bankcards(context.Context, int) ([]map[string]interface{}, error) {
	return s.bankcards, nil
}

func (s fakeUserStore) Banks(context.Context) ([]map[string]interface{}, error) {
	return s.banks, nil
}

func (s fakeUserStore) BankcardByID(context.Context, int, int) (map[string]interface{}, error) {
	if s.bankcardRow != nil {
		return s.bankcardRow, nil
	}
	return map[string]interface{}{}, nil
}

func (s fakeUserStore) CreateBankcard(_ context.Context, uid int, name string, bankname string, cardnum string, isdef int, cardType int) (int, error) {
	if s.createdBankcard != nil {
		s.createdBankcard["uid"] = uid
		s.createdBankcard["name"] = name
		s.createdBankcard["bankname"] = bankname
		s.createdBankcard["cardnum"] = cardnum
		s.createdBankcard["isdef"] = isdef
		s.createdBankcard["type"] = cardType
	}
	return 88, nil
}

func (s fakeUserStore) UpdateBankcard(_ context.Context, uid int, cardID int, name string, bankname string, cardnum string, isdef int, cardType int) (int, error) {
	if s.updatedBankcard != nil {
		s.updatedBankcard["uid"] = uid
		s.updatedBankcard["cardid"] = cardID
		s.updatedBankcard["name"] = name
		s.updatedBankcard["bankname"] = bankname
		s.updatedBankcard["cardnum"] = cardnum
		s.updatedBankcard["isdef"] = isdef
		s.updatedBankcard["type"] = cardType
	}
	return 1, nil
}

func (s fakeUserStore) DeleteBankcard(_ context.Context, uid int, cardID int) (int, error) {
	if s.deletedBankcard != nil {
		s.deletedBankcard["uid"] = uid
		s.deletedBankcard["cardid"] = cardID
	}
	return 1, nil
}

func (s fakeUserStore) SetDefaultBankcard(_ context.Context, uid int, cardID int) error {
	if s.defaultBankcard != nil {
		s.defaultBankcard["uid"] = uid
		s.defaultBankcard["cardid"] = cardID
	}
	return nil
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

func (s fakeUserStore) SendMessage(_ context.Context, senderID int, receiverID int, content string, cid int, now int64) (int, error) {
	if s.sentMessage != nil {
		s.sentMessage["senderID"] = senderID
		s.sentMessage["receiverID"] = receiverID
		s.sentMessage["content"] = content
		s.sentMessage["cid"] = cid
		s.sentMessage["now"] = now
	}
	return 99, nil
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

func (s fakeUserStore) CountWithdraws(context.Context, int) (int, error) {
	if s.withdraws != nil {
		return len(s.withdraws), nil
	}
	return 1, nil
}

func (s fakeUserStore) CountWithdrawsSince(context.Context, int, int64) (int, error) {
	if s.withdrawSinceCount != nil {
		return *s.withdrawSinceCount, nil
	}
	return 0, nil
}

func (s fakeUserStore) Withdraws(context.Context, int, int, int) ([]map[string]interface{}, error) {
	if s.withdraws != nil {
		return s.withdraws, nil
	}
	return []map[string]interface{}{{
		"wdid": "3", "uid": "5", "username": "tester", "wdtype": "0", "withdraw_amount": "12345", "remit_amount": "12000",
		"createtime": "1770000000", "lastupdate": "1770000060", "name": "张三", "cardnum": "abc", "bankname": "支付宝",
		"errmsg": "", "wdstatus": "1", "checkstatus": "0",
	}}, nil
}

func (s fakeUserStore) SumWithdrawAmount(context.Context, int) (int, error) {
	if s.withdrawTotal > 0 {
		return s.withdrawTotal, nil
	}
	return 12345, nil
}

func (s fakeUserStore) SettingExRate(context.Context) (int, error) {
	if s.exrate != nil {
		return *s.exrate, nil
	}
	return 10, nil
}

func (s fakeUserStore) SettingByUUID(_ context.Context, uuid string) (map[string]interface{}, error) {
	if s.settings != nil {
		if row, ok := s.settings[uuid]; ok {
			return row, nil
		}
	}
	return map[string]interface{}{}, nil
}

func (s fakeUserStore) CalldataByUUID(_ context.Context, uuid string) (map[string]interface{}, error) {
	if s.calldata != nil {
		if row, ok := s.calldata[uuid]; ok {
			return row, nil
		}
	}
	return map[string]interface{}{}, nil
}

func (s fakeUserStore) KeylimitCountSince(_ context.Context, key string, since int64) (int, error) {
	if since == 0 && s.keylimitTotalCounts != nil {
		return s.keylimitTotalCounts[key], nil
	}
	if s.keylimitCounts != nil {
		return s.keylimitCounts[key], nil
	}
	return 0, nil
}

func (s fakeUserStore) KeylimitDataSince(_ context.Context, key string, _ int64) (string, error) {
	if s.keylimitData != nil {
		return s.keylimitData[key], nil
	}
	return "", nil
}

func (s fakeUserStore) SetKeylimit(_ context.Context, key string, keynum int, keydata string, now int64) error {
	if s.setKeylimit != nil {
		*s.setKeylimit = map[string]interface{}{
			"key":     key,
			"keynum":  keynum,
			"keydata": keydata,
			"now":     now,
		}
	}
	return nil
}

func (s fakeUserStore) PackageRows(context.Context, string) ([]map[string]interface{}, error) {
	return s.packages, nil
}

func (s fakeUserStore) PackageByID(context.Context, string, int) (map[string]interface{}, error) {
	return s.packageRow, nil
}

func (s fakeUserStore) PaymentChannels(context.Context, bool) ([]map[string]interface{}, error) {
	return s.payments, nil
}

func (s fakeUserStore) CountVODOrders(context.Context, int, *int) (int, error) {
	return len(s.vodOrders), nil
}

func (s fakeUserStore) VODOrders(context.Context, int, *int, int, int, string) ([]map[string]interface{}, error) {
	return s.vodOrders, nil
}

func (s fakeUserStore) VODOrderByID(context.Context, int) (map[string]interface{}, error) {
	return s.vodOrderRow, nil
}

func (s fakeUserStore) LatestVODIssue(context.Context) (map[string]interface{}, error) {
	if s.latestVODIssue != nil {
		return s.latestVODIssue, nil
	}
	return map[string]interface{}{}, nil
}

func (s fakeUserStore) CountVODOrdersByCreateTime(context.Context, int64, int64) (int, error) {
	return len(s.vodOrders), nil
}

func (s fakeUserStore) VODOrdersByCreateTime(context.Context, int64, int64, int, int) ([]map[string]interface{}, error) {
	return s.vodOrders, nil
}

func (s fakeUserStore) SumVODOrderCoins(_ context.Context, _ int, status int) (int, error) {
	if status == 1 {
		return 100, nil
	}
	return 30, nil
}

func (s fakeUserStore) CountVODSupports(context.Context, int) (int, error) {
	return len(s.vodSupports), nil
}

func (s fakeUserStore) VODSupports(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return s.vodSupports, nil
}

func (s fakeUserStore) MaxVODSupport(context.Context, int) (map[string]interface{}, error) {
	if s.maxVODSupport != nil {
		return s.maxVODSupport, nil
	}
	return map[string]interface{}{}, nil
}

func (s fakeUserStore) MyVODSupportCoins(context.Context, int, int) (int, error) {
	return s.myVODSupportCoins, nil
}

func (s fakeUserStore) SumVODSupportCoins(_ context.Context, _ int, onlyFrozen bool) (int, error) {
	if onlyFrozen {
		return 7, nil
	}
	return 11, nil
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

func TestTaskQRLinkRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.TaskQRLink(context.Background(), "", "")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestTaskInviteMatchesEmptyPHPAction(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	retcode, errmsg, err := service.TaskInvite(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	retcode, errmsg, err = service.TaskInvite(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestHighRiskActionEdgeRequiresLoginAndBlocksSuccess(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	retcode, errmsg, err := service.HighRiskActionEdge(context.Background(), "", "成功分支暂未迁移")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	retcode, errmsg, err = service.HighRiskActionEdge(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", "成功分支暂未迁移")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "成功分支暂未迁移" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestTaskRewardEdgePrechecks(t *testing.T) {
	zero := 0
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "0", "sid": "guest"}, missingGuest: true}, "https://res.example.test")
	retcode, errmsg, err := service.TaskSignEdge(context.Background(), "token")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "请登录后操作，客户端游客请先携带信息" {
		t.Fatalf("task sign guest retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{user: map[string]interface{}{"uid": "5", "uniqkey": "12345"}, countCoinLogResult: 1}, "https://res.example.test")
	retcode, errmsg, err = service.TaskInviteCodeInputEdge(context.Background(), "token", "BAD")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "您今天已经保存过了" {
		t.Fatalf("invitecode saved retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{user: map[string]interface{}{"uid": "5", "uniqkey": "12345"}, countCoinLogSince: &zero}, "https://res.example.test")
	retcode, errmsg, err = service.TaskInviteCodeInputEdge(context.Background(), "token", "BAD")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "邀请码不正确" {
		t.Fatalf("invitecode bad retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}, countCoinLogResult: 1}, "https://res.example.test")
	retcode, errmsg, err = service.TaskAdviewClickEdge(context.Background(), "token")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "您今天已经送过了" {
		t.Fatalf("adview retcode=%d errmsg=%q", retcode, errmsg)
	}

	one := 1
	service = NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}, countCoinLogSince: &one}, "https://res.example.test")
	retcode, errmsg, err = service.TaskQRCodeSaveEdge(context.Background(), "token")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "您今天已经保存过了" {
		t.Fatalf("qrcode save saved retcode=%d errmsg=%q", retcode, errmsg)
	}

	awarded := map[string]interface{}{}
	service = NewService(fakeUserStore{
		user:              map[string]interface{}{"uid": "5", "gid": "4"},
		countCoinLogSince: &zero,
		awardedCoins:      &awarded,
	}, "https://res.example.test")
	data, retcode, errmsg, err := service.TaskQRCodeSave(context.Background(), "token")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "保存二维码已送金币: 5" || atoi(data["taskdone"]) != 5 {
		t.Fatalf("qrcode save success retcode=%d errmsg=%q data=%v", retcode, errmsg, data)
	}
	if atoi(awarded["uid"]) != 5 || atoi(awarded["cointype"]) != coinTypeSaveQRCode || atoi(awarded["addcoin"]) != 5 {
		t.Fatalf("qrcode save award mismatch: %v", awarded)
	}
}

func TestTaskRewardSuccessBranches(t *testing.T) {
	zero := 0
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

	signedGuest := map[string]interface{}{}
	service := NewService(fakeUserStore{
		user:              map[string]interface{}{"uid": "0", "sid": "guest", "perms": `{"max.goldcoin.sign.num":"3"}`},
		guest:             map[string]interface{}{"sid": "guest", "goldcoin": "9", "signtime": "0"},
		signedGuest:       &signedGuest,
		countCoinLogSince: &zero,
	}, "https://res.example.test")
	service.now = func() time.Time { return now }
	data, retcode, errmsg, err := service.TaskSign(context.Background(), "token")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" || data != nil {
		t.Fatalf("guest sign retcode=%d errmsg=%q data=%v", retcode, errmsg, data)
	}
	if signedGuest["sid"] != "guest" || atoi(signedGuest["addcoin"]) != 3 {
		t.Fatalf("guest sign mismatch: %v", signedGuest)
	}

	awarded := map[string]interface{}{}
	service = NewService(fakeUserStore{
		user: map[string]interface{}{
			"uid":     "5",
			"gid":     "4",
			"email":   "me@example.com",
			"mobi":    "86.13800138000",
			"uniqkey": "12345",
		},
		countCoinLogSince: &zero,
		awardedCoins:      &awarded,
	}, "https://res.example.test")
	service.now = func() time.Time { return now }
	data, retcode, errmsg, err = service.TaskSign(context.Background(), "token")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" || atoi(data["taskdone"]) != 14 {
		t.Fatalf("user sign retcode=%d errmsg=%q data=%v", retcode, errmsg, data)
	}
	if atoi(awarded["cointype"]) != coinTypeSign || atoi(awarded["addcoin"]) != 14 {
		t.Fatalf("user sign award mismatch: %v", awarded)
	}

	awarded = map[string]interface{}{}
	service = NewService(fakeUserStore{
		user:              map[string]interface{}{"uid": "5", "gid": "4", "uniqkey": "12345"},
		countCoinLogSince: &zero,
		awardedCoins:      &awarded,
	}, "https://res.example.test")
	data, retcode, errmsg, err = service.TaskInviteCodeInput(context.Background(), "token", "9IX")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "保存二维码已送金币: 5" || atoi(data["taskdone"]) != 5 {
		t.Fatalf("invite code retcode=%d errmsg=%q data=%v", retcode, errmsg, data)
	}
	if atoi(awarded["cointype"]) != coinTypeSaveQRCode || atoi(awarded["addcoin"]) != 5 {
		t.Fatalf("invite code award mismatch: %v", awarded)
	}

	awarded = map[string]interface{}{}
	service = NewService(fakeUserStore{
		user:              map[string]interface{}{"uid": "5", "gid": "4"},
		countCoinLogSince: &zero,
		awardedCoins:      &awarded,
	}, "https://res.example.test")
	data, retcode, errmsg, err = service.TaskAdviewClick(context.Background(), "token")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "点击广告已送金币: 6" || atoi(data["taskdone"]) != 6 {
		t.Fatalf("adview retcode=%d errmsg=%q data=%v", retcode, errmsg, data)
	}
	if atoi(awarded["cointype"]) != coinTypeAdViewClick || atoi(awarded["addcoin"]) != 6 {
		t.Fatalf("adview award mismatch: %v", awarded)
	}
}

func TestTaskboxOpenEdgePrechecks(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	retcode, errmsg, err := service.TaskboxOpenEdge(context.Background(), "token", 0)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "任务不存在或已停用" {
		t.Fatalf("taskbox missing retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}, taskboxRow: map[string]interface{}{"taskid": "1", "showtype": "0", "mincoin": "0", "maxcoin": "0"}}, "https://res.example.test")
	retcode, errmsg, err = service.TaskboxOpenEdge(context.Background(), "token", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "宝箱赠送金币为0" {
		t.Fatalf("taskbox zero retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestTaskboxOpenEdgeMysteryTaskFailures(t *testing.T) {
	cases := []struct {
		name    string
		taskRow map[string]interface{}
		logRow  map[string]interface{}
		now     time.Time
		errmsg  string
	}{
		{
			name:    "daily not started",
			taskRow: map[string]interface{}{"taskid": "1022", "showtype": "0", "mincoin": "1", "maxcoin": "2"},
			now:     time.Date(2026, 7, 15, 21, 59, 59, 0, time.UTC),
			errmsg:  "每日神秘宝箱领取时间未开始",
		},
		{
			name:    "daily ended",
			taskRow: map[string]interface{}{"taskid": "1022", "showtype": "0", "mincoin": "1", "maxcoin": "2"},
			now:     time.Date(2026, 7, 15, 22, 5, 0, 0, time.UTC),
			errmsg:  "每日神秘宝箱领取时间已结束",
		},
		{
			name:    "daily claimed",
			taskRow: map[string]interface{}{"taskid": "1022", "showtype": "0", "mincoin": "1", "maxcoin": "2"},
			logRow:  map[string]interface{}{"logid": "1"},
			now:     time.Date(2026, 7, 15, 22, 1, 0, 0, time.UTC),
			errmsg:  "每日神秘宝箱已领过了",
		},
		{
			name:    "weekly not saturday",
			taskRow: map[string]interface{}{"taskid": "1622", "showtype": "0", "mincoin": "1", "maxcoin": "2"},
			now:     time.Date(2026, 7, 15, 22, 1, 0, 0, time.UTC),
			errmsg:  "每周神秘宝箱周六晚开始",
		},
		{
			name:    "weekly not started",
			taskRow: map[string]interface{}{"taskid": "1622", "showtype": "0", "mincoin": "1", "maxcoin": "2"},
			now:     time.Date(2026, 7, 18, 21, 59, 59, 0, time.UTC),
			errmsg:  "每周神秘宝箱领取时间未开始",
		},
		{
			name:    "weekly ended",
			taskRow: map[string]interface{}{"taskid": "1622", "showtype": "0", "mincoin": "1", "maxcoin": "2"},
			now:     time.Date(2026, 7, 18, 22, 5, 0, 0, time.UTC),
			errmsg:  "每周神秘宝箱领取时间已结束",
		},
		{
			name:    "weekly claimed",
			taskRow: map[string]interface{}{"taskid": "1622", "showtype": "0", "mincoin": "1", "maxcoin": "2"},
			logRow:  map[string]interface{}{"logid": "1"},
			now:     time.Date(2026, 7, 18, 22, 1, 0, 0, time.UTC),
			errmsg:  "每周神秘宝箱已领过了",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(fakeUserStore{
				user:       map[string]interface{}{"uid": "5"},
				taskboxRow: tt.taskRow,
				taskboxLog: tt.logRow,
			}, "https://res.example.test")
			service.now = func() time.Time { return tt.now }

			retcode, errmsg, err := service.TaskboxOpenEdge(context.Background(), "token", atoi(tt.taskRow["taskid"]))
			if err != nil {
				t.Fatal(err)
			}
			if retcode != -1 || errmsg != tt.errmsg {
				t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
			}
		})
	}
}

func TestTaskboxOpenEdgePromotionTaskFailures(t *testing.T) {
	service := NewService(fakeUserStore{
		user:       map[string]interface{}{"uid": "5"},
		taskboxRow: map[string]interface{}{"taskid": "3", "showtype": "0", "mincoin": "1", "maxcoin": "2"},
		taskboxLog: map[string]interface{}{"logid": "1"},
		userByID:   map[string]interface{}{"uid": "5", "recommend_total": "3"},
	}, "https://res.example.test")
	retcode, errmsg, err := service.TaskboxOpenEdge(context.Background(), "token", 3)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "推广任务宝箱已领过了" {
		t.Fatalf("claimed retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:       map[string]interface{}{"uid": "5"},
		taskboxRow: map[string]interface{}{"taskid": "3", "showtype": "0", "mincoin": "1", "maxcoin": "2"},
		userByID:   map[string]interface{}{"uid": "5", "recommend_total": "2"},
	}, "https://res.example.test")
	retcode, errmsg, err = service.TaskboxOpenEdge(context.Background(), "token", 3)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "推广人数未达标，继续加油哦" {
		t.Fatalf("insufficient retcode=%d errmsg=%q", retcode, errmsg)
	}

	openedTaskbox := map[string]interface{}{}
	service = NewService(fakeUserStore{
		user:          map[string]interface{}{"uid": "5"},
		taskboxRow:    map[string]interface{}{"taskid": "3", "showtype": "0", "mincoin": "2", "maxcoin": "2"},
		userByID:      map[string]interface{}{"uid": "5", "recommend_total": "3"},
		openedTaskbox: &openedTaskbox,
	}, "https://res.example.test")
	data, retcode, errmsg, err := service.TaskboxOpen(context.Background(), "token", 3)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "宝箱成功开启" || atoi(data["taskdone"]) != 2 {
		t.Fatalf("success retcode=%d errmsg=%q data=%v", retcode, errmsg, data)
	}
	if atoi(openedTaskbox["uid"]) != 5 || atoi(openedTaskbox["taskid"]) != 3 || atoi(openedTaskbox["daykey"]) != 0 || atoi(openedTaskbox["addcoin"]) != 2 {
		t.Fatalf("opened taskbox mismatch: %v", openedTaskbox)
	}
	if openedTaskbox["duplicateMessage"] != "推广任务宝箱已领过了" {
		t.Fatalf("duplicate message mismatch: %v", openedTaskbox)
	}
}

func TestUserContactEdges(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")

	retcode, errmsg, err := service.UserCheckEmailEdge(context.Background(), "token", "bad")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "请输入正确的邮箱地址" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.UserVerifyEmailEdge(context.Background(), "token", "me@example.com", "123456")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "验证码不存在或已失效" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.UserBindMobiEdge(context.Background(), "token", "", "13800138000", "123456")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "手机验证码不正确" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.UserProfileEdge(context.Background(), "token", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "资料设置成功" {
		t.Fatalf("profile retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.UserProfileEdge(context.Background(), "token", 1, "bad!")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "昵称2-8个汉字，英文6-16个字符" {
		t.Fatalf("profile nickname length retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.UserProfileEdge(context.Background(), "token", 1, "abcdef!")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "昵称只允许中英文、数字及下划线组成" {
		t.Fatalf("profile nickname charset retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.UserProfileEdge(context.Background(), "token", 2, "abcdef")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "如需修改昵称，请联系客服修改" {
		t.Fatalf("profile nickname whitelist retcode=%d errmsg=%q", retcode, errmsg)
	}

	updatedProfile := map[string]interface{}{}
	service = NewService(fakeUserStore{
		user:           map[string]interface{}{"uid": "5", "nickname": "oldname"},
		nicknames:      []map[string]interface{}{{"name": "abcdef", "gender": "2"}},
		updatedProfile: &updatedProfile,
	}, "https://res.example.test")
	retcode, errmsg, err = service.UserProfileEdge(context.Background(), "token", 2, "abcdef")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "资料设置成功" {
		t.Fatalf("profile whitelisted success retcode=%d errmsg=%q", retcode, errmsg)
	}
	if updatedProfile["uid"] != 5 || updatedProfile["gender"] != 2 || updatedProfile["nickname"] != "abcdef" {
		t.Fatalf("updated profile=%#v", updatedProfile)
	}

	service = NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	retcode, errmsg, err = service.UserPasswdEdge(context.Background(), "token", "", "123", "123")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "密码6-16位" {
		t.Fatalf("passwd length retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.UserPasswdEdge(context.Background(), "token", "", "123456", "654321")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "两次输入密码不一致" {
		t.Fatalf("passwd confirm retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user: map[string]interface{}{
			"uid":      "5",
			"password": phpPassword("oldpasssalt1234"),
			"salt":     "salt1234",
		},
	}, "https://res.example.test")
	retcode, errmsg, err = service.UserPasswdEdge(context.Background(), "token", "wrong", "123", "123")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "原密码不正确" {
		t.Fatalf("passwd old password retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.UserPasswdEdge(context.Background(), "token", "oldpass", "123", "123")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "密码6-16位" {
		t.Fatalf("passwd valid old password retcode=%d errmsg=%q", retcode, errmsg)
	}

	changedPassword := map[string]interface{}{}
	service = NewService(fakeUserStore{
		user: map[string]interface{}{"uid": "5"},
		userByID: map[string]interface{}{
			"uid":      "5",
			"username": "tester",
			"password": phpPassword("oldpasssalt1234"),
			"salt":     "salt1234",
			"gid":      "4",
			"sysgid":   "0",
			"regtime":  "1700000000",
			"gender":   "1",
			"avatar":   "",
		},
		changedPassword: &changedPassword,
	}, "https://res.example.test")
	data, retcode, errmsg, err := service.UserPasswd(context.Background(), "token", "oldpass", "newpass1", "newpass1")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "密码修改成功" {
		t.Fatalf("passwd success retcode=%d errmsg=%q", retcode, errmsg)
	}
	if data["xxx_api_auth"] == "" {
		t.Fatalf("missing api auth data=%#v", data)
	}
	if changedPassword["uid"] != 5 || changedPassword["password"] == "" || changedPassword["salt"] == "" || changedPassword["sid"] == "" || changedPassword["token"] == "" {
		t.Fatalf("changed password=%#v", changedPassword)
	}
	if changedPassword["password"] == phpPassword("oldpasssalt1234") {
		t.Fatalf("password hash was not changed: %#v", changedPassword)
	}
}

func TestUserBindMobiEdgePrechecks(t *testing.T) {
	ctx := context.Background()

	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5", "mobi": "86.13800138000"}}, "https://res.example.test")
	retcode, errmsg, err := service.UserBindMobiEdge(ctx, "token", "", "13800138000", "123456")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "您已绑定手机" {
		t.Fatalf("bound retcode=%d errmsg=%q", retcode, errmsg)
	}

	boundMobi := map[string]interface{}{}
	service = NewService(fakeUserStore{
		user:           map[string]interface{}{"uid": "5", "mobi": "~old"},
		keylimitCounts: map[string]int{"sms.86.13800138000.123456": 1},
		boundMobi:      &boundMobi,
	}, "https://res.example.test")
	retcode, errmsg, err = service.UserBindMobiEdge(ctx, "token", "", "13800138000", "123456")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "手机验证已确认，绑定成功" {
		t.Fatalf("default prefix retcode=%d errmsg=%q", retcode, errmsg)
	}
	if boundMobi["uid"] != 5 || boundMobi["mobi"] != "86.13800138000" {
		t.Fatalf("bound mobi=%#v", boundMobi)
	}

	boundMobi = map[string]interface{}{}
	service = NewService(fakeUserStore{
		user:           map[string]interface{}{"uid": "5", "mobi": "~old"},
		keylimitCounts: map[string]int{"sms.1.5550001.999999": 1},
		userByMobi:     map[string]interface{}{"uid": "9", "mobi": "1.5550001"},
		boundMobi:      &boundMobi,
	}, "https://res.example.test")
	retcode, errmsg, err = service.UserBindMobiEdge(ctx, "token", " 1 ", " 5550001 ", " 999999 ")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "手机验证已确认，绑定成功" {
		t.Fatalf("existing mobi placeholder retcode=%d errmsg=%q", retcode, errmsg)
	}
	if boundMobi["uid"] != 5 || boundMobi["mobi"] != "1.5550001" {
		t.Fatalf("bound existing mobi=%#v", boundMobi)
	}
}

func TestUserEmailEdges(t *testing.T) {
	ctx := context.Background()
	today := time.Now().Format("20060102")

	cases := []struct {
		name   string
		store  fakeUserStore
		call   func(*Service) (int, string, error)
		errmsg string
	}{
		{
			name:  "checkemail rate limited",
			store: fakeUserStore{user: map[string]interface{}{"uid": "5"}, keylimitCounts: map[string]int{"checkemail.me@example.com." + today: 1}},
			call: func(s *Service) (int, string, error) {
				return s.UserCheckEmailEdge(ctx, "token", "me@example.com")
			},
			errmsg: "发送太频率请稍后重试",
		},
		{
			name:  "checkemail daily limited",
			store: fakeUserStore{user: map[string]interface{}{"uid": "5"}, keylimitTotalCounts: map[string]int{"checkemail.me@example.com." + today: 50}},
			call: func(s *Service) (int, string, error) {
				return s.UserCheckEmailEdge(ctx, "token", "me@example.com")
			},
			errmsg: "系统维护稍后重试",
		},
		{
			name:  "checkemail existing email",
			store: fakeUserStore{user: map[string]interface{}{"uid": "5"}, userByEmail: map[string]interface{}{"uid": "9"}},
			call: func(s *Service) (int, string, error) {
				return s.UserCheckEmailEdge(ctx, "token", "me@example.com")
			},
			errmsg: "邮箱已经被使用了",
		},
		{
			name:  "checkemail available",
			store: fakeUserStore{user: map[string]interface{}{"uid": "5"}},
			call: func(s *Service) (int, string, error) {
				return s.UserCheckEmailEdge(ctx, "token", "me@example.com")
			},
			errmsg: "邮箱可用",
		},
		{
			name: "sendemail missing mail config",
			store: fakeUserStore{
				user:     map[string]interface{}{"uid": "5"},
				settings: map[string]map[string]interface{}{"setting": {"value": `a:1:{s:8:"mailconf";s:0:"";}`}},
			},
			call: func(s *Service) (int, string, error) {
				return s.UserSendEmailEdge(ctx, "token", "me@example.com")
			},
			errmsg: "邮箱功能暂未开启，请稍后重试",
		},
		{
			name:  "sendemail rate limited",
			store: fakeUserStore{user: map[string]interface{}{"uid": "5"}, keylimitCounts: map[string]int{"bindemail.me@example.com." + today: 1}},
			call: func(s *Service) (int, string, error) {
				return s.UserSendEmailEdge(ctx, "token", "me@example.com")
			},
			errmsg: "发送太频率请稍后重试",
		},
		{
			name:  "sendemail existing email",
			store: fakeUserStore{user: map[string]interface{}{"uid": "5"}, userByEmail: map[string]interface{}{"uid": "9"}},
			call: func(s *Service) (int, string, error) {
				return s.UserSendEmailEdge(ctx, "token", "me@example.com")
			},
			errmsg: "邮箱已经被使用了",
		},
		{
			name: "verifyemail existing email",
			store: fakeUserStore{
				user:         map[string]interface{}{"uid": "5"},
				keylimitData: map[string]string{"email.me@example.com.123456": "5.me@example.com"},
				userByEmail:  map[string]interface{}{"uid": "9"},
			},
			call: func(s *Service) (int, string, error) {
				return s.UserVerifyEmailEdge(ctx, "token", "me@example.com", "123456")
			},
			errmsg: "邮箱已经被使用",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService(tc.store, "https://res.example.test")
			retcode, errmsg, err := tc.call(service)
			if err != nil {
				t.Fatal(err)
			}
			wantRetcode := -1
			if tc.name == "checkemail available" {
				wantRetcode = 0
			}
			if retcode != wantRetcode || errmsg != tc.errmsg {
				t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
			}
		})
	}

	verifiedEmail := map[string]interface{}{}
	service := NewService(fakeUserStore{
		user:          map[string]interface{}{"uid": "5"},
		keylimitData:  map[string]string{"email.me@example.com.123456": "5.me@example.com"},
		verifiedEmail: &verifiedEmail,
	}, "https://res.example.test")
	retcode, errmsg, err := service.UserVerifyEmailEdge(ctx, "token", "me@example.com", "123456")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "邮箱验证已确认，绑定成功" {
		t.Fatalf("verifyemail success retcode=%d errmsg=%q", retcode, errmsg)
	}
	if verifiedEmail["uid"] != 5 || verifiedEmail["email"] != "me@example.com" || verifiedEmail["key"] != "email.me@example.com.123456" {
		t.Fatalf("verified email=%#v", verifiedEmail)
	}
}

func TestUCPWriteActionPrecheckEdges(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	zero := 0
	closed := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}, exrate: &zero}, "https://res.example.test")

	retcode, errmsg, err := closed.CoinLogExchangeEdge(context.Background(), "token", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "系统已关闭兑换功能" {
		t.Fatalf("exchange closed retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.CoinLogExchangeEdge(context.Background(), "token", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "请指定兑换类型" {
		t.Fatalf("exchange type retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.CoinLogExchangeEdge(context.Background(), "token", 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "请指定兑换数量" {
		t.Fatalf("exchange num retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.CoinLogExchangeEdge(context.Background(), "token", 1, 9)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "提交金币最小数量为:10" {
		t.Fatalf("exchange min retcode=%d errmsg=%q", retcode, errmsg)
	}

	exchanged := map[string]interface{}{}
	exchangeNow := time.Unix(1700000900, 0)
	service = NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}, exchangedCoins: &exchanged}, "https://res.example.test")
	service.now = func() time.Time { return exchangeNow }
	retcode, errmsg, err = service.CoinLogExchangeEdge(context.Background(), "token", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("exchange coin2rmb retcode=%d errmsg=%q", retcode, errmsg)
	}
	if exchanged["uid"] != 5 || exchanged["extype"] != 1 || exchanged["coinnum"] != 10 || exchanged["amount"] != 100 || exchanged["now"] != exchangeNow.Unix() {
		t.Fatalf("exchange coin2rmb=%#v", exchanged)
	}

	exchanged = map[string]interface{}{}
	service = NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}, exchangedCoins: &exchanged}, "https://res.example.test")
	service.now = func() time.Time { return exchangeNow }
	retcode, errmsg, err = service.CoinLogExchangeEdge(context.Background(), "token", 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("exchange rmb2coin retcode=%d errmsg=%q", retcode, errmsg)
	}
	if exchanged["uid"] != 5 || exchanged["extype"] != 2 || exchanged["coinnum"] != 20 || exchanged["amount"] != 200 || exchanged["now"] != exchangeNow.Unix() {
		t.Fatalf("exchange rmb2coin=%#v", exchanged)
	}

	retcode, errmsg, err = service.VODOrderCreateEdge(context.Background(), "token", "", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "请填写视频番号或者视频名称" {
		t.Fatalf("vodorder create missing retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.VODOrderCreateEdge(context.Background(), "token", "ABC-001", "", 99)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "求片金币不能低于100" {
		t.Fatalf("vodorder create coins retcode=%d errmsg=%q", retcode, errmsg)
	}

	lowGold := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}, quota: map[string]interface{}{"goldcoin": "120"}}, "https://res.example.test")
	retcode, errmsg, err = lowGold.VODOrderCreateEdge(context.Background(), "token", "ABC-001", "", 200)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "金币不足:120" {
		t.Fatalf("vodorder create balance retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.VODOrderSupportEdge(context.Background(), "token", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "您助力的求片记录不存在" {
		t.Fatalf("vodorder support retcode=%d errmsg=%q", retcode, errmsg)
	}

	service.now = func() time.Time { return time.Unix(1000, 0) }
	supportStore := fakeUserStore{
		user:        map[string]interface{}{"uid": "5"},
		quota:       map[string]interface{}{"goldcoin": "10"},
		vodOrderRow: map[string]interface{}{"id": "7", "uid": "8", "start_time": "1200", "stop_time": "1300"},
	}
	support := NewService(supportStore, "https://res.example.test")
	support.now = service.now
	retcode, errmsg, err = support.VODOrderSupportEdge(context.Background(), "token", 7, 1)
	if err != nil {
		t.Fatal(err)
	}
	wantWindow := "该求片助力时间为" + formatUnixTime(1200) + "~" + formatUnixTime(1300)
	if retcode != -1 || errmsg != wantWindow {
		t.Fatalf("vodorder support time retcode=%d errmsg=%q", retcode, errmsg)
	}

	supportStore.vodOrderRow = map[string]interface{}{"id": "7", "uid": "8", "start_time": "900", "stop_time": "1300"}
	supportStore.quota = map[string]interface{}{"goldcoin": "1"}
	support = NewService(supportStore, "https://res.example.test")
	support.now = service.now
	retcode, errmsg, err = support.VODOrderSupportEdge(context.Background(), "token", 7, 2)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "金币不足:1" {
		t.Fatalf("vodorder support balance retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestTaskQRLinkFormatsLinkAndFallsBackFromPID(t *testing.T) {
	service := NewService(fakeUserStore{
		user: map[string]interface{}{"uid": "5", "uniqkey": "12345"},
		settings: map[string]map[string]interface{}{
			"baseset": {"value": "a:1:{s:10:\"inviteUrls\";s:29:\"https://a.test\n-\nhttps://b.test\";}"},
		},
		calldata: map[string]map[string]interface{}{
			"global.qrcode.link": {"type": "code", "content": "https://qr.test?u={inviteUrl}&c={inviteCode}"},
		},
	}, "https://res.example.test")
	service.now = func() time.Time { return time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC) }

	data, retcode, errmsg, err := service.TaskQRLink(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", "bad-pid")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if data["qrlink"] != "https://qr.test?u=https://b.test&c=9IX" {
		t.Fatalf("qrlink = %#v", data["qrlink"])
	}
}

func TestTaskShareAwardsAndFormatsText(t *testing.T) {
	awarded := map[string]interface{}{}
	service := NewService(fakeUserStore{
		user:             map[string]interface{}{"uid": "5", "gid": "4", "uniqkey": "12345"},
		sumCoinLogByType: map[int]int{coinTypeVODShare: 0},
		awardedCoins:     &awarded,
		settings: map[string]map[string]interface{}{
			"baseset": {"value": "a:1:{s:10:\"inviteUrls\";s:14:\"https://a.test\";}"},
		},
		calldata: map[string]map[string]interface{}{
			"global.share.text": {"type": "html", "content": "share {inviteUrl} {inviteCode}"},
		},
	}, "https://res.example.test")
	service.now = func() time.Time { return time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC) }

	data, retcode, errmsg, err := service.TaskShare(context.Background(), "token", "")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if data["sharetext"] != "share https://a.test 9IX" || atoi(data["taskdone"]) != 1 {
		t.Fatalf("share data = %#v", data)
	}
	if atoi(awarded["cointype"]) != coinTypeVODShare || atoi(awarded["addcoin"]) != 1 {
		t.Fatalf("share award mismatch: %v", awarded)
	}
}

func TestTaskQRCodeSetsKeylimitAndRendersPNG(t *testing.T) {
	renderer := &fakeQRCodeRenderer{body: []byte("task-png")}
	setKeylimit := map[string]interface{}{}
	service := NewService(fakeUserStore{
		user:        map[string]interface{}{"uid": "5", "uniqkey": "12345"},
		setKeylimit: &setKeylimit,
		settings: map[string]map[string]interface{}{
			"baseset": {"value": "a:1:{s:10:\"inviteUrls\";s:14:\"https://a.test\";}"},
		},
		calldata: map[string]map[string]interface{}{
			"global.qrcode.link": {"type": "code", "content": "https://qr.test?u={inviteUrl}&c={inviteCode}"},
		},
	}, "https://res.example.test").WithQRCodeRenderer(renderer)
	service.now = func() time.Time { return time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC) }

	body, retcode, errmsg, err := service.TaskQRCode(context.Background(), "token", "")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" || string(body) != "task-png" {
		t.Fatalf("retcode=%d errmsg=%q body=%q", retcode, errmsg, body)
	}
	if setKeylimit["key"] != "task.qrcode.5.20260701" || atoi(setKeylimit["keynum"]) != 1 {
		t.Fatalf("keylimit mismatch: %v", setKeylimit)
	}
	if renderer.content != "https://qr.test?u=https://a.test&c=9IX" {
		t.Fatalf("renderer content = %q", renderer.content)
	}
}

func TestTaskIndexRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.TaskIndex(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestTaskIndexFormatsTaskStats(t *testing.T) {
	service := NewService(fakeUserStore{
		user: map[string]interface{}{
			"uid":             "5",
			"uniqkey":         "12345",
			"username":        "tester",
			"nickname":        "",
			"gid":             "4",
			"sysgid":          "0",
			"regtime":         "1760000000",
			"gender":          "1",
			"avatar":          "",
			"newmsg":          "0",
			"goldcoin":        "9",
			"gold_bean":       "8",
			"recommend_total": "7",
		},
	}, "https://res.example.test")
	service.now = func() time.Time { return time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC) }

	data, retcode, errmsg, err := service.TaskIndex(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	user, ok := data["user"].(map[string]interface{})
	if !ok || user["uid"] != "5" {
		t.Fatalf("user = %#v", data["user"])
	}
	share, ok := data["share"].(map[string]interface{})
	if !ok {
		t.Fatalf("share = %#v", data["share"])
	}
	if share["daynum"] != 1 || share["limit"] != 3 || share["coinnum"] != coinTypeVODShare || share["donenum"] != 1 {
		t.Fatalf("share = %#v", share)
	}
	comment, ok := data["comment"].(map[string]interface{})
	if !ok || comment["daynum"] != 2 || comment["limit"] != 6 || comment["coinnum"] != coinTypeVODComment || comment["donenum"] != 3 {
		t.Fatalf("comment = %#v", data["comment"])
	}
	favorite, ok := data["favorite"].(map[string]interface{})
	if !ok || favorite["daynum"] != 3 || favorite["limit"] != 9 || favorite["coinnum"] != coinTypeVODFavorite || favorite["donenum"] != 4 {
		t.Fatalf("favorite = %#v", data["favorite"])
	}
	play10, ok := data["play10"].(map[string]interface{})
	if !ok || play10["daynum"] != 4 || play10["limit"] != 12 || play10["coinnum"] != coinTypeVODPlay10 || play10["donenum"] != 1 {
		t.Fatalf("play10 = %#v", data["play10"])
	}
	minivoddown, ok := data["minivoddown"].(map[string]interface{})
	if !ok || minivoddown["daynum"] != 7 || minivoddown["limit"] != 21 || minivoddown["coinnum"] != coinTypeMiniVODDownTask || minivoddown["donenum"] != 1 {
		t.Fatalf("minivoddown = %#v", data["minivoddown"])
	}
}

func TestTaskboxQRLinkUsesTaskboxConfig(t *testing.T) {
	service := NewService(fakeUserStore{
		user: map[string]interface{}{"uid": "5", "uniqkey": "12345"},
		settings: map[string]map[string]interface{}{
			"baseset": {"value": "a:1:{s:10:\"inviteUrls\";s:14:\"https://a.test\";}"},
		},
		calldata: map[string]map[string]interface{}{
			"taskbox.qrcode.link": {"type": "code", "content": "https://box.test?u={inviteUrl}&c={inviteCode}"},
		},
	}, "https://res.example.test")
	service.now = func() time.Time { return time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC) }

	data, retcode, errmsg, err := service.TaskboxQRLink(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", "")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if data["qrlink"] != "https://box.test?u=https://a.test&c=9IX" {
		t.Fatalf("qrlink = %#v", data["qrlink"])
	}
}

func TestTaskboxQRCodeRendersTaskboxLink(t *testing.T) {
	renderer := &fakeQRCodeRenderer{body: []byte("fake-png")}
	service := NewService(fakeUserStore{
		user: map[string]interface{}{"uid": "5", "uniqkey": "12345"},
		settings: map[string]map[string]interface{}{
			"baseset": {"value": "a:1:{s:10:\"inviteUrls\";s:14:\"https://a.test\";}"},
		},
		calldata: map[string]map[string]interface{}{
			"taskbox.qrcode.link": {"type": "code", "content": "https://box.test?u={inviteUrl}&c={inviteCode}"},
		},
	}, "https://res.example.test").WithQRCodeRenderer(renderer)
	service.now = func() time.Time { return time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC) }

	body, retcode, errmsg, err := service.TaskboxQRCode(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", "")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if string(body) != "fake-png" {
		t.Fatalf("body = %q", body)
	}
	if renderer.content != "https://box.test?u=https://a.test&c=9IX" {
		t.Fatalf("renderer content = %q", renderer.content)
	}
}

func TestTaskboxShareFormatsText(t *testing.T) {
	service := NewService(fakeUserStore{
		user: map[string]interface{}{"uid": "5", "uniqkey": "12345"},
		settings: map[string]map[string]interface{}{
			"baseset": {"value": "a:1:{s:10:\"inviteUrls\";s:14:\"https://a.test\";}"},
		},
		calldata: map[string]map[string]interface{}{
			"taskbox.share.text": {"type": "html", "content": "go {inviteUrl} {inviteCode}"},
		},
	}, "https://res.example.test")
	service.now = func() time.Time { return time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC) }

	data, retcode, errmsg, err := service.TaskboxShare(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", "")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if data["sharetext"] != "go https://a.test 9IX" {
		t.Fatalf("sharetext = %#v", data["sharetext"])
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

func TestBankcardPostRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	retcode, errmsg, err := service.BankcardPost(context.Background(), "", BankcardPostRequest{Action: "create"})
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestBankcardCreateValidatesName(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")

	retcode, errmsg, err := service.BankcardPost(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", BankcardPostRequest{
		Action:  "create",
		CardNum: "account",
		Type:    1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "姓名长度不正确" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestBankcardCreateSuccessDefaultsAlipay(t *testing.T) {
	created := map[string]interface{}{}
	def := map[string]interface{}{}
	service := NewService(fakeUserStore{
		user:            map[string]interface{}{"uid": "5"},
		createdBankcard: created,
		defaultBankcard: def,
	}, "https://res.example.test")

	retcode, errmsg, err := service.BankcardPost(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", BankcardPostRequest{
		Action:   "create",
		Name:     " 张三 ",
		BankName: "会被覆盖",
		CardNum:  " account ",
		IsDef:    1,
		Type:     1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "操作成功" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if created["uid"] != 5 || created["name"] != "张三" || created["bankname"] != "支付宝" || created["cardnum"] != "account" || created["isdef"] != 1 || created["type"] != 1 {
		t.Fatalf("created = %#v", created)
	}
	if def["uid"] != 5 || def["cardid"] != 88 {
		t.Fatalf("default = %#v", def)
	}
}

func TestBankcardModifyMissing(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")

	retcode, errmsg, err := service.BankcardPost(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", BankcardPostRequest{
		Action:  "modify",
		CardID:  7,
		Name:    "张三",
		CardNum: "account",
		Type:    3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "修改的记录不存在" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestBankcardDeleteSuccess(t *testing.T) {
	deleted := map[string]interface{}{}
	service := NewService(fakeUserStore{
		user:            map[string]interface{}{"uid": "5"},
		deletedBankcard: deleted,
	}, "https://res.example.test")

	retcode, errmsg, err := service.BankcardDelete(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 9)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "操作成功" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if deleted["uid"] != 5 || deleted["cardid"] != 9 {
		t.Fatalf("deleted = %#v", deleted)
	}
}

func TestVIPPackageIndexRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.VIPPkgIndex(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestVIPPackageIndexFormatsRowsAndPayments(t *testing.T) {
	service := NewService(fakeUserStore{
		user: map[string]interface{}{"uid": "5"},
		settings: map[string]map[string]interface{}{
			"setting": {"value": `a:1:{s:10:"safepayurl";s:20:"https://safe.example";}`},
		},
		packages: []map[string]interface{}{{
			"pkgid":          "1",
			"pkgname":        "VIP",
			"daylen":         "30",
			"showtype":       "0",
			"rmbprice":       "6800",
			"coinprice":      "100",
			"recommend":      "1",
			"bonus_vip_days": "2",
			"memo":           "memo",
		}},
		payments: []map[string]interface{}{{
			"channame": "余额支付",
			"chanlogo": "",
			"dscr":     "",
			"payways": []map[string]interface{}{{
				"payname":       "余额",
				"paylogo":       "",
				"dscr":          "",
				"paycode":       "balance.pay",
				"trxamount_min": "1",
				"trxamount_max": "999",
				"allow_paytypes": map[int][]string{
					1: {"ALL"},
				},
			}},
		}},
	}, "https://res.example.test")

	data, retcode, errmsg, err := service.VIPPkgIndex(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	rows := data["pkgrows"].([]map[string]interface{})
	if rows[0]["rmbprice"] != "68.00" || rows[0]["daylen"] != 30 || rows[0]["coinprice"] != 100 {
		t.Fatalf("pkgrows = %#v", rows)
	}
	payments := data["payments"].([]map[string]interface{})
	if len(payments) != 1 || payments[0]["channame"] != "余额支付" {
		t.Fatalf("payments = %#v", payments)
	}
	if data["safepayurl"] != "https://safe.example" {
		t.Fatalf("safepayurl = %#v", data["safepayurl"])
	}
}

func TestPackageOrderEdges(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")

	retcode, errmsg, err := service.VIPPkgCoinOrderEdge(context.Background(), "token", 99)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "套餐不存在或未启用" {
		t.Fatalf("vip missing retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:       map[string]interface{}{"uid": "5"},
		packageRow: map[string]interface{}{"pkgid": "1", "showtype": "1", "coinprice": "100"},
	}, "https://res.example.test")
	retcode, errmsg, err = service.VIPPkgCoinOrderEdge(context.Background(), "token", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "套餐不存在或未启用" {
		t.Fatalf("vip stopped retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:       map[string]interface{}{"uid": "5"},
		quota:      map[string]interface{}{"goldcoin": "99"},
		packageRow: map[string]interface{}{"pkgid": "1", "showtype": "0", "coinprice": "100"},
	}, "https://res.example.test")
	retcode, errmsg, err = service.VIPPkgCoinOrderEdge(context.Background(), "token", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "金币不足，快做任务获取金币吧！" {
		t.Fatalf("vip balance retcode=%d errmsg=%q", retcode, errmsg)
	}

	vipNow := time.Unix(1700000000, 0)
	upgradedVIP := map[string]interface{}{}
	service = NewService(fakeUserStore{
		user:        map[string]interface{}{"uid": "5", "sysgid": "4"},
		quota:       map[string]interface{}{"goldcoin": "100"},
		packageRow:  map[string]interface{}{"pkgid": "1", "showtype": "0", "coinprice": "100", "daylen": "30"},
		upgradedVIP: &upgradedVIP,
	}, "https://res.example.test")
	service.now = func() time.Time { return vipNow }
	data, retcode, errmsg, err := service.VIPPkgCoinOrder(context.Background(), "token", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "您已成功升级尊贵会员" {
		t.Fatalf("vip coinorder success retcode=%d errmsg=%q", retcode, errmsg)
	}
	wantExpiry := vipNow.Add(30 * 24 * time.Hour).Unix()
	if data["deduct_coin"] != 100 || data["expiry_date"] != formatMinuteTime(wantExpiry) {
		t.Fatalf("vip coinorder data=%#v", data)
	}
	if upgradedVIP["uid"] != 5 || upgradedVIP["deductCoin"] != 100 || upgradedVIP["vipGID"] != 6 || upgradedVIP["expiry"] != wantExpiry {
		t.Fatalf("upgraded vip=%#v", upgradedVIP)
	}

	upgradedVIP = map[string]interface{}{}
	currentExpiry := vipNow.Add(10 * 24 * time.Hour).Unix()
	service = NewService(fakeUserStore{
		user:        map[string]interface{}{"uid": "5", "sysgid": "6", "sysgid_exptime": fmt.Sprint(currentExpiry)},
		quota:       map[string]interface{}{"goldcoin": "100"},
		packageRow:  map[string]interface{}{"pkgid": "1", "showtype": "0", "coinprice": "100", "daylen": "30"},
		upgradedVIP: &upgradedVIP,
	}, "https://res.example.test")
	service.now = func() time.Time { return vipNow }
	_, retcode, errmsg, err = service.VIPPkgCoinOrder(context.Background(), "token", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "您已成功升级尊贵会员" {
		t.Fatalf("vip coinorder extend retcode=%d errmsg=%q", retcode, errmsg)
	}
	if upgradedVIP["expiry"] != currentExpiry+30*86400 {
		t.Fatalf("extended vip=%#v", upgradedVIP)
	}

	service = NewService(fakeUserStore{
		user:       map[string]interface{}{"uid": "5"},
		quota:      map[string]interface{}{"goldcoin": "20"},
		packageRow: map[string]interface{}{"pkgid": "2", "showtype": "0", "rmbprice": "300"},
	}, "https://res.example.test")
	retcode, errmsg, err = service.BeanPkgCoinOrderEdge(context.Background(), "token", 2)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "金币不足，快做任务获取金币吧！" {
		t.Fatalf("bean balance retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:       map[string]interface{}{"uid": "5"},
		quota:      map[string]interface{}{"goldcoin": "30"},
		packageRow: map[string]interface{}{"pkgid": "2", "showtype": "0", "rmbprice": "300"},
	}, "https://res.example.test")
	retcode, errmsg, err = service.BeanPkgCoinOrderEdge(context.Background(), "token", 2)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "您已成功兑换金豆" {
		t.Fatalf("bean success edge retcode=%d errmsg=%q", retcode, errmsg)
	}

	boughtBeans := map[string]interface{}{}
	beanNow := time.Unix(1700000600, 0)
	service = NewService(fakeUserStore{
		user:        map[string]interface{}{"uid": "5"},
		quota:       map[string]interface{}{"goldcoin": "30"},
		packageRow:  map[string]interface{}{"pkgid": "2", "showtype": "0", "rmbprice": "300"},
		boughtBeans: &boughtBeans,
	}, "https://res.example.test")
	service.now = func() time.Time { return beanNow }
	data, retcode, errmsg, err = service.BeanPkgCoinOrder(context.Background(), "token", 2)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "您已成功兑换金豆" {
		t.Fatalf("bean success retcode=%d errmsg=%q", retcode, errmsg)
	}
	if data["deduct_coin"] != 30 {
		t.Fatalf("bean data=%#v", data)
	}
	if boughtBeans["uid"] != 5 || boughtBeans["deductCoin"] != 30 || boughtBeans["addBeans"] != 30 || boughtBeans["now"] != beanNow.Unix() {
		t.Fatalf("bought beans=%#v", boughtBeans)
	}

	service = NewService(fakeUserStore{
		user:       map[string]interface{}{"uid": "5"},
		packageRow: map[string]interface{}{"pkgid": "3", "showtype": "0", "rmbprice": "3800"},
	}, "https://res.example.test")
	retcode, errmsg, err = service.VIPPkgPlaceOrderEdge(context.Background(), "token", 3, "wappay3.1")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "套餐仅支持金币兑换" {
		t.Fatalf("vip placeorder rmb retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")
	retcode, errmsg, err = service.CoinPkgPlaceOrderEdge(context.Background(), "token", 3, "wappay3.1")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "套餐不存在或未启用" {
		t.Fatalf("coin placeorder missing retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:       map[string]interface{}{"uid": "5"},
		packageRow: map[string]interface{}{"pkgid": "3", "showtype": "0", "rmbprice": "1200"},
		payments: []map[string]interface{}{{
			"disabled": "0",
			"payways": []map[string]interface{}{{
				"paycode":        "wappay3.1",
				"trxamount_min":  "1000",
				"trxamount_max":  "2000",
				"allow_paytypes": map[int][]string{2: []string{"ALL"}},
			}},
		}},
	}, "https://res.example.test")
	retcode, errmsg, err = service.CoinPkgPlaceOrderEdge(context.Background(), "token", 3, "bad.pay")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "支付方式错误或不被允许" {
		t.Fatalf("coin placeorder paycode retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.CoinPkgPlaceOrderEdge(context.Background(), "token", 3, "wappay3.1")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "支付下单成功分支暂未迁移" {
		t.Fatalf("coin placeorder pending retcode=%d errmsg=%q", retcode, errmsg)
	}

	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	setting := map[string]map[string]interface{}{
		"setting": {"value": `a:8:{s:7:"ordercd";i:10;s:12:"unpaidorders";i:2;s:13:"successorders";i:0;s:7:"regdays";i:3;s:8:"viewvods";i:5;s:13:"orderdaylimit";i:2;s:18:"randomPaywayStatus";i:1;s:6:"exrate";i:10;}`},
	}
	service = NewService(fakeUserStore{
		user:                map[string]interface{}{"uid": "5", "regtime": fmt.Sprint(now.Add(-10 * 24 * time.Hour).Unix())},
		packageRow:          map[string]interface{}{"pkgid": "3", "showtype": "0", "rmbprice": "1200"},
		settings:            setting,
		paymentStatusCounts: map[string]int{fmt.Sprintf("0:%d", now.Unix()-600): 2},
		payments: []map[string]interface{}{{
			"payways": []map[string]interface{}{{"paycode": "wappay3.1", "allow_paytypes": map[int][]string{1: []string{"ALL"}}}},
		}},
	}, "https://res.example.test")
	service.now = func() time.Time { return now }
	retcode, errmsg, err = service.VIPPkgPlaceOrderEdge(context.Background(), "token", 3, "wappay3.1")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "你有未支付订单，10分钟内无法提交新订单" {
		t.Fatalf("vip order cooldown retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:       map[string]interface{}{"uid": "5", "regtime": fmt.Sprint(now.Unix())},
		packageRow: map[string]interface{}{"pkgid": "3", "showtype": "0", "rmbprice": "1200"},
		settings:   setting,
		payments: []map[string]interface{}{{
			"payways": []map[string]interface{}{{"paycode": "wappay3.1", "allow_paytypes": map[int][]string{1: []string{"ALL"}}}},
		}},
	}, "https://res.example.test")
	service.now = func() time.Time { return now }
	retcode, errmsg, err = service.VIPPkgPlaceOrderEdge(context.Background(), "token", 3, "wappay3.1")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "系统限制注册3天后方可充值VIP" {
		t.Fatalf("vip regdays retcode=%d errmsg=%q", retcode, errmsg)
	}

	zeroViews := 0
	service = NewService(fakeUserStore{
		user:         map[string]interface{}{"uid": "5", "regtime": fmt.Sprint(now.Add(-10 * 24 * time.Hour).Unix())},
		packageRow:   map[string]interface{}{"pkgid": "3", "showtype": "0", "rmbprice": "1200"},
		settings:     setting,
		vodPlayCount: &zeroViews,
		payments: []map[string]interface{}{{
			"payways": []map[string]interface{}{{"paycode": "wappay3.1", "allow_paytypes": map[int][]string{1: []string{"ALL"}}}},
		}},
	}, "https://res.example.test")
	service.now = func() time.Time { return now }
	retcode, errmsg, err = service.VIPPkgPlaceOrderEdge(context.Background(), "token", 3, "wappay3.1")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "系统限制观影5部后方可充值VIP" {
		t.Fatalf("vip viewvods retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:                map[string]interface{}{"uid": "5", "regtime": fmt.Sprint(now.Add(-10 * 24 * time.Hour).Unix())},
		packageRow:          map[string]interface{}{"pkgid": "3", "showtype": "0", "rmbprice": "1200"},
		settings:            setting,
		paymentStatusCounts: map[string]int{fmt.Sprintf("0:%d", dayStartUnix(now)): 2},
		payments: []map[string]interface{}{{
			"payways": []map[string]interface{}{{"paycode": "wappay3.1", "allow_paytypes": map[int][]string{3: []string{"ALL"}}}},
		}},
	}, "https://res.example.test")
	service.now = func() time.Time { return now }
	retcode, errmsg, err = service.BeanPkgPlaceOrderEdge(context.Background(), "token", 3, "wappay3.1")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "您今天购买金豆订单数已达上限，请明天再来" {
		t.Fatalf("bean day limit retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:       map[string]interface{}{"uid": "5", "regtime": fmt.Sprint(now.Add(-10 * 24 * time.Hour).Unix())},
		packageRow: map[string]interface{}{"pkgid": "3", "showtype": "0", "rmbprice": "1200"},
		settings:   setting,
	}, "https://res.example.test")
	service.now = func() time.Time { return now }
	retcode, errmsg, err = service.BeanPkgPlaceOrderEdge(context.Background(), "token", 3, "all.alipay")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "该支付方式当前没有可用通道" {
		t.Fatalf("bean random no channel retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestUpgradeEdgePrechecks(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5", "sysgid": "6"}}, "https://res.example.test")
	retcode, errmsg, err := service.UpgradeEdge(context.Background(), "token", 7)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "您已经是尊贵会员" {
		t.Fatalf("vip retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{user: map[string]interface{}{"uid": "5", "sysgid": "4"}}, "https://res.example.test")
	retcode, errmsg, err = service.UpgradeEdge(context.Background(), "token", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "请选择一个时长" {
		t.Fatalf("day retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.UpgradeEdge(context.Background(), "token", 3650)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "终身尊贵VIP暂停升级" {
		t.Fatalf("lifetime retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:  map[string]interface{}{"uid": "5", "sysgid": "4"},
		quota: map[string]interface{}{"goldcoin": "99"},
	}, "https://res.example.test")
	retcode, errmsg, err = service.UpgradeEdge(context.Background(), "token", 7)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "金币不足，快做任务获取金币吧！" {
		t.Fatalf("gold retcode=%d errmsg=%q", retcode, errmsg)
	}

	upgraded := map[string]interface{}{}
	service = NewService(fakeUserStore{
		user:        map[string]interface{}{"uid": "5", "sysgid": "4"},
		quota:       map[string]interface{}{"goldcoin": "100"},
		upgradedVIP: &upgraded,
	}, "https://res.example.test")
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	data, retcode, errmsg, err := service.Upgrade(context.Background(), "token", 7)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "您已成功尊贵会员" || atoi(data["deduct_coin"]) != 100 || data["expiry_date"] != "2026-07-08 20:00" {
		t.Fatalf("upgrade retcode=%d errmsg=%q data=%v", retcode, errmsg, data)
	}
	if atoi(upgraded["uid"]) != 5 || atoi(upgraded["deductCoin"]) != 100 || atoi(upgraded["vipGID"]) != 6 || atoi64(upgraded["expiry"]) != now.Add(7*24*time.Hour).Unix() {
		t.Fatalf("upgrade write mismatch: %v", upgraded)
	}
}

func TestCoinPackageIndexFormatsBonusCoins(t *testing.T) {
	service := NewService(fakeUserStore{
		user: map[string]interface{}{"uid": "5"},
		packages: []map[string]interface{}{{
			"pkgid":          "2",
			"pkgname":        "金币",
			"showtype":       "0",
			"rmbprice":       "1200",
			"recommend":      "0",
			"bonus_vip_days": "0",
			"bonus_coins":    "300",
		}},
	}, "https://res.example.test")

	data, retcode, errmsg, err := service.CoinPkgIndex(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	rows := data["pkgrows"].([]map[string]interface{})
	if rows[0]["rmbprice"] != "12.00" || rows[0]["bonus_coins"] != 300 {
		t.Fatalf("pkgrows = %#v", rows)
	}
}

func TestVODOrderMyOrdersRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.VODOrderMyOrders(context.Background(), "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestVODOrderIndexRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.VODOrderIndex(context.Background(), "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestVODOrderIndexReturnsRanking(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	service := NewService(fakeUserStore{
		user: map[string]interface{}{"uid": "5"},
		settings: map[string]map[string]interface{}{
			"setting": {"value": `a:1:{s:16:"vod_order_period";i:7;}`},
		},
		latestVODIssue: map[string]interface{}{"issue": time.Date(2026, 7, 1, 0, 0, 0, 0, loc).Unix()},
		vodOrders: []map[string]interface{}{
			{"id": "9", "uid": "6", "coins": "100", "support_coins": "30"},
		},
		maxVODSupport:     map[string]interface{}{"uid": "7", "total_coins": "88"},
		myVODSupportCoins: 12,
		userByID: map[string]interface{}{
			"uid":             "7",
			"username":        "helper",
			"nickname":        "Helper Nick",
			"gid":             "4",
			"sysgid":          "0",
			"regtime":         "1760000000",
			"avatar":          "",
			"newmsg":          "0",
			"recommend_total": "0",
		},
	}, "https://res.example.test")
	service.now = func() time.Time { return time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC) }

	data, retcode, errmsg, err := service.VODOrderIndex(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if data["issue"] != "260708" {
		t.Fatalf("issue = %#v", data["issue"])
	}
	rows := data["data"].([]map[string]interface{})
	if len(rows) != 1 {
		t.Fatalf("rows = %#v", rows)
	}
	if rows[0]["my_support"] != 12 {
		t.Fatalf("my_support = %#v", rows[0]["my_support"])
	}
	top := rows[0]["top"].(map[string]interface{})
	if top["uid"] != "7" || top["total_coins"] != "88" || top["username"] != "helper" || top["nickname"] != "helper" {
		t.Fatalf("top = %#v", top)
	}
}

func TestVODOrderMyOrdersReturnsTotals(t *testing.T) {
	service := NewService(fakeUserStore{
		user:      map[string]interface{}{"uid": "5"},
		vodOrders: []map[string]interface{}{{"id": "1", "uid": "5", "coins": "100", "status": "0"}},
	}, "https://res.example.test")

	data, retcode, errmsg, err := service.VODOrderMyOrders(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if data["total_cost"] != 111 || data["current_frozen"] != 37 {
		t.Fatalf("totals = %#v", data)
	}
	if len(data["data"].([]map[string]interface{})) != 1 {
		t.Fatalf("data = %#v", data["data"])
	}
}

func TestVODOrderMySupportsReturnsRows(t *testing.T) {
	service := NewService(fakeUserStore{
		user:        map[string]interface{}{"uid": "5"},
		vodSupports: []map[string]interface{}{{"void": "9", "coins": "12"}},
	}, "https://res.example.test")

	data, retcode, errmsg, err := service.VODOrderMySupports(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if len(data["data"].([]map[string]interface{})) != 1 {
		t.Fatalf("data = %#v", data["data"])
	}
}

func TestVODOrderHistoryOrdersReturnsRows(t *testing.T) {
	service := NewService(fakeUserStore{
		user:      map[string]interface{}{"uid": "5"},
		vodOrders: []map[string]interface{}{{"id": "2", "status": "1"}},
	}, "https://res.example.test")

	data, retcode, errmsg, err := service.VODOrderHistoryOrders(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if len(data["data"].([]map[string]interface{})) != 1 {
		t.Fatalf("data = %#v", data["data"])
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

func TestFeedbackCreateRequiresLogin(t *testing.T) {
	service := NewService(&fakeUserStore{}, "")

	retcode, errmsg, err := service.FeedbackCreate(context.Background(), "", FeedbackCreateRequest{Content: "hello"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("response=%d %q", retcode, errmsg)
	}
}

func TestFeedbackCreateValidatesContent(t *testing.T) {
	service := NewService(&fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "")

	retcode, errmsg, err := service.FeedbackCreate(context.Background(), "token", FeedbackCreateRequest{Content: ""})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if retcode != -1 || errmsg != "内容最多250个字符" {
		t.Fatalf("response=%d %q", retcode, errmsg)
	}
}

func TestFeedbackCreateValidatesPaymentOwner(t *testing.T) {
	service := NewService(&fakeUserStore{user: map[string]interface{}{"uid": "5"}, paymentRow: map[string]interface{}{"payid": "9", "uid": "6"}}, "")

	retcode, errmsg, err := service.FeedbackCreate(context.Background(), "token", FeedbackCreateRequest{CID: 5, Content: "hello", PayID: 9})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if retcode != -1 || errmsg != "请选择订单信息" {
		t.Fatalf("response=%d %q", retcode, errmsg)
	}
}

func TestFeedbackCreateSuccess(t *testing.T) {
	created := domain.FeedbackCreateInput{}
	store := &fakeUserStore{user: map[string]interface{}{"uid": "5"}, paymentRow: map[string]interface{}{"payid": "9", "uid": "5"}, createdFeedback: &created}
	service := NewService(store, "")
	service.now = func() time.Time { return time.Unix(1700000000, 0) }

	retcode, errmsg, err := service.FeedbackCreate(context.Background(), "token", FeedbackCreateRequest{
		CID:        5,
		Content:    "hello\n",
		PayID:      9,
		PayName:    " ali ",
		PayAccount: " acc ",
		Device:     " ios ",
		LongIDs:    " 1 ",
		ShortIDs:   " 2 ",
		IP:         "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if retcode != 0 || errmsg != "信息已反馈" {
		t.Fatalf("response=%d %q", retcode, errmsg)
	}
	if created.UID != 5 || created.Content != "hello" || created.PayID != 9 || created.CreatedAt != 1700000000 {
		t.Fatalf("created=%#v", created)
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

func TestWithdrawIndexRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.WithdrawIndex(context.Background(), "")
	if err != nil {
		t.Fatalf("withdraw index: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestWithdrawIndexFormatsRows(t *testing.T) {
	service := NewService(fakeUserStore{
		user: map[string]interface{}{"uid": "5"},
		settings: map[string]map[string]interface{}{
			"setting": {
				"value": `a:7:{s:6:"exrate";i:10;s:8:"topupmin";i:5000;s:19:"alipay_withdraw_min";i:1000;s:19:"alipay_withdraw_max";i:200000;s:21:"bankcard_withdraw_min";i:3000;s:21:"bankcard_withdraw_max";i:500000;}`,
			},
			"game.setting": {
				"value": `a:2:{s:11:"withdrawmin";i:6000;s:12:"withdrawrate";d:0.08;}`,
			},
		},
		bankcards: []map[string]interface{}{{"cardid": "1", "uid": "5", "name": "张三", "bankname": "支付宝", "cardnum": "abc", "isdef": "1", "type": "1"}},
	}, "https://res.example.test")

	data, retcode, errmsg, err := service.WithdrawIndex(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965")
	if err != nil {
		t.Fatalf("withdraw index: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if data.GoldCoin != 625 || data.ExRate != 10 || data.TopupMin != "50.00" {
		t.Fatalf("unexpected coin settings %#v", data)
	}
	if data.Coin2RMB != "62.50" || data.Max2RMB != "70.50" {
		t.Fatalf("unexpected rmb fields %#v", data)
	}
	if data.GameWithdrawMin != 6000 || data.GameWithdrawRate != 0.08 {
		t.Fatalf("unexpected game withdraw settings %#v", data)
	}
	if data.AlipayWithdrawMin != 1000 || data.BankcardWithdrawMax != 500000 {
		t.Fatalf("unexpected withdraw settings %#v", data)
	}
	if len(data.CardRows) != 1 || data.CardRows[0]["bankname"] != "支付宝" {
		t.Fatalf("unexpected card rows %#v", data.CardRows)
	}
}

func TestWithdrawListingRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	_, retcode, errmsg, err := service.WithdrawListing(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("withdraw listing: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected auth response %d %q", retcode, errmsg)
	}
}

func TestWithdrawListingFormatsRows(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}, withdrawTotal: 43210}, "https://res.example.test")

	data, retcode, errmsg, err := service.WithdrawListing(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 1)
	if err != nil {
		t.Fatalf("withdraw listing: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if data.WithdrawTotal != "432.10" || data.PageInfo["page_url"] != "/ucp/withdraw/listing?page=[?]" {
		t.Fatalf("unexpected listing data %#v", data)
	}
	if len(data.Rows) != 1 || data.Rows[0]["withdraw_amount"] != "123.45" || data.Rows[0]["createtime"] != formatWithdrawTime(1770000000) {
		t.Fatalf("unexpected rows %#v", data.Rows)
	}
}

func TestWithdrawCreateEdgePrechecks(t *testing.T) {
	baseSettings := map[string]map[string]interface{}{
		"setting": {
			"value": `a:7:{s:8:"topupmin";i:5000;s:14:"withdraw_limit";i:2;s:19:"alipay_withdraw_min";i:1000;s:19:"alipay_withdraw_max";i:200000;s:21:"bankcard_withdraw_min";i:3000;s:21:"bankcard_withdraw_max";i:500000;}`,
		},
		"game.setting": {
			"value": `a:1:{s:11:"withdrawmin";i:6000;}`,
		},
	}
	service := NewService(fakeUserStore{
		user:     map[string]interface{}{"uid": "5", "perms": `{}`},
		settings: baseSettings,
	}, "https://res.example.test")

	retcode, errmsg, err := service.WithdrawCreateEdge(context.Background(), "token", 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "请填写提现金额" {
		t.Fatalf("amount retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.WithdrawCreateEdge(context.Background(), "token", 0, 0, 1000000000)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "提现金额异常" {
		t.Fatalf("large retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.WithdrawCreateEdge(context.Background(), "token", 0, 1, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "提现金额最小为60.00元" {
		t.Fatalf("game min retcode=%d errmsg=%q", retcode, errmsg)
	}

	retcode, errmsg, err = service.WithdrawCreateEdge(context.Background(), "token", 0, 0, 4000)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "提现金额最小为50.00元" {
		t.Fatalf("min retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:     map[string]interface{}{"uid": "5", "withdraw_deny": "1", "perms": `{}`},
		settings: baseSettings,
	}, "https://res.example.test")
	retcode, errmsg, err = service.WithdrawCreateEdge(context.Background(), "token", 0, 0, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "您已被限制提现" {
		t.Fatalf("deny retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:     map[string]interface{}{"uid": "5", "recommend_total": "1", "perms": `{"min.withdraw.recommend.num":"2"}`},
		settings: baseSettings,
	}, "https://res.example.test")
	retcode, errmsg, err = service.WithdrawCreateEdge(context.Background(), "token", 0, 0, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "提现最少需邀请2人" {
		t.Fatalf("recommend retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:     map[string]interface{}{"uid": "5", "recommend_total": "2", "perms": `{"min.withdraw.recommend.num":"2"}`},
		settings: baseSettings,
	}, "https://res.example.test")
	retcode, errmsg, err = service.WithdrawCreateEdge(context.Background(), "token", 0, 0, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "请选择一个收款账号" {
		t.Fatalf("card retcode=%d errmsg=%q", retcode, errmsg)
	}

	two := 2
	service = NewService(fakeUserStore{
		user:               map[string]interface{}{"uid": "5", "recommend_total": "2", "perms": `{"min.withdraw.recommend.num":"2"}`},
		settings:           baseSettings,
		bankcardRow:        map[string]interface{}{"cardid": "7", "type": "1"},
		withdrawSinceCount: &two,
	}, "https://res.example.test")
	retcode, errmsg, err = service.WithdrawCreateEdge(context.Background(), "token", 7, 0, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "当天提现次数不能超过2次" {
		t.Fatalf("limit retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:        map[string]interface{}{"uid": "5", "recommend_total": "2", "perms": `{"min.withdraw.recommend.num":"2"}`},
		settings:    baseSettings,
		bankcardRow: map[string]interface{}{"cardid": "7", "type": "1"},
		account:     map[string]interface{}{"uid": "5", "available_balance": "5000"},
	}, "https://res.example.test")
	retcode, errmsg, err = service.WithdrawCreateEdge(context.Background(), "token", 7, 0, 250000)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "支付宝提现范围为 ¥10.00~2000.00" {
		t.Fatalf("alipay range retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:        map[string]interface{}{"uid": "5", "recommend_total": "2", "perms": `{"min.withdraw.recommend.num":"2"}`},
		settings:    baseSettings,
		bankcardRow: map[string]interface{}{"cardid": "8", "type": "2"},
	}, "https://res.example.test")
	retcode, errmsg, err = service.WithdrawCreateEdge(context.Background(), "token", 8, 0, 600000)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "银行卡提现范围为 ¥30.00~5000.00" {
		t.Fatalf("bankcard range retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:        map[string]interface{}{"uid": "5", "recommend_total": "2", "perms": `{"min.withdraw.recommend.num":"2"}`},
		settings:    baseSettings,
		bankcardRow: map[string]interface{}{"cardid": "7", "type": "1"},
		account:     map[string]interface{}{"uid": "5", "game_available_balance": "3000", "available_balance": "800"},
	}, "https://res.example.test")
	retcode, errmsg, err = service.WithdrawCreateEdge(context.Background(), "token", 7, 1, 7000)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "余额不足" {
		t.Fatalf("game balance retcode=%d errmsg=%q", retcode, errmsg)
	}

	wideSettings := map[string]map[string]interface{}{
		"setting": {
			"value": `a:7:{s:8:"topupmin";i:5000;s:14:"withdraw_limit";i:2;s:6:"exrate";i:0;s:19:"alipay_withdraw_min";i:1000;s:19:"alipay_withdraw_max";i:999999999;s:21:"bankcard_withdraw_min";i:3000;s:21:"bankcard_withdraw_max";i:999999999;}`,
		},
		"game.setting": {
			"value": `a:1:{s:11:"withdrawmin";i:6000;}`,
		},
	}
	service = NewService(fakeUserStore{
		user:        map[string]interface{}{"uid": "5", "recommend_total": "2", "perms": `{"min.withdraw.recommend.num":"2"}`},
		settings:    wideSettings,
		bankcardRow: map[string]interface{}{"cardid": "7", "type": "1"},
		account:     map[string]interface{}{"uid": "5", "available_balance": "800"},
	}, "https://res.example.test")
	retcode, errmsg, err = service.WithdrawCreateEdge(context.Background(), "token", 7, 0, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "系统已关闭兑换功能" {
		t.Fatalf("exchange closed retcode=%d errmsg=%q", retcode, errmsg)
	}

	wideSettings["setting"]["value"] = `a:7:{s:8:"topupmin";i:5000;s:14:"withdraw_limit";i:2;s:6:"exrate";i:100000;s:19:"alipay_withdraw_min";i:1000;s:19:"alipay_withdraw_max";i:999999999;s:21:"bankcard_withdraw_min";i:3000;s:21:"bankcard_withdraw_max";i:999999999;}`
	service = NewService(fakeUserStore{
		user:        map[string]interface{}{"uid": "5", "recommend_total": "2", "perms": `{"min.withdraw.recommend.num":"2"}`},
		settings:    wideSettings,
		bankcardRow: map[string]interface{}{"cardid": "7", "type": "1"},
		account:     map[string]interface{}{"uid": "5", "available_balance": "800"},
	}, "https://res.example.test")
	retcode, errmsg, err = service.WithdrawCreateEdge(context.Background(), "token", 7, 0, 500000)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "兑换数量100万以上请分次兑换" {
		t.Fatalf("exchange too large retcode=%d errmsg=%q", retcode, errmsg)
	}

	service = NewService(fakeUserStore{
		user:        map[string]interface{}{"uid": "5", "recommend_total": "2", "perms": `{"min.withdraw.recommend.num":"2"}`},
		settings:    baseSettings,
		bankcardRow: map[string]interface{}{"cardid": "7", "type": "1"},
		account:     map[string]interface{}{"uid": "5", "available_balance": "5000"},
	}, "https://res.example.test")
	retcode, errmsg, err = service.WithdrawCreateEdge(context.Background(), "token", 7, 0, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "提现申请成功分支暂未迁移" {
		t.Fatalf("pending retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestWithdrawRule(t *testing.T) {
	service := NewService(fakeUserStore{
		calldata: map[string]map[string]interface{}{
			"withdraw.rule": {"type": "html", "content": " <p>规则</p> "},
		},
	}, "https://res.example.test")

	data, err := service.WithdrawRule(context.Background())
	if err != nil {
		t.Fatalf("withdraw rule: %v", err)
	}
	if data["content"] != "<p>规则</p>" {
		t.Fatalf("unexpected data %#v", data)
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

func TestMsgSendRequiresLogin(t *testing.T) {
	service := NewService(fakeUserStore{}, "https://res.example.test")

	retcode, errmsg, err := service.MsgSend(context.Background(), "", 9, "hello", false)
	if err != nil {
		t.Fatalf("msg send: %v", err)
	}
	if retcode != -9999 || errmsg != "您还没有登录" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestMsgSendValidatesContent(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")

	retcode, errmsg, err := service.MsgSend(context.Background(), "token", 9, "", false)
	if err != nil {
		t.Fatalf("msg send: %v", err)
	}
	if retcode != -1 || errmsg != "请填写信息内容" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestMsgSendWithoutConversationKeepsPHPGroupBug(t *testing.T) {
	service := NewService(fakeUserStore{user: map[string]interface{}{"uid": "5"}}, "https://res.example.test")

	retcode, errmsg, err := service.MsgSend(context.Background(), "token", 0, "hello", true)
	if err != nil {
		t.Fatalf("msg send: %v", err)
	}
	if retcode != -1 || errmsg != "请选择一个用户" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestMsgSendReplySuccess(t *testing.T) {
	sent := map[string]interface{}{}
	store := fakeUserStore{user: map[string]interface{}{"uid": "5"}, sentMessage: sent}
	service := NewService(store, "https://res.example.test")
	service.now = func() time.Time { return time.Unix(1700000000, 0) }

	retcode, errmsg, err := service.MsgSend(context.Background(), "token", 9, "hello", false)
	if err != nil {
		t.Fatalf("msg send: %v", err)
	}
	if retcode != 0 || errmsg != "发送成功" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if sent["senderID"] != 5 || sent["receiverID"] != 7 || sent["content"] != "hello" || sent["cid"] != 9 || sent["now"] != int64(1700000000) {
		t.Fatalf("sent=%#v", sent)
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
