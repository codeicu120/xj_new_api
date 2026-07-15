package payment

import (
	"context"
	"testing"
)

type fakeStore struct {
	user    map[string]interface{}
	payment map[string]interface{}
}

func (s fakeStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s fakeStore) PaymentByID(context.Context, int) (map[string]interface{}, error) {
	return s.payment, nil
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
