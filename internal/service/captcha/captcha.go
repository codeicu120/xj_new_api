package captcha

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"xj_comp/internal/domain"
)

type SecretGenerator interface {
	Generate(style int) (string, error)
}

type RandomSecretGenerator struct{}

func (RandomSecretGenerator) Generate(int) (string, error) {
	buf := make([]byte, 36)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate captcha secret: %w", err)
	}
	return hex.EncodeToString(buf), nil
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
