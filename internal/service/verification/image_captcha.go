package verification

import "context"

// ImageCaptchaVerifier adapts the PHP-compatible encrypted captcha secret to
// verification endpoints. The other verification providers remain disabled
// until their server-side clients are configured.
type ImageCaptchaVerifier struct {
	verify func(secret, code string) bool
}

func NewImageCaptchaVerifier(verify func(secret, code string) bool) ImageCaptchaVerifier {
	return ImageCaptchaVerifier{verify: verify}
}

func (v ImageCaptchaVerifier) VerifyImage(_ context.Context, key, code string) bool {
	return v.verify != nil && v.verify(key, code)
}

func (ImageCaptchaVerifier) VerifyGoogle(context.Context, string) bool {
	return false
}

func (ImageCaptchaVerifier) VerifyTencent(context.Context, string, string, string) bool {
	return false
}

func (ImageCaptchaVerifier) VerifySelf(context.Context, string, string) bool {
	return false
}
