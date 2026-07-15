package index

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var ErrCoverNotFound = errors.New("cover not found")

type CoverStore interface {
	SettingByUUID(ctx context.Context, uuid string) (map[string]interface{}, error)
}

type CoverCache interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
}

type CoverFetcher interface {
	FetchCover(ctx context.Context, baseURL string, pic string) (string, error)
}

type CoverService struct {
	store   CoverStore
	cache   CoverCache
	fetcher CoverFetcher
}

func NewCoverService(store CoverStore, cache CoverCache, fetcher CoverFetcher) *CoverService {
	if cache == nil {
		cache = NewMemoryCoverCache()
	}
	if fetcher == nil {
		fetcher = HTTPTimeoutCoverFetcher{Client: &http.Client{Timeout: 2 * time.Second}}
	}
	return &CoverService{store: store, cache: cache, fetcher: fetcher}
}

func (s *CoverService) GetCover(ctx context.Context, pic string) (string, error) {
	pic = strings.TrimSpace(pic)
	if pic == "" {
		return "", ErrCoverNotFound
	}
	if cached, ok, err := s.cache.Get(ctx, pic); err != nil {
		return "", err
	} else if ok && cached != "" {
		return cached, nil
	}
	if len(pic) < 42 {
		return "", ErrCoverNotFound
	}

	baseURL, err := s.coverURL(ctx)
	if err != nil {
		return "", err
	}
	raw, err := s.fetcher.FetchCover(ctx, baseURL, pic)
	if err != nil || raw == "" {
		return "", ErrCoverNotFound
	}
	encrypted, err := encryptCover(raw, pic)
	if err != nil {
		return "", ErrCoverNotFound
	}
	if err := s.cache.Set(ctx, pic, encrypted, 24*time.Hour); err != nil {
		return "", err
	}
	return encrypted, nil
}

func (s *CoverService) coverURL(ctx context.Context) (string, error) {
	row, err := s.store.SettingByUUID(ctx, "setting")
	if err != nil {
		return "", err
	}
	setting := parsePHPSerializedMap(str(row["value"]))
	if value := strings.TrimSpace(str(setting["getCoverUrl"])); value != "" {
		return value, nil
	}
	return "http://172.22.0.7:8026/coverpic", nil
}

func encryptCover(plain string, pic string) (string, error) {
	if len(pic) < 42 {
		return "", fmt.Errorf("pic too short for cover key")
	}
	key := pic[10:42]
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	data := pkcs7PadBytes([]byte(plain), block.BlockSize())
	out := make([]byte, len(data))
	cipher.NewCBCEncrypter(block, []byte(key[:16])).CryptBlocks(out, data)
	return base64.StdEncoding.EncodeToString(out), nil
}

func pkcs7PadBytes(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	out := make([]byte, 0, len(data)+padding)
	out = append(out, data...)
	for i := 0; i < padding; i++ {
		out = append(out, byte(padding))
	}
	return out
}

type HTTPTimeoutCoverFetcher struct {
	Client *http.Client
}

func (f HTTPTimeoutCoverFetcher) FetchCover(ctx context.Context, baseURL string, pic string) (string, error) {
	endpoint, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	query := endpoint.Query()
	query.Set("pic", pic)
	endpoint.RawQuery = query.Encode()
	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", ErrCoverNotFound
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

type MemoryCoverCache struct {
	mu   sync.Mutex
	rows map[string]coverCacheRow
}

type coverCacheRow struct {
	value     string
	expiresAt time.Time
}

func NewMemoryCoverCache() *MemoryCoverCache {
	return &MemoryCoverCache{rows: map[string]coverCacheRow{}}
}

func (c *MemoryCoverCache) Get(_ context.Context, key string) (string, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	row, ok := c.rows[key]
	if !ok {
		return "", false, nil
	}
	if !row.expiresAt.IsZero() && time.Now().After(row.expiresAt) {
		delete(c.rows, key)
		return "", false, nil
	}
	return row.value, true, nil
}

func (c *MemoryCoverCache) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	expiresAt := time.Time{}
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	c.rows[key] = coverCacheRow{value: value, expiresAt: expiresAt}
	return nil
}
