package server

import (
	"context"

	ucpRepo "xj_comp/internal/repository/ucp"
	userRepo "xj_comp/internal/repository/user"
)

type ucpStore struct {
	user *userRepo.Repository
	ucp  *ucpRepo.Repository
}

func (s ucpStore) UserBySession(ctx context.Context, sid string) (map[string]interface{}, error) {
	return s.user.UserBySession(ctx, sid)
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
