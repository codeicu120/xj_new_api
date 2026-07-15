package ucp

import (
	"context"
	"strconv"

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
