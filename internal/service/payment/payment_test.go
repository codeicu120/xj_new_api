package payment

import (
	"context"
	"testing"
)

type fakeStore struct {
	user     map[string]interface{}
	payment  map[string]interface{}
	channels []map[string]interface{}
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
