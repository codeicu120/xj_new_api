package aiundress

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
	vodService "xj_comp/internal/service/vod"
)

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type Store interface {
	Count(ctx context.Context, uid int, module int) (int, error)
	List(ctx context.Context, uid int, module int, total int, page int, pageSize int) ([]map[string]interface{}, error)
	SettingByUUID(ctx context.Context, uuid string) (string, error)
}

type ExternalClient interface {
	PostJSON(ctx context.Context, path string, payload map[string]interface{}) (ExternalResponse, error)
}

type ExternalResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ResourceListInput = domain.AIUndressResourceListInput

type HTTPExternalClient struct {
	baseURL string
	key     string
	client  *http.Client
}

func NewHTTPExternalClient(host string, key string, timeout time.Duration) *HTTPExternalClient {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &HTTPExternalClient{
		baseURL: "https://" + strings.TrimRight(host, "/"),
		key:     strings.TrimSpace(key),
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *HTTPExternalClient) PostJSON(ctx context.Context, path string, payload map[string]interface{}) (ExternalResponse, error) {
	if c == nil || c.client == nil || c.baseURL == "https://" || c.key == "" {
		return ExternalResponse{}, fmt.Errorf("aiundress external config incomplete")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return ExternalResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return ExternalResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("channel_key", c.key)
	resp, err := c.client.Do(req)
	if err != nil {
		return ExternalResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ExternalResponse{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ExternalResponse{}, err
	}
	var decoded ExternalResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return ExternalResponse{}, err
	}
	return decoded, nil
}

type Service struct {
	auth            AuthStore
	store           Store
	resourceBaseURL string
	env             string
	externalClient  ExternalClient
	now             func() time.Time
}

func NewService(auth AuthStore, store Store, resourceBaseURL string, env ...string) *Service {
	envValue := ""
	if len(env) > 0 {
		envValue = env[0]
	}
	return &Service{
		auth:            auth,
		store:           store,
		resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"),
		env:             strings.ToLower(strings.TrimSpace(envValue)),
		now:             time.Now,
	}
}

func (s *Service) WithExternalClient(client ExternalClient) *Service {
	s.externalClient = client
	return s
}

func (s *Service) Listing(ctx context.Context, token string, page int, module int) (domain.AIUndressListingData, int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return domain.AIUndressListingData{}, -1, "获取AI记录失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.AIUndressListingData{}, -1, "请先登录", nil
	}
	const pageSize = 10
	total, err := s.store.Count(ctx, uid, module)
	if err != nil {
		return domain.AIUndressListingData{}, -1, "获取AI记录失败", err
	}
	rows, err := s.store.List(ctx, uid, module, total, page, pageSize)
	if err != nil {
		return domain.AIUndressListingData{}, -1, "获取AI记录失败", err
	}
	baseURL, err := s.aiResourceBaseURL(ctx)
	if err != nil {
		return domain.AIUndressListingData{}, -1, "获取AI记录失败", err
	}
	for _, row := range rows {
		row["image"] = s.resourceURL(baseURL, row["image"])
		row["output"] = s.resourceURL(baseURL, row["output"])
	}
	return domain.AIUndressListingData{
		Rows:     rows,
		PageInfo: vodService.PageInfo(total, pageSize, page, "/aiundress/listing?page=[?]"),
	}, 0, "", nil
}

func (s *Service) ModuleList(ctx context.Context) (domain.AIUndressExternalData, int, string, error) {
	return s.externalRequest(ctx, "/cps/getModuleList", map[string]interface{}{})
}

func (s *Service) ResourceTypeList(ctx context.Context, module string) (domain.AIUndressExternalData, int, string, error) {
	return s.externalRequest(ctx, "/cps/resourceTypeList", map[string]interface{}{
		"module": module,
	})
}

func (s *Service) ResourceList(ctx context.Context, input domain.AIUndressResourceListInput) (domain.AIUndressExternalData, int, string, error) {
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	return s.externalRequest(ctx, "/cps/resourceList", map[string]interface{}{
		"module":    input.Module,
		"typeId":    input.TypeID,
		"pageSize":  pageSize,
		"current":   input.Current,
		"id":        "",
		"name":      "",
		"sortField": "",
		"sortType":  "",
	})
}

func (s *Service) externalRequest(ctx context.Context, path string, payload map[string]interface{}) (domain.AIUndressExternalData, int, string, error) {
	if s.externalClient == nil {
		return domain.AIUndressExternalData{}, -1, "请求失败", nil
	}
	result, err := s.externalClient.PostJSON(ctx, path, payload)
	if err != nil {
		return domain.AIUndressExternalData{}, -1, "请求失败", nil
	}
	if result.Code != 200 {
		return domain.AIUndressExternalData{}, -1, fmt.Sprintf("请求失败[%d]:%s", result.Code, result.Message), nil
	}
	return domain.AIUndressExternalData{Data: result.Data}, 0, "", nil
}

func (s *Service) userByToken(ctx context.Context, token string) (map[string]interface{}, error) {
	sid := userRepo.CleanToken(token)
	if sid == "" || s.auth == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	user, err := s.auth.UserBySession(ctx, sid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	return user, nil
}

func (s *Service) aiResourceBaseURL(ctx context.Context) (string, error) {
	if s.env == "test" {
		return "https://pub-21fd0f8233a7476797cc1786f4cabea9.r2.dev", nil
	}
	if s.store == nil {
		return s.resourceBaseURL, nil
	}
	raw, err := s.store.SettingByUUID(ctx, "setting")
	if err != nil {
		return "", err
	}
	value := serializedString(raw, "resurl_h5_ai")
	if value == "" {
		return s.resourceBaseURL, nil
	}
	now := s.now()
	if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		now = now.In(loc)
	}
	value = strings.ReplaceAll(value, "{rand}", now.Format("2006010215"))
	return strings.TrimRight(value, "/"), nil
}

func (s *Service) resourceURL(baseURL string, value interface{}) interface{} {
	path := strings.TrimSpace(fmt.Sprint(value))
	if path == "" || path == "<nil>" {
		return path
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		return path
	}
	return baseURL + "/" + strings.TrimLeft(path, "/")
}

func serializedString(raw string, key string) string {
	pattern := regexp.MustCompile(`s:\d+:"` + regexp.QuoteMeta(key) + `";s:\d+:"([^"]*)"`)
	match := pattern.FindStringSubmatch(raw)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
