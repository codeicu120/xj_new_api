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

func TestCORSPreflight(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/login", nil)
	req.Header.Set("Origin", "https://app.example.test")
	req.Header.Set("Access-Control-Request-Method", "POST")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.test" {
		t.Fatalf("unexpected allow origin %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("unexpected allow credentials %q", got)
	}
}

func TestCORSHeadersOnNormalRequest(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "https://app.example.test")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.test" {
		t.Fatalf("unexpected allow origin %q", got)
	}
}

func TestLegacyPlaceholder(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/__missing_placeholder__", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
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

	for _, path := range []string{"/sms", "/sms/index", "/email", "/email/index", "/aiundress/index", "/playlog", "/playlog/index", "/downlog", "/downlog/index", "/favorite", "/favorite/index", "/minifavorite", "/minifavorite/index", "/v2/minifavorite", "/v2/minifavorite/index"} {
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

func TestAIUndressExternalRoutesReturnRequestFailedWhenConfigMissing(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/aiundress/moduleList", "/aiundress/resourceTypeList?module=4", "/aiundress/resourceList?module=4&typeId=2&page=1&pageSize=20"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected status %d, got %d", path, http.StatusOK, rec.Code)
		}
		if servedBy := rec.Header().Get("X-Served-By"); servedBy != "newbie" {
			t.Fatalf("%s: expected X-Served-By newbie, got %q", path, servedBy)
		}
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s decode response: %v", path, err)
		}
		if body.RetCode != -1 || body.ErrMsg != "请求失败" {
			t.Fatalf("%s unexpected response %#v", path, body)
		}
	}
}

func TestRespondFailureRoutes(t *testing.T) {
	router := newTestRouter()

	cases := map[string]string{
		"/respond/shangfu":  "failed",
		"/respond/pay12":    "Err",
		"/respond/newpaykf": "FAILED",
	}
	for path, want := range cases {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected status %d, got %d", path, http.StatusOK, rec.Code)
		}
		if got := rec.Body.String(); got != want {
			t.Fatalf("%s: expected %q, got %q", path, want, got)
		}
	}
}

func TestRespondChan1InvalidToken(t *testing.T) {
	router := newTestRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/respond/chan1", strings.NewReader("mobi=86.18812345678&token=bad"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != 1 || body.ErrMsg != "校验失败" {
		t.Fatalf("unexpected response %#v", body)
	}
}

func TestUserAuthEdgeRoutes(t *testing.T) {
	router := newTestRouter()

	cases := map[string]struct {
		path    string
		retcode int
		errmsg  string
	}{
		"register":    {path: "/register", retcode: -1, errmsg: "请同意用户协议"},
		"v2register":  {path: "/v2/register", retcode: -1, errmsg: "请同意用户协议"},
		"v2login":     {path: "/v2/login", retcode: -1, errmsg: "用户名未注册"},
		"forgot":      {path: "/forgot", retcode: -1, errmsg: "手机号码填写不正确"},
		"v2forgot":    {path: "/v2/forgot", retcode: -1, errmsg: "请填写手机号码或者邮箱"},
		"delete":      {path: "/delete", retcode: -9999, errmsg: "您还没有登录"},
		"changePhone": {path: "/changePhone", retcode: -9999, errmsg: "请登录后操作"},
		"taskInvite":  {path: "/ucp/task/invite", retcode: -9999, errmsg: "您还没有登录"},
		"ucpProfile":  {path: "/ucp/user/profile", retcode: -9999, errmsg: "您还没有登录"},
		"ucpPasswd":   {path: "/ucp/user/passwd", retcode: -9999, errmsg: "您还没有登录"},
		"aiUpload":    {path: "/aiundress/upload", retcode: -1, errmsg: "请先登录"},
		"aiUndress":   {path: "/aiundress/undress", retcode: -1, errmsg: "请先登录"},
		"aiDelete":    {path: "/aiundress/delete", retcode: -1, errmsg: "请先登录"},
	}
	for name, tc := range cases {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, tc.path, nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected status %d, got %d", name, http.StatusOK, rec.Code)
		}
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s decode response: %v", name, err)
		}
		if body.RetCode != tc.retcode || body.ErrMsg != tc.errmsg {
			t.Fatalf("%s unexpected response %#v", name, body)
		}
	}
}

func TestUCPHighRiskRoutesRequireLogin(t *testing.T) {
	router := newTestRouter()

	paths := []string{
		"/ucp/upgrade",
		"/ucp/task/qrcode",
		"/ucp/task/qrcodeSave",
		"/ucp/task/invitecodeInput",
		"/ucp/task/adviewClick",
		"/ucp/taskbox/taskboxopen",
		"/ucp/taskbox/qrcode",
		"/ucp/user/checkemail",
		"/ucp/user/sendemail",
		"/ucp/user/verifyemail",
		"/ucp/user/bindmobi",
		"/ucp/withdraw/create",
		"/ucp/vippkg/placeorder",
		"/ucp/vippkg/coinorder",
		"/ucp/coinpkg/placeorder",
		"/ucp/beanpkg/placeorder",
		"/ucp/beanpkg/coinorder",
		"/ucp/vodorder/create",
		"/ucp/vodorder/support",
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ucp/task/share", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("/ucp/task/share: expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var shareBody legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &shareBody); err != nil {
		t.Fatalf("/ucp/task/share decode response: %v", err)
	}
	if shareBody.RetCode != 0 {
		t.Fatalf("/ucp/task/share unexpected response %#v", shareBody)
	}

	for _, path := range paths {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected status %d, got %d", path, http.StatusOK, rec.Code)
		}
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s decode response: %v", path, err)
		}
		if body.RetCode != -9999 || body.ErrMsg != "您还没有登录" {
			t.Fatalf("%s unexpected response %#v", path, body)
		}
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/ucp/coinlog/exchange", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("/ucp/coinlog/exchange: expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("/ucp/coinlog/exchange decode response: %v", err)
	}
	if body.RetCode != -1 || body.ErrMsg != "系统已关闭兑换功能" {
		t.Fatalf("/ucp/coinlog/exchange unexpected response %#v", body)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/ucp/task/sign", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("/ucp/task/sign: expected status %d, got %d", http.StatusOK, rec.Code)
	}
	body = legacyjson.Response{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("/ucp/task/sign decode response: %v", err)
	}
	if body.RetCode != -1 || body.ErrMsg != "请登录后操作，客户端游客请先携带信息" {
		t.Fatalf("/ucp/task/sign unexpected response %#v", body)
	}
}

func TestV2MiniFavoriteRequiresLogin(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/v2/minifavorite/listing", "/v2/minifavorite/add?vodid=9", "/v2/minifavorite/remove?vodid=9"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected status %d, got %d", path, http.StatusOK, rec.Code)
		}
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s decode response: %v", path, err)
		}
		if body.RetCode != -9999 || body.ErrMsg != "请登录后操作" {
			t.Fatalf("%s unexpected response %#v", path, body)
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

func TestV2CaptchaReq(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/captcha/req", nil)
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
		t.Fatalf("expected data object, got %T", body.Data)
	}
	picURL, ok := data["picurl"].(string)
	if !ok {
		t.Fatalf("expected picurl string, got %T", data["picurl"])
	}
	if !strings.HasPrefix(picURL, "data%3Aimage%2Fpng%3Bbase64%2C") {
		t.Fatalf("unexpected v2 picurl %q", picURL[:min(len(picURL), 40)])
	}
	if data["captcha_key"] == "" {
		t.Fatalf("expected captcha_key, got %#v", data["captcha_key"])
	}
}

func TestCaptchaPicInvalidSecret(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/captcha/pic", "/captcha/pic?bad", "/captcha/picx", "/captcha/picx?bad", "/v2/captcha/pic", "/v2/captcha/pic?bad", "/v2/captcha/picx", "/v2/captcha/picx?bad"} {
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

func TestV2CaptchaVerify(t *testing.T) {
	router := newTestRouter()

	reqRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/captcha/req", nil)
	router.ServeHTTP(reqRec, req)

	var reqBody legacyjson.Response
	if err := json.Unmarshal(reqRec.Body.Bytes(), &reqBody); err != nil {
		t.Fatalf("decode req response: %v", err)
	}
	key := reqBody.Data.(map[string]interface{})["captcha_key"].(string)

	rec := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v2/captcha/verify", strings.NewReader("captcha_key="+key+"&captcha_code=wrong"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(rec, req)
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode wrong response: %v", err)
	}
	if body.RetCode != -1 || body.ErrMsg != "验证失败" {
		t.Fatalf("unexpected wrong response %#v", body)
	}
}

func TestCaptchaVerifyRoute(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/captcha/verify", strings.NewReader("captcha_key=bad&captcha_code=wrong"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != -1 || body.ErrMsg != "验证失败" {
		t.Fatalf("unexpected response %#v", body)
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

func TestV2CaptchaTestPNG(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/captcha/test", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "image/png" {
		t.Fatalf("expected image/png content type, got %q", contentType)
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

func TestCommunityShowMissingTopic(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/community/show?tid=0", nil)
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
	if body.RetCode != -1 || body.ErrMsg != "记录不存在或已删除" {
		t.Fatalf("unexpected response %#v", body)
	}
}

func TestCommunityPublicReadRoutes(t *testing.T) {
	router := newTestRouter()

	tests := []struct {
		path    string
		retcode int
		errmsg  string
	}{
		{path: "/community/categories", retcode: 0},
		{path: "/community/slides", retcode: 0},
		{path: "/community/search", retcode: -1, errmsg: "请输入关键词"},
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
			t.Fatalf("%s unexpected response %#v", tt.path, body)
		}
	}
}

func TestCommunityUpRoutesRequireLogin(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/community/attention?tid=9", "/community/up?tid=9", "/community/up_comment?cid=1", "/community/comment?tid=9&content=hello", "/community/post?title=a&content=b"} {
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
		if body.RetCode != -9999 || body.ErrMsg != "请登录后操作" {
			t.Fatalf("%s unexpected response %#v", path, body)
		}
	}
}

func TestTaskboxShareRoute(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ucp/taskbox/share", nil)
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

func TestOneGoBetRequiresLogin(t *testing.T) {
	router := newTestRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/onego/bet", strings.NewReader("quantity=0"))
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != -9999 || body.ErrMsg != "您还没有登录" {
		t.Fatalf("expected login error, got %#v", body)
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

func TestCommentPostRouteRequiresLogin(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/comment/post", strings.NewReader("vodid=1&content=hello"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
	if body.RetCode != -9999 || body.ErrMsg != "请注册会员并登录APP才可以发表评论噢" {
		t.Fatalf("unexpected response %#v", body)
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

func TestUCPTaskIndexRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/ucp/task", "/ucp/task/index"} {
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
			if body.RetCode != -9999 || body.ErrMsg != "您还没有登录" {
				t.Fatalf("unexpected response %#v", body)
			}
		})
	}
}

func TestUCPVODOrderIndexRequiresLoginWithoutToken(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/ucp/vodorder", "/ucp/vodorder/index"} {
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
			if body.RetCode != -9999 || body.ErrMsg != "您还没有登录" {
				t.Fatalf("unexpected response %#v", body)
			}
		})
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

	for _, path := range []string{"/ucp/withdraw", "/ucp/withdraw/index?wdtype=1", "/ucp/withdraw/listing"} {
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

func TestUCPWithdrawRuleRoute(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ucp/withdraw/rule", nil)
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
		t.Fatalf("unexpected response %#v", body)
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

func TestPaymentPaywaysRouteMissingPayment(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/payment/payways", nil)
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
	if body.RetCode != -1 || body.ErrMsg != "记录不存在或已支付" {
		t.Fatalf("unexpected response %#v", body)
	}
}

func TestPaymentChPaywayRouteMissingPayment(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/payment/chpayway", nil)
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
	if body.RetCode != -1 || body.ErrMsg != "记录不存在或已支付" {
		t.Fatalf("unexpected response %#v", body)
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
		{path: "/payment/shangfu", retcode: 0, errmsg: "支付成功回调"},
		{path: "/payment/wappay3", retcode: 0, errmsg: "支付成功回调"},
		{path: "/payment/wappay4", retcode: 0, errmsg: "支付成功回调"},
		{path: "/payment/wappay4a", retcode: 0, errmsg: "支付成功回调"},
		{path: "/payment/wappay5", retcode: 0, errmsg: "支付成功回调"},
		{path: "/payment/hawpay", retcode: 0, errmsg: "支付成功回调"},
		{path: "/payment/easypay", retcode: 0, errmsg: "支付成功回调"},
		{path: "/payment/pay6", retcode: 0, errmsg: "支付成功回调"},
		{path: "/payment/reqpay", retcode: -1, errmsg: "记录不存在或已支付"},
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

func TestPaymentPay12ReqErrorHTML(t *testing.T) {
	router := newTestRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/payment/pay12req", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if servedBy := rec.Header().Get("X-Served-By"); servedBy != "newbie" {
		t.Fatalf("expected X-Served-By newbie, got %q", servedBy)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("expected html content type, got %q", ct)
	}
	if body := rec.Body.String(); !strings.Contains(body, "记录不存在或已支付") {
		t.Fatalf("expected pay error message in body, got %q", body)
	}
}

func TestGameHighRiskRoutesRequireLogin(t *testing.T) {
	router := newTestRouter()
	for _, path := range []string{
		"/game/wali/topup",
		"/game/wali/withdraw",
		"/game/wali/enter",
		"/game/lottery/topup",
		"/game/lottery/withdraw",
		"/game/lottery/enter",
		"/game/lottery/balance",
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s expected status %d, got %d", path, http.StatusOK, rec.Code)
		}
		var body legacyjson.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s decode response: %v", path, err)
		}
		if body.RetCode != -9999 || body.ErrMsg != "您还没有登录" {
			t.Fatalf("%s expected not-login response, got %#v", path, body)
		}
	}
}

func TestInviteBindEdgeRoutes(t *testing.T) {
	router := newTestRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/invite/bind", strings.NewReader(""))
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != -9999 || body.ErrMsg != "您还没有登录" {
		t.Fatalf("expected not-login response, got %#v", body)
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

func TestMiniVODReqMediaRoutes(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/minivod/reqplay/0", "/minivod/reqdown/0"} {
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
		if body.RetCode != 1 || body.ErrMsg != "记录不存在或已被删除" {
			t.Fatalf("%s unexpected response %#v", path, body)
		}
	}
}

func TestMiniVODThrowCoinRequiresLogin(t *testing.T) {
	router := newTestRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/minivod/throwcoin/1", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var body legacyjson.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RetCode != -9999 || body.ErrMsg != "需登录后方可使用投币功能" {
		t.Fatalf("unexpected response %#v", body)
	}
}

func TestMiniVODReqListRoute(t *testing.T) {
	router := newTestRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/minivod/reqlist", nil)
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
		t.Fatalf("unexpected response %#v", body)
	}
}

func TestVODReqMediaRoutes(t *testing.T) {
	router := newTestRouter()

	for _, path := range []string{"/vod/reqplay/0", "/vod/reqdown/0", "/v2/vod/reqplay/0", "/v2/vod/reqdown/0"} {
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
		if body.RetCode != 1 || body.ErrMsg != "记录不存在或已被删除" {
			t.Fatalf("%s unexpected response %#v", path, body)
		}
	}
}

func TestStarLiveEdgeRoutes(t *testing.T) {
	router := newTestRouter()
	tests := []struct {
		path string
		body string
		msg  string
	}{
		{path: "/starLive/gameBet", body: `{"memberId":"tourist-member-123456"}`, msg: "游客用户请先登录"},
		{path: "/starLive/gameWin", body: `{"memberId":""}`, msg: "未知用户"},
		{path: "/starLive/translate", body: `{"memberId":"tourist-member-123456"}`, msg: "游客用户请先登录"},
		{path: "/starLive/tryAgain", body: `{"busiType":9}`, msg: "未知业务类型"},
	}

	for _, tt := range tests {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s expected status %d, got %d", tt.path, http.StatusOK, rec.Code)
		}
		var body struct {
			Code int                    `json:"code"`
			Data map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s decode response: %v", tt.path, err)
		}
		if body.Code != -1 || body.Data["msg"] != tt.msg {
			t.Fatalf("%s unexpected response %#v", tt.path, body)
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
