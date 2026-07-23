package verification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/mail"
	"regexp"
	"strings"
	"sync"
	"time"

	userRepo "xj_comp/internal/repository/user"
)

var ErrLoginRequired = errors.New("login required")

type Store interface {
	SettingByUUID(ctx context.Context, uuid string) (map[string]interface{}, error)
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type Limiter interface {
	Count(ctx context.Context, key string, window time.Duration) (int, error)
	Incr(ctx context.Context, key string, ttl time.Duration, value string) error
}

type CaptchaVerifier interface {
	VerifyImage(ctx context.Context, key string, code string) bool
	VerifyGoogle(ctx context.Context, ticket string) bool
	VerifyTencent(ctx context.Context, ticket string, randstr string, ip string) bool
	VerifySelf(ctx context.Context, endpoint string, token string) bool
}

type SMSSender interface {
	SendSMS(ctx context.Context, platform int, config string, mobi string, code string, sendCount int) error
}

type MailSender interface {
	SendMail(ctx context.Context, config map[string]interface{}, email string, subject string, body string) error
}

type Service struct {
	store    Store
	limiter  Limiter
	captcha  CaptchaVerifier
	sms      SMSSender
	mail     MailSender
	now      func() time.Time
	randCode func() string
}

func NewService(store Store, limiter Limiter, captcha CaptchaVerifier, sms SMSSender, mail MailSender) *Service {
	if limiter == nil {
		limiter = NewMemoryLimiter()
	}
	if captcha == nil {
		captcha = RejectCaptchaVerifier{}
	}
	if sms == nil {
		sms = DisabledSMSSender{}
	}
	if mail == nil {
		mail = DisabledMailSender{}
	}
	return &Service{
		store:   store,
		limiter: limiter,
		captcha: captcha,
		sms:     sms,
		mail:    mail,
		now:     time.Now,
		randCode: func() string {
			return fmt.Sprintf("%06d", rand.Intn(1_000_000))
		},
	}
}

type SendSMSRequest struct {
	Action      string
	Token       string
	MobiPrefix  string
	Mobi        string
	CaptchaKey  string
	CaptchaCode string
	GTicket     string
	TXTicket    string
	TXRandstr   string
	SelfToken   string
	SendCount   int
	UserAgent   string
	ClientIP    string
}

type SendEmailRequest struct {
	Email       string
	CaptchaKey  string
	CaptchaCode string
	GTicket     string
	TXTicket    string
	TXRandstr   string
	SelfToken   string
	SendCount   int
	ClientIP    string
}

func (s *Service) SendV(ctx context.Context, req SendSMSRequest) (string, error) {
	if androidVersion(req.UserAgent) > 0 && androidVersion(req.UserAgent) < 416 {
		return "当前版本短信功能已关闭，请下载最新版本", nil
	}
	mobi, ok := normalizeMobi(req.MobiPrefix, req.Mobi)
	if !ok {
		return "手机号码填写不正确", nil
	}
	setting, err := s.setting(ctx)
	if err != nil {
		return "", err
	}
	limit := atoi(setting["smsphonelimit"])
	if limit == 0 {
		limit = 5
	}
	if atoi(setting["smscaptcha"]) == 0 {
		return s.sendSMS(ctx, "sms.sendv", mobi, req.SendCount, limit, req.ClientIP, setting)
	}
	if msg := s.verifyCaptcha(ctx, req.CaptchaKey, req.CaptchaCode, req.GTicket, req.TXTicket, req.TXRandstr, req.SelfToken, req.ClientIP, str(setting["selfcaptchaurl"])); msg != "" {
		return msg, nil
	}
	return s.sendSMS(ctx, "sms.sendv", mobi, req.SendCount, limit, req.ClientIP, setting)
}

func (s *Service) SendU(ctx context.Context, req SendSMSRequest) (string, error) {
	user, err := s.store.UserBySession(ctx, userRepo.CleanToken(req.Token))
	if err != nil {
		return "", err
	}
	if atoi(user["uid"]) == 0 {
		return "", ErrLoginRequired
	}
	mobi := str(user["mobi"])
	if strings.TrimSpace(req.Mobi) != "" {
		var ok bool
		mobi, ok = normalizeMobi(req.MobiPrefix, req.Mobi)
		if !ok {
			return "手机号码填写不正确", nil
		}
	}
	setting, err := s.setting(ctx)
	if err != nil {
		return "", err
	}
	return s.sendSMS(ctx, "sms.sendu", mobi, req.SendCount, 10, req.ClientIP, setting)
}

func (s *Service) SendEmail(ctx context.Context, req SendEmailRequest) (string, error) {
	email := strings.TrimSpace(req.Email)
	if !validEmail(email) {
		return "邮箱格式填写不正确", nil
	}
	setting, err := s.setting(ctx)
	if err != nil {
		return "", err
	}
	limit := atoi(setting["smsphonelimit"])
	if limit == 0 {
		limit = 5
	}
	if atoi(setting["smscaptcha"]) != 0 {
		if msg := s.verifyCaptcha(ctx, req.CaptchaKey, req.CaptchaCode, req.GTicket, req.TXTicket, req.TXRandstr, req.SelfToken, req.ClientIP, str(setting["selfcaptchaurl"])); msg != "" {
			return msg, nil
		}
	}
	return s.sendMail(ctx, email, req.SendCount, limit, setting)
}

func (s *Service) verifyCaptcha(ctx context.Context, key string, code string, gTicket string, txTicket string, txRandstr string, selfToken string, ip string, selfURL string) string {
	mode := 1
	if gTicket != "" {
		mode = 2
	}
	if txTicket != "" {
		mode = 3
	}
	if selfToken != "" {
		mode = 4
	}
	switch mode {
	case 1:
		if key == "" {
			return "未提供图形验证码串"
		}
		if code == "" {
			return "请提供写图形验证码"
		}
		if !s.captcha.VerifyImage(ctx, key, code) {
			return "验证码不正确，点击图片可刷新验证码"
		}
	case 2:
		if !s.captcha.VerifyGoogle(ctx, gTicket) {
			return "人机验证失败，请重新验证"
		}
	case 3:
		if txTicket == "" || txRandstr == "" {
			return "未提供腾讯验证票据"
		}
		if !s.captcha.VerifyTencent(ctx, txTicket, txRandstr, ip) {
			return "腾讯验证失败，请重新验证"
		}
	case 4:
		if !s.captcha.VerifySelf(ctx, selfURL, selfToken) {
			return "自建验证验证失败，请重新验证"
		}
	}
	return ""
}

func (s *Service) sendSMS(ctx context.Context, keyPrefix string, mobi string, sendCount int, limit int, ip string, setting map[string]interface{}) (string, error) {
	key := keyPrefix + "." + mobi + "." + s.now().Format("2006-01-02")
	if count, err := s.limiter.Count(ctx, key, time.Minute); err != nil {
		return "", err
	} else if count > 0 {
		return "发送太频率请稍后重试", nil
	}
	if count, err := s.limiter.Count(ctx, key, 24*time.Hour); err != nil {
		return "", err
	} else if limit > 0 && count >= limit {
		return "索取验证码次数过于频繁，请明天再试", nil
	}
	code := s.randCode()
	platform := atoi(setting["smsplatform"])
	if !strings.HasPrefix(mobi, "86.") {
		platform = atoi(setting["smsplatforminternational"])
	}
	if err := s.sms.SendSMS(ctx, platform, str(setting["smsconfig"]), mobi, code, sendCount); err != nil {
		return "发送失败，请重试" + err.Error(), nil
	}
	if err := s.limiter.Incr(ctx, key, 24*time.Hour, ""); err != nil {
		return "", err
	}
	if err := s.limiter.Incr(ctx, "sms."+mobi+"."+code, 10*time.Minute, mobi+"."+code); err != nil {
		return "", err
	}
	if ip != "" {
		if err := s.limiter.Incr(ctx, "sms.send.ip."+ip, 24*time.Hour, ""); err != nil {
			return "", err
		}
	}
	return "短信已成功发送", nil
}

func (s *Service) sendMail(ctx context.Context, email string, sendCount int, limit int, setting map[string]interface{}) (string, error) {
	key := "email.send." + email + "." + s.now().Format("2006-01-02")
	if count, err := s.limiter.Count(ctx, key, time.Minute); err != nil {
		return "", err
	} else if count > 0 {
		return "发送太频率请稍后重试", nil
	}
	if count, err := s.limiter.Count(ctx, key, 24*time.Hour); err != nil {
		return "", err
	} else if limit > 0 && count >= limit {
		return "索取验证码次数过于频繁，请明天再试", nil
	}
	conf := map[string]interface{}{}
	if err := json.Unmarshal([]byte(str(setting["mailconf"])), &conf); err != nil || len(conf) == 0 {
		return "邮箱功能暂未开启，请稍后重试", nil
	}
	code := s.randCode()
	if err := s.mail.SendMail(ctx, conf, email, "您的邮箱信息", "验证码为："+code+"，10分钟内有效，感谢您的使用！"); err != nil {
		return "发送失败，请重试", nil
	}
	_ = sendCount
	if err := s.limiter.Incr(ctx, key, 24*time.Hour, ""); err != nil {
		return "", err
	}
	return "验证码已发送至您的邮箱，请10分钟内验证并确认", s.limiter.Incr(ctx, "email."+email+"."+code, 10*time.Minute, email+"."+code)
}

func (s *Service) setting(ctx context.Context) (map[string]interface{}, error) {
	row, err := s.store.SettingByUUID(ctx, "setting")
	if err != nil {
		return nil, err
	}
	return parsePHPSerializedMap(str(row["value"])), nil
}

func normalizeMobi(prefix string, mobi string) (string, bool) {
	prefix = strings.TrimSpace(prefix)
	mobi = strings.TrimSpace(mobi)
	if prefix == "" {
		prefix = "+86"
	}
	full := prefix + mobi
	if ok, _ := regexp.MatchString(`^\+?86[1-9][0-9]{10}$`, full); !ok {
		return "", false
	}
	return strings.Trim(strings.TrimSpace(prefix+"."+mobi), "+"), true
}

func validEmail(value string) bool {
	_, err := mail.ParseAddress(value)
	return err == nil && strings.Contains(value, "@")
}

func androidVersion(ua string) int {
	if !strings.Contains(strings.ToLower(ua), "android") {
		return 0
	}
	digits := regexp.MustCompile(`\D`).ReplaceAllString(ua, "")
	return atoi(digits)
}

type RejectCaptchaVerifier struct{}

func (RejectCaptchaVerifier) VerifyImage(context.Context, string, string) bool { return false }
func (RejectCaptchaVerifier) VerifyGoogle(context.Context, string) bool        { return false }
func (RejectCaptchaVerifier) VerifyTencent(context.Context, string, string, string) bool {
	return false
}
func (RejectCaptchaVerifier) VerifySelf(context.Context, string, string) bool { return false }

type DisabledSMSSender struct{}

func (DisabledSMSSender) SendSMS(context.Context, int, string, string, string, int) error {
	return errors.New("")
}

type DisabledMailSender struct{}

func (DisabledMailSender) SendMail(context.Context, map[string]interface{}, string, string, string) error {
	return errors.New("disabled")
}

type MemoryLimiter struct {
	mu   sync.Mutex
	rows map[string][]limitRow
}

type limitRow struct {
	createdAt time.Time
	value     string
}

func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{rows: map[string][]limitRow{}}
}

func (l *MemoryLimiter) Count(_ context.Context, key string, window time.Duration) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	rows := l.rows[key]
	kept := rows[:0]
	for _, row := range rows {
		if window <= 0 || now.Sub(row.createdAt) <= window {
			kept = append(kept, row)
		}
	}
	l.rows[key] = kept
	return len(kept), nil
}

func (l *MemoryLimiter) Incr(_ context.Context, key string, _ time.Duration, value string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rows[key] = append(l.rows[key], limitRow{createdAt: time.Now(), value: value})
	return nil
}

func parsePHPSerializedMap(value string) map[string]interface{} {
	out := map[string]interface{}{}
	for i := 0; i < len(value); {
		key, next, ok := readPHPString(value, i)
		if !ok {
			i++
			continue
		}
		i = next
		if strings.HasPrefix(value[i:], "s:") {
			val, next, ok := readPHPString(value, i)
			if !ok {
				i++
				continue
			}
			out[key] = val
			i = next
			continue
		}
		if strings.HasPrefix(value[i:], "i:") || strings.HasPrefix(value[i:], "d:") {
			end := strings.Index(value[i:], ";")
			if end < 0 {
				break
			}
			out[key] = atoi(value[i+2 : i+end])
			i += end + 1
			continue
		}
		if strings.HasPrefix(value[i:], "N;") {
			out[key] = nil
			i += 2
			continue
		}
	}
	return out
}

func readPHPString(value string, start int) (string, int, bool) {
	idx := strings.Index(value[start:], "s:")
	if idx < 0 {
		return "", start, false
	}
	i := start + idx + 2
	colon := strings.Index(value[i:], ":")
	if colon < 0 {
		return "", start, false
	}
	n := atoi(value[i : i+colon])
	i += colon + 1
	if i >= len(value) || value[i] != '"' {
		return "", start, false
	}
	i++
	if i+n > len(value) {
		return "", start, false
	}
	out := value[i : i+n]
	i += n
	if i+2 <= len(value) && value[i:i+2] == "\";" {
		i += 2
	}
	return out, i, true
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
