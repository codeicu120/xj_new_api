package server

import (
	"context"

	indexRepo "xj_comp/internal/repository/index"
	ucpRepo "xj_comp/internal/repository/ucp"
	userRepo "xj_comp/internal/repository/user"
)

type indexStore struct {
	user  *userRepo.Repository
	ucp   *ucpRepo.Repository
	index *indexRepo.SettingsRepository
}

func (s indexStore) UserBySession(ctx context.Context, sid string) (map[string]interface{}, error) {
	return s.user.UserBySession(ctx, sid)
}

func (s indexStore) Groups(ctx context.Context) ([]map[string]interface{}, error) {
	return s.user.Groups(ctx)
}

func (s indexStore) Quota(ctx context.Context, uid int) (map[string]interface{}, error) {
	return s.ucp.Quota(ctx, uid)
}

func (s indexStore) Goldbean(ctx context.Context, uid int) (map[string]interface{}, error) {
	return s.ucp.Goldbean(ctx, uid)
}

func (s indexStore) SettingByUUID(ctx context.Context, uuid string) (map[string]interface{}, error) {
	return s.index.SettingByUUID(ctx, uuid)
}

func (s indexStore) CalldataByUUID(ctx context.Context, uuid string) (map[string]interface{}, error) {
	return s.index.CalldataByUUID(ctx, uuid)
}
