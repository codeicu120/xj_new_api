package payment

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"strconv"
	"strings"
	"time"

	userRepo "xj_comp/internal/repository/user"
)

type Store interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
	PaymentByID(ctx context.Context, payid int) (map[string]interface{}, error)
	PaymentChannels(ctx context.Context, gameOnly bool) ([]map[string]interface{}, error)
	UpdatePaymentPayway(ctx context.Context, payid int, payway string, paycode string) (int, error)
}

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: time.Now}
}

func (s *Service) Unpaid(_ context.Context) map[string]interface{} {
	return map[string]interface{}{
		"total_count": 0,
	}
}

func (s *Service) SuccessMessage(_ context.Context) string {
	return "支付成功回调"
}

func (s *Service) FailedMessage(_ context.Context) string {
	return "支付失败回调"
}

func (s *Service) ReqPay(ctx context.Context, token string, payID int) (int, string, error) {
	if s.store == nil {
		return -1, "记录不存在或已支付", nil
	}
	user, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -1, "请求支付失败", err
	}
	payrow, err := s.store.PaymentByID(ctx, payID)
	if err != nil {
		return -1, "请求支付失败", err
	}
	if len(payrow) == 0 || atoi(payrow["ispaid"]) > 0 {
		return -1, "记录不存在或已支付", nil
	}
	if s.now().Unix() > int64(atoi(payrow["createtime"])+3600) {
		return -1, "支付已过期", nil
	}
	if atoi(payrow["uid"]) > 0 && atoi(user["uid"]) != atoi(payrow["uid"]) {
		return -1, "此项目需要本人操作", nil
	}
	payway := str(payrow["payway"])
	if knownReqPayPayway(payway) && atoi(payrow["nocheck"]) == 0 {
		channels, err := s.store.PaymentChannels(ctx, false)
		if err != nil {
			return -1, "请求支付失败", err
		}
		if !paymentCodeAllowed(channels, atoi(payrow["paytype"]), payway+"."+str(payrow["paycode"])) {
			return -1, "此项目不能使用此支付方式", nil
		}
	}
	return -1, "支付请求成功分支暂未迁移", nil
}

func (s *Service) PayErrorHTML(_ context.Context, errmsg string) string {
	if errmsg == "" {
		errmsg = "记录不存在或已支付"
	}
	return `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no"><title>支付错误</title></head><body>` + html.EscapeString(errmsg) + `</body></html>`
}

func (s *Service) SuccessHTML(_ context.Context) string {
	return `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no"><title>支付成功</title></head><body><ul><li><strong>支付成功</strong></li></ul></body></html>`
}

func (s *Service) QRCodeHTML(_ context.Context, qrlink string) string {
	return `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no"><title>微信扫码支付</title><script type="text/javascript" src="/qrcode.min.js"></script></head><body><ul><li>1、截图保存并打开微信</li><li>2、微信右上角扫一扫选择相册</li><li>付款中如遇警示弹窗，请点击继续支付</li><li><div id="qrcode"></div></li></ul><script type="text/javascript">new QRCode(document.getElementById("qrcode"), "` + html.EscapeString(qrlink) + `");</script></body></html>`
}

func (s *Service) SubmitHTML(_ context.Context, gateway string, rawParams string) string {
	values := parseQuery(rawParams)
	formdata := strings.Builder{}
	for key, vals := range values {
		for _, value := range vals {
			formdata.WriteString(`<input type="hidden" name="`)
			formdata.WriteString(html.EscapeString(key))
			formdata.WriteString(`" value="`)
			formdata.WriteString(html.EscapeString(value))
			formdata.WriteString(`" />`)
		}
	}
	return `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no"><title>正在跳转</title></head><body><form method="post" action="` + html.EscapeString(gateway) + `" id="form">` + formdata.String() + `</form><script type="text/javascript">document.getElementById('form').submit();</script></body></html>`
}

func (s *Service) PaymentHTML(ctx context.Context, payID int) (string, error) {
	if payID <= 0 || s.store == nil {
		return "", nil
	}
	payrow, err := s.store.PaymentByID(ctx, payID)
	if err != nil {
		return "", err
	}
	return str(payrow["payhtml"]), nil
}

func (s *Service) Query(ctx context.Context, token string, payID int) (map[string]interface{}, int, string, error) {
	if s.store == nil {
		return nil, -1, "获取支付订单失败", nil
	}
	user, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -1, "获取支付订单失败", err
	}
	payrow, err := s.store.PaymentByID(ctx, payID)
	if err != nil {
		return nil, -1, "获取支付订单失败", err
	}
	if len(payrow) == 0 || (atoi(payrow["uid"]) > 0 && atoi(user["uid"]) != atoi(payrow["uid"])) {
		return nil, -1, "无权限", nil
	}
	rows := processPaymentRows([]map[string]interface{}{payrow})
	return map[string]interface{}{"payrow": rows[0]}, 0, "", nil
}

func (s *Service) Payways(ctx context.Context, token string, payID int) (map[string]interface{}, int, string, error) {
	if s.store == nil {
		return nil, -1, "记录不存在或已支付", nil
	}
	user, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -1, "获取支付方式失败", err
	}
	payrow, err := s.store.PaymentByID(ctx, payID)
	if err != nil {
		return nil, -1, "获取支付方式失败", err
	}
	if len(payrow) == 0 || atoi(payrow["ispaid"]) > 0 {
		return nil, -1, "记录不存在或已支付", nil
	}
	if atoi(payrow["uid"]) > 0 && atoi(user["uid"]) != atoi(payrow["uid"]) {
		return nil, -1, "无权限", nil
	}
	channels, err := s.store.PaymentChannels(ctx, false)
	if err != nil {
		return nil, -1, "获取支付方式失败", err
	}
	rows := processPaymentRows([]map[string]interface{}{payrow})
	return map[string]interface{}{
		"payrow":   rows[0],
		"payments": filterPaymentChannels(channels, atoi(payrow["paytype"])),
	}, 0, "", nil
}

func (s *Service) ChangePayway(ctx context.Context, token string, payID int, rawPaycode string) (int, string, error) {
	if s.store == nil {
		return -1, "记录不存在或已支付", nil
	}
	user, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -1, "修改支付方式失败", err
	}
	payrow, err := s.store.PaymentByID(ctx, payID)
	if err != nil {
		return -1, "修改支付方式失败", err
	}
	if len(payrow) == 0 || atoi(payrow["ispaid"]) > 0 {
		return -1, "记录不存在或已支付", nil
	}
	if atoi(payrow["uid"]) > 0 && atoi(user["uid"]) != atoi(payrow["uid"]) {
		return -1, "此项目需要本人操作", nil
	}
	paycode := strings.TrimSpace(rawPaycode)
	channels, err := s.store.PaymentChannels(ctx, false)
	if err != nil {
		return -1, "修改支付方式失败", err
	}
	if !paymentCodeAllowed(channels, atoi(payrow["paytype"]), paycode) {
		return -1, "支付方式错误或不被允许", nil
	}
	parts := strings.SplitN(paycode, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return -1, "支付方式错误或不被允许", nil
	}
	affected, err := s.store.UpdatePaymentPayway(ctx, atoi(payrow["payid"]), parts[0], parts[1])
	if err != nil {
		return -1, "修改支付方式失败", err
	}
	if affected == 0 {
		return -1, "记录不存在或已支付", nil
	}
	return 0, "支付方式已修改", nil
}

func (s *Service) authenticatedUser(ctx context.Context, token string) (map[string]interface{}, error) {
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

func paymentCodeAllowed(channels []map[string]interface{}, payType int, paycode string) bool {
	if paycode == "" {
		return false
	}
	for _, channel := range filterPaymentChannels(channels, payType) {
		payways, _ := channel["payways"].([]map[string]interface{})
		for _, payway := range payways {
			if str(payway["paycode"]) == paycode {
				return true
			}
		}
	}
	return false
}

func knownReqPayPayway(payway string) bool {
	switch payway {
	case "walletpay", "shangfu", "wappay1", "wappay2", "wappay3", "wappay4", "wappay4a", "wappay5",
		"hawpay", "easypay", "pay6", "pay7", "pay8", "pay9", "pay10", "pay10a", "pay10b", "pay11", "pay12",
		"gpay1", "gpay2", "newpaykf", "newpayrq", "newpayhw", "newpaywm", "newpaygcash", "newpayupi",
		"newpayyhk", "newpaytm", "newpayok", "newpayuz", "newpaygcashqr", "newpaympesa", "newpaymoov",
		"newpaymtn", "newpayusdt", "newpaybakong", "newpaymomo", "newpayviettel", "newpayzalo",
		"newpaybb", "newpayphonepe", "newpayindianbank", "newpaytruemoney", "newpaypromptpay",
		"newpaypromptpayqr", "newpaygopay", "newpayovo", "newpaylinkaja", "newpayqris", "newpayjazzcash",
		"newpayeasypaisa":
		return true
	default:
		return false
	}
}

func mapOrEmpty(value interface{}) map[string]interface{} {
	if typed, ok := value.(map[string]interface{}); ok {
		return typed
	}
	return map[string]interface{}{}
}

func parseQuery(raw string) url.Values {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return url.Values{}
	}
	return values
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

func atoi(value interface{}) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(str(value)))
	return parsed
}

func atoi64(value interface{}) int64 {
	parsed, _ := strconv.ParseInt(strings.TrimSpace(str(value)), 10, 64)
	return parsed
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
