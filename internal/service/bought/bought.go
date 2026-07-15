package bought

import (
	"context"
	"fmt"

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
	vodService "xj_comp/internal/service/vod"
)

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type Store interface {
	Items(ctx context.Context, uid int, page int, pageSize int) (int, []map[string]interface{}, error)
	Delete(ctx context.Context, uid int, vodid int) error
}

type VODProcessor interface {
	ProcessRows(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error)
}

type Service struct {
	auth         AuthStore
	store        Store
	vodProcessor VODProcessor
}

func NewService(auth AuthStore, store Store, vodProcessor ...VODProcessor) *Service {
	service := &Service{auth: auth, store: store}
	if len(vodProcessor) > 0 {
		service.vodProcessor = vodProcessor[0]
	}
	return service
}

func (s *Service) Listing(ctx context.Context, token string, page int, isH5Request bool) (domain.BoughtListingData, int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return domain.BoughtListingData{}, -1, "获取已购影片失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.BoughtListingData{}, -9999, "请登录后操作", nil
	}
	const pageSize = 20
	total, rows, err := s.store.Items(ctx, uid, page, pageSize)
	if err != nil {
		return domain.BoughtListingData{}, -1, "获取已购影片失败", err
	}
	if s.vodProcessor != nil {
		rows, err = s.vodProcessor.ProcessRows(ctx, rows, isH5Request)
		if err != nil {
			return domain.BoughtListingData{}, -1, "获取已购影片失败", err
		}
	}
	return domain.BoughtListingData{
		Rows:     rows,
		PageInfo: vodService.PageInfo(total, pageSize, page, "/bought/listing?page=[?]"),
	}, 0, "", nil
}

func (s *Service) Delete(ctx context.Context, token string, vodids []int) (int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -1, "删除已购影片失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "请登录后操作", nil
	}
	for _, vodid := range vodids {
		if err := s.store.Delete(ctx, uid, vodid); err != nil {
			return -1, "删除已购影片失败", err
		}
	}
	return 0, "", nil
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
