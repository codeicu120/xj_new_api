package attach

import (
	"context"
	"fmt"
	"regexp"

	userRepo "xj_comp/internal/repository/user"
)

type Store interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
	UpdateAvatar(ctx context.Context, uid int, avatarID string) error
}

type Service struct {
	store Store
}

var avatarIDPattern = regexp.MustCompile(`^[0-9]+$`)

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) UpAvatar(ctx context.Context, token string, avatarID string) (int, string, error) {
	user, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -1, "获取用户失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "您还没有登录", nil
	}
	if !avatarIDPattern.MatchString(avatarID) {
		return -1, "请选择系统头像", nil
	}
	if err := s.store.UpdateAvatar(ctx, uid, avatarID); err != nil {
		return -1, "更新头像失败", err
	}
	return 0, "", nil
}

func (s *Service) authenticatedUser(ctx context.Context, token string) (map[string]interface{}, error) {
	sid := userRepo.CleanToken(token)
	if s.store == nil {
		return map[string]interface{}{"uid": "0", "sid": sid}, nil
	}
	user, err := s.store.UserBySession(ctx, sid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return map[string]interface{}{"uid": "0", "sid": sid}, nil
	}
	return user, nil
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
