package game

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	userRepo "xj_comp/internal/repository/user"
)

type LotteryPlatformStore interface {
	PlatformBySlug(ctx context.Context, slug string) (map[string]interface{}, error)
}

type LotteryEnterClient interface {
	EnterGame(ctx context.Context, cfg LotteryConfig, req LotteryEnterRequest) (string, error)
	Balance(ctx context.Context, cfg LotteryConfig, uid int) (LotteryBalance, error)
}

type LotteryConfig struct {
	PlatformID int
	APIURL     string
	Agent      string
	EncryptKey string
	SignKey    string
	Platform   string
	Lang       string
}

type LotteryEnterRequest struct {
	UID   int
	LotID int
}

type LotteryBalance struct {
	Status       interface{}
	Balance      string
	Transferable string
	Currency     interface{}
}

type LotteryHTTP struct {
	client *http.Client
	now    func() time.Time
}

func NewLotteryHTTP(timeout time.Duration) *LotteryHTTP {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &LotteryHTTP{client: &http.Client{Timeout: timeout}, now: time.Now}
}

func (h *LotteryHTTP) EnterGame(ctx context.Context, cfg LotteryConfig, req LotteryEnterRequest) (string, error) {
	data, err := h.post(ctx, cfg, "/api/player/enterGame", []lotteryParam{
		{key: "platform", value: cfg.Platform},
		{key: "uid", value: fmt.Sprint(req.UID)},
		{key: "lotId", value: fmt.Sprint(req.LotID)},
		{key: "lang", value: cfg.Lang},
	})
	if err != nil {
		return "", err
	}
	gameURL := fmt.Sprint(data["gameUrl"])
	if gameURL == "" || gameURL == "<nil>" {
		return "", fmt.Errorf("lottery gameUrl missing")
	}
	return gameURL, nil
}

func (h *LotteryHTTP) Balance(ctx context.Context, cfg LotteryConfig, uid int) (LotteryBalance, error) {
	data, err := h.post(ctx, cfg, "/api/player/balance", []lotteryParam{
		{key: "platform", value: cfg.Platform},
		{key: "uid", value: fmt.Sprint(uid)},
	})
	if err != nil {
		return LotteryBalance{}, err
	}
	return lotteryBalanceFromResult(data), nil
}

func (h *LotteryHTTP) post(ctx context.Context, cfg LotteryConfig, path string, params []lotteryParam) (map[string]interface{}, error) {
	if h.now == nil {
		h.now = time.Now
	}
	param, err := encryptLotteryParam(params, cfg.EncryptKey)
	if err != nil {
		return nil, err
	}
	timestamp := h.now().Unix() * 1000
	form := url.Values{}
	form.Set("agent", cfg.Agent)
	form.Set("timestamp", fmt.Sprint(timestamp))
	form.Set("param", url.QueryEscape(param))
	form.Set("sign", lotterySign(cfg.Agent, timestamp, param, cfg.SignKey))

	endpoint := strings.TrimRight(cfg.APIURL, "/") + path
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := h.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lottery status %d", resp.StatusCode)
	}
	var decoded struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, err
	}
	if decoded.Code != 0 {
		return nil, fmt.Errorf("lottery code %d", decoded.Code)
	}
	return decoded.Data, nil
}

type LotteryService struct {
	store     LotteryPlatformStore
	authStore WaliAuthStore
	client    LotteryEnterClient
}

type lotteryConfigJSON struct {
	APIURL     string `json:"apiUrl"`
	Agent      string `json:"agent"`
	EncryptKey string `json:"encryptKey"`
	SignKey    string `json:"signKey"`
	Platform   string `json:"platform"`
}

type lotteryParam struct {
	key   string
	value string
}

func NewLotteryService(store LotteryPlatformStore, authStore WaliAuthStore, client LotteryEnterClient) *LotteryService {
	if client == nil {
		client = NewLotteryHTTP(5 * time.Second)
	}
	return &LotteryService{store: store, authStore: authStore, client: client}
}

func (s *LotteryService) Enter(ctx context.Context, token string, lotidValue string) (string, int, string, error) {
	user, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return "", -1, "获取用户失败", err
	}
	uid := atoi(fmt.Sprint(user["uid"]))
	if uid == 0 {
		return "", -9999, "您还没有登录", nil
	}

	cfg, err := s.config(ctx)
	if err != nil {
		return "", -1, "进入游戏失败", nil
	}
	lotID := atoi(lotidValue)
	gameURL, err := s.client.EnterGame(ctx, cfg, LotteryEnterRequest{UID: uid, LotID: lotID})
	if err != nil {
		return "", -1, "进入游戏失败", nil
	}

	if historyStore, ok := s.store.(WaliHistoryStore); ok && historyStore != nil {
		history, err := historyStore.GameHistoryByUniqueKey(ctx, uid, cfg.PlatformID, lotID)
		if err != nil {
			return "", -1, "进入游戏失败", err
		}
		if id := atoi(fmt.Sprint(history["id"])); id > 0 {
			if err := historyStore.DeleteGameHistory(ctx, id); err != nil {
				return "", -1, "进入游戏失败", err
			}
		}
		if _, err := historyStore.SaveGameHistory(ctx, uid, cfg.PlatformID, lotID); err != nil {
			return "", -1, "进入游戏失败", err
		}
	}

	return gameURL, 0, "", nil
}

func (s *LotteryService) Balance(ctx context.Context, token string) (map[string]interface{}, int, string, error) {
	user, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -1, "获取用户失败", err
	}
	uid := atoi(fmt.Sprint(user["uid"]))
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}

	cfg, err := s.config(ctx)
	if err != nil {
		return nil, -1, "查询余额失败", nil
	}
	balance, err := s.client.Balance(ctx, cfg, uid)
	if err != nil {
		return nil, -1, "查询余额失败", nil
	}
	return map[string]interface{}{
		"status":       balance.Status,
		"balance":      balance.Balance,
		"transferable": balance.Transferable,
		"currency":     balance.Currency,
	}, 0, "", nil
}

func (s *LotteryService) authenticatedUser(ctx context.Context, token string) (map[string]interface{}, error) {
	sid := userRepo.CleanToken(token)
	if s.authStore == nil {
		return map[string]interface{}{"uid": "0", "sid": sid}, nil
	}
	user, err := s.authStore.UserBySession(ctx, sid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return map[string]interface{}{"uid": "0", "sid": sid}, nil
	}
	return user, nil
}

func (s *LotteryService) config(ctx context.Context) (LotteryConfig, error) {
	if s.store == nil {
		return LotteryConfig{}, fmt.Errorf("lottery platform unavailable")
	}
	row, err := s.store.PlatformBySlug(ctx, "lottery")
	if err != nil {
		return LotteryConfig{}, err
	}
	if len(row) == 0 || atoi(fmt.Sprint(row["status"])) != 1 {
		return LotteryConfig{}, fmt.Errorf("lottery platform disabled")
	}
	var payload lotteryConfigJSON
	if err := json.Unmarshal([]byte(fmt.Sprint(row["config_json"])), &payload); err != nil {
		return LotteryConfig{}, err
	}
	cfg := LotteryConfig{
		PlatformID: atoi(fmt.Sprint(row["id"])),
		APIURL:     payload.APIURL,
		Agent:      payload.Agent,
		EncryptKey: payload.EncryptKey,
		SignKey:    payload.SignKey,
		Platform:   payload.Platform,
		Lang:       "zh-CN",
	}
	if cfg.PlatformID <= 0 || cfg.APIURL == "" || cfg.Agent == "" || cfg.EncryptKey == "" || cfg.SignKey == "" || cfg.Platform == "" {
		return LotteryConfig{}, fmt.Errorf("lottery platform config incomplete")
	}
	return cfg, nil
}

func encryptLotteryParam(params []lotteryParam, key string) (string, error) {
	if len(key) < aes.BlockSize {
		return "", fmt.Errorf("lottery aes key too short")
	}
	parts := make([]string, 0, len(params))
	for _, param := range params {
		parts = append(parts, url.QueryEscape(param.key)+"="+url.QueryEscape(param.value))
	}
	plain, err := url.QueryUnescape(strings.Join(parts, "&"))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	padded := pkcs7Pad([]byte(plain), block.BlockSize())
	encrypted := make([]byte, len(padded))
	iv := []byte(key[:aes.BlockSize])
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(encrypted, padded)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func lotterySign(agent string, timestamp int64, param string, signKey string) string {
	sum := md5.Sum([]byte(agent + fmt.Sprint(timestamp) + param + signKey))
	return hex.EncodeToString(sum[:])
}

func lotteryBalanceFromResult(data map[string]interface{}) LotteryBalance {
	return LotteryBalance{
		Status:       data["status"],
		Balance:      formatLotteryCents(data["totalMoney"]),
		Transferable: formatLotteryCents(data["freeMoney"]),
		Currency:     data["currency"],
	}
}

func formatLotteryCents(value interface{}) string {
	amount, _ := strconv.ParseFloat(fmt.Sprint(value), 64)
	return fmt.Sprintf("%.2f", amount/100)
}
