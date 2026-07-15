package user

import (
	"context"

	userRepo "xj_comp/internal/repository/user"
)

type LogoutStore interface {
	Logout(ctx context.Context, sid string) error
}

type LogoutService struct {
	store LogoutStore
}

func NewLogoutService(store LogoutStore) *LogoutService {
	return &LogoutService{store: store}
}

func (s *LogoutService) Logout(ctx context.Context, token string) error {
	if s.store == nil {
		return nil
	}
	return s.store.Logout(ctx, userRepo.CleanToken(token))
}
