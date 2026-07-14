package sendfile

import (
	"context"
	"errors"
	"fmt"

	userRepo "xj_comp/internal/repository/user"
)

var ErrVODNotFound = errors.New("vod not found")

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type VODStore interface {
	VODByID(ctx context.Context, vodID int) (map[string]interface{}, error)
}

type Service struct {
	authStore AuthStore
	vodStore  VODStore
}

func NewService(authStore AuthStore, vodStore VODStore) *Service {
	return &Service{authStore: authStore, vodStore: vodStore}
}

func (s *Service) Play(ctx context.Context, token string, vodID int) (int, string, error) {
	user, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -1, "获取用户失败", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "请登录后操作", nil
	}

	row, err := s.vodStore.VODByID(ctx, vodID)
	if err != nil {
		return -1, "获取视频失败", err
	}
	if len(row) == 0 || atoi(row["showtype"]) > 0 {
		return -1, "记录不存在或已被删除", ErrVODNotFound
	}
	return 0, "", nil
}

func (s *Service) authenticatedUser(ctx context.Context, token string) (map[string]interface{}, error) {
	sid := userRepo.CleanToken(token)
	user, err := s.authStore.UserBySession(ctx, sid)
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
