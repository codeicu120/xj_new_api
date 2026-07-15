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

func (s *Service) UserProfileEdge(ctx context.Context, token string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	return -1, "资料设置成功分支暂未迁移", nil
}

func (s *Service) UserPasswdEdge(ctx context.Context, token string, password string, passwordConfirm string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if len(password) < 6 || len(password) > 16 {
		return -1, "密码6-16位", nil
	}
	if password != passwordConfirm {
		return -1, "两次输入密码不一致", nil
	}
	return -1, "密码修改成功分支暂未迁移", nil
}

func (s *Service) CoinLogExchangeEdge(ctx context.Context, token string, extype int, exnum int) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if extype != 1 && extype != 2 {
		return -1, "请指定兑换类型", nil
	}
	if exnum == 0 {
		return -1, "请指定兑换数量", nil
	}
	if exnum > 1000000 {
		return -1, "兑换数量100万以上请分次兑换", nil
	}
	return -1, "金币兑换成功分支暂未迁移", nil
}

func (s *Service) VODOrderCreateEdge(ctx context.Context, token string, vodserial string, vodname string, coins int) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if strings.TrimSpace(vodserial) == "" && strings.TrimSpace(vodname) == "" {
		return -1, "请填写视频番号或者视频名称", nil
	}
	if coins < 100 {
		return -1, "求片金币不能低于100", nil
	}
	return -1, "求片创建成功分支暂未迁移", nil
}

func (s *Service) VODOrderSupportEdge(ctx context.Context, token string, orderID int) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if orderID <= 0 {
		return -1, "您助力的求片记录不存在", nil
	}
	return -1, "求片助力成功分支暂未迁移", nil
}

func validEmail(email string) bool {
	email = strings.TrimSpace(email)
	at := strings.Index(email, "@")
	dot := strings.LastIndex(email, ".")
	return at > 0 && dot > at+1 && dot < len(email)-1
}
