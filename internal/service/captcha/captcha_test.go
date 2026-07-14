package captcha

import "testing"

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
