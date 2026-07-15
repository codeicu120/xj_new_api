package history

import (
	"context"
	"fmt"
	"time"

	"xj_comp/internal/domain"
	historyRepo "xj_comp/internal/repository/history"
	userRepo "xj_comp/internal/repository/user"
	vodService "xj_comp/internal/service/vod"
)

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type Store interface {
	Items(ctx context.Context, kind historyRepo.Kind, uid int, sid string, page int, pageSize int, timeline int, now int64) (int, []map[string]interface{}, error)
	Remove(ctx context.Context, kind historyRepo.Kind, uid int, sid string, vodid int) (int, error)
}

type VODProcessor interface {
	ProcessRows(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error)
}

type Service struct {
	auth         AuthStore
	store        Store
	vodProcessor VODProcessor
	now          func() time.Time
}

func NewService(auth AuthStore, store Store, vodProcessor VODProcessor) *Service {
	return &Service{
		auth:         auth,
		store:        store,
		vodProcessor: vodProcessor,
		now:          time.Now,
	}
}

func (s *Service) Listing(ctx context.Context, token string, kind historyRepo.Kind, page int, timeline int, isH5Request bool) (domain.HistoryListingData, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return domain.HistoryListingData{}, err
	}
	now := s.now().Unix()
	const pageSize = 20
	total, rows, err := s.store.Items(ctx, kind, atoi(user["uid"]), str(user["sid"]), page, pageSize, timeline, now)
	if err != nil {
		return domain.HistoryListingData{}, err
	}
	if s.vodProcessor != nil {
		rows, err = s.vodProcessor.ProcessRows(ctx, rows, isH5Request)
		if err != nil {
			return domain.HistoryListingData{}, err
		}
	}
	timeField := "playtime"
	baseURL := "/playlog/listing"
	if kind == historyRepo.KindDown {
		timeField = "downtime"
		baseURL = "/downlog/listing"
	}
	for _, row := range rows {
		row[timeField] = legacyRelativeTime(atoi64(row[timeField]), now)
	}
	return domain.HistoryListingData{
		Rows:     rows,
		PageInfo: vodService.PageInfo(total, pageSize, page, fmt.Sprintf("%s?timeline=%d&page=[?]", baseURL, timeline)),
	}, nil
}

func (s *Service) Remove(ctx context.Context, token string, kind historyRepo.Kind, vodids []int) (string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return "", err
	}
	rowCount := 0
	for _, vodid := range vodids {
		count, err := s.store.Remove(ctx, kind, atoi(user["uid"]), str(user["sid"]), vodid)
		if err != nil {
			return "", err
		}
		rowCount += count
	}
	return fmt.Sprintf("已删除%d项", rowCount), nil
}

func (s *Service) userByToken(ctx context.Context, token string) (map[string]interface{}, error) {
	sid := userRepo.CleanToken(token)
	if sid == "" || s.auth == nil {
		return map[string]interface{}{"uid": "0", "sid": sid}, nil
	}
	user, err := s.auth.UserBySession(ctx, sid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return map[string]interface{}{"uid": "0", "sid": sid}, nil
	}
	if _, ok := user["sid"]; !ok {
		user["sid"] = sid
	}
	return user, nil
}

func legacyRelativeTime(ts int64, now int64) string {
	if ts <= 0 {
		ts = 0
	}
	if ts > now-86400*30 {
		diff := now - ts
		if diff < 0 {
			diff = 0
		}
		switch {
		case diff >= 86400:
			return fmt.Sprintf("%d天前", diff/86400)
		case diff >= 3600:
			return fmt.Sprintf("%d小时前", diff/3600)
		case diff >= 60:
			return fmt.Sprintf("%d分钟前", diff/60)
		default:
			return fmt.Sprintf("%d秒前", diff)
		}
	}
	return time.Unix(ts, 0).Format("2006-01-02")
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}

func atoi64(value interface{}) int64 {
	var n int64
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
