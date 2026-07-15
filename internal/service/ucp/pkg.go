package ucp

import (
	"context"
	"fmt"
)

func (s *Service) VIPPkgIndex(ctx context.Context, token string) (map[string]interface{}, int, string, error) {
	return s.pkgIndex(ctx, token, "vip", 1, false)
}

func (s *Service) CoinPkgIndex(ctx context.Context, token string) (map[string]interface{}, int, string, error) {
	return s.pkgIndex(ctx, token, "coin", 2, true)
}

func (s *Service) BeanPkgIndex(ctx context.Context, token string) (map[string]interface{}, int, string, error) {
	return s.pkgIndex(ctx, token, "bean", 3, false)
}

func (s *Service) pkgIndex(ctx context.Context, token string, kind string, payType int, gameOnly bool) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	rows, err := s.store.PackageRows(ctx, kind)
	if err != nil {
		return nil, -1, "获取套餐失败", err
	}
	channels, err := s.store.PaymentChannels(ctx, gameOnly)
	if err != nil {
		return nil, -1, "获取支付方式失败", err
	}
	settingRow, err := s.store.SettingByUUID(ctx, "setting")
	if err != nil {
		return nil, -1, "获取支付方式失败", err
	}
	setting := parseTaskPHPSerializedMap(str(settingRow["value"]))
	return map[string]interface{}{
		"pkgrows":    processPackageRows(kind, rows),
		"payments":   filterPaymentChannels(channels, payType),
		"safepayurl": str(setting["safepayurl"]),
	}, 0, "", nil
}

func processPackageRows(kind string, rows []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		item := map[string]interface{}{
			"pkgid":          row["pkgid"],
			"pkgname":        row["pkgname"],
			"memo":           row["memo"],
			"showtype":       atoi(row["showtype"]),
			"rmbprice":       rmbString(row["rmbprice"]),
			"recommend":      atoi(row["recommend"]),
			"bonus_vip_days": atoi(row["bonus_vip_days"]),
		}
		if kind == "vip" {
			item["daylen"] = atoi(row["daylen"])
			item["coinprice"] = atoi(row["coinprice"])
		} else {
			item["bonus_coins"] = atoi(row["bonus_coins"])
		}
		out = append(out, item)
	}
	return out
}

func filterPaymentChannels(channels []map[string]interface{}, payType int) []map[string]interface{} {
	out := []map[string]interface{}{}
	for _, channel := range channels {
		if atoi(channel["disabled"]) > 0 {
			continue
		}
		payways, _ := channel["payways"].([]map[string]interface{})
		filtered := []map[string]interface{}{}
		for _, payway := range payways {
			if !paywayAllowsType(payway, payType) {
				continue
			}
			filtered = append(filtered, map[string]interface{}{
				"payname":       payway["payname"],
				"paylogo":       payway["paylogo"],
				"dscr":          payway["dscr"],
				"paycode":       payway["paycode"],
				"trxamount_min": atoi(payway["trxamount_min"]),
				"trxamount_max": atoi(payway["trxamount_max"]),
				"extras":        mapOrEmpty(payway["extras"]),
			})
		}
		if len(filtered) == 0 {
			continue
		}
		item := map[string]interface{}{
			"channame": channel["channame"],
			"chanlogo": channel["chanlogo"],
			"dscr":     channel["dscr"],
			"payways":  filtered,
		}
		for _, key := range []string{"appId", "appSecret", "appKey", "notifyUrl"} {
			if value := str(channel[key]); value != "" {
				item[key] = value
			}
		}
		out = append(out, item)
	}
	return out
}

func paywayAllowsType(payway map[string]interface{}, payType int) bool {
	allow, ok := payway["allow_paytypes"].(map[int][]string)
	if !ok || len(allow) == 0 {
		return true
	}
	platforms, ok := allow[payType]
	if !ok {
		return false
	}
	for _, platform := range platforms {
		if platform == "ALL" || platform == "" {
			return true
		}
	}
	return false
}

func mapOrEmpty(value interface{}) interface{} {
	if value == nil {
		return map[string]interface{}{}
	}
	return value
}

func rmbString(value interface{}) string {
	return fmt.Sprintf("%.2f", float64(atoi(value))/100)
}
