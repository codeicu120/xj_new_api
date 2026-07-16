package ucp

import (
	"context"
	"fmt"
	"strings"
	"time"
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

func (s *Service) VIPPkgCoinOrderEdge(ctx context.Context, token string, pkgID int) (int, string, error) {
	_, retcode, errmsg, err := s.VIPPkgCoinOrder(ctx, token, pkgID)
	return retcode, errmsg, err
}

func (s *Service) BeanPkgCoinOrderEdge(ctx context.Context, token string, pkgID int) (int, string, error) {
	_, retcode, errmsg, err := s.BeanPkgCoinOrder(ctx, token, pkgID)
	return retcode, errmsg, err
}

func (s *Service) VIPPkgCoinOrder(ctx context.Context, token string, pkgID int) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	pkg, err := s.store.PackageByID(ctx, "vip", pkgID)
	if err != nil {
		return nil, -1, "套餐购买失败", err
	}
	if len(pkg) == 0 || atoi(pkg["showtype"]) != 0 {
		return nil, -1, "套餐不存在或未启用", nil
	}
	deductCoin := atoi(pkg["coinprice"])
	dayLen := atoi(pkg["daylen"])
	now := s.now()
	expiry := now.Add(time.Duration(dayLen) * 24 * time.Hour).Unix()
	const superVIPGID = 6
	if atoi(user["sysgid"]) == superVIPGID && atoi64(user["sysgid_exptime"]) > now.Unix() {
		expiry = atoi64(user["sysgid_exptime"]) + int64(dayLen)*86400
	}
	quota, err := s.store.Quota(ctx, uid)
	if err != nil {
		return nil, -1, "套餐购买失败", err
	}
	if atoi(quota["goldcoin"]) < deductCoin {
		return nil, -1, "金币不足，快做任务获取金币吧！", nil
	}
	if err := s.store.UpgradeVIP(ctx, uid, deductCoin, superVIPGID, expiry, now.Unix()); err != nil {
		return nil, -1, "套餐购买失败", err
	}
	return map[string]interface{}{
		"deduct_coin": deductCoin,
		"expiry_date": formatMinuteTime(expiry),
	}, 0, "您已成功升级尊贵会员", nil
}

func (s *Service) BeanPkgCoinOrder(ctx context.Context, token string, pkgID int) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	pkg, err := s.store.PackageByID(ctx, "bean", pkgID)
	if err != nil {
		return nil, -1, "套餐购买失败", err
	}
	if len(pkg) == 0 || atoi(pkg["showtype"]) != 0 {
		return nil, -1, "套餐不存在或未启用", nil
	}
	deductCoin, err := s.coinOrderPrice(ctx, "bean", pkg)
	if err != nil {
		return nil, -1, "套餐购买失败", err
	}
	quota, err := s.store.Quota(ctx, uid)
	if err != nil {
		return nil, -1, "套餐购买失败", err
	}
	if atoi(quota["goldcoin"]) < deductCoin {
		return nil, -1, "金币不足，快做任务获取金币吧！", nil
	}
	if err := s.store.BuyBeansWithCoins(ctx, uid, deductCoin, deductCoin, s.now().Unix()); err != nil {
		return nil, -1, "套餐购买失败", err
	}
	return map[string]interface{}{"deduct_coin": deductCoin}, 0, "您已成功兑换金豆", nil
}

func (s *Service) VIPPkgPlaceOrderEdge(ctx context.Context, token string, pkgID int, paycode string) (int, string, error) {
	return s.pkgPlaceOrderEdge(ctx, token, "vip", pkgID, paycode)
}

func (s *Service) CoinPkgPlaceOrderEdge(ctx context.Context, token string, pkgID int, paycode string) (int, string, error) {
	return s.pkgPlaceOrderEdge(ctx, token, "coin", pkgID, paycode)
}

func (s *Service) BeanPkgPlaceOrderEdge(ctx context.Context, token string, pkgID int, paycode string) (int, string, error) {
	return s.pkgPlaceOrderEdge(ctx, token, "bean", pkgID, paycode)
}

func (s *Service) pkgPlaceOrderEdge(ctx context.Context, token string, kind string, pkgID int, paycode string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	pkg, err := s.store.PackageByID(ctx, kind, pkgID)
	if err != nil {
		return -1, "套餐下单失败", err
	}
	if len(pkg) == 0 || atoi(pkg["showtype"]) != 0 {
		return -1, "套餐不存在或未启用", nil
	}
	if kind == "vip" && atoi(pkg["rmbprice"]) == 3800 {
		return -1, "套餐仅支持金币兑换", nil
	}
	settingRow, err := s.store.SettingByUUID(ctx, "setting")
	if err != nil {
		return -1, "套餐下单失败", err
	}
	setting := parseTaskPHPSerializedMap(str(settingRow["value"]))
	if kind == "vip" || kind == "bean" {
		if retcode, errmsg, blocked, err := s.checkPackageOrderLimits(ctx, kind, user, setting); err != nil || blocked {
			return retcode, errmsg, err
		}
	}
	channels, err := s.store.PaymentChannels(ctx, false)
	if err != nil {
		return -1, "套餐下单失败", err
	}
	paycode = strings.TrimSpace(paycode)
	if isRandomPaycode(paycode) && atoi(setting["randomPaywayStatus"]) != 0 && !randomPaymentCodeAvailable(channels, packagePayType(kind), paycode, atoi(pkg["rmbprice"])) {
		return -1, "该支付方式当前没有可用通道", nil
	}
	if !paymentCodeAllowed(channels, packagePayType(kind), paycode, atoi(pkg["rmbprice"])) {
		return -1, "支付方式错误或不被允许", nil
	}
	return -1, "支付下单成功分支暂未迁移", nil
}

func (s *Service) checkPackageOrderLimits(ctx context.Context, kind string, user map[string]interface{}, setting map[string]interface{}) (int, string, bool, error) {
	uid := atoi(user["uid"])
	now := s.now()
	cdSeconds := atoi(setting["ordercd"]) * 60
	if cdSeconds > 0 {
		unpaid, err := s.store.CountPaymentsByStatusSince(ctx, uid, 0, now.Unix()-int64(cdSeconds))
		if err != nil {
			return -1, "套餐下单失败", true, err
		}
		if unpaid >= atoi(setting["unpaidorders"]) {
			return -1, fmt.Sprintf("你有未支付订单，%d分钟内无法提交新订单", atoi(setting["ordercd"])), true, nil
		}
	}
	if kind == "vip" {
		newUserLimit := true
		successOrders := atoi(setting["successorders"])
		if successOrders > 0 {
			succeeded, err := s.store.CountPaymentsByStatusSince(ctx, uid, 1, 0)
			if err != nil {
				return -1, "套餐下单失败", true, err
			}
			if successOrders <= succeeded {
				newUserLimit = false
			}
		}
		if newUserLimit {
			regDays := atoi(setting["regdays"])
			if regDays > 0 && atoi64(user["regtime"]) > now.Unix()-int64(regDays)*86400 {
				return -1, fmt.Sprintf("系统限制注册%d天后方可充值VIP", regDays), true, nil
			}
			viewVODs := atoi(setting["viewvods"])
			if viewVODs > 0 {
				viewed, err := s.store.CountVODPlayLogsSince(ctx, uid, 0)
				if err != nil {
					return -1, "套餐下单失败", true, err
				}
				if viewed < viewVODs {
					return -1, fmt.Sprintf("系统限制观影%d部后方可充值VIP", viewVODs), true, nil
				}
			}
		}
	}
	orderDayLimit := atoi(setting["orderdaylimit"])
	if orderDayLimit > 0 {
		unpaid, err := s.store.CountPaymentsByStatusSince(ctx, uid, 0, dayStartUnix(now))
		if err != nil {
			return -1, "套餐下单失败", true, err
		}
		if unpaid >= orderDayLimit {
			if kind == "vip" {
				return -1, "您今天购买会员订单数已达上限，请明天再来", true, nil
			}
			return -1, "您今天购买金豆订单数已达上限，请明天再来", true, nil
		}
	}
	return 0, "", false, nil
}

func (s *Service) pkgCoinOrderEdge(ctx context.Context, token string, kind string, pkgID int) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "您还没有登录", nil
	}
	pkg, err := s.store.PackageByID(ctx, kind, pkgID)
	if err != nil {
		return -1, "套餐购买失败", err
	}
	if len(pkg) == 0 || atoi(pkg["showtype"]) != 0 {
		return -1, "套餐不存在或未启用", nil
	}
	coinPrice, err := s.coinOrderPrice(ctx, kind, pkg)
	if err != nil {
		return -1, "套餐购买失败", err
	}
	quota, err := s.store.Quota(ctx, uid)
	if err != nil {
		return -1, "套餐购买失败", err
	}
	if atoi(quota["goldcoin"]) < coinPrice {
		return -1, "金币不足，快做任务获取金币吧！", nil
	}
	return -1, "金币购买成功分支暂未迁移", nil
}

func (s *Service) coinOrderPrice(ctx context.Context, kind string, pkg map[string]interface{}) (int, error) {
	if kind == "vip" {
		return atoi(pkg["coinprice"]), nil
	}
	exrate, err := s.store.SettingExRate(ctx)
	if err != nil {
		return 0, err
	}
	return int(float64(atoi(pkg["rmbprice"])) / 100 * float64(exrate)), nil
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

func packagePayType(kind string) int {
	switch kind {
	case "vip":
		return 1
	case "coin":
		return 2
	case "bean":
		return 3
	default:
		return 0
	}
}

func paymentCodeAllowed(channels []map[string]interface{}, payType int, paycode string, amount int) bool {
	if payType == 0 || paycode == "" {
		return false
	}
	for _, channel := range channels {
		if atoi(channel["disabled"]) > 0 {
			continue
		}
		payways, _ := channel["payways"].([]map[string]interface{})
		for _, payway := range payways {
			if str(payway["paycode"]) != paycode {
				continue
			}
			if !paywayAllowsType(payway, payType) {
				continue
			}
			minAmount := atoi(payway["trxamount_min"])
			maxAmount := atoi(payway["trxamount_max"])
			if minAmount > 0 && amount < minAmount {
				continue
			}
			if maxAmount > 0 && amount > maxAmount {
				continue
			}
			return true
		}
	}
	return false
}

func isRandomPaycode(paycode string) bool {
	return paycode == "all.alipay" || paycode == "all.wechat"
}

func randomPaymentCodeAvailable(channels []map[string]interface{}, payType int, paycode string, amount int) bool {
	for _, channel := range channels {
		if atoi(channel["disabled"]) > 0 {
			continue
		}
		payways, _ := channel["payways"].([]map[string]interface{})
		for _, payway := range payways {
			code := str(payway["paycode"])
			if paycode == "all.alipay" && !strings.Contains(strings.ToLower(code), "alipay") {
				continue
			}
			if paycode == "all.wechat" && !strings.Contains(strings.ToLower(code), "wechat") && !strings.Contains(strings.ToLower(code), "wx") {
				continue
			}
			if !paymentCodeAllowed([]map[string]interface{}{{"payways": []map[string]interface{}{payway}}}, payType, code, amount) {
				continue
			}
			return true
		}
	}
	return false
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
