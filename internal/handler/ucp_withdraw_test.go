package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseRMBInput(t *testing.T) {
	tests := map[string]int{
		"5.6":     560,
		"5.60":    560,
		"0.01":    1,
		" 12 ":    1200,
		"":        0,
		"invalid": 0,
		"5.601":   0,
		"1e3":     0,
		"NaN":     0,
		"-1":      0,
	}
	for input, want := range tests {
		if got := parseRMBInput(input); got != want {
			t.Fatalf("parseRMBInput(%q)=%d want %d", input, got, want)
		}
	}
}

func TestWithdrawCreateValuesReadsJSONBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("POST", "/ucp/withdraw/create", strings.NewReader(
		`{"cardid":1711305,"wdtype":0,"withdraw_amount":5.6}`,
	))
	ctx.Request.Header.Set("Content-Type", "application/json")

	cardID, wdType, amount := withdrawCreateValues(ctx)
	if cardID != "1711305" || wdType != "0" || amount != "5.6" {
		t.Fatalf("values=%q,%q,%q", cardID, wdType, amount)
	}
	if got := parseRMBInput(amount); got != 560 {
		t.Fatalf("amount cents=%d", got)
	}
}
