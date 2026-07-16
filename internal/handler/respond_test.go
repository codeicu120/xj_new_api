package handler

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/config"
	respondService "xj_comp/internal/service/respond"
)

func TestRespondProviderPay7ValidSignatureFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := respondService.NewService(nil, respondService.WithRegistry(respondService.NewRegistryFromConfig([]config.RespondProviderConfig{
		{
			Action:            "pay7",
			EchoOK:            "OK",
			EchoErr:           "failed",
			Verifier:          "md5_form_query_secret_suffix",
			Secret:            "pay7-secret",
			SignField:         "sign",
			StatusField:       "trade_status",
			SuccessStatus:     "TRADE_SUCCESS",
			OrderIDField:      "out_trade_no",
			OutTradeIDField:   "trade_no",
			AmountField:       "money",
			AccountingEnabled: false,
		},
	})))
	router := gin.New()
	router.POST("/respond/pay7", NewRespondHandler(service).Provider("pay7", "failed"))

	form := pay7HandlerFixtureForm("pay7-secret")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/respond/pay7", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Body.String(); got != "failed" {
		t.Fatalf("expected fail-closed echoErr failed, got %q", got)
	}
	if rec.Body.String() == "OK" || rec.Body.String() == "SUCCESS" {
		t.Fatalf("must not echo provider success while accounting is disabled")
	}
}

func pay7HandlerFixtureForm(secret string) url.Values {
	form := url.Values{
		"pid":          {"merchant-007"},
		"type":         {"alipay"},
		"out_trade_no": {"67890"},
		"name":         {"购买套餐"},
		"money":        {"25.00"},
		"trade_no":     {"pay7-third-456"},
		"trade_status": {"TRADE_SUCCESS"},
		"sign_type":    {"MD5"},
	}
	form.Set("sign", pay7HandlerFixtureSign(form, secret))
	return form
}

func pay7HandlerFixtureSign(form url.Values, secret string) string {
	keys := make([]string, 0, len(form))
	for key, vals := range form {
		if key == "sign" || key == "sign_type" || len(vals) != 1 || vals[0] == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for i, key := range keys {
		if i > 0 {
			builder.WriteByte('&')
		}
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(form.Get(key))
	}
	builder.WriteString(secret)
	sum := md5.Sum([]byte(builder.String()))
	return hex.EncodeToString(sum[:])
}
