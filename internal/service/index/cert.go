package index

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const defaultGetCertURL = "https://api.apkcdn.cc/api/get_cert"

var errCertNotFound = errors.New("cert not found")

type SettingsStore interface {
	SettingValue(ctx context.Context) (string, error)
}

type CertHTTPClient interface {
	Get(ctx context.Context, rawURL string) ([]byte, error)
}

type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient(timeout time.Duration) *HTTPClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &HTTPClient{client: &http.Client{Timeout: timeout}}
}

func (c *HTTPClient) Get(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

type CertService struct {
	store  SettingsStore
	client CertHTTPClient
}

func NewCertService(store SettingsStore, client CertHTTPClient) *CertService {
	if client == nil {
		client = NewHTTPClient(5 * time.Second)
	}
	return &CertService{store: store, client: client}
}

func (s *CertService) GetCertUUID(ctx context.Context, uuid string) (interface{}, error) {
	endpoint := defaultGetCertURL
	if s.store != nil {
		value, err := s.store.SettingValue(ctx)
		if err != nil {
			return nil, err
		}
		if configured := serializedString(value, "getCertUrl"); configured != "" {
			endpoint = configured
		}
	}

	body, err := s.client.Get(ctx, endpoint+"?uuid="+url.QueryEscape(uuid))
	if err != nil {
		return nil, errCertNotFound
	}

	var decoded struct {
		Code int         `json:"code"`
		Data interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, errCertNotFound
	}
	if decoded.Code != 0 {
		return nil, errCertNotFound
	}
	return decoded.Data, nil
}

func IsCertNotFound(err error) bool {
	return errors.Is(err, errCertNotFound)
}

func serializedString(value string, key string) string {
	if value == "" || key == "" {
		return ""
	}
	pattern := regexp.MustCompile(`s:\d+:"` + regexp.QuoteMeta(key) + `";s:\d+:"([^"]*)"`)
	matches := pattern.FindStringSubmatch(value)
	if len(matches) != 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}
