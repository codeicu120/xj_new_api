package respond

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"xj_comp/internal/config"
)

func TestVerifyProviderMissingConfigReturnsEchoErr(t *testing.T) {
	service := NewService(nil)

	result := service.VerifyProvider(context.Background(), CallbackRequest{Action: "shangfu"}, "failed")

	if result.Echo != "failed" {
		t.Fatalf("expected echoErr failed, got %q", result.Echo)
	}
	if !errors.Is(result.Reason, ErrMissingConfig) {
		t.Fatalf("expected ErrMissingConfig, got %v", result.Reason)
	}
}

func TestVerifyProviderInvalidSignatureReturnsEchoErr(t *testing.T) {
	service := NewService(nil, WithRegistry(NewRegistryFromConfig([]config.RespondProviderConfig{
		shangfuFixtureConfig("fixture-secret", false),
	})))
	form := shangfuFixtureForm("fixture-secret")
	form.Set("app_sign", "bad")

	result := service.VerifyProvider(context.Background(), CallbackRequest{Action: "shangfu", Form: form}, "failed")

	if result.Echo != "failed" {
		t.Fatalf("expected echoErr failed, got %q", result.Echo)
	}
	if !errors.Is(result.Reason, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", result.Reason)
	}
}

func TestMD5FormVerifierFixtureSuccessReturnsEchoErrWhenAccountingDisabled(t *testing.T) {
	service := NewService(nil, WithRegistry(NewRegistryFromConfig([]config.RespondProviderConfig{
		shangfuFixtureConfig("fixture-secret", false),
	})))

	result := service.VerifyProvider(context.Background(), CallbackRequest{
		Action: "shangfu",
		Form:   shangfuFixtureForm("fixture-secret"),
	}, "failed")

	if result.Echo != "failed" {
		t.Fatalf("expected accounting disabled to echoErr failed, got %q", result.Echo)
	}
	if result.Echo == "success" {
		t.Fatalf("must not echo provider success while accounting is disabled")
	}
	if !errors.Is(result.Reason, ErrAccountingDisabled) {
		t.Fatalf("expected ErrAccountingDisabled, got %v", result.Reason)
	}
}

func TestPay7MD5FormVerifierFixtureSuccessReturnsEchoErrWhenAccountingDisabled(t *testing.T) {
	service := NewService(nil, WithRegistry(NewRegistryFromConfig([]config.RespondProviderConfig{
		pay7FixtureConfig("pay7-secret", false),
	})))

	result := service.VerifyProvider(context.Background(), CallbackRequest{
		Action: "pay7",
		Form:   pay7FixtureForm("pay7-secret"),
	}, "failed")

	if result.Echo != "failed" {
		t.Fatalf("expected accounting disabled to echoErr failed, got %q", result.Echo)
	}
	if result.Echo == "OK" || result.Echo == "SUCCESS" {
		t.Fatalf("must not echo provider success while accounting is disabled")
	}
	if !errors.Is(result.Reason, ErrAccountingDisabled) {
		t.Fatalf("expected ErrAccountingDisabled, got %v", result.Reason)
	}
}

func TestMD5FormVerifierFixtureExtractsPayment(t *testing.T) {
	verifier := MD5FormVerifier{
		Secret:          "fixture-secret",
		SignField:       "app_sign",
		StatusField:     "status",
		SuccessStatus:   "1",
		OrderIDField:    "user_trade_no",
		OutTradeIDField: "trade_no",
		AmountField:     "amount",
	}

	payment, err := verifier.Verify(context.Background(), CallbackRequest{
		Action: "shangfu",
		Form:   shangfuFixtureForm("fixture-secret"),
	})

	if err != nil {
		t.Fatalf("verify fixture: %v", err)
	}
	if payment.PayID != 12345 || payment.OutTradeID != "third-789" || payment.PayAmount != 1234 {
		t.Fatalf("unexpected verified payment %#v", payment)
	}
}

func TestPay7MD5FormVerifierFixtureExtractsPayment(t *testing.T) {
	verifier := MD5FormQuerySecretSuffixVerifier{
		Secret:          "pay7-secret",
		SignField:       "sign",
		StatusField:     "trade_status",
		SuccessStatus:   "TRADE_SUCCESS",
		OrderIDField:    "out_trade_no",
		OutTradeIDField: "trade_no",
		AmountField:     "money",
	}

	payment, err := verifier.Verify(context.Background(), CallbackRequest{
		Action: "pay7",
		Form:   pay7FixtureForm("pay7-secret"),
	})

	if err != nil {
		t.Fatalf("verify fixture: %v", err)
	}
	if payment.PayID != 67890 || payment.OutTradeID != "pay7-third-456" || payment.PayAmount != 2500 {
		t.Fatalf("unexpected verified payment %#v", payment)
	}
}

func shangfuFixtureConfig(secret string, accountingEnabled bool) config.RespondProviderConfig {
	return config.RespondProviderConfig{
		Action:            "shangfu",
		EchoOK:            "success",
		EchoErr:           "failed",
		Verifier:          verifierMD5FormConcatSortedValues,
		Secret:            secret,
		SignField:         "app_sign",
		StatusField:       "status",
		SuccessStatus:     "1",
		OrderIDField:      "user_trade_no",
		OutTradeIDField:   "trade_no",
		AmountField:       "amount",
		AccountingEnabled: accountingEnabled,
	}
}

func shangfuFixtureForm(secret string) url.Values {
	form := url.Values{
		"app_id":        {"app-001"},
		"user_trade_no": {"12345"},
		"subject":       {"vip"},
		"amount":        {"12.34"},
		"trade_no":      {"third-789"},
		"status":        {"1"},
	}
	verifier := MD5FormVerifier{
		Secret:    secret,
		SignField: "app_sign",
	}
	form.Set("app_sign", verifier.sign(form))
	return form
}

func pay7FixtureConfig(secret string, accountingEnabled bool) config.RespondProviderConfig {
	return config.RespondProviderConfig{
		Action:            "pay7",
		EchoOK:            "OK",
		EchoErr:           "failed",
		Verifier:          verifierMD5FormQuerySecretSuffix,
		Secret:            secret,
		SignField:         "sign",
		StatusField:       "trade_status",
		SuccessStatus:     "TRADE_SUCCESS",
		OrderIDField:      "out_trade_no",
		OutTradeIDField:   "trade_no",
		AmountField:       "money",
		AccountingEnabled: accountingEnabled,
	}
}

func pay7FixtureForm(secret string) url.Values {
	form := url.Values{
		"pid":           {"merchant-007"},
		"type":          {"alipay"},
		"out_trade_no":  {"67890"},
		"name":          {"购买套餐"},
		"money":         {"25.00"},
		"trade_no":      {"pay7-third-456"},
		"trade_status":  {"TRADE_SUCCESS"},
		"sign_type":     {"MD5"},
		"empty_ignored": {""},
	}
	verifier := MD5FormQuerySecretSuffixVerifier{
		Secret:    secret,
		SignField: "sign",
	}
	form.Set("sign", verifier.sign(form))
	return form
}
