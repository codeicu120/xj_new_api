package captcha

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"image/png"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

type staticSecretGenerator struct {
	secret string
}

func (g staticSecretGenerator) Generate(int) (string, error) {
	return g.secret, nil
}

func TestServiceReqPicX(t *testing.T) {
	service := NewService(1, 0, staticSecretGenerator{secret: "abc123"})

	data, err := service.Req()
	if err != nil {
		t.Fatalf("req captcha: %v", err)
	}

	if data.PicURL != "/captcha/picx?abc123" {
		t.Fatalf("unexpected picurl %q", data.PicURL)
	}
	if data.SMSCaptcha != 1 {
		t.Fatalf("unexpected smscaptcha %d", data.SMSCaptcha)
	}
}

func TestServiceReqPic(t *testing.T) {
	service := NewService(0, 1, staticSecretGenerator{secret: "def456"})

	data, err := service.Req()
	if err != nil {
		t.Fatalf("req captcha: %v", err)
	}

	if data.PicURL != "/captcha/pic?def456" {
		t.Fatalf("unexpected picurl %q", data.PicURL)
	}
	if data.SMSCaptcha != 0 {
		t.Fatalf("unexpected smscaptcha %d", data.SMSCaptcha)
	}
}

func TestServiceReqV2ReturnsBase64ImageAndKey(t *testing.T) {
	service := NewService(1, 0, nil)

	data, err := service.ReqV2()
	if err != nil {
		t.Fatalf("req v2 captcha: %v", err)
	}
	if data.SMSCaptcha != 1 {
		t.Fatalf("unexpected smscaptcha %d", data.SMSCaptcha)
	}
	if data.CaptchaKey == "" {
		t.Fatal("expected captcha key")
	}
	decoded, err := url.QueryUnescape(data.PicURL)
	if err != nil {
		t.Fatalf("unescape picurl: %v", err)
	}
	if !strings.HasPrefix(decoded, "data:image/png;base64,") {
		t.Fatalf("unexpected picurl prefix %q", decoded[:min(len(decoded), 32)])
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(decoded, "data:image/png;base64,"))
	if err != nil {
		t.Fatalf("decode png base64: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	if img.Bounds().Dx() != 100 || img.Bounds().Dy() != 34 {
		t.Fatalf("unexpected captcha image size %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestServicePNGRejectsInvalidSecret(t *testing.T) {
	service := NewService(1, 0, nil)

	if _, err := service.PNG("bad"); err != ErrInvalidSecret {
		t.Fatalf("expected invalid secret error, got %v", err)
	}
}

func TestUnpackSecretMatchesPHPEncryptFormat(t *testing.T) {
	secret := "41474d494d677339576a384264564a6a434449475067746f566a52565a56646b413251494f514932"

	code, err := unpackSecret(secret)
	if err != nil {
		t.Fatalf("unpack php secret: %v", err)
	}
	if code != "1234" {
		t.Fatalf("unexpected code %q", code)
	}
}

func TestServiceVerify(t *testing.T) {
	service := NewService(1, 0, nil)
	data, err := service.Req()
	if err != nil {
		t.Fatalf("req captcha: %v", err)
	}
	secret := data.PicURL[strings.Index(data.PicURL, "?")+1:]
	code, err := unpackSecret(secret)
	if err != nil {
		t.Fatalf("unpack generated secret: %v", err)
	}
	if !service.Verify(secret, code) {
		t.Fatal("expected verify success")
	}
	if service.Verify(secret, code+"x") {
		t.Fatal("expected verify failure")
	}
}

func TestServiceReqSecretCanRenderPNG(t *testing.T) {
	service := NewService(1, 0, nil)
	data, err := service.Req()
	if err != nil {
		t.Fatalf("req captcha: %v", err)
	}
	secret := data.PicURL[strings.Index(data.PicURL, "?")+1:]
	body, err := service.PNG(secret)
	if err != nil {
		t.Fatalf("render png: %v", err)
	}
	if len(body) < 24 || string(body[:8]) != "\x89PNG\r\n\x1a\n" {
		t.Fatalf("expected PNG body, got %q", body[:min(len(body), 8)])
	}
}

func TestServicePNGUsesSecretCode(t *testing.T) {
	service := NewService(1, 0, nil)
	left := secretForTestCode(t, "1111")
	right := secretForTestCode(t, "2222")

	leftBody, err := service.PNG(left)
	if err != nil {
		t.Fatalf("render left png: %v", err)
	}
	rightBody, err := service.PNG(right)
	if err != nil {
		t.Fatalf("render right png: %v", err)
	}
	if bytes.Equal(leftBody, rightBody) {
		t.Fatal("expected different captcha images for different codes")
	}
}

func secretForTestCode(t *testing.T, code string) string {
	t.Helper()
	encrypted, err := phpEncrypt(code+"."+strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10), "28ea4")
	if err != nil {
		t.Fatalf("encrypt secret: %v", err)
	}
	return hex.EncodeToString([]byte(encrypted))
}
