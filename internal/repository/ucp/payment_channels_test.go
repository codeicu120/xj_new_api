package ucp

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPaymentChannelsIncludesEnabledNewpayRQAlipay(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT value FROM settings WHERE uuid=?")).
		WithArgs("payment.chansetting").
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(
			`a:1:{s:8:"paycodes";a:1:{i:0;s:15:"newpayrq.alipay";}}`,
		))

	channels, err := NewRepository(db).PaymentChannels(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(channels) != 1 || channels[0]["channame"] != "瑞奇支付" {
		t.Fatalf("channels=%#v", channels)
	}
	payways := channels[0]["payways"].([]map[string]interface{})
	if len(payways) != 1 || payways[0]["paycode"] != "newpayrq.alipay" {
		t.Fatalf("payways=%#v", payways)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPaymentChannelsDoesNotMatchPaycodeOutsideSelectedList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT value FROM settings WHERE uuid=?")).
		WithArgs("payment.chansetting").
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(
			`a:2:{s:8:"paycodes";a:0:{}s:8:"paynames";a:1:{s:15:"newpayrq.alipay";s:3:"foo";}}`,
		))
	channels, err := NewRepository(db).PaymentChannels(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(channels) != 0 {
		t.Fatalf("channels=%#v", channels)
	}
}

func TestPaymentChannelsOmitsUnselectedNewpayRQ(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT value FROM settings WHERE uuid=?")).
		WithArgs("payment.chansetting").
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(`a:0:{}`))

	channels, err := NewRepository(db).PaymentChannels(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(channels) != 0 {
		t.Fatalf("channels=%#v", channels)
	}
}
