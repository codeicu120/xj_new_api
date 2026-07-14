package ucp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
)

var paymentTypeNames = map[int]string{
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
	18: "游戏其它类型",
	19: "游戏系统增加",
	20: "游戏系统减少",
	21: "购买金豆",
}

var paymentWayNames = map[string]string{
	"walletpay": "余额支付",
	"safepay":   "人工代付",
	"wxpay":     "微信支付",
	"alipay":    "支付宝",
	"shangfu":   "shangfu",
	"wappay1":   "wappay1",
	"wappay2":   "wappay2",
	"wappay3":   "wappay3",
	"wappay4":   "wappay4",
	"wappay5":   "wappay5",
	"hawpay":    "hawpay",
	"easypay":   "easypay",
	"chan1":     "chan1",
	"pay6":      "pay6",
	"pay7":      "pay7",
	"pay8":      "pay8",
	"pay9":      "pay9",
	"pay10":     "pay10",
	"pay10a":    "pay10a",
	"pay11":     "pay11",
	"pay12":     "pay12",
	"pay13":     "pay13",
	"pay14":     "pay14",
	"pay15":     "pay15",
	"gpay1":     "gpay1",
	"gpay2":     "gpay2",
	"newpay":    "newpay",
	"newpayff":  "newpayff",
	"newpayxxx": "newpayxxx",
	"newpayqk":  "newpayqk",
}

var paymentCodeNames = map[string]string{
	"walletpay.walletpay": "余额支付",
	"safepay.safepay":     "客服支付",
	"shangfu.alipay_wap":  "支付宝",
	"shangfu.alipay_scan": "支付宝扫码",
	"shangfu.union_wap":   "云闪付",
	"wappay1.ali_jyes":    "支付宝jyes",
	"wappay1.wx_jyes":     "微信支付jyes",
	"wappay1.ali_nxys":    "支付宝nxys",
	"wappay1.wx_nxys":     "微信支付nxys",
	"wappay1.ali_gpay":    "支付宝普通红包",
	"wappay1.ali_bank":    "支付宝转卡",
	"wappay1.wx_fix_v2":   "微信固码V2",
	"wappay1.ali_gm":      "支付宝个码",
	"wappay1.union_wap":   "云闪付",
	"wappay2.1":           "支付宝",
	"wappay2.2":           "微信支付",
	"wappay2.3":           "银联卡",
	"wappay3.1":           "支付宝",
	"newpay.unionpayapi":  "银联卡",
	"newpayff.65":         "银行卡转账",
	"newpayxxx.2":         "支付宝",
	"newpayqk.31":         "支付宝",
}

func (s *Service) PaymentListing(ctx context.Context, token string, page int) (domain.UCPPaymentListingData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPPaymentListingData{}, -1, "获取用户失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPPaymentListingData{}, -9999, "您还没有登录", nil
	}

	pageSize := 20
	total, err := s.store.CountPayments(ctx, uid)
	if err != nil {
		return domain.UCPPaymentListingData{}, -1, "获取支付记录失败", err
	}
	page = normalizePage(total, pageSize, page)
	rows, err := s.store.Payments(ctx, uid, page, pageSize)
	if err != nil {
		return domain.UCPPaymentListingData{}, -1, "获取支付记录失败", err
	}
	return domain.UCPPaymentListingData{
		Rows:     processPaymentRows(rows),
		PageInfo: pageInfo(total, pageSize, page, "/ucp/payment/listing?page=[?]"),
	}, 0, "", nil
}

func (s *Service) SafePayLog(ctx context.Context, token string) (domain.UCPSafePayLogData, int, string, error) {
	user, err := s.authenticatedPaymentUser(ctx, token)
	if err != nil {
		return domain.UCPSafePayLogData{}, -1, "获取用户失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.UCPSafePayLogData{}, -9999, "您还没有登录", nil
	}

	since := s.now().Unix() - 86400*7
	rows, err := s.store.SafePayLogs(ctx, uid, since, 10)
	if err != nil {
		return domain.UCPSafePayLogData{}, -1, "获取支付记录失败", err
	}
	return domain.UCPSafePayLogData{PayRows: processPaymentRows(rows)}, 0, "", nil
}

func (s *Service) authenticatedPaymentUser(ctx context.Context, token string) (map[string]interface{}, error) {
	sid := userRepo.CleanToken(strings.TrimSpace(token))
	if sid == "" {
		return map[string]interface{}{"uid": "0"}, nil
	}
	user, err := s.store.UserBySession(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("load payment user: %w", err)
	}
	if user == nil {
		user = map[string]interface{}{"uid": "0"}
	}
	return user, nil
}

func processPaymentRows(rows []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		payway := str(row["payway"])
		paycode := str(row["paycode"])
		out = append(out, map[string]interface{}{
			"payid":        str(row["payid"]),
			"paytype":      mapString(paymentTypeNames, atoi(row["paytype"])),
			"payway":       payway,
			"paycode":      paycode,
			"payway_name":  mapString(paymentWayNames, payway),
			"paycode_name": mapString(paymentCodeNames, payway+"."+paycode),
			"itemname":     str(row["itemname"]),
			"trx_amount":   formatRMB(atoi(row["trx_amount"])),
			"pay_amount":   formatRMB(atoi(row["pay_amount"])),
			"uid":          str(row["uid"]),
			"createtime":   formatUnixMinute(atoi64(row["createtime"])),
			"ispaid":       atoi(row["ispaid"]),
			"paidtime":     formatOptionalUnixMinute(atoi64(row["paidtime"])),
			"out_trxid":    str(row["out_trxid"]),
		})
	}
	return out
}

func mapString[K comparable](items map[K]string, key K) interface{} {
	if value, ok := items[key]; ok {
		return value
	}
	return nil
}

func formatRMB(cents int) string {
	return fmt.Sprintf("%.2f", float64(cents)/100)
}

func formatUnixMinute(ts int64) string {
	if ts <= 0 {
		return "1970-01-01 08:00"
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return time.Unix(ts, 0).In(loc).Format("2006-01-02 15:04")
}

func formatOptionalUnixMinute(ts int64) string {
	if ts == 0 {
		return ""
	}
	return formatUnixMinute(ts)
}
