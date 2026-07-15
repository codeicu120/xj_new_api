package game

import (
	"context"
	"crypto/aes"
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

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
)

type WaliPlatformStore interface {
	PlatformByID(ctx context.Context, id int) (map[string]interface{}, error)
}

type WaliSettingStore interface {
	Setting(ctx context.Context, key string) (string, error)
}

type WaliAuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type WaliHTTPClient interface {
	Get(ctx context.Context, rawURL string) ([]byte, error)
}

type WaliHTTP struct {
	client *http.Client
}

func NewWaliHTTP(timeout time.Duration) *WaliHTTP {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &WaliHTTP{client: &http.Client{Timeout: timeout}}
}

func (h *WaliHTTP) Get(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

type WaliService struct {
	store     WaliPlatformStore
	authStore WaliAuthStore
	client    WaliHTTPClient
	now       func() time.Time
}

type waliConfig struct {
	URL     string `json:"url"`
	Account string `json:"account"`
	AESKey  string `json:"aesKey"`
	SignKey string `json:"signKey"`
	AgentID string `json:"agentId"`
}

func NewWaliService(store WaliPlatformStore, authStore WaliAuthStore, client WaliHTTPClient) *WaliService {
	if client == nil {
		client = NewWaliHTTP(5 * time.Second)
	}
	return &WaliService{store: store, authStore: authStore, client: client, now: time.Now}
}

func (s *WaliService) Ping(ctx context.Context) (domain.GameWaliData, error) {
	data, err := s.sendRequest(ctx, "ping", map[string]string{"text": "helloThere"})
	if err != nil {
		return domain.GameWaliData{}, err
	}
	return domain.GameWaliData{Data: data}, nil
}

func (s *WaliService) Balance(ctx context.Context, token string) (domain.GameWaliData, int, string, error) {
	user, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return domain.GameWaliData{}, -1, "获取用户失败", err
	}
	uid := atoi(fmt.Sprint(user["uid"]))
	if uid == 0 {
		return domain.GameWaliData{}, -9999, "您还没有登录", nil
	}

	data, err := s.sendRequest(ctx, "getBalance", map[string]string{"uid": fmt.Sprint(uid)})
	if err != nil {
		return domain.GameWaliData{}, -1, "查询余额失败", nil
	}
	return domain.GameWaliData{Data: map[string]interface{}{
		"status":       data["status"],
		"balance":      data["balance"],
		"transferable": data["transferable"],
	}}, 0, "", nil
}

func (s *WaliService) ActionEdge(ctx context.Context, token string, pendingMessage string) (int, string, error) {
	user, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -1, "获取用户失败", err
	}
	if atoi(fmt.Sprint(user["uid"])) == 0 {
		return -9999, "您还没有登录", nil
	}
	if pendingMessage == "" {
		pendingMessage = "游戏成功分支暂未迁移"
	}
	return -1, pendingMessage, nil
}

func (s *WaliService) TopupEdge(ctx context.Context, token string, amount string, pendingMessage string) (int, string, error) {
	user, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -1, "获取用户失败", err
	}
	if atoi(fmt.Sprint(user["uid"])) == 0 {
		return -9999, "您还没有登录", nil
	}
	limit, err := s.gameCoinLimit(ctx)
	if err != nil {
		return -1, "游戏上分失败", err
	}
	coins := atoi(amount)
	if coins == 0 || coins < limit {
		return -1, fmt.Sprintf("转入金币不能低于%d", limit), nil
	}
	if pendingMessage == "" {
		pendingMessage = "游戏上分成功分支暂未迁移"
	}
	return -1, pendingMessage, nil
}

func (s *WaliService) WithdrawEdge(ctx context.Context, token string, amount string, pendingMessage string) (int, string, error) {
	user, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -1, "获取用户失败", err
	}
	if atoi(fmt.Sprint(user["uid"])) == 0 {
		return -9999, "您还没有登录", nil
	}
	if atof(amount) <= 0 {
		return -1, "金额输入不正确", nil
	}
	if pendingMessage == "" {
		pendingMessage = "游戏下分成功分支暂未迁移"
	}
	return -1, pendingMessage, nil
}

func (s *WaliService) gameCoinLimit(ctx context.Context) (int, error) {
	store, ok := s.store.(WaliSettingStore)
	if !ok || store == nil {
		return 0, nil
	}
	value, err := store.Setting(ctx, "gamecoinlimit")
	if err != nil {
		return 0, err
	}
	return atoi(value), nil
}

func (s *WaliService) authenticatedUser(ctx context.Context, token string) (map[string]interface{}, error) {
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

func (s *WaliService) sendRequest(ctx context.Context, action string, data map[string]string) (map[string]interface{}, error) {
	cfg, err := s.config(ctx)
	if err != nil {
		return nil, err
	}
	param, err := encryptWaliParams(data, cfg.AESKey)
	if err != nil {
		return nil, err
	}
	unixTime := s.now().Unix()
	sign := waliSign(param, unixTime, cfg.SignKey)
	rawURL := strings.TrimRight(cfg.URL, "/") + "/" + action + "?a=" + url.QueryEscape(cfg.Account) + "&t=" + fmt.Sprint(unixTime) + "&p=" + url.QueryEscape(param) + "&k=" + sign

	body, err := s.client.Get(ctx, rawURL)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("wali code %d", decoded.Code)
	}
	if decoded.Data == nil {
		decoded.Data = map[string]interface{}{}
	}
	decoded.Data["code"] = decoded.Code
	decoded.Data["msg"] = decoded.Msg
	return decoded.Data, nil
}

func (s *WaliService) config(ctx context.Context) (waliConfig, error) {
	if s.store == nil {
		return waliConfig{}, fmt.Errorf("wali platform unavailable")
	}
	row, err := s.store.PlatformByID(ctx, 1)
	if err != nil {
		return waliConfig{}, err
	}
	if len(row) == 0 || atoi(fmt.Sprint(row["status"])) != 1 {
		return waliConfig{}, fmt.Errorf("wali platform disabled")
	}
	var cfg waliConfig
	if err := json.Unmarshal([]byte(fmt.Sprint(row["config_json"])), &cfg); err != nil {
		return waliConfig{}, err
	}
	if cfg.URL == "" || cfg.Account == "" || cfg.AESKey == "" || cfg.SignKey == "" || cfg.AgentID == "" {
		return waliConfig{}, fmt.Errorf("wali platform config incomplete")
	}
	return cfg, nil
}

func encryptWaliParams(data map[string]string, key string) (string, error) {
	values := url.Values{}
	for k, v := range data {
		values.Set(k, v)
	}
	plain := values.Encode()
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	padded := pkcs7Pad([]byte(plain), block.BlockSize())
	encrypted := make([]byte, len(padded))
	for start := 0; start < len(padded); start += block.BlockSize() {
		block.Encrypt(encrypted[start:start+block.BlockSize()], padded[start:start+block.BlockSize()])
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	out := make([]byte, 0, len(data)+padding)
	out = append(out, data...)
	for i := 0; i < padding; i++ {
		out = append(out, byte(padding))
	}
	return out
}

func waliSign(param string, unixTime int64, signKey string) string {
	sum := md5.Sum([]byte(param + fmt.Sprint(unixTime) + signKey))
	return hex.EncodeToString(sum[:])
}

func atof(value string) float64 {
	n, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return n
}
