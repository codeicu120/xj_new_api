package user

import (
	"context"
	"fmt"
	"strings"

	userRepo "xj_comp/internal/repository/user"
)

type AuthEdgeStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type AuthEdgeService struct {
	store AuthEdgeStore
}

type AuthEdgeRequest struct {
	Token      string
	AUP        int
	Step       string
	Mobi       string
	Email      string
	Username   string
	MobiPrefix string
}

func NewAuthEdgeService(store AuthEdgeStore) *AuthEdgeService {
	return &AuthEdgeService{store: store}
}

func (s *AuthEdgeService) Register(ctx context.Context, req AuthEdgeRequest) (int, string, error) {
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
	return -1, "登录成功分支暂未迁移", nil
}

func (s *AuthEdgeService) Forgot(req AuthEdgeRequest, v2 bool) (int, string) {
	if v2 && strings.TrimSpace(req.Mobi) == "" && strings.TrimSpace(req.Email) == "" {
		return -1, "请填写手机号码或者邮箱"
	}
	if !v2 && !validMainlandMobile(req.MobiPrefix, req.Mobi) {
		return -1, "手机号码填写不正确"
	}
	if strings.TrimSpace(req.Step) == "" {
		return -1, "无效的操作"
	}
	return -1, "密码重置成功分支暂未迁移"
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
	return -1, "手机号更换成功分支暂未迁移", nil
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

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
