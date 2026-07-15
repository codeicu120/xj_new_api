package favorite

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"xj_comp/internal/domain"
	favoriteRepo "xj_comp/internal/repository/favorite"
	userRepo "xj_comp/internal/repository/user"
	vodService "xj_comp/internal/service/vod"
)

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type Store interface {
	Items(ctx context.Context, kind favoriteRepo.Kind, uid int, page int, pageSize int, keyword string) (int, []map[string]interface{}, error)
	Remove(ctx context.Context, kind favoriteRepo.Kind, uid int, vodid int) (int, error)
	VODByID(ctx context.Context, vodid int) (map[string]interface{}, error)
	Count(ctx context.Context, kind favoriteRepo.Kind, uid int, vodid int, since int64) (int, error)
	Add(ctx context.Context, kind favoriteRepo.Kind, uid int, vodid int, now int64) error
	UsersByIDs(ctx context.Context, ids []int) ([]map[string]interface{}, error)
}

type VODProcessor interface {
	ProcessRows(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error)
	ProcessRowsPlain(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error)
	ProcessMiniRows(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error)
	ProcessMiniRowsFullPrice(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error)
}

type Service struct {
	auth         AuthStore
	store        Store
	vodProcessor VODProcessor
	now          func() time.Time
	resourceBase string
}

func NewService(auth AuthStore, store Store, vodProcessor VODProcessor) *Service {
	return &Service{auth: auth, store: store, vodProcessor: vodProcessor, now: time.Now}
}

func (s *Service) WithResourceBaseURL(base string) *Service {
	s.resourceBase = strings.TrimRight(base, "/")
	return s
}

func (s *Service) Listing(ctx context.Context, token string, kind favoriteRepo.Kind, page int, keyword string, isH5Request bool) (domain.HistoryListingData, int, string, error) {
	return s.listing(ctx, token, kind, page, keyword, isH5Request, false)
}

func (s *Service) MiniV2Listing(ctx context.Context, token string, page int, keyword string, isH5Request bool) (domain.HistoryListingData, int, string, error) {
	data, retcode, errmsg, err := s.listing(ctx, token, favoriteRepo.KindMini, page, keyword, isH5Request, true)
	if err != nil || retcode != 0 {
		return data, retcode, errmsg, err
	}
	rows, err := s.wrapMiniRowsWithUsers(ctx, data.Rows)
	if err != nil {
		return domain.HistoryListingData{}, -1, "获取收藏失败", err
	}
	data.Rows = rows
	return data, 0, "", nil
}

func (s *Service) listing(ctx context.Context, token string, kind favoriteRepo.Kind, page int, keyword string, isH5Request bool, miniKeyword bool) (domain.HistoryListingData, int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return domain.HistoryListingData{}, -1, "获取收藏失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.HistoryListingData{}, -9999, "请登录后操作", nil
	}
	const pageSize = 20
	effectiveKeyword := keyword
	if kind == favoriteRepo.KindMini && !miniKeyword {
		effectiveKeyword = ""
	}
	total, rows, err := s.store.Items(ctx, kind, uid, page, pageSize, effectiveKeyword)
	if err != nil {
		return domain.HistoryListingData{}, -1, "获取收藏失败", err
	}
	if s.vodProcessor != nil {
		if kind == favoriteRepo.KindMini {
			rows, err = s.vodProcessor.ProcessMiniRowsFullPrice(ctx, rows, isH5Request)
			for _, row := range rows {
				row["isfavorite"] = 1
			}
		} else {
			rows, err = s.vodProcessor.ProcessRowsPlain(ctx, rows, isH5Request)
		}
		if err != nil {
			return domain.HistoryListingData{}, -1, "获取收藏失败", err
		}
	}
	baseURL := "/favorite/listing?page=[?]"
	if kind == favoriteRepo.KindMini {
		baseURL = "/minifavorite/listing?page=[?]"
		if miniKeyword && keyword != "" {
			baseURL += "&wd=" + keyword
		}
	} else if keyword != "" {
		baseURL = "/favorite/listing?page=[?]&wd=" + keyword
	}
	return domain.HistoryListingData{
		Rows:     rows,
		PageInfo: vodService.PageInfo(total, pageSize, page, baseURL),
	}, 0, "", nil
}

func (s *Service) wrapMiniRowsWithUsers(ctx context.Context, rows []map[string]interface{}) ([]map[string]interface{}, error) {
	users, err := s.store.UsersByIDs(ctx, rowIDs(rows, "authorid"))
	if err != nil {
		return nil, err
	}
	userByID := map[string]map[string]interface{}{}
	for _, user := range users {
		userByID[str(user["uid"])] = processUser(user, s.resourceBase)
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		var user interface{}
		if found, ok := userByID[str(row["authorid"])]; ok {
			user = found
		}
		out = append(out, map[string]interface{}{"vodrow": row, "user": user})
	}
	return out, nil
}

func (s *Service) Remove(ctx context.Context, token string, kind favoriteRepo.Kind, vodids []int) (int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -1, "删除收藏失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "请登录后操作", nil
	}
	rowCount := 0
	for _, vodid := range vodids {
		count, err := s.store.Remove(ctx, kind, uid, vodid)
		if err != nil {
			return -1, "删除收藏失败", err
		}
		rowCount += count
	}
	return 0, fmt.Sprintf("已删除%d项", rowCount), nil
}

func (s *Service) Add(ctx context.Context, token string, kind favoriteRepo.Kind, vodid int) (map[string]interface{}, int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return nil, -1, "添加收藏失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "请登录后操作", nil
	}
	row, err := s.store.VODByID(ctx, vodid)
	if err != nil {
		return nil, -1, "添加收藏失败", err
	}
	showtype := atoi(row["showtype"])
	if len(row) == 0 || (kind == favoriteRepo.KindVOD && showtype > 0) || (kind == favoriteRepo.KindMini && showtype != 1) {
		return nil, -1, "记录不存在或已被删除", nil
	}
	count, err := s.store.Count(ctx, kind, uid, vodid, 0)
	if err != nil {
		return nil, -1, "添加收藏失败", err
	}
	if count > 0 {
		return nil, -1, "您已经收藏过了", nil
	}
	if err := s.store.Add(ctx, kind, uid, vodid, s.now().Unix()); err != nil {
		return nil, -1, "添加收藏失败", err
	}
	return map[string]interface{}{}, 0, "已收藏", nil
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

func rowIDs(rows []map[string]interface{}, key string) []int {
	seen := map[int]bool{}
	ids := []int{}
	for _, row := range rows {
		id := atoi(row[key])
		if id > 0 && !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	return ids
}

func processUser(row map[string]interface{}, base string) map[string]interface{} {
	avatar := str(row["avatar"])
	avatarURL := ""
	if avatar != "" {
		if strings.HasPrefix(avatar, "http://") || strings.HasPrefix(avatar, "https://") {
			avatarURL = avatar
		} else if _, err := strconv.Atoi(avatar); err == nil {
			avatarURL = avatar
		} else if strings.HasPrefix(avatar, "avatar/") {
			avatarURL = strings.TrimRight(base, "/") + "/C1/" + strings.TrimLeft(avatar, "/")
		} else {
			avatarURL = strings.TrimRight(base, "/") + "/C1/avatar/" + strings.TrimLeft(avatar, "/")
		}
	}
	return map[string]interface{}{
		"uid":        str(row["uid"]),
		"username":   str(row["username"]),
		"nickname":   str(row["nickname"]),
		"avatar":     avatar,
		"avatar_url": avatarURL,
		"gender":     str(row["gender"]),
	}
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
