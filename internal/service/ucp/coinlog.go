package ucp

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"xj_comp/internal/domain"
)

var coinLogTypeNames = map[int]string{
	0:   "未知类型进账",
	1:   "签到送金币",
	2:   "分享送金币",
	3:   "评论送金币",
	4:   "收藏送金币",
	5:   "观影送金币",
	6:   "保存二维码送金币",
	7:   "点击广告送金币",
	8:   "人民币转金币",
	9:   "注册送金币",
	10:  "注册本人奖",
	11:  "注册一级奖",
	12:  "注册二级奖",
	13:  "注册三级奖",
	14:  "活跃本人奖",
	15:  "活跃一级奖",
	16:  "活跃二级奖",
	17:  "活跃三级奖",
	18:  "人工增加",
	19:  "宝箱收入",
	20:  "购买金币",
	21:  "投币收入",
	22:  "视频下载奖",
	23:  "激励视频奖",
	24:  "连续签到奖",
	25:  "小视频任务奖",
	32:  "邀请送金币",
	100: "未知扣减",
	101: "观影扣减",
	102: "下载扣减",
	103: "升级尊贵扣减",
	104: "金币转人民币",
	105: "人工扣减",
	106: "投币支出",
	201: "邀请好友赠送vip天数",
}

var (
	phoneWithCountryPattern = regexp.MustCompile(`^86\.(\d{3})(\d{4})(\d{4})$`)
	phonePattern            = regexp.MustCompile(`^(\d{3})(\d{4})(\d{4})$`)
)

var coinLogBonusListTypes = []int{0, 1, 9, 2, 3, 4, 5, 6, 7, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 32, 22}

func (s *Service) CoinLogIndex(ctx context.Context, token string) (domain.UCPCoinLogIndexData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPCoinLogIndexData{}, -1, "获取用户失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPCoinLogIndexData{}, -9999, "您还没有登录", nil
	}

	account, err := s.store.Account(ctx, uid)
	if err != nil {
		return domain.UCPCoinLogIndexData{}, -1, "获取金币日志失败", err
	}
	quota, err := s.store.Quota(ctx, uid)
	if err != nil {
		return domain.UCPCoinLogIndexData{}, -1, "获取金币日志失败", err
	}
	exrate, err := s.store.SettingExRate(ctx)
	if err != nil {
		return domain.UCPCoinLogIndexData{}, -1, "获取金币日志失败", err
	}
	logRows, err := s.store.CoinLogs(ctx, uid, 1, 10)
	if err != nil {
		return domain.UCPCoinLogIndexData{}, -1, "获取金币日志失败", err
	}

	return domain.UCPCoinLogIndexData{
		Account:  processAccountRow(account),
		GoldCoin: atoi(quota["goldcoin"]),
		ExRate:   exrate,
		LogRows:  processCoinLogRows(logRows, s.now()),
	}, 0, "", nil
}

func (s *Service) CoinLogBonusLog(ctx context.Context, token string, page int) (domain.UCPCoinLogBonusData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPCoinLogBonusData{}, -1, "获取用户失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPCoinLogBonusData{}, -9999, "您还没有登录", nil
	}

	pageSize := 20
	total, err := s.store.CountCoinLogsByTypes(ctx, uid, coinLogBonusListTypes)
	if err != nil {
		return domain.UCPCoinLogBonusData{}, -1, "获取金币日志失败", err
	}
	page = normalizePage(total, pageSize, page)
	rows, err := s.store.CoinLogsByTypes(ctx, uid, coinLogBonusListTypes, page, pageSize, "addtime DESC")
	if err != nil {
		return domain.UCPCoinLogBonusData{}, -1, "获取金币日志失败", err
	}
	addInfo, err := s.store.CoinBonusStats(ctx, uid)
	if err != nil {
		return domain.UCPCoinLogBonusData{}, -1, "获取金币日志失败", err
	}

	return domain.UCPCoinLogBonusData{
		LogRows:  processCoinLogRows(rows, s.now()),
		AddInfo:  processCoinBonusStats(addInfo),
		PageInfo: pageInfo(total, pageSize, page, "/ucp/coinlog/bonuslog?page=[?]"),
	}, 0, "", nil
}

func (s *Service) CoinLogInviteLog(ctx context.Context, token string, page int) (domain.UCPCoinLogListingData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPCoinLogListingData{}, -1, "获取用户失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPCoinLogListingData{}, -9999, "您还没有登录", nil
	}

	pageSize := 20
	coinTypes := []int{201, 32, 11}
	total, err := s.store.CountCoinLogsByTypes(ctx, uid, coinTypes)
	if err != nil {
		return domain.UCPCoinLogListingData{}, -1, "获取金币日志失败", err
	}
	page = normalizePage(total, pageSize, page)
	rows, err := s.store.CoinLogsByTypes(ctx, uid, coinTypes, page, pageSize, "addtime DESC")
	if err != nil {
		return domain.UCPCoinLogListingData{}, -1, "获取金币日志失败", err
	}

	return domain.UCPCoinLogListingData{
		LogRows:  processCoinLogRows(rows, s.now()),
		PageInfo: pageInfo(total, pageSize, page, "/ucp/coinlog/invitelog?page=[?]"),
	}, 0, "", nil
}

func processCoinBonusStats(row map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"inviteTotal": atoi(row["inviteTotal"]),
		"activeTotal": atoi(row["activeTotal"]),
		"bonusTotal":  atoi(row["bonusTotal"]),
	}
}

func processCoinLogRows(rows []map[string]interface{}, now time.Time) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		cointype := mapString(coinLogTypeNames, atoi(row["cointype"]))
		if cointype == nil {
			cointype = "--"
		}
		out = append(out, map[string]interface{}{
			"logid":            str(row["logid"]),
			"uid":              str(row["uid"]),
			"cointype":         cointype,
			"coinnum":          str(row["coinnum"]),
			"balance":          str(row["balance"]),
			"addtime":          formatCoinLogTime(atoi64(row["addtime"]), now),
			"remark":           str(row["remark"]),
			"invited_uid":      str(row["invited_uid"]),
			"mobi":             maskPhone(row["mobi"]),
			"invited_username": str(row["invited_username"]),
		})
	}
	return out
}

func formatCoinLogTime(ts int64, now time.Time) string {
	nowSec := now.Unix()
	if ts > nowSec-86400*30 {
		return formatCoinAgo(nowSec - ts)
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return time.Unix(ts, 0).In(loc).Format("2006-01-02")
}

func formatCoinAgo(seconds int64) string {
	if seconds < 0 {
		seconds = 0
	}
	days := seconds / 86400
	seconds -= days * 86400
	hours := seconds / 3600
	seconds -= hours * 3600
	minutes := seconds / 60
	seconds -= minutes * 60
	switch {
	case days > 0:
		return fmt.Sprintf("%d天前 ", days)
	case hours > 0:
		return fmt.Sprintf("%d小时前 ", hours)
	case minutes > 0:
		return fmt.Sprintf("%d分钟前 ", minutes)
	default:
		return fmt.Sprintf("%d秒前", seconds)
	}
}

func maskPhone(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	phone := str(value)
	if matches := phoneWithCountryPattern.FindStringSubmatch(phone); len(matches) == 4 {
		return "86." + matches[1] + "****" + matches[3]
	}
	if matches := phonePattern.FindStringSubmatch(phone); len(matches) == 4 {
		return matches[1] + "****" + matches[3]
	}
	return phone
}
