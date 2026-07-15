package captcha

import (
	"strings"
	"testing"
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
