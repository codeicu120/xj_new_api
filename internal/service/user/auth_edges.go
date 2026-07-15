package user

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	userRepo "xj_comp/internal/repository/user"
)

type AuthEdgeStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type AuthEdgeLookupStore interface {
	UserByMobi(ctx context.Context, mobi string) (map[string]interface{}, error)
	UserByEmail(ctx context.Context, email string) (map[string]interface{}, error)
	UserByUsername(ctx context.Context, username string) (map[string]interface{}, error)
}

type AuthEdgeService struct {
	store AuthEdgeStore
}

var usernamePattern = regexp.MustCompile(`^[\p{Han}a-z0-9_]+$`)

type AuthEdgeRequest struct {
	Token      string
	AUP        int
	Step       string
	Mobi       string
	Email      string
	Username   string
	Password   string
	MobiPrefix string
	RegType    int
	LoginType  int
}

func NewAuthEdgeService(store AuthEdgeStore) *AuthEdgeService {
	return &AuthEdgeService{store: store}
}

func (s *AuthEdgeService) Register(ctx context.Context, req AuthEdgeRequest, v2 bool) (int, string, error) {
	user, err := s.userByToken(ctx, req.Token)
	if err != nil {
		return -1, "注册失败", err
	}
	if atoi(user["uid"]) > 0 {
		return -1, "用户已登录", nil
	}
	if req.AUP != 1 {
		return -1, "请同意用户协议", nil
	}
	if !v2 {
		if !validMainlandMobile(req.MobiPrefix, req.Mobi) {
			return -1, "手机号码填写不正确", nil
		}
		if errmsg, err := s.checkMobiRegistration(ctx, normalizedMobi(req.MobiPrefix, req.Mobi)); err != nil {
			return -1, "注册失败", err
		} else if errmsg != "" {
			return -1, errmsg, nil
		}
	}
	if v2 && req.RegType == 1 {
		if errmsg, err := s.checkUsernameRegistration(ctx, req.Username); err != nil {
			return -1, "注册失败", err
		} else if errmsg != "" {
			return -1, errmsg, nil
		}
	}
	if v2 && req.RegType == 2 && !validMainlandMobile(req.MobiPrefix, req.Mobi) {
		return -1, "手机号码填写不正确", nil
	}
	if v2 && req.RegType == 2 {
		if errmsg, err := s.checkMobiRegistration(ctx, normalizedMobi(req.MobiPrefix, req.Mobi)); err != nil {
			return -1, "注册失败", err
		} else if errmsg != "" {
			return -1, errmsg, nil
		}
	}
	if v2 && req.RegType == 3 && !validEmail(req.Email) {
		return -1, "请输入正确邮箱地址", nil
	}
	if v2 && req.RegType == 3 {
		if errmsg, err := s.checkEmailRegistration(ctx, req.Email); err != nil {
			return -1, "注册失败", err
		} else if errmsg != "" {
			return -1, errmsg, nil
		}
	}
	if v2 && req.RegType == 1 && !validPassword(req.Password) {
		return -1, "密码6-16位", nil
	}
	return -1, "注册成功分支暂未迁移", nil
}

func (s *AuthEdgeService) Login(ctx context.Context, req AuthEdgeRequest, v2 bool) (int, string, error) {
	user, err := s.userByToken(ctx, req.Token)
	if err != nil {
		return -1, "登录失败", err
	}
	if atoi(user["uid"]) > 0 {
		return -1, "用户已登录", nil
	}
	if v2 && strings.TrimSpace(req.Mobi) == "" && strings.TrimSpace(req.Email) == "" && strings.TrimSpace(req.Username) == "" {
		return -1, "用户名未注册", nil
	}
	if v2 {
		row, err := s.lookupLoginUser(ctx, req)
		if err != nil {
			return -1, "登录失败", err
		}
		if len(row) == 0 {
			switch {
			case strings.TrimSpace(req.Mobi) != "":
				return -1, "手机号码未注册", nil
			case strings.TrimSpace(req.Email) != "":
				return -1, "邮箱未注册", nil
			default:
				return -1, "用户名未注册", nil
			}
		}
		if req.LoginType != 1 && strings.TrimSpace(req.Password) == "" {
			return -1, "密码不能为空", nil
		}
	}
	return -1, "登录成功分支暂未迁移", nil
}

func (s *AuthEdgeService) Forgot(ctx context.Context, req AuthEdgeRequest, v2 bool) (int, string, error) {
	if v2 && strings.TrimSpace(req.Mobi) == "" && strings.TrimSpace(req.Email) == "" {
		return -1, "请填写手机号码或者邮箱", nil
	}
	if v2 && strings.TrimSpace(req.Email) != "" && strings.TrimSpace(req.Mobi) == "" && !validEmail(req.Email) {
		return -1, "请输入正确邮箱地址", nil
	}
	if !v2 && !validMainlandMobile(req.MobiPrefix, req.Mobi) || v2 && strings.TrimSpace(req.Mobi) != "" && !validMainlandMobile(req.MobiPrefix, req.Mobi) {
		return -1, "手机号码填写不正确", nil
	}
	if strings.TrimSpace(req.Step) == "" {
		return -1, "无效的操作", nil
	}
	if req.Step == "step1" {
		row, err := s.lookupForgotUser(ctx, req, v2)
		if err != nil {
			return -1, "密码重置失败", err
		}
		if len(row) == 0 {
			if v2 && strings.TrimSpace(req.Email) != "" && strings.TrimSpace(req.Mobi) == "" {
				return -1, "输入的邮箱不存在", nil
			}
			return -1, "输入的手机号码不存在", nil
		}
		return 0, "step1->step2", nil
	}
	return -1, "密码重置成功分支暂未迁移", nil
}

func (s *AuthEdgeService) Delete(ctx context.Context, token string) (int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	return -1, "账号注销成功分支暂未迁移", nil
}

func (s *AuthEdgeService) ChangePhone(ctx context.Context, req AuthEdgeRequest) (int, string, error) {
	user, err := s.userByToken(ctx, req.Token)
	if err != nil {
		return -9999, "请登录后操作", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "请登录后操作", nil
	}
	if !validMainlandMobile(req.MobiPrefix, req.Mobi) {
		return -1, "手机号码填写不正确", nil
	}
	if req.Step != "step1" && req.Step != "step2" {
		return -1, "步骤错误", nil
	}
	mobi := normalizedMobi(req.MobiPrefix, req.Mobi)
	if strings.TrimSpace(fmt.Sprint(user["mobi"])) == mobi {
		return -1, "更换的手机号和当前手机号相同！", nil
	}
	row, err := s.lookupMobi(ctx, mobi)
	if err != nil {
		return -1, "手机号更换失败", err
	}
	if len(row) > 0 {
		return -1, "手机号已经存在", nil
	}
	if req.Step == "step1" {
		return 0, "step1->step2", nil
	}
	return -1, "手机号更换成功分支暂未迁移", nil
}

func (s *AuthEdgeService) lookupLoginUser(ctx context.Context, req AuthEdgeRequest) (map[string]interface{}, error) {
	switch {
	case strings.TrimSpace(req.Mobi) != "":
		return s.lookupMobi(ctx, normalizedMobi(req.MobiPrefix, req.Mobi))
	case strings.TrimSpace(req.Email) != "":
		return s.lookupEmail(ctx, strings.TrimSpace(req.Email))
	default:
		return s.lookupUsername(ctx, strings.TrimSpace(req.Username))
	}
}

func (s *AuthEdgeService) checkMobiRegistration(ctx context.Context, mobi string) (string, error) {
	if strings.TrimSpace(mobi) == "" {
		return "手机号码不能为空", nil
	}
	row, err := s.lookupMobi(ctx, mobi)
	if err != nil {
		return "", err
	}
	if len(row) > 0 {
		return "手机号码已被注册", nil
	}
	return "", nil
}

func (s *AuthEdgeService) checkEmailRegistration(ctx context.Context, email string) (string, error) {
	if !validEmail(email) {
		return "请输入正确邮箱地址", nil
	}
	row, err := s.lookupEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		return "", err
	}
	if len(row) > 0 {
		return "该邮箱已经被注册，您可以通过邮箱找回密码", nil
	}
	return "", nil
}

func (s *AuthEdgeService) checkUsernameRegistration(ctx context.Context, username string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(username))
	switch {
	case value == "":
		return "请填写用户名", nil
	case isDigits(value):
		return "用户名不能是纯数字", nil
	case len(value) < 6 || len(value) > 24 || utf8.RuneCountInString(value) > 16:
		return "用户名2-8个汉字，英文6-16个字符", nil
	case !usernamePattern.MatchString(value):
		return "用户名只允许中英文、数字及下划线组成", nil
	}
	row, err := s.lookupUsername(ctx, value)
	if err != nil {
		return "", err
	}
	if len(row) > 0 {
		return "用户名已存在", nil
	}
	return "", nil
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func (s *AuthEdgeService) lookupForgotUser(ctx context.Context, req AuthEdgeRequest, v2 bool) (map[string]interface{}, error) {
	if v2 && strings.TrimSpace(req.Email) != "" && strings.TrimSpace(req.Mobi) == "" {
		return s.lookupEmail(ctx, strings.TrimSpace(req.Email))
	}
	return s.lookupMobi(ctx, normalizedMobi(req.MobiPrefix, req.Mobi))
}

func (s *AuthEdgeService) lookupMobi(ctx context.Context, mobi string) (map[string]interface{}, error) {
	lookup, ok := s.store.(AuthEdgeLookupStore)
	if !ok || lookup == nil {
		return map[string]interface{}{}, nil
	}
	row, err := lookup.UserByMobi(ctx, mobi)
	if row == nil {
		row = map[string]interface{}{}
	}
	return row, err
}

func (s *AuthEdgeService) lookupEmail(ctx context.Context, email string) (map[string]interface{}, error) {
	lookup, ok := s.store.(AuthEdgeLookupStore)
	if !ok || lookup == nil {
		return map[string]interface{}{}, nil
	}
	row, err := lookup.UserByEmail(ctx, email)
	if row == nil {
		row = map[string]interface{}{}
	}
	return row, err
}

func (s *AuthEdgeService) lookupUsername(ctx context.Context, username string) (map[string]interface{}, error) {
	lookup, ok := s.store.(AuthEdgeLookupStore)
	if !ok || lookup == nil {
		return map[string]interface{}{}, nil
	}
	row, err := lookup.UserByUsername(ctx, username)
	if row == nil {
		row = map[string]interface{}{}
	}
	return row, err
}

func (s *AuthEdgeService) userByToken(ctx context.Context, token string) (map[string]interface{}, error) {
	if s.store == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	sid := userRepo.CleanToken(token)
	if sid == "" {
		return map[string]interface{}{"uid": "0"}, nil
	}
	user, err := s.store.UserBySession(ctx, sid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	return user, nil
}

func validMainlandMobile(prefix string, mobi string) bool {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "+86"
	}
	raw := strings.TrimSpace(mobi)
	if prefix != "+86" && prefix != "86" {
		return raw != ""
	}
	if len(raw) != 11 || raw[0] != '1' {
		return false
	}
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func validEmail(email string) bool {
	email = strings.TrimSpace(email)
	at := strings.Index(email, "@")
	dot := strings.LastIndex(email, ".")
	return at > 0 && dot > at+1 && dot < len(email)-1
}

func validPassword(password string) bool {
	n := len(password)
	return n >= 6 && n <= 16
}

func normalizedMobi(prefix string, mobi string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "+86"
	}
	return strings.Trim(strings.TrimSpace(prefix)+"."+strings.TrimSpace(mobi), "+")
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
