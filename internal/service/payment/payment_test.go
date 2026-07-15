package payment

import (
	"context"
	"strings"
	"testing"
)

type fakeStore struct {
	user           map[string]interface{}
	payment        map[string]interface{}
	channels       []map[string]interface{}
	updateAffected int
	updatedPayway  *string
	updatedPaycode *string
}

func (s fakeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s fakeStore) PaymentByID(context.Context, int) (map[string]interface{}, error) {
	return s.payment, nil
}

func (s fakeStore) PaymentChannels(context.Context, bool) ([]map[string]interface{}, error) {
	return s.channels, nil
}

func (s fakeStore) UpdatePaymentPayway(_ context.Context, _ int, payway string, paycode string) (int, error) {
	if s.updatedPayway != nil {
		*s.updatedPayway = payway
	}
	if s.updatedPaycode != nil {
		*s.updatedPaycode = paycode
	}
	return s.updateAffected, nil
}

func TestUnpaidAlwaysReturnsZeroTotal(t *testing.T) {
	service := NewService(nil)

	data := service.Unpaid(context.Background())

	if data["total_count"] != 0 {
		t.Fatalf("expected total_count 0, got %v", data["total_count"])
	}
	if len(data) != 1 {
		t.Fatalf("expected only total_count, got %#v", data)
	}
}

func TestCallbackMessages(t *testing.T) {
	service := NewService(nil)

	if got := service.SuccessMessage(context.Background()); got != "支付成功回调" {
		t.Fatalf("unexpected success message %q", got)
	}
	if got := service.FailedMessage(context.Background()); got != "支付失败回调" {
		t.Fatalf("unexpected failed message %q", got)
	}
}

func TestPaymentHTMLHelpers(t *testing.T) {
	service := NewService(fakeStore{payment: map[string]interface{}{"payhtml": "<form>ok</form>"}})

	if html := service.SuccessHTML(context.Background()); !containsAll(html, "支付成功", "<html") {
		t.Fatalf("unexpected success html %q", html)
	}
	if html := service.QRCodeHTML(context.Background(), `we"chat`); !containsAll(html, "QRCode", "we&#34;chat") {
		t.Fatalf("unexpected qrcode html %q", html)
	}
	if html := service.SubmitHTML(context.Background(), `https://pay.example/g`, "a=1&b=%E4%B8%AD"); !containsAll(html, `action="https://pay.example/g"`, `name="a" value="1"`, `name="b" value="中"`) {
		t.Fatalf("unexpected submit html %q", html)
	}
	payHTML, err := service.PaymentHTML(context.Background(), 10)
	if err != nil {
		t.Fatalf("payment html: %v", err)
	}
	if payHTML != "<form>ok</form>" {
		t.Fatalf("unexpected payment html %q", payHTML)
	}
}

func containsAll(value string, needles ...string) bool {
	for _, needle := range needles {
		if !strings.Contains(value, needle) {
			return false
		}
	}
	return true
}

func TestQueryRequiresPaymentAccess(t *testing.T) {
	service := NewService(fakeStore{
		user:    map[string]interface{}{"uid": "5"},
		payment: map[string]interface{}{"payid": "10", "uid": "6"},
	})

	_, retcode, errmsg, err := service.Query(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 10)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if retcode != -1 || errmsg != "无权限" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestQueryReturnsProcessedPayment(t *testing.T) {
	service := NewService(fakeStore{
		user: map[string]interface{}{"uid": "5"},
		payment: map[string]interface{}{
			"payid":      "10",
			"paytype":    "8",
			"payway":     "safepay",
			"paycode":    "safepay",
			"itemname":   "60天",
			"trx_amount": "6800",
			"pay_amount": "6800",
			"uid":        "5",
			"createtime": "1762948948",
			"ispaid":     "1",
			"paidtime":   "1762949999",
			"out_trxid":  "abc",
		},
	})

	data, retcode, errmsg, err := service.Query(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 10)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	payrow := data["payrow"].(map[string]interface{})
	if payrow["paytype"] != "购买套餐" || payrow["payway_name"] != "人工代付" || payrow["trx_amount"] != "68.00" {
		t.Fatalf("unexpected payrow %#v", payrow)
	}
}

func TestPaywaysRejectsMissingOrPaidPayment(t *testing.T) {
	for _, payment := range []map[string]interface{}{
		{},
		{"payid": "10", "ispaid": "1"},
	} {
		service := NewService(fakeStore{payment: payment})

		_, retcode, errmsg, err := service.Payways(context.Background(), "", 10)
		if err != nil {
			t.Fatalf("payways: %v", err)
		}
		if retcode != -1 || errmsg != "记录不存在或已支付" {
			t.Fatalf("unexpected response %d %q", retcode, errmsg)
		}
	}
}

func TestPaywaysRequiresPaymentOwner(t *testing.T) {
	service := NewService(fakeStore{
		user:    map[string]interface{}{"uid": "5"},
		payment: map[string]interface{}{"payid": "10", "uid": "6", "ispaid": "0"},
	})

	_, retcode, errmsg, err := service.Payways(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 10)
	if err != nil {
		t.Fatalf("payways: %v", err)
	}
	if retcode != -1 || errmsg != "无权限" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestPaywaysReturnsFilteredChannels(t *testing.T) {
	service := NewService(fakeStore{
		user: map[string]interface{}{"uid": "5"},
		payment: map[string]interface{}{
			"payid":      "10",
			"paytype":    "8",
			"payway":     "safepay",
			"paycode":    "safepay",
			"itemname":   "60天",
			"trx_amount": "6800",
			"pay_amount": "6800",
			"uid":        "5",
			"createtime": "1762948948",
			"ispaid":     "0",
		},
		channels: []map[string]interface{}{
			{
				"channame": "余额",
				"payways": []map[string]interface{}{
					{
						"payname":        "余额支付",
						"paycode":        "walletpay.walletpay",
						"trxamount_min":  "100",
						"trxamount_max":  "9999",
						"allow_paytypes": map[int][]string{8: {"ALL"}},
					},
					{
						"payname":        "游戏支付",
						"paycode":        "game.pay",
						"allow_paytypes": map[int][]string{13: {"ALL"}},
					},
				},
			},
			{"channame": "禁用", "disabled": "1", "payways": []map[string]interface{}{{"payname": "x"}}},
		},
	})

	data, retcode, errmsg, err := service.Payways(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 10)
	if err != nil {
		t.Fatalf("payways: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if data["payrow"].(map[string]interface{})["paytype"] != "购买套餐" {
		t.Fatalf("payrow = %#v", data["payrow"])
	}
	channels := data["payments"].([]map[string]interface{})
	if len(channels) != 1 {
		t.Fatalf("channels = %#v", channels)
	}
	payways := channels[0]["payways"].([]map[string]interface{})
	if len(payways) != 1 || payways[0]["paycode"] != "walletpay.walletpay" || payways[0]["trxamount_min"] != 100 {
		t.Fatalf("payways = %#v", payways)
	}
}

func TestChangePaywayRejectsMissingOrPaidPayment(t *testing.T) {
	for _, payment := range []map[string]interface{}{
		{},
		{"payid": "10", "ispaid": "1"},
	} {
		service := NewService(fakeStore{payment: payment})

		retcode, errmsg, err := service.ChangePayway(context.Background(), "", 10, "walletpay.walletpay")
		if err != nil {
			t.Fatalf("change payway: %v", err)
		}
		if retcode != -1 || errmsg != "记录不存在或已支付" {
			t.Fatalf("unexpected response %d %q", retcode, errmsg)
		}
	}
}

func TestChangePaywayRequiresPaymentOwner(t *testing.T) {
	service := NewService(fakeStore{
		user:    map[string]interface{}{"uid": "5"},
		payment: map[string]interface{}{"payid": "10", "uid": "6", "ispaid": "0"},
	})

	retcode, errmsg, err := service.ChangePayway(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 10, "walletpay.walletpay")
	if err != nil {
		t.Fatalf("change payway: %v", err)
	}
	if retcode != -1 || errmsg != "此项目需要本人操作" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestChangePaywayRejectsInvalidPaycode(t *testing.T) {
	service := NewService(fakeStore{
		user:    map[string]interface{}{"uid": "5"},
		payment: map[string]interface{}{"payid": "10", "uid": "5", "ispaid": "0", "paytype": "8"},
		channels: []map[string]interface{}{{
			"payways": []map[string]interface{}{{
				"paycode":        "walletpay.walletpay",
				"allow_paytypes": map[int][]string{8: {"ALL"}},
			}},
		}},
	})

	retcode, errmsg, err := service.ChangePayway(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 10, "safepay.safepay")
	if err != nil {
		t.Fatalf("change payway: %v", err)
	}
	if retcode != -1 || errmsg != "支付方式错误或不被允许" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestChangePaywayUpdatesPayment(t *testing.T) {
	var updatedPayway string
	var updatedPaycode string
	service := NewService(fakeStore{
		user:    map[string]interface{}{"uid": "5"},
		payment: map[string]interface{}{"payid": "10", "uid": "5", "ispaid": "0", "paytype": "8"},
		channels: []map[string]interface{}{{
			"payways": []map[string]interface{}{{
				"paycode":        "walletpay.walletpay",
				"allow_paytypes": map[int][]string{8: {"ALL"}},
			}},
		}},
		updateAffected: 1,
		updatedPayway:  &updatedPayway,
		updatedPaycode: &updatedPaycode,
	})

	retcode, errmsg, err := service.ChangePayway(context.Background(), "3235306637393062613731656332623964333835356634323464623232353965", 10, "walletpay.walletpay")
	if err != nil {
		t.Fatalf("change payway: %v", err)
	}
	if retcode != 0 || errmsg != "支付方式已修改" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
	if updatedPayway != "walletpay" || updatedPaycode != "walletpay" {
		t.Fatalf("unexpected update %q %q", updatedPayway, updatedPaycode)
	}
}
