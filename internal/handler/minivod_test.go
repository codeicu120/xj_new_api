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
	minivodRepo "xj_comp/internal/repository/minivod"
	minivodService "xj_comp/internal/service/minivod"
)

type miniVODHandlerStore struct {
	vod    map[string]interface{}
	l2sMap map[string]interface{}
}

func (s miniVODHandlerStore) Categories(context.Context) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) Areas(context.Context) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) Years(context.Context) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) Servers(context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"srvid": "3", "srvhost": "https://cdn.test"}}, nil
}
func (s miniVODHandlerStore) TagsByNames(context.Context, []string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) Count(context.Context, minivodRepo.Filter, int64) (int, error) {
	return 0, nil
}
func (s miniVODHandlerStore) List(context.Context, minivodRepo.Filter, int, int, int, string, int64) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) CountByAuthor(context.Context, int) (int, error) { return 0, nil }
func (s miniVODHandlerStore) ListByAuthor(context.Context, int, int, int, int, string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) Random(context.Context, int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) VODByID(context.Context, int) (map[string]interface{}, error) {
	return s.vod, nil
}
func (s miniVODHandlerStore) UserByID(context.Context, int) (map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) UserQuota(context.Context, int) (map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) SimilarVODsByTagIDs(context.Context, []int, int, int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) RandomVODsExcept(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) Setting(context.Context, string) (string, error) { return "", nil }
func (s miniVODHandlerStore) UsersByIDs(context.Context, []int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) VODsByIDs(context.Context, []int, bool) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) PendingViewLogs(context.Context, int, string, int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) PullViewLogs(context.Context, int, string) (int, error) {
	return 0, nil
}
func (s miniVODHandlerStore) MarkViewLogsShown(context.Context, int, string, []int, int64) error {
	return nil
}
func (s miniVODHandlerStore) MiniVODAdCallRows(context.Context) ([]map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) UpDownByUser(context.Context, int, int) (map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) DeleteUpDown(context.Context, int, int) error { return nil }
func (s miniVODHandlerStore) SaveUpDown(context.Context, int, int, int, int64) (int, error) {
	return 0, nil
}
func (s miniVODHandlerStore) IncrementVODCounter(context.Context, int, string, int) error {
	return nil
}
func (s miniVODHandlerStore) RecountUpDown(context.Context, int) error { return nil }
func (s miniVODHandlerStore) FavoriteCount(context.Context, int, int) (int, error) {
	return 0, nil
}
func (s miniVODHandlerStore) MiniViewLog(context.Context, int, string, int) (map[string]interface{}, error) {
	return nil, nil
}
func (s miniVODHandlerStore) CountMiniViewLogsSince(context.Context, int, string, int64, int) (int, error) {
	return 0, nil
}
func (s miniVODHandlerStore) RecordMiniMedia(context.Context, int, string, int, bool, int, int64) error {
	return nil
}
func (s miniVODHandlerStore) ReqTaskCoin(context.Context, int, string, int, int64) (int, string, error) {
	return 0, "", nil
}
func (s miniVODHandlerStore) LongToShortMapByLongID(context.Context, int) (map[string]interface{}, error) {
	if s.l2sMap != nil {
		return s.l2sMap, nil
	}
	return map[string]interface{}{}, nil
}

type miniVODHandlerProcessor struct{}

func (miniVODHandlerProcessor) ProcessRowsFullPrice(context.Context, []map[string]interface{}, bool) ([]map[string]interface{}, error) {
	return nil, nil
}

func (miniVODHandlerProcessor) ProcessMiniRowsFullPrice(context.Context, []map[string]interface{}, bool) ([]map[string]interface{}, error) {
	return nil, nil
}

type miniVODHandlerFetcher map[string]string

func (f miniVODHandlerFetcher) Fetch(_ context.Context, url string) (string, error) {
	return f[url], nil
}

func TestMiniVODParseLongM3U8Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := minivodService.NewService(miniVODHandlerStore{
		vod: map[string]interface{}{"vodid": "9", "showtype": "0", "play_url": "a/index.m3u8", "play_srvid": "3"},
	}, miniVODHandlerProcessor{}, "").WithM3U8Fetcher(miniVODHandlerFetcher{
		"https://cdn.test/a/index.m3u8":     "#EXTM3U\nchild/index.m3u8\n",
		"https://cdn.test/child/index.m3u8": "#EXTM3U\n#EXTINF:8,\nseg0.ts\n",
	})
	router := gin.New()
	router.GET("/minivod/parselong/:vodid/index.m3u8", NewMiniVODHandler(service).ParseLongM3U8)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/minivod/parselong/9/index.m3u8", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "vnd.apple.mpegurl") {
		t.Fatalf("unexpected content-type %q", contentType)
	}
	if body := rec.Body.String(); !strings.Contains(body, "https://cdn.test/seg0.ts") || strings.Contains(body, "retcode") {
		t.Fatalf("unexpected body %q", body)
	}
}

func TestMiniVODParseLongM3U8KeepsReqLongErrorJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := minivodService.NewService(miniVODHandlerStore{
		vod: map[string]interface{}{"vodid": "9", "showtype": "1", "play_url": "a/index.m3u8", "play_srvid": "3"},
	}, miniVODHandlerProcessor{}, "")
	router := gin.New()
	router.GET("/minivod/parselong/:vodid/index.m3u8", NewMiniVODHandler(service).ParseLongM3U8)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/minivod/parselong/9/index.m3u8", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != 1 || body.ErrMsg != "记录不存在或已被删除" {
		t.Fatalf("unexpected response %#v", body)
	}
}
