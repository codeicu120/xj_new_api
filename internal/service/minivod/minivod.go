package minivod

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"xj_comp/internal/domain"
	minivodRepo "xj_comp/internal/repository/minivod"
	userRepo "xj_comp/internal/repository/user"
	vodService "xj_comp/internal/service/vod"
)

const sampleParams = "$cateid:0-$areaid:0-$yearid:0-$tagid:0-$definition:0-$duration:0-$freetype:0-$mosaic:0-$langvoice:0-$orderby:0-$page:1"

var paramKeys = []string{"cateid", "areaid", "yearid", "tagid", "definition", "duration", "freetype", "mosaic", "langvoice", "orderby", "page"}

type Store interface {
	Categories(ctx context.Context) ([]map[string]interface{}, error)
	Areas(ctx context.Context) ([]map[string]interface{}, error)
	Years(ctx context.Context) ([]map[string]interface{}, error)
	Servers(ctx context.Context) ([]map[string]interface{}, error)
	TagsByNames(ctx context.Context, names []string) ([]map[string]interface{}, error)
	Count(ctx context.Context, filter minivodRepo.Filter, now int64) (int, error)
	List(ctx context.Context, filter minivodRepo.Filter, total int, page int, pageSize int, orderBy string, now int64) ([]map[string]interface{}, error)
	CountByAuthor(ctx context.Context, authorID int) (int, error)
	ListByAuthor(ctx context.Context, authorID int, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error)
	Random(ctx context.Context, pageSize int) ([]map[string]interface{}, error)
	VODByID(ctx context.Context, vodID int) (map[string]interface{}, error)
	UserByID(ctx context.Context, uid int) (map[string]interface{}, error)
	SimilarVODsByTagIDs(ctx context.Context, tagIDs []int, excludeID int, pageSize int) ([]map[string]interface{}, error)
	RandomVODsExcept(ctx context.Context, pageSize int, excludeID int, cateID int) ([]map[string]interface{}, error)
	Setting(ctx context.Context, key string) (string, error)
	UsersByIDs(ctx context.Context, ids []int) ([]map[string]interface{}, error)
	VODsByIDs(ctx context.Context, ids []int, orderByField bool) ([]map[string]interface{}, error)
	PendingViewLogs(ctx context.Context, uid int, sid string, limit int) ([]map[string]interface{}, error)
	UpDownByUser(ctx context.Context, uid int, vodID int) (map[string]interface{}, error)
	DeleteUpDown(ctx context.Context, uid int, vodID int) error
	SaveUpDown(ctx context.Context, uid int, vodID int, updown int, now int64) (int, error)
	IncrementVODCounter(ctx context.Context, vodID int, field string, delta int) error
	RecountUpDown(ctx context.Context, vodID int) error
	FavoriteCount(ctx context.Context, uid int, vodID int) (int, error)
	MiniViewLog(ctx context.Context, uid int, sid string, vodID int) (map[string]interface{}, error)
	CountMiniViewLogsSince(ctx context.Context, uid int, sid string, since int64, action int) (int, error)
	ReqTaskCoin(ctx context.Context, uid int, sid string, logid int, now int64) (int, string, error)
}

type VODProcessor interface {
	ProcessRowsFullPrice(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error)
	ProcessMiniRowsFullPrice(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error)
}

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type VoteLimiter interface {
	Seen(ctx context.Context, key string) (bool, error)
	Mark(ctx context.Context, key string) error
	Delete(ctx context.Context, key string) error
}

var (
	ErrVODNotFound    = errors.New("minivod not found")
	ErrAuthorNotFound = errors.New("minivod author not found")
)

type Service struct {
	store           Store
	vodProcessor    VODProcessor
	auth            AuthStore
	limiter         VoteLimiter
	now             func() time.Time
	resourceBaseURL string
}

type ListingRequest struct {
	Action      string
	PathParams  string
	QueryPage   string
	IsH5Request bool
}

func NewService(store Store, vodProcessor VODProcessor, resourceBaseURL string) *Service {
	return &Service{store: store, vodProcessor: vodProcessor, limiter: newMemoryVoteLimiter(), now: time.Now, resourceBaseURL: strings.TrimRight(resourceBaseURL, "/")}
}

func (s *Service) WithAuth(auth AuthStore) *Service {
	s.auth = auth
	return s
}

func (s *Service) Listing(ctx context.Context, req ListingRequest) (domain.MiniVODListingData, error) {
	params := parseParams(req.PathParams)
	if atoi(params["page"]) == 0 {
		params["page"] = req.QueryPage
		if params["page"] == "" {
			params["page"] = "0"
		}
	}
	now := s.now().Unix()
	categories, err := s.store.Categories(ctx)
	if err != nil {
		return domain.MiniVODListingData{}, err
	}
	areas, err := s.store.Areas(ctx)
	if err != nil {
		return domain.MiniVODListingData{}, err
	}
	years, err := s.store.Years(ctx)
	if err != nil {
		return domain.MiniVODListingData{}, err
	}

	filter, orderBy, err := s.filter(ctx, req.Action, params, categories, now)
	if err != nil {
		return domain.MiniVODListingData{}, err
	}
	const pageSize = 16
	total := 0
	rows := []map[string]interface{}{}
	if req.Action == "recommend" {
		rows, err = s.store.Random(ctx, pageSize)
	} else {
		total, err = s.store.Count(ctx, filter, now)
		if err == nil {
			rows, err = s.store.List(ctx, filter, total, atoi(params["page"]), pageSize, orderBy, now)
		}
	}
	if err != nil {
		return domain.MiniVODListingData{}, err
	}
	vodRows := rows
	if s.vodProcessor != nil {
		vodRows, err = s.vodProcessor.ProcessMiniRowsFullPrice(ctx, rows, req.IsH5Request)
		if err != nil {
			return domain.MiniVODListingData{}, err
		}
	}
	richRows, err := s.richRows(ctx, req.Action, params, vodRows)
	if err != nil {
		return domain.MiniVODListingData{}, err
	}
	pageURL := "/minivod/" + req.Action + "-" + buildParams(params, map[string]string{"page": "[?]"})
	return domain.MiniVODListingData{
		Now:          now,
		Action:       req.Action,
		SampleParams: sampleParams,
		Params:       params,
		Rows:         richRows,
		VODRows:      vodRows,
		PageInfo:     vodService.PageInfo(total, pageSize, atoi(params["page"]), pageURL),
		Orders:       optionRows([][2]interface{}{{1, "最多好评"}, {2, "最多播放"}, {3, "最高评分"}}),
		Categories:   categories,
		Areas:        areas,
		Years:        years,
		Definitions:  optionRows([][2]interface{}{{1, "标清"}, {2, "高清"}}),
		Durations:    optionRows([][2]interface{}{{1, "长片"}, {2, "短片"}}),
		FreeTypes:    optionRows([][2]interface{}{{1, "免费"}, {2, "会员"}}),
		Mosaics:      optionRows([][2]interface{}{{1, "有码"}, {2, "无码"}}),
		LangVoices:   optionRows([][2]interface{}{{1, "中文字幕"}, {2, "国语对白"}, {3, "其它"}}),
	}, nil
}

func (s *Service) ReqList(ctx context.Context, token string, isH5Request bool) (map[string]interface{}, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	logs, err := s.store.PendingViewLogs(ctx, atoi(user["uid"]), str(user["sid"]), 10)
	if err != nil {
		return nil, err
	}
	vodIDs := rowIDs(logs, "vodid")
	rows, err := s.store.VODsByIDs(ctx, vodIDs, true)
	if err != nil {
		return nil, err
	}
	tagRows, err := s.store.TagsByNames(ctx, collectTagNames(rows))
	if err != nil {
		return nil, err
	}
	_ = tagRows
	vodRows := rows
	if s.vodProcessor != nil {
		vodRows, err = s.vodProcessor.ProcessMiniRowsFullPrice(ctx, rows, isH5Request)
		if err != nil {
			return nil, err
		}
	}
	if atoi(user["uid"]) > 0 {
		for _, row := range vodRows {
			count, err := s.store.FavoriteCount(ctx, atoi(user["uid"]), atoi(row["vodid"]))
			if err != nil {
				return nil, err
			}
			row["isfavorite"] = boolInt(count > 0)
		}
	} else {
		for _, row := range vodRows {
			row["isfavorite"] = 0
		}
	}
	users, err := s.store.UsersByIDs(ctx, rowIDs(vodRows, "authorid"))
	if err != nil {
		return nil, err
	}
	userByID := map[string]map[string]interface{}{}
	for _, item := range users {
		userByID[str(item["uid"])] = processUser(item, s.resourceBaseURL)
	}
	out := []map[string]interface{}{}
	for _, row := range vodRows {
		var author interface{}
		if found, ok := userByID[str(row["authorid"])]; ok {
			author = found
		}
		out = append(out, map[string]interface{}{"vodrow": row, "user": author})
	}
	return map[string]interface{}{"rows": out}, nil
}

func (s *Service) Show(ctx context.Context, vodID int, isH5Request bool) (domain.MiniVODShowData, error) {
	categories, err := s.store.Categories(ctx)
	if err != nil {
		return domain.MiniVODShowData{}, err
	}
	row, err := s.store.VODByID(ctx, vodID)
	if err != nil {
		return domain.MiniVODShowData{}, err
	}
	if len(row) == 0 || atoi(row["showtype"]) != 1 {
		return domain.MiniVODShowData{}, ErrVODNotFound
	}
	user, err := s.store.UserByID(ctx, atoi(row["authorid"]))
	if err != nil {
		return domain.MiniVODShowData{}, err
	}
	if len(user) == 0 {
		return domain.MiniVODShowData{}, ErrAuthorNotFound
	}
	tagRows, err := s.store.TagsByNames(ctx, collectTagNames([]map[string]interface{}{row}))
	if err != nil {
		return domain.MiniVODShowData{}, err
	}
	similarRows, err := s.store.SimilarVODsByTagIDs(ctx, tagIDs(tagRows), vodID, 11)
	if err != nil {
		return domain.MiniVODShowData{}, err
	}
	filteredSimilar := []map[string]interface{}{}
	for _, similar := range similarRows {
		if atoi(similar["vodid"]) == vodID {
			continue
		}
		filteredSimilar = append(filteredSimilar, similar)
		if len(filteredSimilar) >= 10 {
			break
		}
	}
	if len(filteredSimilar) < 10 {
		fillRows, err := s.store.RandomVODsExcept(ctx, 10-len(filteredSimilar), vodID, 0)
		if err != nil {
			return domain.MiniVODShowData{}, err
		}
		filteredSimilar = append(filteredSimilar, fillRows...)
		if len(filteredSimilar) > 10 {
			filteredSimilar = filteredSimilar[:10]
		}
	}
	likeRows, err := s.store.RandomVODsExcept(ctx, 5, vodID, atoi(row["cateid"]))
	if err != nil {
		return domain.MiniVODShowData{}, err
	}
	vodRows, similarRowsOut, likeRowsOut := []map[string]interface{}{row}, filteredSimilar, likeRows
	if s.vodProcessor != nil {
		vodRows, err = s.vodProcessor.ProcessRowsFullPrice(ctx, vodRows, isH5Request)
		if err != nil {
			return domain.MiniVODShowData{}, err
		}
		similarRowsOut, err = s.vodProcessor.ProcessRowsFullPrice(ctx, filteredSimilar, isH5Request)
		if err != nil {
			return domain.MiniVODShowData{}, err
		}
		likeRowsOut, err = s.vodProcessor.ProcessRowsFullPrice(ctx, likeRows, isH5Request)
		if err != nil {
			return domain.MiniVODShowData{}, err
		}
	}
	vodRow := map[string]interface{}{}
	if len(vodRows) > 0 {
		vodRow = vodRows[0]
	}
	return domain.MiniVODShowData{
		VODRow:      vodRow,
		Categories:  categoryParents(categories, atoi(row["cateid"])),
		SimilarRows: similarRowsOut,
		LikeRows:    likeRowsOut,
		VODUser:     processUser(user, s.resourceBaseURL),
	}, nil
}

func (s *Service) AuthorListing(ctx context.Context, authorID int, page int, isH5Request bool) (domain.MiniAuthorListingData, error) {
	user, err := s.store.UserByID(ctx, authorID)
	if err != nil {
		return domain.MiniAuthorListingData{}, err
	}
	if len(user) == 0 {
		return domain.MiniAuthorListingData{}, ErrAuthorNotFound
	}
	const pageSize = 16
	total, err := s.store.CountByAuthor(ctx, authorID)
	if err != nil {
		return domain.MiniAuthorListingData{}, err
	}
	rows, err := s.store.ListByAuthor(ctx, authorID, total, page, pageSize, "utimestamp DESC")
	if err != nil {
		return domain.MiniAuthorListingData{}, err
	}
	if s.vodProcessor != nil {
		rows, err = s.vodProcessor.ProcessMiniRowsFullPrice(ctx, rows, isH5Request)
		if err != nil {
			return domain.MiniAuthorListingData{}, err
		}
	}
	return domain.MiniAuthorListingData{
		Now:      s.now().Unix(),
		UserRow:  processUserFull(user, s.resourceBaseURL, s.now().Unix()),
		VODRows:  rows,
		PageInfo: vodService.PageInfo(total, pageSize, page, ""),
		Orders:   optionRows([][2]interface{}{{1, "最多好评"}, {2, "最多播放"}, {3, "最高评分"}}),
	}, nil
}

func (s *Service) Vote(ctx context.Context, token string, vodID int, up bool) (int, string, error) {
	row, err := s.store.VODByID(ctx, vodID)
	if err != nil {
		return -1, "小视频操作失败", err
	}
	if len(row) == 0 || atoi(row["showtype"]) != 1 {
		return -1, "记录不存在或已被删除", nil
	}
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -1, "小视频操作失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return s.voteGuest(ctx, str(row["vodid"]), up)
	}
	return s.voteUser(ctx, uid, atoi(row["vodid"]), up)
}

func (s *Service) ReqLong(ctx context.Context, token string, vodID int) (string, int, string, error) {
	row, err := s.store.VODByID(ctx, vodID)
	if err != nil {
		return "", -1, "请求小视频长片地址失败", err
	}
	if len(row) == 0 || atoi(row["showtype"]) > 0 {
		return "", 1, "记录不存在或已被删除", nil
	}
	httpURL := sanitizePlayURL(str(row["play_url"]))
	if httpURL == "" {
		return "", 2, "播放地址不存在", nil
	}
	servers, err := s.store.Servers(ctx)
	if err != nil {
		return "", -1, "请求小视频长片地址失败", err
	}
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return "", -1, "请求小视频长片地址失败", err
	}
	httpURL = signCDNURL(httpURL, row, servers, user, s.now().Unix())
	if !hasURLScheme(httpURL) {
		host := serverHost(servers, atoi(row["play_srvid"]))
		httpURL = strings.TrimRight(host, "/") + "/" + strings.TrimLeft(httpURL, "/")
	}
	return httpURL, 0, "", nil
}

func (s *Service) ReqCoin(ctx context.Context, token string, logid int) (int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -1, "领取小视频任务金币失败", err
	}
	retcode, errmsg, err := s.store.ReqTaskCoin(ctx, atoi(user["uid"]), str(user["sid"]), logid, s.now().Unix())
	if err != nil {
		return -1, "领取小视频任务金币失败", err
	}
	return retcode, errmsg, nil
}

func (s *Service) ThrowCoinEdge(ctx context.Context, token string) (int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -9999, "需登录后方可使用投币功能", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "需登录后方可使用投币功能", nil
	}
	return -1, "小视频投币成功分支暂未迁移", nil
}

func (s *Service) ReqPlay(ctx context.Context, token string, vodID int, playIndex int) (map[string]interface{}, int, string, error) {
	return s.reqMedia(ctx, token, vodID, playIndex, true)
}

func (s *Service) ReqDown(ctx context.Context, token string, vodID int, playIndex int) (map[string]interface{}, int, string, error) {
	return s.reqMedia(ctx, token, vodID, playIndex, false)
}

func (s *Service) reqMedia(ctx context.Context, token string, vodID int, playIndex int, play bool) (map[string]interface{}, int, string, error) {
	row, err := s.store.VODByID(ctx, vodID)
	if err != nil {
		return nil, -1, mediaFailMessage(play), err
	}
	if len(row) == 0 || atoi(row["showtype"]) != 1 {
		return map[string]interface{}{}, 1, "记录不存在或已被删除", nil
	}
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return nil, -1, mediaFailMessage(play), err
	}
	if retcode, errmsg := s.checkMiniPerm(row, user); retcode != 0 {
		return map[string]interface{}{}, retcode, errmsg, nil
	}
	mediaURL, price, serverID := mediaSource(row, playIndex, play)
	if mediaURL == "" {
		if play {
			return map[string]interface{}{}, 2, "播放地址不存在", nil
		}
		return map[string]interface{}{}, 2, "下载地址不存在", nil
	}
	if atoi(user["uid"]) == 0 && str(user["sid"]) == "" {
		if play {
			return map[string]interface{}{}, -9999, "客户端游客请先携带信息", nil
		}
		return map[string]interface{}{}, -9999, "客户端游客请先携带信息", nil
	}
	servers, err := s.store.Servers(ctx)
	if err != nil {
		return nil, -1, mediaFailMessage(play), err
	}
	httpURL := sanitizePlayURL(mediaURL)
	if play {
		httpURL = signCDNURL(httpURL, row, servers, user, s.now().Unix())
	}
	if !hasURLScheme(httpURL) {
		host := strings.TrimRight(serverHost(servers, serverID), "/")
		httpURL = host + "/" + strings.TrimLeft(httpURL, "/")
	}
	data := map[string]interface{}{}
	if play {
		isFavorite, err := s.store.FavoriteCount(ctx, atoi(user["uid"]), vodID)
		if err != nil {
			return nil, -1, mediaFailMessage(play), err
		}
		data["isfavorite"] = boolInt(isFavorite > 0)
		data["iszan"] = 0
		if atoi(user["uid"]) > 0 {
			updown, err := s.store.UpDownByUser(ctx, atoi(user["uid"]), vodID)
			if err != nil {
				return nil, -1, mediaFailMessage(play), err
			}
			data["iszan"] = atoi(updown["updown"])
		}
		data["playtask"] = map[string]interface{}{"playnum": 0, "tasknum": 0, "taskcoin": 0, "logid": 0}
	}
	viewrow, err := s.store.MiniViewLog(ctx, atoi(user["uid"]), str(user["sid"]), vodID)
	if err != nil {
		return nil, -1, mediaFailMessage(play), err
	}
	viewedField := "playtime"
	if !play {
		viewedField = "downtime"
	}
	if atoi(viewrow[viewedField]) > 0 {
		data["httpurl"] = httpURL
		if play {
			data["httpurls"] = []map[string]interface{}{{"hdtype": "默认", "httpurl": httpURL}}
			data["jumpId"] = 0
			data["jumpOffset"] = 0
			return data, 0, "已观看过继续提供", nil
		}
		return data, 0, "本周已下载过继续提供", nil
	}
	now := s.now().Unix()
	if price == 0 || (atoi64(row["free_sdate"]) < now && now < atoi64(row["free_edate"])) {
		data["httpurl"] = httpURL
		if play {
			data["httpurls"] = []map[string]interface{}{{"hdtype": "默认", "httpurl": httpURL}}
			data["jumpId"] = 0
			data["jumpOffset"] = 0
			if price == 0 {
				return data, 0, "免费观看", nil
			}
			return data, 0, "限时免费观看", nil
		}
		if price == 0 {
			return data, 0, "免费观看提供下载", nil
		}
		return data, 0, "限时免费观看提供下载", nil
	}
	daynumKey := "max.minivod.play.daynum"
	action := 1
	if !play {
		daynumKey = "max.minivod.down.daynum"
		action = 2
	}
	daycount, err := s.store.CountMiniViewLogsSince(ctx, atoi(user["uid"]), str(user["sid"]), dayStart(s.now), action)
	if err != nil {
		return nil, -1, mediaFailMessage(play), err
	}
	if daycount < getMiniPermInt(user["perms"], daynumKey) {
		data["httpurl"] = httpURL
		if play {
			data["httpurls"] = []map[string]interface{}{{"hdtype": "默认", "httpurl": httpURL}}
			data["jumpId"] = 0
			data["jumpOffset"] = 0
			if atoi(user["uid"]) > 0 {
				return data, 0, "用户权限范围内免费播放", nil
			}
			return data, 0, "游客权限范围内免费播放", nil
		}
		if atoi(user["uid"]) > 0 {
			return data, 0, "用户权限范围内免费下载", nil
		}
		return data, 0, "游客权限范围内免费下载", nil
	}
	if play {
		data["httpurl_preview"] = httpURL + "?300"
		if atoi(user["uid"]) > 0 {
			return data, 4, "今日观影次数已用完，是否去免费增加次数？", nil
		}
		return data, 3, "今日观看次数已看完，请点击免费注册会员获取更多影片观看次数。", nil
	}
	if atoi(user["uid"]) > 0 {
		return data, 4, "今日下载次数已用完，是否去免费增加次数？", nil
	}
	return data, 3, "今日下载次数已用完，请点击免费注册会员获取更多影片下载次数。", nil
}

func (s *Service) voteGuest(ctx context.Context, vodID string, up bool) (int, string, error) {
	key := "vod.updown." + vodID + ".guest"
	seen, err := s.limiter.Seen(ctx, key)
	if err != nil {
		return -1, "小视频操作失败", err
	}
	if up {
		if seen {
			if err := s.store.IncrementVODCounter(ctx, atoi(vodID), "upnum", -1); err != nil {
				return -1, "小视频操作失败", err
			}
			_ = s.store.RecountUpDown(ctx, atoi(vodID))
			_ = s.limiter.Delete(ctx, key)
			return 0, "已取消赞", nil
		}
		if err := s.store.IncrementVODCounter(ctx, atoi(vodID), "upnum", 1); err != nil {
			return -1, "小视频操作失败", err
		}
		_ = s.store.RecountUpDown(ctx, atoi(vodID))
		_ = s.limiter.Mark(ctx, key)
		return 0, "已赞", nil
	}
	if seen {
		return -1, "您已经赞/踩过了", nil
	}
	if err := s.store.IncrementVODCounter(ctx, atoi(vodID), "downnum", 1); err != nil {
		return -1, "小视频操作失败", err
	}
	_ = s.store.RecountUpDown(ctx, atoi(vodID))
	_ = s.limiter.Mark(ctx, key)
	return 0, "已踩", nil
}

func (s *Service) voteUser(ctx context.Context, uid int, vodID int, up bool) (int, string, error) {
	target := 2
	message := "已踩"
	counter := "downnum"
	if up {
		target = 1
		message = "已赞"
		counter = "upnum"
	}
	item, err := s.store.UpDownByUser(ctx, uid, vodID)
	if err != nil {
		return -1, "小视频操作失败", err
	}
	if len(item) > 0 {
		if err := s.store.DeleteUpDown(ctx, uid, vodID); err != nil {
			return -1, "小视频操作失败", err
		}
		current := atoi(item["updown"])
		if current == target {
			if err := s.store.IncrementVODCounter(ctx, vodID, counter, -1); err != nil {
				return -1, "小视频操作失败", err
			}
			_ = s.store.RecountUpDown(ctx, vodID)
			if up {
				return 0, "已取消赞", nil
			}
			return 0, "已取消踩", nil
		}
	}
	id, err := s.store.SaveUpDown(ctx, uid, vodID, target, s.now().Unix())
	if err != nil {
		return -1, "小视频操作失败", err
	}
	if id == 0 {
		if up {
			return -1, "您已经赞过了", nil
		}
		return -1, "您已经踩过了", nil
	}
	if err := s.store.IncrementVODCounter(ctx, vodID, counter, 1); err != nil {
		return -1, "小视频操作失败", err
	}
	if !up {
		_ = s.store.IncrementVODCounter(ctx, vodID, "upnum", -1)
	}
	_ = s.store.RecountUpDown(ctx, vodID)
	return 0, message, nil
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
	return user, nil
}

func (s *Service) filter(ctx context.Context, action string, params map[string]string, categories []map[string]interface{}, now int64) (minivodRepo.Filter, string, error) {
	filter := minivodRepo.Filter{
		CateIDs:    descendantCategoryIDs(categories, atoi(params["cateid"])),
		AreaID:     atoi(params["areaid"]),
		YearID:     atoi(params["yearid"]),
		TagIDs:     splitIDs(params["tagid"]),
		Definition: atoi(params["definition"]),
		Duration:   atoi(params["duration"]),
		FreeOnly:   atoi(params["freetype"]) > 0,
		Mosaic:     atoi(params["mosaic"]),
		LangVoice:  atoi(params["langvoice"]),
	}
	orderBy := ""
	switch action {
	case "recommend":
		filter.Recommend = true
	case "hot":
		orderBy = "playcount_week DESC"
	case "latest":
		orderBy = "vodid DESC"
	case "topnew":
		orderBy = "RAND()"
	case "topday":
		orderBy = "upnum_day DESC,scorenum DESC"
	case "topweek":
		orderBy = "upnum_week DESC,playcount_total DESC"
	case "topmonth":
		orderBy = "upnum_month DESC, upnum DESC"
	case "topzan", "topcomment", "topplay", "topcoin":
		key := map[string]string{"topzan": "minivod.zan_vodids", "topcomment": "minivod.comment_vodids", "topplay": "minivod.play_vodids", "topcoin": "minivod.coin_vodids"}[action]
		value, err := s.store.Setting(ctx, key)
		if err != nil {
			return filter, "", err
		}
		filter.TopIDs = splitIDs(value)
		orderBy = "FIELD(vodid, " + intListSQL(filter.TopIDs) + ")"
	}
	if orderBy == "" {
		switch atoi(params["orderby"]) {
		case 1:
			orderBy = "upnum DESC"
		case 2:
			orderBy = "playcount_total DESC"
		case 3:
			orderBy = "scorenum DESC"
		default:
			orderBy = "utimestamp DESC"
		}
	}
	return filter, orderBy, nil
}

func (s *Service) richRows(ctx context.Context, action string, params map[string]string, vodRows []map[string]interface{}) ([]map[string]interface{}, error) {
	if !needsUserRows(action, params) {
		return []map[string]interface{}{}, nil
	}
	users, err := s.store.UsersByIDs(ctx, rowIDs(vodRows, "authorid"))
	if err != nil {
		return nil, err
	}
	userByID := map[string]map[string]interface{}{}
	for _, user := range users {
		userByID[str(user["uid"])] = processUser(user, s.resourceBaseURL)
	}
	out := []map[string]interface{}{}
	for _, row := range vodRows {
		var user interface{}
		if found, ok := userByID[str(row["authorid"])]; ok {
			user = found
		}
		out = append(out, map[string]interface{}{"vodrow": row, "user": user})
	}
	return out, nil
}

func needsUserRows(action string, params map[string]string) bool {
	switch action {
	case "topzan", "topcomment", "topplay", "topcoin", "topnew", "topday", "topweek", "topmonth", "latest":
		return true
	}
	return params["tagid"] != "" && params["tagid"] != "0"
}

func parseParams(raw string) map[string]string {
	params := map[string]string{}
	defaults := []string{"0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "1"}
	values := []string{}
	if raw != "" {
		values = strings.Split(raw, "-")
	}
	for i, key := range paramKeys {
		value := defaults[i]
		if i < len(values) && values[i] != "" {
			value = values[i]
		}
		params[key] = value
	}
	return params
}

func buildParams(params map[string]string, replace map[string]string) string {
	values := make([]string, 0, len(paramKeys))
	for _, key := range paramKeys {
		value := params[key]
		if next, ok := replace[key]; ok {
			value = next
		}
		values = append(values, value)
	}
	return strings.Join(values, "-")
}

func descendantCategoryIDs(categories []map[string]interface{}, parent int) []int {
	if parent <= 0 {
		return nil
	}
	result := []int{parent}
	changed := true
	for changed {
		changed = false
		for _, row := range categories {
			id := atoi(row["cateid"])
			pid := atoi(row["parentid"])
			if containsInt(result, pid) && !containsInt(result, id) {
				result = append(result, id)
				changed = true
			}
		}
	}
	return result
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

func processUserFull(row map[string]interface{}, base string, now int64) map[string]interface{} {
	out := processUser(row, base)
	sysgidExp := atoi64(row["sysgid_exptime"])
	duetime := ""
	dueday := ""
	if sysgidExp > 0 {
		duetime = formatTimestamp(sysgidExp)
		if sysgidExp > now {
			dueday = "未过期"
		} else {
			dueday = "已过期"
		}
	}
	out["uniqkey"] = strings.ToUpper(strconv.FormatInt(atoi64(row["uniqkey"]), 36))
	out["mobi"] = str(row["mobi"])
	out["email"] = str(row["email"])
	out["sysgid"] = str(row["sysgid"])
	out["gid"] = str(row["gid"])
	out["gids"] = nil
	out["gicon"] = ""
	out["isvip"] = 0
	out["regtime"] = formatTimestamp(atoi64(row["regtime"]))
	out["newmsg"] = str(row["newmsg"])
	out["goldcoin"] = atoi(row["goldcoin"])
	out["gold_bean"] = atoi(row["gold_bean"])
	out["duetime"] = duetime
	out["dueday"] = dueday
	out["recommend_total"] = atoi(row["recommend_total"])
	return out
}

func formatTimestamp(ts int64) string {
	if ts <= 0 {
		return "1970-01-01 08:00:00"
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return time.Unix(ts, 0).In(loc).Format("2006-01-02 15:04:05")
}

func optionRows(items [][2]interface{}) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		rows = append(rows, map[string]interface{}{"keyid": item[0], "value": item[1]})
	}
	return rows
}

func rowIDs(rows []map[string]interface{}, key string) []int {
	ids := []int{}
	for _, row := range rows {
		id := atoi(row[key])
		if id > 0 && !containsInt(ids, id) {
			ids = append(ids, id)
		}
	}
	return ids
}

func splitIDs(value string) []int {
	ids := []int{}
	for _, part := range strings.Split(value, ",") {
		id := atoi(strings.TrimSpace(part))
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

func intListSQL(ids []int) string {
	parts := []string{}
	for _, id := range ids {
		if id > 0 {
			parts = append(parts, fmt.Sprint(id))
		}
	}
	if len(parts) == 0 {
		return "NULL"
	}
	return strings.Join(parts, ",")
}

func containsInt(values []int, want int) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func collectTagNames(rows []map[string]interface{}) []string {
	seen := map[string]struct{}{}
	names := []string{}
	for _, row := range rows {
		for _, field := range []string{"tags", "actor_tags"} {
			for _, part := range strings.Split(str(row[field]), ",") {
				name := strings.TrimSpace(part)
				if name == "" {
					continue
				}
				if _, ok := seen[name]; ok {
					continue
				}
				seen[name] = struct{}{}
				names = append(names, name)
			}
		}
	}
	return names
}

func tagIDs(rows []map[string]interface{}) []int {
	ids := []int{}
	for _, row := range rows {
		id := atoi(row["tagid"])
		if id > 0 && !containsInt(ids, id) {
			ids = append(ids, id)
		}
	}
	return ids
}

func sanitizePlayURL(value string) string {
	return strings.NewReplacer("\n", "", "\r", "", "\t", "", "'", "", `"`, "").Replace(value)
}

func signCDNURL(httpURL string, row map[string]interface{}, servers []map[string]interface{}, user map[string]interface{}, now int64) string {
	playServerID := atoi(row["play_srvid"])
	for _, server := range servers {
		if playServerID != atoi(server["srvid"]) || str(server["cdnkey"]) == "" || str(server["cdnparam"]) == "" {
			continue
		}
		if !strings.HasPrefix(httpURL, "/") && !hasURLScheme(httpURL) {
			httpURL = "/" + httpURL
		}
		actor := str(user["uid"])
		if atoi(actor) == 0 {
			actor = str(user["sid"])
		}
		if strings.Contains(str(server["cdnparam"]), "tx") {
			sign := fmt.Sprintf("%d-%s-0-", now, actor)
			sum := md5.Sum([]byte(httpURL + "-" + sign + str(server["cdnkey"])))
			return httpURL + "?" + str(server["cdnparam"]) + "=" + sign + hex.EncodeToString(sum[:])
		}
		sign := fmt.Sprintf("%d-0-%s-", now, actor)
		sum := sha256.Sum256([]byte(httpURL + "-" + sign + str(server["cdnkey"])))
		return httpURL + "?" + str(server["cdnparam"]) + "=" + sign + hex.EncodeToString(sum[:])
	}
	return httpURL
}

func serverHost(servers []map[string]interface{}, serverID int) string {
	for _, server := range servers {
		if atoi(server["srvid"]) == serverID {
			return str(server["srvhost"])
		}
	}
	return ""
}

func hasURLScheme(value string) bool {
	if index := strings.Index(value, "://"); index > 1 && index <= 5 {
		return true
	}
	return false
}

func (s *Service) checkMiniPerm(row map[string]interface{}, user map[string]interface{}) (int, string) {
	perms := user["perms"]
	if atoi(row["isvip"]) == 1 && getMiniPermInt(perms, "allow.minivod.vip") != 1 {
		return 5, "VIP独享内容，请升级"
	}
	if atoi(row["islimit"]) > 0 && getMiniPermInt(perms, "allow.minivod.limit") != 1 {
		return 6, "此内容仅提供给高级别用户，请升级或做任务推广吧"
	}
	if atoi(row["islimitv3"]) > 0 && getMiniPermInt(perms, "allow.minivod.limitv3") != 1 {
		return 6, "此内容仅提供给高级别用户，请升级或做任务推广吧"
	}
	return 0, ""
}

func mediaFailMessage(play bool) string {
	if play {
		return "请求小视频播放地址失败"
	}
	return "请求小视频下载地址失败"
}

func mediaSource(row map[string]interface{}, playIndex int, play bool) (string, int, int) {
	if playIndex < 0 {
		playIndex = 0
	}
	if playIndex > 0 {
		field := "playlist"
		if !play {
			field = "downlist"
		}
		if item := playlistItem(str(row[field]), playIndex); len(item) > 0 {
			price := atoi(item["view_price"])
			if price == -1 {
				price = atoi(row["view_price"])
			}
			serverID := atoi(row["play_srvid"])
			if !play {
				serverID = atoi(row["down_srvid"])
			}
			return str(item["play_url"]), price, serverID
		}
		return "", 0, 0
	}
	if play {
		return str(row["play_url"]), atoi(row["view_price"]), atoi(row["play_srvid"])
	}
	return str(row["down_url"]), atoi(row["view_price"]), atoi(row["down_srvid"])
}

func playlistItem(raw string, index int) map[string]interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" || index <= 0 {
		return map[string]interface{}{}
	}
	var rows []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &rows); err == nil && index <= len(rows) {
		return rows[index-1]
	}
	return map[string]interface{}{}
}

func dayStart(now func() time.Time) int64 {
	t := now()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

func boolInt(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

func getMiniPermInt(perms interface{}, key string) int {
	switch typed := perms.(type) {
	case map[string]interface{}:
		return atoi(typed[key])
	case string:
		values := map[string]interface{}{}
		if err := json.Unmarshal([]byte(typed), &values); err == nil {
			return atoi(values[key])
		}
	}
	return 0
}

func categoryParents(categories []map[string]interface{}, cateID int) []map[string]interface{} {
	byID := map[int]map[string]interface{}{}
	for _, row := range categories {
		byID[atoi(row["cateid"])] = row
	}
	stack := []map[string]interface{}{}
	for cateID > 0 {
		row, ok := byID[cateID]
		if !ok {
			break
		}
		stack = append([]map[string]interface{}{{
			"cateid":    str(row["cateid"]),
			"catename":  str(row["catename"]),
			"itemcount": str(row["itemcount"]),
		}}, stack...)
		cateID = atoi(row["parentid"])
	}
	return stack
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

type memoryVoteLimiter struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

func newMemoryVoteLimiter() *memoryVoteLimiter {
	return &memoryVoteLimiter{seen: map[string]struct{}{}}
}

func (l *memoryVoteLimiter) Seen(_ context.Context, key string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, ok := l.seen[key]
	return ok, nil
}

func (l *memoryVoteLimiter) Mark(_ context.Context, key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.seen[key] = struct{}{}
	return nil
}

func (l *memoryVoteLimiter) Delete(_ context.Context, key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.seen, key)
	return nil
}
