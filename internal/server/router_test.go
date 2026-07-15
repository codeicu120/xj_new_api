package server

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"xj_comp/internal/config"
	"xj_comp/internal/handler"
	"xj_comp/internal/legacyjson"
	attachService "xj_comp/internal/service/attach"
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

func TestAttachIndexEmptyResponse(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/attach", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "" {
		t.Fatalf("expected empty body, got %q", rec.Body.String())
	}
}

func TestSMSAndEmailIndexEmptyResponse(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/sms", "/sms/index", "/email", "/email/index", "/aiundress/index", "/playlog", "/playlog/index", "/downlog", "/downlog/index", "/favorite", "/favorite/index", "/minifavorite", "/minifavorite/index"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected status %d, got %d", path, http.StatusOK, rec.Code)
		}
		if rec.Body.String() != "" {
			t.Fatalf("%s: expected empty body, got %q", path, rec.Body.String())
		}
	}
}

func TestAttachUpAvatarRequiresLogin(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/attach/upavatar", strings.NewReader("avatarid=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != -9999 || body.ErrMsg != "您还没有登录" {
		t.Fatalf("unexpected response %+v", body)
	}
}

func TestAttachUpAvatarRejectsInvalidAvatarIDWithPHPErrorShape(t *testing.T) {
	service := attachService.NewService(&attachTestStore{user: map[string]interface{}{"uid": "8"}})
	router := gin.New()
	router.POST("/attach/upavatar", handler.NewAttachHandler(service).UpAvatar)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/attach/upavatar", strings.NewReader("avatarid=abc"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["retcode"] != float64(-1) || body["errmsg"] != "请选择系统头像" {
		t.Fatalf("unexpected response %+v", body)
	}
	if _, ok := body["data"]; ok {
		t.Fatal("expected data to be omitted")
	}
}

type attachTestStore struct {
	user map[string]interface{}
}

func (s *attachTestStore) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return s.user, nil
}

func (s *attachTestStore) UpdateAvatar(context.Context, int, string) error {
	return nil
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

func TestCaptchaPicInvalidSecret(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/captcha/pic", "/captcha/pic?bad", "/captcha/picx", "/captcha/picx?bad"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s expected status %d, got %d", path, http.StatusNotFound, rec.Code)
		}
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s decode response: %v", path, err)
		}
		if body.RetCode != -4 || body.ErrMsg != "验证码无效" {
			t.Fatalf("%s unexpected response %#v", path, body)
		}
	}
}

func TestCaptchaPicFromReqSecret(t *testing.T) {
	router := newTestRouter()

	reqRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/captcha/req", nil)
	router.ServeHTTP(reqRec, req)

	var reqBody legacyjson.Response
	if err := json.Unmarshal(reqRec.Body.Bytes(), &reqBody); err != nil {
		t.Fatalf("decode req response: %v", err)
	}
	data := reqBody.Data.(map[string]interface{})
	picURL := data["picurl"].(string)

	rec := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, picURL, nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "image/png" {
		t.Fatalf("expected image/png content type, got %q", contentType)
	}
	body := rec.Body.Bytes()
	if len(body) < 24 || string(body[:8]) != "\x89PNG\r\n\x1a\n" {
		t.Fatalf("expected PNG response, got %q", body[:min(len(body), 8)])
	}
	if width := binary.BigEndian.Uint32(body[16:20]); width != 100 {
		t.Fatalf("expected PNG width 100, got %d", width)
	}
	if height := binary.BigEndian.Uint32(body[20:24]); height != 34 {
		t.Fatalf("expected PNG height 34, got %d", height)
	}
}

func TestTestCaptchaPNG(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if servedBy := rec.Header().Get("X-Served-By"); servedBy != "newbie" {
		t.Fatalf("expected X-Served-By newbie, got %q", servedBy)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "image/png" {
		t.Fatalf("expected image/png content type, got %q", contentType)
	}
	body := rec.Body.Bytes()
	if len(body) < 24 || string(body[:8]) != "\x89PNG\r\n\x1a\n" {
		t.Fatalf("expected PNG response, got %q", body[:min(len(body), 8)])
	}
	if width := binary.BigEndian.Uint32(body[16:20]); width != 100 {
		t.Fatalf("expected PNG width 100, got %d", width)
	}
	if height := binary.BigEndian.Uint32(body[20:24]); height != 34 {
		t.Fatalf("expected PNG height 34, got %d", height)
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

func TestOneGoRulesNotOpenWithoutMySQL(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/onego/rules", nil)
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
	if body.RetCode != -1 || body.ErrMsg != "系统尚未开放该活动" {
		t.Fatalf("unexpected response %#v", body)
	}
}

func TestOneGoRootUsesRules(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/onego", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != -1 || body.ErrMsg != "系统尚未开放该活动" {
		t.Fatalf("unexpected response %#v", body)
	}
}

func TestOneGoEmptyIndexRoute(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/onego/index", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "text/html" {
		t.Fatalf("expected text/html content type, got %q", contentType)
	}
	if body := rec.Body.String(); body != "" {
		t.Fatalf("expected empty body, got %q", body)
	}
}

func TestOneGoRoomsNotOpenWithoutMySQL(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/onego/rooms", nil)
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
	if body.RetCode != -1 || body.ErrMsg != "系统尚未开放该活动" {
		t.Fatalf("unexpected response %#v", body)
	}
}

func TestOneGoCurrentNotOpenWithoutMySQL(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/onego/current?roomid=1", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != -1 || body.ErrMsg != "系统尚未开放该活动" {
		t.Fatalf("unexpected response %#v", body)
	}
}

func TestOneGoLastNoDataWithoutMySQL(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/onego/last", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != -1 || body.ErrMsg != "暂无数据" {
		t.Fatalf("unexpected response %#v", body)
	}
}

func TestOneGoHash(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/onego/hash?plaintext=abc", nil)
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
	payload, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected nested data object, got %T", data["data"])
	}
	if payload["hash_code"] != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("unexpected hash_code %v", payload["hash_code"])
	}
	if payload["hash_number"] != "120015" {
		t.Fatalf("unexpected hash_number %v", payload["hash_number"])
	}
}

func TestOneGoHashPostAndMissingPlaintext(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/onego/hash", strings.NewReader("plaintext=abc"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/onego/hash", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != -1 || body.ErrMsg != "请传入参数" {
		t.Fatalf("unexpected response %#v", body)
	}
}

func TestOneGoLuckyWithoutMySQL(t *testing.T) {
	router := newTestRouter()

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/onego/lucky?page=1"},
		{method: http.MethodPost, path: "/onego/lucky", body: "page=2"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
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
			rows, ok := data["data"].([]interface{})
			if !ok {
				t.Fatalf("expected nested data array, got %T", data["data"])
			}
			if len(rows) != 0 {
				t.Fatalf("expected empty ranks, got %#v", rows)
			}
		})
	}
}

func TestOneGoMarqueeNoDataWithoutMySQL(t *testing.T) {
	router := newTestRouter()

	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/onego/marquee"},
		{method: http.MethodPost, path: "/onego/marquee"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
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
			if body.RetCode != -1 || body.ErrMsg != "暂无数据" {
				t.Fatalf("unexpected response %#v", body)
			}
		})
	}
}

func TestSpecialListingWithoutMySQL(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/special/listing", "/special/listing-1-3-0?page=2"} {
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
			if _, ok := data["actorrows"].([]interface{}); !ok {
				t.Fatalf("expected actorrows array, got %T", data["actorrows"])
			}
		})
	}
}

func TestSpecialIndexEmpty(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/special/index", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if body := rec.Body.String(); body != "" {
		t.Fatalf("expected empty body, got %q", body)
	}
	if servedBy := rec.Header().Get("X-Served-By"); servedBy != "newbie" {
		t.Fatalf("expected X-Served-By newbie, got %q", servedBy)
	}
}

func TestSpecialDetailNotFoundWithoutMySQL(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/special/detail/123", "/special/detail/123-1"} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, path, nil)
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
			if body.RetCode != -1 || body.ErrMsg != "记录不存在或已被删除" {
				t.Fatalf("unexpected response %#v", body)
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

func TestVODErrorReportRoutesRequireParams(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/vod/errorreport", "/v2/vod/errorreport"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader("vodid=1"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s expected status %d, got %d", path, http.StatusOK, rec.Code)
		}
		if servedBy := rec.Header().Get("X-Served-By"); servedBy != "newbie" {
			t.Fatalf("%s expected X-Served-By newbie, got %q", path, servedBy)
		}
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s decode response: %v", path, err)
		}
		if body.RetCode != -9999 || body.ErrMsg != "缺少参数" {
			t.Fatalf("%s expected missing params, got %#v", path, body)
		}
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

func TestV2VODShowUsesRealHandler(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/vod/show/1", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != -1 || body.ErrMsg != "记录不存在或已删除" {
		t.Fatalf("unexpected response %+v", body)
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

func TestSearchRoutes(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/search", "/search?wd=test&page=1", "/search?wd=test&free=1&page=1"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s expected status %d, got %d", path, http.StatusOK, rec.Code)
		}
		if servedBy := rec.Header().Get("X-Served-By"); servedBy != "newbie" {
			t.Fatalf("%s expected X-Served-By newbie, got %q", path, servedBy)
		}
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s decode response: %v", path, err)
		}
		if body.RetCode != 0 {
			t.Fatalf("%s expected retcode 0, got %d", path, body.RetCode)
		}
	}
}

func TestMiniSearchRoutes(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/minisearch", "/minisearch?wd=test&page=1"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s expected status %d, got %d", path, http.StatusOK, rec.Code)
		}
		if servedBy := rec.Header().Get("X-Served-By"); servedBy != "newbie" {
			t.Fatalf("%s expected X-Served-By newbie, got %q", path, servedBy)
		}
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s decode response: %v", path, err)
		}
		if body.RetCode != 0 {
			t.Fatalf("%s expected retcode 0, got %d", path, body.RetCode)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/minisearch", strings.NewReader("wd=test&page=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /minisearch expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("POST /minisearch decode response: %v", err)
	}
	data, ok := body.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", body.Data)
	}
	pageinfo, ok := data["pageinfo"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected pageinfo object, got %T", data["pageinfo"])
	}
	if pageinfo["page_url"] != "/search?wd=test&page=[?]" {
		t.Fatalf("unexpected page_url %v", pageinfo["page_url"])
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

func TestLotteryGameListFrequentRequiresLogin(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/game/lottery/gameList?category_id=5", nil)
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

func TestLotteryGameListOrdinaryCategory(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/game/lottery/gameList?category_id=1", nil)
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

func TestUCPRollTitle(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ucp/rolltitle", nil)
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
	if _, ok := data["messages"].([]interface{}); !ok {
		t.Fatalf("expected messages array, got %T", data["messages"])
	}
}

func TestUCPPaymentListingRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/ucp/payment/listing?page=1"},
		{method: http.MethodPost, path: "/ucp/payment/listing", body: "page=1"},
		{method: http.MethodGet, path: "/ucp/payment"},
		{method: http.MethodGet, path: "/ucp/payment/index"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
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
			if body.RetCode != -9999 {
				t.Fatalf("expected retcode -9999, got %d", body.RetCode)
			}
			if body.ErrMsg != "您还没有登录" {
				t.Fatalf("unexpected errmsg %q", body.ErrMsg)
			}
		})
	}
}

func TestUCPSafePayLogRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ucp/payment/safepaylog", strings.NewReader(""))
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
	if body.RetCode != -9999 {
		t.Fatalf("expected retcode -9999, got %d", body.RetCode)
	}
	if body.ErrMsg != "您还没有登录" {
		t.Fatalf("unexpected errmsg %q", body.ErrMsg)
	}
}

func TestUCPAccountIndexRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/ucp/account", "/ucp/account/index"} {
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
			if body.RetCode != -9999 {
				t.Fatalf("expected retcode -9999, got %d", body.RetCode)
			}
			if body.ErrMsg != "您还没有登录" {
				t.Fatalf("unexpected errmsg %q", body.ErrMsg)
			}
		})
	}
}

func TestUCPBalanceLogRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/ucp/account/balancelog?page=1"},
		{method: http.MethodPost, path: "/ucp/account/balancelog", body: "page=1"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
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
			if body.RetCode != -9999 {
				t.Fatalf("expected retcode -9999, got %d", body.RetCode)
			}
			if body.ErrMsg != "您还没有登录" {
				t.Fatalf("unexpected errmsg %q", body.ErrMsg)
			}
		})
	}
}

func TestUCPWithdrawIndexRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/ucp/withdraw", "/ucp/withdraw/index?wdtype=1"} {
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
			if body.RetCode != -9999 {
				t.Fatalf("expected retcode -9999, got %d", body.RetCode)
			}
			if body.ErrMsg != "您还没有登录" {
				t.Fatalf("unexpected errmsg %q", body.ErrMsg)
			}
		})
	}
}

func TestUCPCoinLogIndexRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/ucp/coinlog"},
		{method: http.MethodGet, path: "/ucp/coinlog/index"},
		{method: http.MethodPost, path: "/ucp/coinlog/index"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
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
			if body.RetCode != -9999 {
				t.Fatalf("expected retcode -9999, got %d", body.RetCode)
			}
			if body.ErrMsg != "您还没有登录" {
				t.Fatalf("unexpected errmsg %q", body.ErrMsg)
			}
		})
	}
}

func TestUCPCoinLogInviteLogRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/ucp/coinlog/invitelog?page=1"},
		{method: http.MethodPost, path: "/ucp/coinlog/invitelog", body: "page=1"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
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
			if body.RetCode != -9999 {
				t.Fatalf("expected retcode -9999, got %d", body.RetCode)
			}
			if body.ErrMsg != "您还没有登录" {
				t.Fatalf("unexpected errmsg %q", body.ErrMsg)
			}
		})
	}
}

func TestUCPCoinLogBonusLogRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/ucp/coinlog/bonuslog?page=1"},
		{method: http.MethodPost, path: "/ucp/coinlog/bonuslog", body: "page=1"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
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
			if body.RetCode != -9999 {
				t.Fatalf("expected retcode -9999, got %d", body.RetCode)
			}
			if body.ErrMsg != "您还没有登录" {
				t.Fatalf("unexpected errmsg %q", body.ErrMsg)
			}
		})
	}
}

func TestUCPAffCenterRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/ucp/affcenter"},
		{method: http.MethodPost, path: "/ucp/affcenter"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
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
			if body.RetCode != -9999 {
				t.Fatalf("expected retcode -9999, got %d", body.RetCode)
			}
			if body.ErrMsg != "您还没有登录" {
				t.Fatalf("unexpected errmsg %q", body.ErrMsg)
			}
		})
	}
}

func TestUCPIndexGuestMissingRowOmitsData(t *testing.T) {
	router := newTestRouter()

	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/ucp/index"},
		{method: http.MethodPost, path: "/ucp/index"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
			}
			if servedBy := rec.Header().Get("X-Served-By"); servedBy != "newbie" {
				t.Fatalf("expected X-Served-By newbie, got %q", servedBy)
			}
			var body map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body["retcode"] != float64(-1) {
				t.Fatalf("expected retcode -1, got %#v", body["retcode"])
			}
			if body["errmsg"] != "请登录后操作，客户端游客请先携带信息" {
				t.Fatalf("unexpected errmsg %q", body["errmsg"])
			}
			if _, ok := body["data"]; ok {
				t.Fatalf("expected data to be omitted, got %#v", body["data"])
			}
		})
	}
}

func TestUCPFeedbackRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ucp/feedback?page=1", nil)
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
	if body.RetCode != -9999 {
		t.Fatalf("expected retcode -9999, got %d", body.RetCode)
	}
	if body.ErrMsg != "您还没有登录" {
		t.Fatalf("unexpected errmsg %q", body.ErrMsg)
	}
}

func TestUCPFeedbackPostNotHandledByListing(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ucp/feedback", strings.NewReader("content=test"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err == nil && body.RetCode == 0 {
			t.Fatalf("POST /ucp/feedback should not be handled by listing, got %#v", body)
		}
	}
}

func TestUCPFeedbackIndexRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ucp/feedback/index", nil)
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
	if body.RetCode != -9999 {
		t.Fatalf("expected retcode -9999, got %d", body.RetCode)
	}
	if body.ErrMsg != "您还没有登录" {
		t.Fatalf("unexpected errmsg %q", body.ErrMsg)
	}
}

func TestUCPFeedbackIndexPostNotHandled(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ucp/feedback/index", strings.NewReader(""))
	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err == nil && body.RetCode == 0 {
			t.Fatalf("POST /ucp/feedback/index should not be handled, got %#v", body)
		}
	}
}

func TestUCPFeedbackNewListingRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ucp/feedback/listing?type=1&page=1", nil)
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
	if body.RetCode != -9999 {
		t.Fatalf("expected retcode -9999, got %d", body.RetCode)
	}
	if body.ErrMsg != "您还没有登录" {
		t.Fatalf("unexpected errmsg %q", body.ErrMsg)
	}
}

func TestUCPFeedbackNewListingPostNotHandled(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ucp/feedback/listing", strings.NewReader("type=1&page=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err == nil && body.RetCode == 0 {
			t.Fatalf("POST /ucp/feedback/listing should not be handled, got %#v", body)
		}
	}
}

func TestUCPFeedbackDetailRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ucp/feedback/detail?id=1917132", nil)
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
	if body.RetCode != -9999 {
		t.Fatalf("expected retcode -9999, got %d", body.RetCode)
	}
	if body.ErrMsg != "您还没有登录" {
		t.Fatalf("unexpected errmsg %q", body.ErrMsg)
	}
}

func TestUCPFeedbackDetailPostNotHandled(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ucp/feedback/detail", strings.NewReader("id=1917132"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err == nil && body.RetCode == 0 {
			t.Fatalf("POST /ucp/feedback/detail should not be handled, got %#v", body)
		}
	}
}

func TestUCPMsgRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/ucp/msg?page=1", "/ucp/msg/index?page=1"} {
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
			if body.RetCode != -9999 {
				t.Fatalf("expected retcode -9999, got %d", body.RetCode)
			}
			if body.ErrMsg != "您还没有登录" {
				t.Fatalf("unexpected errmsg %q", body.ErrMsg)
			}
		})
	}
}

func TestUCPMsgPostNotHandledByListing(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ucp/msg", strings.NewReader("page=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err == nil && body.RetCode == 0 {
			t.Fatalf("POST /ucp/msg should not be handled by listing, got %#v", body)
		}
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

func TestPaymentUnpaidRoute(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/payment/unpaid", nil)
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
	if body.RetCode != 0 || body.ErrMsg != "" {
		t.Fatalf("expected ok response, got %#v", body)
	}
	data, ok := body.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data map, got %#v", body.Data)
	}
	if data["total_count"] != float64(0) {
		t.Fatalf("expected total_count 0, got %#v", data["total_count"])
	}
}

func TestPaymentQueryRoutesNoAccess(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/payment/index", "/payment/query"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s expected status %d, got %d", path, http.StatusOK, rec.Code)
		}
		if servedBy := rec.Header().Get("X-Served-By"); servedBy != "newbie" {
			t.Fatalf("%s expected X-Served-By newbie, got %q", path, servedBy)
		}
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s decode response: %v", path, err)
		}
		if body.RetCode != -1 || body.ErrMsg != "无权限" {
			t.Fatalf("%s expected no access, got %#v", path, body)
		}
	}
}

func TestPaymentCallbackStatusRoutes(t *testing.T) {
	router := newTestRouter()

	tests := []struct {
		path    string
		retcode int
		errmsg  string
	}{
		{path: "/payment/success", retcode: 0, errmsg: "支付成功回调"},
		{path: "/payment/failed", retcode: -1, errmsg: "支付失败回调"},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s expected status %d, got %d", tt.path, http.StatusOK, rec.Code)
		}
		if servedBy := rec.Header().Get("X-Served-By"); servedBy != "newbie" {
			t.Fatalf("%s expected X-Served-By newbie, got %q", tt.path, servedBy)
		}
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s decode response: %v", tt.path, err)
		}
		if body.RetCode != tt.retcode || body.ErrMsg != tt.errmsg {
			t.Fatalf("%s expected retcode=%d errmsg=%q, got %#v", tt.path, tt.retcode, tt.errmsg, body)
		}
	}
}

func TestCommentEmptyIndexRoutes(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/comment", "/comment/index"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s expected status %d, got %d", path, http.StatusOK, rec.Code)
		}
		if body := rec.Body.String(); body != "" {
			t.Fatalf("%s expected empty body, got %q", path, body)
		}
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
