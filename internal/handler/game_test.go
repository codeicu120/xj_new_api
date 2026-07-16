package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/legacyjson"
	gameService "xj_comp/internal/service/game"
)

type gameHandlerWaliStore struct {
	row map[string]interface{}
}

func (s gameHandlerWaliStore) PlatformByID(context.Context, int) (map[string]interface{}, error) {
	return s.row, nil
}

type gameHandlerWaliAuthStore struct {
	user map[string]interface{}
}

func (s gameHandlerWaliAuthStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

type gameHandlerWaliClient struct {
	body []byte
}

func (c gameHandlerWaliClient) Get(context.Context, string) ([]byte, error) {
	return c.body, nil
}

type gameHandlerLotteryStore struct {
	row map[string]interface{}
}

func (s gameHandlerLotteryStore) PlatformBySlug(context.Context, string) (map[string]interface{}, error) {
	return s.row, nil
}

type gameHandlerLotteryClient struct {
	gameURL string
	balance gameService.LotteryBalance
}

func (c gameHandlerLotteryClient) EnterGame(context.Context, gameService.LotteryConfig, gameService.LotteryEnterRequest) (string, error) {
	return c.gameURL, nil
}

func (c gameHandlerLotteryClient) Balance(context.Context, gameService.LotteryConfig, int) (gameService.LotteryBalance, error) {
	return c.balance, nil
}

func TestGameHandlerWaliEnterReturnsGameURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	waliService := gameService.NewWaliService(
		gameHandlerWaliStore{row: map[string]interface{}{
			"status":      "1",
			"config_json": `{"url":"https://wali.example/api","account":"acct","aesKey":"1234567890abcdef","signKey":"sign","agentId":"agent"}`,
		}},
		gameHandlerWaliAuthStore{user: map[string]interface{}{"uid": "9"}},
		gameHandlerWaliClient{body: []byte(`{"code":0,"data":{"gameUrl":"https://play.example/table"}}`)},
	)
	router := gin.New()
	router.POST("/game/wali/enter", NewGameHandler(nil, nil, nil, nil, waliService).WaliEnter)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/game/wali/enter", strings.NewReader("game=12"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("x-cookie-auth", "token")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if servedBy := rec.Header().Get("X-Served-By"); servedBy != "newbie" {
		t.Fatalf("expected X-Served-By newbie, got %q", servedBy)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != 0 || body.ErrMsg != "" || body.Data != "https://play.example/table" {
		t.Fatalf("unexpected response %#v", body)
	}
}

func TestGameHandlerLotteryEnterReturnsGameURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lotteryService := gameService.NewLotteryService(
		gameHandlerLotteryStore{row: map[string]interface{}{
			"id":          "8",
			"status":      "1",
			"config_json": `{"apiUrl":"https://lottery.example/api","agent":"agent","encryptKey":"1234567890abcdef","signKey":"sign","platform":"xjlottery"}`,
		}},
		gameHandlerWaliAuthStore{user: map[string]interface{}{"uid": "9"}},
		gameHandlerLotteryClient{gameURL: "https://lottery.example/table"},
	)
	router := gin.New()
	router.POST("/game/lottery/enter", NewGameHandler(nil, nil, nil, nil, nil, lotteryService).LotteryEnter)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/game/lottery/enter", strings.NewReader("lotid=12"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("x-cookie-auth", "token")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if servedBy := rec.Header().Get("X-Served-By"); servedBy != "newbie" {
		t.Fatalf("expected X-Served-By newbie, got %q", servedBy)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != 0 || body.ErrMsg != "" || body.Data != "https://lottery.example/table" {
		t.Fatalf("unexpected response %#v", body)
	}
}

func TestGameHandlerLotteryBalanceReturnsDataWrapper(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lotteryService := gameService.NewLotteryService(
		gameHandlerLotteryStore{row: map[string]interface{}{
			"id":          "8",
			"status":      "1",
			"config_json": `{"apiUrl":"https://lottery.example/api","agent":"agent","encryptKey":"1234567890abcdef","signKey":"sign","platform":"xjlottery"}`,
		}},
		gameHandlerWaliAuthStore{user: map[string]interface{}{"uid": "9"}},
		gameHandlerLotteryClient{balance: gameService.LotteryBalance{Status: float64(10), Balance: "12.34", Transferable: "5.60", Currency: "CNY"}},
	)
	router := gin.New()
	router.POST("/game/lottery/balance", NewGameHandler(nil, nil, nil, nil, nil, lotteryService).LotteryBalance)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/game/lottery/balance", nil)
	req.Header.Set("x-cookie-auth", "token")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := body.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected data %#v", body.Data)
	}
	row, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected nested data %#v", data)
	}
	if body.RetCode != 0 || row["balance"] != "12.34" || row["transferable"] != "5.60" || row["currency"] != "CNY" {
		t.Fatalf("unexpected response %#v", body)
	}
}
