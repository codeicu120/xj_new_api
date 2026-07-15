package invite

import (
	"context"
	"fmt"
	"strconv"

	userRepo "xj_comp/internal/repository/user"
)

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type Store interface {
	RecordRecommend(ctx context.Context, uid int) (map[string]interface{}, error)
}

type Service struct {
	auth  AuthStore
	store Store
}

func NewService(auth AuthStore, store Store) *Service {
	return &Service{auth: auth, store: store}
}

func (s *Service) Info(ctx context.Context, token string) (map[string]interface{}, int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return nil, -1, "获取邀请码失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	row, err := s.store.RecordRecommend(ctx, uid)
	if err != nil {
		return nil, -1, "获取邀请码失败", err
	}
	if len(row) == 0 {
		return map[string]interface{}{"data": nil}, 0, "", nil
	}
	key := ""
	if uniqkey := atoi(row["uniqkey"]); uniqkey > 0 {
		key = strconv.FormatInt(int64(uniqkey), 36)
	}
	return map[string]interface{}{"data": key}, 0, "", nil
}

func (s *Service) BindEdge(ctx context.Context, token string, inviteCode string) (int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -1, "绑定邀请码失败", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if inviteCode == "" {
		return -1, "请输入邀请码", nil
	}
	return -1, "邀请码绑定成功分支暂未迁移", nil
}

func (s *Service) userByToken(ctx context.Context, token string) (map[string]interface{}, error) {
	sid := userRepo.CleanToken(token)
	if sid == "" || s.auth == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	user, err := s.auth.UserBySession(ctx, sid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	return user, nil
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
