package ucp

import (
	"context"
	"fmt"
	"time"

	"xj_comp/internal/domain"
)

var balanceLogPayTypeNames = map[int]string{
	0:  "其它类型",
	1:  "系统增加",
	2:  "系统扣减",
	3:  "充值",
	4:  "提现",
	5:  "转账",
	6:  "官方服务",
	7:  "购买金币",
	8:  "购买套餐",
	9:  "金币转人民币",
	10: "人民币转金币",
	11: "人工增加",
	12: "人工扣减",
	13: "游戏充值",
	14: "游戏提现",
	15: "游戏划拨",
	16: "游戏人工增加",
	17: "游戏人工扣减",
	21: "购买金豆",
}

func (s *Service) AccountIndex(ctx context.Context, token string) (domain.UCPAccountIndexData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPAccountIndexData{}, -1, "获取用户失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPAccountIndexData{}, -9999, "您还没有登录", nil
	}

	exrate, err := s.store.SettingExRate(ctx)
	if err != nil {
		return domain.UCPAccountIndexData{}, -1, "获取账户信息失败", err
	}
	account, err := s.store.Account(ctx, uid)
	if err != nil {
		return domain.UCPAccountIndexData{}, -1, "获取账户信息失败", err
	}
	quota, err := s.store.Quota(ctx, uid)
	if err != nil {
		return domain.UCPAccountIndexData{}, -1, "获取账户信息失败", err
	}
	logRows, err := s.store.BalanceLogs(ctx, uid, 1, 10)
	if err != nil {
		return domain.UCPAccountIndexData{}, -1, "获取账户信息失败", err
	}

	goldCoin := atoi(quota["goldcoin"])
	coin2RMB := 0
	if exrate > 0 {
		coin2RMB = (goldCoin * 100) / exrate
	}
	max2RMB := atoi(account["available_balance"]) + coin2RMB

	return domain.UCPAccountIndexData{
		Account:  processAccountRow(account),
		GoldCoin: goldCoin,
		ExRate:   exrate,
		Coin2RMB: formatRMB(coin2RMB),
		Max2RMB:  formatRMB(max2RMB),
		LogRows:  processBalanceLogRows(logRows, s.now()),
	}, 0, "", nil
}

func (s *Service) BalanceLog(ctx context.Context, token string, page int) (domain.UCPBalanceLogData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPBalanceLogData{}, -1, "获取用户失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPBalanceLogData{}, -9999, "您还没有登录", nil
	}

	pageSize := 20
	total, err := s.store.CountBalanceLogs(ctx, uid)
	if err != nil {
		return domain.UCPBalanceLogData{}, -1, "获取余额日志失败", err
	}
	page = normalizePage(total, pageSize, page)
	rows, err := s.store.BalanceLogs(ctx, uid, page, pageSize)
	if err != nil {
		return domain.UCPBalanceLogData{}, -1, "获取余额日志失败", err
	}
	return domain.UCPBalanceLogData{
		LogRows:  processBalanceLogRows(rows, s.now()),
		PageInfo: pageInfo(total, pageSize, page, "/ucp/account/balancelog?page=[?]"),
	}, 0, "", nil
}

func processAccountRow(row map[string]interface{}) map[string]interface{} {
	if len(row) == 0 {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"uid":                    str(row["uid"]),
		"balance":                formatRMB(atoi(row["balance"])),
		"frozen":                 formatRMB(atoi(row["frozen"])),
		"deposit":                formatRMB(atoi(row["deposit"])),
		"game_balance":           atoi(row["game_balance"]),
		"game_frozen":            atoi(row["game_frozen"]),
		"available_balance":      formatRMB(atoi(row["available_balance"])),
		"game_available_balance": atoi(row["game_available_balance"]),
	}
}

func processBalanceLogRows(rows []map[string]interface{}, now time.Time) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		paytype := mapString(balanceLogPayTypeNames, atoi(row["paytype"]))
		if paytype == nil {
			paytype = "--"
		}
		out = append(out, map[string]interface{}{
			"trxid":   str(row["trxid"]),
			"paytype": paytype,
			"uid":     str(row["uid"]),
			"trxin":   formatRMB(atoi(row["trxin"])),
			"trxout":  formatRMB(atoi(row["trxout"])),
			"balance": formatRMB(atoi(row["balance"])),
			"trxtime": formatBalanceLogTime(atoi64(row["trxtime"]), now),
			"remark":  str(row["remark"]),
		})
	}
	return out
}

func formatBalanceLogTime(ts int64, now time.Time) string {
	nowSec := now.Unix()
	if ts > nowSec-86400*30 {
		return formatAgo(nowSec - ts)
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return time.Unix(ts, 0).In(loc).Format("2006-01-02")
}

func formatAgo(seconds int64) string {
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
		return fmt.Sprintf("%d天前", days)
	case hours > 0:
		return fmt.Sprintf("%d小时前", hours)
	case minutes > 0:
		return fmt.Sprintf("%d分钟前", minutes)
	default:
		return fmt.Sprintf("%d秒前", seconds)
	}
}
