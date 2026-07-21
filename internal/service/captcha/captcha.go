package captcha

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"xj_comp/internal/domain"
)

var ErrInvalidSecret = errors.New("invalid captcha secret")

type SecretGenerator interface {
	Generate(style int) (string, error)
}

type RandomSecretGenerator struct{}

func (RandomSecretGenerator) Generate(style int) (string, error) {
	code, err := randomCode(style)
	if err != nil {
		return "", fmt.Errorf("generate captcha secret: %w", err)
	}
	plain := code + "." + strconv.FormatInt(time.Now().Add(10*time.Minute).Unix(), 10)
	encrypted, err := phpEncrypt(plain, "28ea4")
	if err != nil {
		return "", fmt.Errorf("encrypt captcha secret: %w", err)
	}
	return hex.EncodeToString([]byte(encrypted)), nil
}

type Service struct {
	smsCaptcha      int
	captchaStyle    int
	secretGenerator SecretGenerator
}

func NewService(smsCaptcha int, captchaStyle int, secretGenerator SecretGenerator) *Service {
	if secretGenerator == nil {
		secretGenerator = RandomSecretGenerator{}
	}
	return &Service{
		smsCaptcha:      smsCaptcha,
		captchaStyle:    captchaStyle,
		secretGenerator: secretGenerator,
	}
}

func (s *Service) Req() (domain.CaptchaReqData, error) {
	secret, err := s.secretGenerator.Generate(s.captchaStyle)
	if err != nil {
		return domain.CaptchaReqData{}, err
	}

	action := "picx"
	if s.captchaStyle == 1 {
		action = "pic"
	}

	return domain.CaptchaReqData{
		PicURL:     "/captcha/" + action + "?" + secret,
		SMSCaptcha: s.smsCaptcha,
	}, nil
}

func (s *Service) ReqV2() (domain.CaptchaReqV2Data, error) {
	secret, err := s.secretGenerator.Generate(s.captchaStyle)
	if err != nil {
		return domain.CaptchaReqV2Data{}, err
	}
	body, err := s.PNG(secret)
	if err != nil {
		return domain.CaptchaReqV2Data{}, err
	}
	return domain.CaptchaReqV2Data{
		PicURL:     url.QueryEscape("data:image/png;base64," + base64.StdEncoding.EncodeToString(body)),
		SMSCaptcha: s.smsCaptcha,
		CaptchaKey: secret,
	}, nil
}

func (s *Service) PNG(secret string) ([]byte, error) {
	code, err := unpackSecret(secret)
	if err != nil {
		return nil, err
	}
	return NewTestImageService().PNGForCode(code)
}

func (s *Service) Verify(secret, code string) bool {
	actual, err := unpackSecret(secret)
	if err != nil {
		return false
	}
	return code == actual
}

func unpackSecret(secret string) (string, error) {
	raw, err := hex.DecodeString(strings.TrimSpace(secret))
	if err != nil {
		return "", ErrInvalidSecret
	}
	plain, err := phpDecrypt(string(raw), "28ea4")
	if err != nil {
		return "", ErrInvalidSecret
	}
	parts := strings.SplitN(plain, ".", 2)
	if len(parts) != 2 {
		return "", ErrInvalidSecret
	}
	code := parts[0]
	expires, err := strconv.ParseInt(parts[1], 10, 64)
	if code == "" || err != nil || expires < time.Now().Unix() {
		return "", ErrInvalidSecret
	}
	return code, nil
}

func phpEncrypt(content string, key string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	random := fmt.Sprintf("%x", md5.Sum(randomBytes))
	contentBytes := []byte(content)
	encrypted := make([]byte, 0, len(contentBytes)*2)
	for i, b := range contentBytes {
		k := i
		if k >= 32 {
			k = 0
		}
		r := random[k]
		encrypted = append(encrypted, r, b^r)
	}
	keyMD5 := fmt.Sprintf("%x", md5.Sum([]byte(key)))
	second := make([]byte, len(encrypted))
	for i, b := range encrypted {
		k := i
		if k >= 32 {
			k = 0
		}
		second[i] = b ^ keyMD5[k]
	}
	return base64.StdEncoding.EncodeToString(second), nil
}

func phpDecrypt(content string, key string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", err
	}
	keyMD5 := fmt.Sprintf("%x", md5.Sum([]byte(key)))
	first := make([]byte, len(decoded))
	for i, b := range decoded {
		k := i
		if k >= 32 {
			k = 0
		}
		first[i] = b ^ keyMD5[k]
	}
	if len(first)%2 != 0 {
		return "", ErrInvalidSecret
	}
	plain := make([]byte, 0, len(first)/2)
	for i := 0; i < len(first); i += 2 {
		plain = append(plain, first[i+1]^first[i])
	}
	return string(plain), nil
}

func randomCode(style int) (string, error) {
	const digits = "0123456789"
	size := 4
	if style == 1 {
		size = 2
	}
	buf := make([]byte, size)
	random := make([]byte, size)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = digits[int(random[i])%len(digits)]
	}
	return string(buf), nil
}
