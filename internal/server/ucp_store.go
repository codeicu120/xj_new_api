package server

import (
	"context"

	"xj_comp/internal/domain"
	indexRepo "xj_comp/internal/repository/index"
	ucpRepo "xj_comp/internal/repository/ucp"
	userRepo "xj_comp/internal/repository/user"
)

type ucpStore struct {
	user  *userRepo.Repository
	ucp   *ucpRepo.Repository
	index *indexRepo.SettingsRepository
}

func (s ucpStore) UserBySession(ctx context.Context, sid string) (map[string]interface{}, error) {
	return s.user.UserBySession(ctx, sid)
}

func (s ucpStore) UserByID(ctx context.Context, uid int) (map[string]interface{}, error) {
	return s.user.UserByID(ctx, uid)
}

func (s ucpStore) BotByID(ctx context.Context, uid int) (map[string]interface{}, error) {
	return s.user.BotByID(ctx, uid)
}

func (s ucpStore) Bankcards(ctx context.Context, uid int) ([]map[string]interface{}, error) {
	return s.ucp.Bankcards(ctx, uid)
}

func (s ucpStore) Banks(ctx context.Context) ([]map[string]interface{}, error) {
	return s.ucp.Banks(ctx)
}

func (s ucpStore) BankcardByID(ctx context.Context, uid int, cardID int) (map[string]interface{}, error) {
	return s.ucp.BankcardByID(ctx, uid, cardID)
}

func (s ucpStore) CreateBankcard(ctx context.Context, uid int, name string, bankname string, cardnum string, isdef int, cardType int) (int, error) {
	return s.ucp.CreateBankcard(ctx, uid, name, bankname, cardnum, isdef, cardType)
}

func (s ucpStore) UpdateBankcard(ctx context.Context, uid int, cardID int, name string, bankname string, cardnum string, isdef int, cardType int) (int, error) {
	return s.ucp.UpdateBankcard(ctx, uid, cardID, name, bankname, cardnum, isdef, cardType)
}

func (s ucpStore) DeleteBankcard(ctx context.Context, uid int, cardID int) (int, error) {
	return s.ucp.DeleteBankcard(ctx, uid, cardID)
}

func (s ucpStore) SetDefaultBankcard(ctx context.Context, uid int, cardID int) error {
	return s.ucp.SetDefaultBankcard(ctx, uid, cardID)
}

func (s ucpStore) Groups(ctx context.Context) ([]map[string]interface{}, error) {
	return s.user.Groups(ctx)
}

func (s ucpStore) CountRecommended(ctx context.Context, uid int) (int, error) {
	return s.user.CountRecommended(ctx, uid)
}

func (s ucpStore) RecommendedUsers(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	return s.user.RecommendedUsers(ctx, uid, page, pageSize)
}

func (s ucpStore) RollTitles(ctx context.Context) ([]map[string]interface{}, error) {
	return s.ucp.RollTitles(ctx)
}

func (s ucpStore) Posters(ctx context.Context) ([]map[string]interface{}, error) {
	return s.ucp.Posters(ctx)
}

func (s ucpStore) Taskboxes(ctx context.Context) ([]map[string]interface{}, error) {
	return s.ucp.Taskboxes(ctx)
}

func (s ucpStore) TaskboxLog(ctx context.Context, uid int, taskID int, dayKey int) (map[string]interface{}, error) {
	return s.ucp.TaskboxLog(ctx, uid, taskID, dayKey)
}

func (s ucpStore) TaskboxCompletedLogs(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	return s.ucp.TaskboxCompletedLogs(ctx, limit)
}

func (s ucpStore) CountTaskboxLogs(ctx context.Context, uid int) (int, error) {
	return s.ucp.CountTaskboxLogs(ctx, uid)
}

func (s ucpStore) TaskboxLogs(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	return s.ucp.TaskboxLogs(ctx, uid, page, pageSize)
}

func (s ucpStore) CountPayments(ctx context.Context, uid int) (int, error) {
	return s.ucp.CountPayments(ctx, uid)
}

func (s ucpStore) Payments(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	return s.ucp.Payments(ctx, uid, page, pageSize)
}

func (s ucpStore) SafePayLogs(ctx context.Context, uid int, since int64, limit int) ([]map[string]interface{}, error) {
	return s.ucp.SafePayLogs(ctx, uid, since, limit)
}

func (s ucpStore) PaymentsSince(ctx context.Context, uid int, since int64, limit int) ([]map[string]interface{}, error) {
	return s.ucp.PaymentsSince(ctx, uid, since, limit)
}

func (s ucpStore) Account(ctx context.Context, uid int) (map[string]interface{}, error) {
	return s.ucp.Account(ctx, uid)
}

func (s ucpStore) Quota(ctx context.Context, uid int) (map[string]interface{}, error) {
	return s.ucp.Quota(ctx, uid)
}

func (s ucpStore) Goldbean(ctx context.Context, uid int) (map[string]interface{}, error) {
	return s.ucp.Goldbean(ctx, uid)
}

func (s ucpStore) CountVODPlayLogsSince(ctx context.Context, uid int, since int64) (int, error) {
	return s.ucp.CountVODPlayLogsSince(ctx, uid, since)
}

func (s ucpStore) CountVODDownLogsSince(ctx context.Context, uid int, since int64) (int, error) {
	return s.ucp.CountVODDownLogsSince(ctx, uid, since)
}

func (s ucpStore) GuestBySID(ctx context.Context, sid string) (map[string]interface{}, error) {
	return s.ucp.GuestBySID(ctx, sid)
}

func (s ucpStore) CountGuestVODPlayLogsSince(ctx context.Context, sid string, since int64) (int, error) {
	return s.ucp.CountGuestVODPlayLogsSince(ctx, sid, since)
}

func (s ucpStore) CountGuestVODDownLogsSince(ctx context.Context, sid string, since int64) (int, error) {
	return s.ucp.CountGuestVODDownLogsSince(ctx, sid, since)
}

func (s ucpStore) CountMiniVODViewLogsSince(ctx context.Context, uid int, since int64, action int) (int, error) {
	return s.ucp.CountMiniVODViewLogsSince(ctx, uid, since, action)
}

func (s ucpStore) CountGuestMiniVODViewLogsSince(ctx context.Context, sid string, since int64, action int) (int, error) {
	return s.ucp.CountGuestMiniVODViewLogsSince(ctx, sid, since, action)
}

func (s ucpStore) CountCoinLogsSinceByType(ctx context.Context, uid int, coinType int, since int64) (int, error) {
	return s.ucp.CountCoinLogsSinceByType(ctx, uid, coinType, since)
}

func (s ucpStore) SumCoinLogsSinceByType(ctx context.Context, uid int, coinType int, since int64) (int, error) {
	return s.ucp.SumCoinLogsSinceByType(ctx, uid, coinType, since)
}

func (s ucpStore) CountVODCommentsSince(ctx context.Context, uid int, since int64, unique bool) (int, error) {
	return s.ucp.CountVODCommentsSince(ctx, uid, since, unique)
}

func (s ucpStore) CountVODFavoritesSince(ctx context.Context, uid int, since int64) (int, error) {
	return s.ucp.CountVODFavoritesSince(ctx, uid, since)
}

func (s ucpStore) CountFeedbacks(ctx context.Context, uid int) (int, error) {
	return s.ucp.CountFeedbacks(ctx, uid)
}

func (s ucpStore) Feedbacks(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	return s.ucp.Feedbacks(ctx, uid, page, pageSize)
}

func (s ucpStore) CountFeedbacksByType(ctx context.Context, uid int, feedbackType int) (int, error) {
	return s.ucp.CountFeedbacksByType(ctx, uid, feedbackType)
}

func (s ucpStore) FeedbacksByType(ctx context.Context, uid int, feedbackType int, page int, pageSize int) ([]map[string]interface{}, error) {
	return s.ucp.FeedbacksByType(ctx, uid, feedbackType, page, pageSize)
}

func (s ucpStore) FeedbackByID(ctx context.Context, id int) (map[string]interface{}, error) {
	return s.ucp.FeedbackByID(ctx, id)
}

func (s ucpStore) CountFeedbacksSince(ctx context.Context, uid int, since int64) (int, error) {
	return s.ucp.CountFeedbacksSince(ctx, uid, since)
}

func (s ucpStore) CreateFeedback(ctx context.Context, input domain.FeedbackCreateInput) (int, error) {
	return s.ucp.CreateFeedback(ctx, input)
}

func (s ucpStore) PaymentByID(ctx context.Context, payid int) (map[string]interface{}, error) {
	return s.ucp.PaymentByID(ctx, payid)
}

func (s ucpStore) AttachByIDs(ctx context.Context, ids []int) ([]map[string]interface{}, error) {
	return s.ucp.AttachByIDs(ctx, ids)
}

func (s ucpStore) CountMsgConversations(ctx context.Context, uid int) (int, error) {
	return s.ucp.CountMsgConversations(ctx, uid)
}

func (s ucpStore) MsgConversations(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	return s.ucp.MsgConversations(ctx, uid, page, pageSize)
}

func (s ucpStore) MsgConversation(ctx context.Context, uid int, cid int) (map[string]interface{}, error) {
	return s.ucp.MsgConversation(ctx, uid, cid)
}

func (s ucpStore) CountMessages(ctx context.Context, uid int, cid int) (int, error) {
	return s.ucp.CountMessages(ctx, uid, cid)
}

func (s ucpStore) Messages(ctx context.Context, uid int, cid int, page int, pageSize int) ([]map[string]interface{}, error) {
	return s.ucp.Messages(ctx, uid, cid, page, pageSize)
}

func (s ucpStore) SetMsgRead(ctx context.Context, uid int, cid int) error {
	return s.ucp.SetMsgRead(ctx, uid, cid)
}

func (s ucpStore) CleanMsgRead(ctx context.Context, uid int) error {
	return s.ucp.CleanMsgRead(ctx, uid)
}

func (s ucpStore) DeleteMsgConversations(ctx context.Context, uid int, cids []int) error {
	return s.ucp.DeleteMsgConversations(ctx, uid, cids)
}

func (s ucpStore) SendMessage(ctx context.Context, senderID int, receiverID int, content string, cid int, now int64) (int, error) {
	return s.ucp.SendMessage(ctx, senderID, receiverID, content, cid, now)
}

func (s ucpStore) BalanceLogs(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	return s.ucp.BalanceLogs(ctx, uid, page, pageSize)
}

func (s ucpStore) CoinLogs(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	return s.ucp.CoinLogs(ctx, uid, page, pageSize)
}

func (s ucpStore) CountCoinLogsByTypes(ctx context.Context, uid int, coinTypes []int) (int, error) {
	return s.ucp.CountCoinLogsByTypes(ctx, uid, coinTypes)
}

func (s ucpStore) CoinLogsByTypes(ctx context.Context, uid int, coinTypes []int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error) {
	return s.ucp.CoinLogsByTypes(ctx, uid, coinTypes, page, pageSize, orderBy)
}

func (s ucpStore) CoinBonusStats(ctx context.Context, uid int) (map[string]interface{}, error) {
	return s.ucp.CoinBonusStats(ctx, uid)
}

func (s ucpStore) CountBalanceLogs(ctx context.Context, uid int) (int, error) {
	return s.ucp.CountBalanceLogs(ctx, uid)
}

func (s ucpStore) SettingExRate(ctx context.Context) (int, error) {
	return s.ucp.SettingExRate(ctx)
}

func (s ucpStore) SettingByUUID(ctx context.Context, uuid string) (map[string]interface{}, error) {
	return s.index.SettingByUUID(ctx, uuid)
}

func (s ucpStore) CalldataByUUID(ctx context.Context, uuid string) (map[string]interface{}, error) {
	return s.index.CalldataByUUID(ctx, uuid)
}

func (s ucpStore) PackageRows(ctx context.Context, kind string) ([]map[string]interface{}, error) {
	return s.ucp.PackageRows(ctx, kind)
}

func (s ucpStore) PaymentChannels(ctx context.Context, gameOnly bool) ([]map[string]interface{}, error) {
	return s.ucp.PaymentChannels(ctx, gameOnly)
}

func (s ucpStore) CountVODOrders(ctx context.Context, uid int, status *int) (int, error) {
	return s.ucp.CountVODOrders(ctx, uid, status)
}

func (s ucpStore) VODOrders(ctx context.Context, uid int, status *int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error) {
	return s.ucp.VODOrders(ctx, uid, status, page, pageSize, orderBy)
}

func (s ucpStore) LatestVODIssue(ctx context.Context) (map[string]interface{}, error) {
	return s.ucp.LatestVODIssue(ctx)
}

func (s ucpStore) CountVODOrdersByCreateTime(ctx context.Context, start int64, end int64) (int, error) {
	return s.ucp.CountVODOrdersByCreateTime(ctx, start, end)
}

func (s ucpStore) VODOrdersByCreateTime(ctx context.Context, start int64, end int64, page int, pageSize int) ([]map[string]interface{}, error) {
	return s.ucp.VODOrdersByCreateTime(ctx, start, end, page, pageSize)
}

func (s ucpStore) SumVODOrderCoins(ctx context.Context, uid int, status int) (int, error) {
	return s.ucp.SumVODOrderCoins(ctx, uid, status)
}

func (s ucpStore) CountVODSupports(ctx context.Context, uid int) (int, error) {
	return s.ucp.CountVODSupports(ctx, uid)
}

func (s ucpStore) VODSupports(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	return s.ucp.VODSupports(ctx, uid, page, pageSize)
}

func (s ucpStore) MaxVODSupport(ctx context.Context, orderID int) (map[string]interface{}, error) {
	return s.ucp.MaxVODSupport(ctx, orderID)
}

func (s ucpStore) MyVODSupportCoins(ctx context.Context, orderID int, uid int) (int, error) {
	return s.ucp.MyVODSupportCoins(ctx, orderID, uid)
}

func (s ucpStore) SumVODSupportCoins(ctx context.Context, uid int, onlyFrozen bool) (int, error) {
	return s.ucp.SumVODSupportCoins(ctx, uid, onlyFrozen)
}
