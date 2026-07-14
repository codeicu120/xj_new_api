package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"xj_comp/internal/config"
	"xj_comp/internal/legacyjson"
)

func TestHealthz(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("expected status ok, got %q", body.Status)
	}
	if body.Service != "xj-comp-api" {
		t.Fatalf("expected service xj-comp-api, got %q", body.Service)
	}
}

func TestReadyz(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestLegacyPlaceholder(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v2/login", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected status %d, got %d", http.StatusNotImplemented, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["legacy_handler"] != "c.apiv2.user.login" {
		t.Fatalf("unexpected legacy handler %q", body["legacy_handler"])
	}
}

func TestSysAvatar(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sysavatar", nil)
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
	if body.RetCode != 0 {
		t.Fatalf("expected retcode 0, got %d", body.RetCode)
	}

	data, ok := body.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", body.Data)
	}
	sysavatar, ok := data["sysavatar"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected sysavatar object, got %T", data["sysavatar"])
	}
	man, ok := sysavatar["man"].([]interface{})
	if !ok {
		t.Fatalf("expected man array, got %T", sysavatar["man"])
	}
	if len(man) != 9 {
		t.Fatalf("expected 9 man avatars, got %d", len(man))
	}
	if man[0] != "https://static.example.test/sysavatar/man/1.png" {
		t.Fatalf("unexpected first man avatar %v", man[0])
	}
}

func TestCaptchaReq(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/captcha/req", nil)
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
	if body.RetCode != 0 {
		t.Fatalf("expected retcode 0, got %d", body.RetCode)
	}

	data, ok := body.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", body.Data)
	}
	picURL, ok := data["picurl"].(string)
	if !ok {
		t.Fatalf("expected picurl string, got %T", data["picurl"])
	}
	if !strings.HasPrefix(picURL, "/captcha/picx?") {
		t.Fatalf("unexpected picurl %q", picURL)
	}
	if data["smscaptcha"] != float64(1) {
		t.Fatalf("expected smscaptcha 1, got %v", data["smscaptcha"])
	}
}

func TestIPLoc(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/iploc/8.8.8.8", nil)
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
	if body.RetCode != 0 {
		t.Fatalf("expected retcode 0, got %d", body.RetCode)
	}
	data, ok := body.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", body.Data)
	}
	if data["data"] != "" {
		t.Fatalf("expected empty test ip location, got %v", data["data"])
	}
}

func TestV2SOList(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/so/list?channel=xj&arm=v8a&version=510", nil)
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
	if body.RetCode != 0 {
		t.Fatalf("expected retcode 0, got %d", body.RetCode)
	}
	data, ok := body.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", body.Data)
	}
	if _, ok := data["data"]; !ok {
		t.Fatal("expected nested data field")
	}
	if data["data"] != nil {
		t.Fatalf("expected nil data without mysql, got %#v", data["data"])
	}
}

func TestV2AmazingListingRoutes(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{
		"/v2/amazing/listing",
		"/v2/amazing/listing-3-0-1",
		"/v2/amazing/recommend",
		"/v2/amazing/hot",
		"/v2/amazing/latest",
	} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
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
			if body.RetCode != 0 {
				t.Fatalf("expected retcode 0, got %d", body.RetCode)
			}
			data, ok := body.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("expected data object, got %T", body.Data)
			}
			if _, ok := data["rows"].([]interface{}); !ok {
				t.Fatalf("expected rows array, got %T", data["rows"])
			}
			if _, ok := data["pageinfo"].(map[string]interface{}); !ok {
				t.Fatalf("expected pageinfo object, got %T", data["pageinfo"])
			}
		})
	}
}

func TestV2VODListingRoutes(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{
		"/v2/vod/listing",
		"/v2/vod/listing-0-0-0-0-0-0-0-0-0-2",
		"/v2/vod/recommend",
		"/v2/vod/hot",
		"/v2/vod/latest",
	} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
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
			if body.RetCode != 0 {
				t.Fatalf("expected retcode 0, got %d", body.RetCode)
			}
			data, ok := body.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("expected data object, got %T", body.Data)
			}
			if _, ok := data["vodrows"].([]interface{}); !ok {
				t.Fatalf("expected vodrows array, got %T", data["vodrows"])
			}
		})
	}
}

func TestLegacyVODListingRoutes(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{
		"/vod/listing",
		"/vod/listing-0-0-0-0-0-0-0-0-0-2",
		"/vod/recommend",
		"/vod/hot",
		"/vod/latest",
	} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
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
			if body.RetCode != 0 {
				t.Fatalf("expected retcode 0, got %d", body.RetCode)
			}
			data, ok := body.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("expected data object, got %T", body.Data)
			}
			if _, ok := data["vodrows"].([]interface{}); !ok {
				t.Fatalf("expected vodrows array, got %T", data["vodrows"])
			}
		})
	}
}

func TestLegacyVODShowRoute(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/vod/show/1", nil)
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
	if body.RetCode != -1 {
		t.Fatalf("expected missing db retcode -1, got %d", body.RetCode)
	}
	if body.ErrMsg != "记录不存在或已删除" {
		t.Fatalf("unexpected errmsg %q", body.ErrMsg)
	}
}

func TestSendfileRoutes(t *testing.T) {
	router := newTestRouter()

	tests := []struct {
		path        string
		status      int
		contentType string
		retcode     float64
		errmsg      string
		jsonBody    bool
	}{
		{path: "/sendfile/play/test", status: http.StatusOK, retcode: -9999, errmsg: "请登录后操作", jsonBody: true},
		{path: "/sendfile/down/test", status: http.StatusOK, contentType: "text/html", jsonBody: false},
		{path: "/sendfile/play/test.m3u8", status: http.StatusNotFound, jsonBody: false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			router.ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Fatalf("expected status %d, got %d", tt.status, rec.Code)
			}
			if tt.contentType != "" && rec.Header().Get("Content-Type") != tt.contentType {
				t.Fatalf("expected content-type %q, got %q", tt.contentType, rec.Header().Get("Content-Type"))
			}
			if !tt.jsonBody {
				return
			}
			var body map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body["retcode"] != tt.retcode || body["errmsg"] != tt.errmsg {
				t.Fatalf("unexpected response %#v", body)
			}
		})
	}
}

func TestLegacyVODPreviewRoute(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/vod/preView/1/index.m3u8", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Header().Get("Content-Type") != "vnd.apple.mpegurl" {
		t.Fatalf("unexpected content-type %q", rec.Header().Get("Content-Type"))
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty body without mysql, got %q", rec.Body.String())
	}
}

func TestCommentListingRoute(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/comment/listing-61494-0-1", nil)
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
	if body.RetCode != -1 {
		t.Fatalf("expected missing db retcode -1, got %d", body.RetCode)
	}
	if body.ErrMsg != "记录不存在或已被删除" {
		t.Fatalf("unexpected errmsg %q", body.ErrMsg)
	}
}

func TestV2VODShowPlaceholderIsNotListing(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/vod/show/1", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected status %d, got %d", http.StatusNotImplemented, rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["legacy_handler"] != "c.apiv2.vod.show" {
		t.Fatalf("unexpected legacy handler %q", body["legacy_handler"])
	}
}

func TestGetLikeRows(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/getLikeRows?pagesize=100", nil)
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
	if body.RetCode != 0 {
		t.Fatalf("expected retcode 0, got %d", body.RetCode)
	}
	data, ok := body.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", body.Data)
	}
	if _, ok := data["likerows"].([]interface{}); !ok {
		t.Fatalf("expected likerows array, got %T", data["likerows"])
	}
}

func TestGameReadRoutes(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{
		"/game/games",
		"/game/games?platform_id=1&category_id=2",
		"/game/broadcasts",
		"/game/wali/gameList",
		"/game/wali/gameList?category_id=2",
	} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
			}
			var body legacyjson.Response
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.RetCode != 0 {
				t.Fatalf("expected retcode 0, got %d", body.RetCode)
			}
		})
	}
}

func TestWaliGameListFrequentRequiresLogin(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/game/wali/gameList?category_id=5", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != -9999 {
		t.Fatalf("expected retcode -9999, got %d", body.RetCode)
	}
	if body.ErrMsg != "您还没有登录" {
		t.Fatalf("unexpected errmsg %q", body.ErrMsg)
	}
}

func TestUCPMyAffRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ucp/myaff", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != -9999 {
		t.Fatalf("expected retcode -9999, got %d", body.RetCode)
	}
}

func TestNoRoute(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func newTestRouter() http.Handler {
	return NewRouter(Options{
		Config: config.Config{
			Env:             "test",
			ResourceBaseURL: "https://static.example.test",
			SMSCaptcha:      1,
			CaptchaStyle:    0,
			IPDBPath:        "/path/not-found.ipdb",
			MySQLDSN:        "",
			GameResourceURL: "https://image.xjdev.one",
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
}
