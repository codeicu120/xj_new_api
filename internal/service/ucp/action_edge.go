package ucp

import (
	"context"
	"strings"
)

func (s *Service) HighRiskActionEdge(ctx context.Context, token string, pendingMessage string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if pendingMessage == "" {
		pendingMessage = "成功分支暂未迁移"
	}
	return -1, pendingMessage, nil
}

func (s *Service) UserEmailEdge(ctx context.Context, token string, email string, pendingMessage string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if !validEmail(email) {
		return -1, "请输入正确的邮箱地址", nil
	}
	if pendingMessage == "" {
		pendingMessage = "邮箱成功分支暂未迁移"
	}
	return -1, pendingMessage, nil
}

func (s *Service) UserVerifyEmailEdge(ctx context.Context, token string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	return -1, "验证码不存在或已失效", nil
}

func (s *Service) UserBindMobiEdge(ctx context.Context, token string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if mobi := str(user["mobi"]); mobi != "" && !strings.HasPrefix(mobi, "~") {
		return -1, "您已绑定手机", nil
	}
	return -1, "手机验证码不正确", nil
}

func validEmail(email string) bool {
	email = strings.TrimSpace(email)
	at := strings.Index(email, "@")
	dot := strings.LastIndex(email, ".")
	return at > 0 && dot > at+1 && dot < len(email)-1
}
