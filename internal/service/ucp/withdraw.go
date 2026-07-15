package ucp

import (
	"context"
	"strconv"
	"strings"
	"time"

	"xj_comp/internal/domain"
)

func (s *Service) WithdrawIndex(ctx context.Context, token string) (domain.UCPWithdrawIndexData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPWithdrawIndexData{}, -1, "获取提现信息失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPWithdrawIndexData{}, -9999, "您还没有登录", nil
	}

	settingRow, err := s.store.SettingByUUID(ctx, "setting")
	if err != nil {
		return domain.UCPWithdrawIndexData{}, -1, "获取提现信息失败", err
	}
	setting := parseTaskPHPSerializedMap(str(settingRow["value"]))
	gameSettingRow, err := s.store.SettingByUUID(ctx, "game.setting")
	if err != nil {
		return domain.UCPWithdrawIndexData{}, -1, "获取提现信息失败", err
	}
	gameSetting := parseTaskPHPSerializedMap(str(gameSettingRow["value"]))
	account, err := s.store.Account(ctx, uid)
	if err != nil {
		return domain.UCPWithdrawIndexData{}, -1, "获取提现信息失败", err
	}
	cardRows, err := s.store.Bankcards(ctx, uid)
	if err != nil {
		return domain.UCPWithdrawIndexData{}, -1, "获取提现信息失败", err
	}
	quota, err := s.store.Quota(ctx, uid)
	if err != nil {
		return domain.UCPWithdrawIndexData{}, -1, "获取提现信息失败", err
	}

	exrate := atoi(setting["exrate"])
	if exrate == 0 {
		exrate, err = s.store.SettingExRate(ctx)
		if err != nil {
			return domain.UCPWithdrawIndexData{}, -1, "获取提现信息失败", err
		}
	}
	goldCoin := atoi(quota["goldcoin"])
	coin2RMB := 0
	if exrate > 0 {
		coin2RMB = (goldCoin * 100) / exrate
	}
	max2RMB := atoi(account["available_balance"]) + coin2RMB
	gameWithdrawRate, _ := strconv.ParseFloat(str(gameSetting["withdrawrate"]), 64)

	return domain.UCPWithdrawIndexData{
		Account:             processAccountRow(account),
		CardRows:            cardRows,
		GoldCoin:            goldCoin,
		ExRate:              exrate,
		TopupMin:            formatRMB(atoi(setting["topupmin"])),
		Coin2RMB:            formatRMB(coin2RMB),
		Max2RMB:             formatRMB(max2RMB),
		GameWithdrawMin:     atoi(gameSetting["withdrawmin"]),
		GameWithdrawRate:    gameWithdrawRate,
		AlipayWithdrawMin:   atoi(setting["alipay_withdraw_min"]),
		AlipayWithdrawMax:   atoi(setting["alipay_withdraw_max"]),
		BankcardWithdrawMin: atoi(setting["bankcard_withdraw_min"]),
		BankcardWithdrawMax: atoi(setting["bankcard_withdraw_max"]),
	}, 0, "", nil
}

func (s *Service) WithdrawListing(ctx context.Context, token string, page int) (domain.UCPWithdrawListingData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPWithdrawListingData{}, -1, "获取提现列表失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPWithdrawListingData{}, -9999, "您还没有登录", nil
	}
	const pageSize = 20
	total, err := s.store.CountWithdraws(ctx, uid)
	if err != nil {
		return domain.UCPWithdrawListingData{}, -1, "获取提现列表失败", err
	}
	page = normalizePage(total, pageSize, page)
	rows, err := s.store.Withdraws(ctx, uid, page, pageSize)
	if err != nil {
		return domain.UCPWithdrawListingData{}, -1, "获取提现列表失败", err
	}
	withdrawTotal, err := s.store.SumWithdrawAmount(ctx, uid)
	if err != nil {
		return domain.UCPWithdrawListingData{}, -1, "获取提现列表失败", err
	}
	return domain.UCPWithdrawListingData{
		Rows:          processWithdrawRows(rows),
		WithdrawTotal: formatRMB(withdrawTotal),
		PageInfo:      pageInfo(total, pageSize, page, "/ucp/withdraw/listing?page=[?]"),
	}, 0, "", nil
}

func (s *Service) WithdrawRule(ctx context.Context) (map[string]interface{}, error) {
	row, err := s.store.CalldataByUUID(ctx, "withdraw.rule")
	if err != nil {
		return nil, err
	}
	content := ""
	if str(row["type"]) == "html" {
		content = strings.TrimSpace(str(row["content"]))
	}
	return map[string]interface{}{"content": content}, nil
}

func processWithdrawRows(rows []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"wdid":            str(row["wdid"]),
			"uid":             str(row["uid"]),
			"username":        str(row["username"]),
			"wdtype":          str(row["wdtype"]),
			"withdraw_amount": formatRMB(atoi(row["withdraw_amount"])),
			"remit_amount":    formatRMB(atoi(row["remit_amount"])),
			"createtime":      formatWithdrawTime(atoi64(row["createtime"])),
			"lastupdate":      formatWithdrawTime(atoi64(row["lastupdate"])),
			"name":            str(row["name"]),
			"cardnum":         str(row["cardnum"]),
			"bankname":        str(row["bankname"]),
			"errmsg":          str(row["errmsg"]),
			"wdstatus":        str(row["wdstatus"]),
			"checkstatus":     str(row["checkstatus"]),
		})
	}
	return out
}

func formatWithdrawTime(ts int64) string {
	if ts <= 0 {
		return time.Unix(0, 0).Format("2006-01-02 15:04:05")
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}
