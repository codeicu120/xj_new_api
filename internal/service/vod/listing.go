package vod

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
	vodRepo "xj_comp/internal/repository/vod"
	"xj_comp/internal/service/resourceurl"
)

const sampleParams = "$cateid:0-$areaid:0-$yearid:0-$definition:0-$duration:0-$freetype:0-$mosaic:0-$langvoice:0-$orderby:0-$page:1"

var paramKeys = []string{"cateid", "areaid", "yearid", "definition", "duration", "freetype", "mosaic", "langvoice", "orderby", "page"}

var ErrVODNotFound = errors.New("vod not found")

type ListingStore interface {
	Categories(ctx context.Context) ([]map[string]interface{}, error)
	Areas(ctx context.Context) ([]map[string]interface{}, error)
	Years(ctx context.Context) ([]map[string]interface{}, error)
	Servers(ctx context.Context) ([]map[string]interface{}, error)
	CountVODs(ctx context.Context, filter vodRepo.ListingFilter) (int, error)
	ListVODs(ctx context.Context, filter vodRepo.ListingFilter, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error)
	RandomVODs(ctx context.Context, pageSize int) ([]map[string]interface{}, error)
	RandomVODsExcept(ctx context.Context, pageSize int, excludeID int, cateID int) ([]map[string]interface{}, error)
	VODByID(ctx context.Context, vodID int) (map[string]interface{}, error)
	VODsByIDs(ctx context.Context, ids []int, orderBy string) ([]map[string]interface{}, error)
	SimilarVODsByTagIDs(ctx context.Context, tagIDs []int, excludeID int, since int64, pageSize int) ([]map[string]interface{}, error)
	TagsByNames(ctx context.Context, names []string) ([]map[string]interface{}, error)
	CalldataByUUID(ctx context.Context, uuid string) (map[string]interface{}, error)
	VODsByIDsLimited(ctx context.Context, ids []int, freeOnly bool, limit int, orderByField bool) ([]map[string]interface{}, error)
	SearchVODs(ctx context.Context, keyword string, freeOnly bool, limit int) ([]map[string]interface{}, error)
	SearchLog(ctx context.Context, keyword string) (map[string]interface{}, error)
	UpsertSearchLog(ctx context.Context, keyword string, now int64, total int, vodIDs []int) error
	IncrementSearchLog(ctx context.Context, keyword string, previous int64, now int64) error
	TopSearchVODIDs(ctx context.Context) (string, error)
	MiniVODsByIDsLimited(ctx context.Context, ids []int, limit int, orderByField bool) ([]map[string]interface{}, error)
	MiniSearchVODs(ctx context.Context, keyword string, limit int) ([]map[string]interface{}, error)
	MiniSearchLog(ctx context.Context, keyword string) (map[string]interface{}, error)
	UpsertMiniSearchLog(ctx context.Context, keyword string, now int64, total int, vodIDs []int) error
	IncrementMiniSearchLog(ctx context.Context, keyword string, previous int64, now int64) error
	TopMiniSearchVODIDs(ctx context.Context) (string, error)
	MiniVODsByIDs(ctx context.Context, ids []int, orderBy string) ([]map[string]interface{}, error)
	BreakingVOD(ctx context.Context, cateID int, since int64) (map[string]interface{}, error)
	VODErrorByUID(ctx context.Context, uid string, vodID int) (map[string]interface{}, error)
	SaveVODError(ctx context.Context, input vodRepo.ErrorReportInput) (int, error)
	UpDownByUser(ctx context.Context, uid int, vodID int) (map[string]interface{}, error)
	DeleteUpDown(ctx context.Context, uid int, vodID int) error
	SaveUpDown(ctx context.Context, uid int, vodID int, updown int, now int64) (int, error)
	IncrementVODCounter(ctx context.Context, vodID int, field string, delta int) error
	RecountUpDown(ctx context.Context, vodID int) error
	FavoriteCount(ctx context.Context, uid int, vodID int) (int, error)
	BoughtCount(ctx context.Context, uid int, vodID int) (int, error)
	PlayLogCount(ctx context.Context, uid int, sid string, vodID int, playIndex int, since int64) (int, error)
	DownLogCount(ctx context.Context, uid int, sid string, vodID int, playIndex int, since int64) (int, error)
	RecordVODPlay(ctx context.Context, uid int, sid string, vodID int, playIndex int, deduct int, updateTime bool, now int64) error
	RecordVODDown(ctx context.Context, uid int, sid string, vodID int, playIndex int, deduct int, updateTime bool, now int64) error
}

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type AuthGroupStore interface {
	Groups(ctx context.Context) ([]map[string]interface{}, error)
}

type VoteLimiter interface {
	Seen(ctx context.Context, key string) (bool, error)
	Mark(ctx context.Context, key string) error
	Delete(ctx context.Context, key string) error
}

type ListingService struct {
	store           ListingStore
	resourceBaseURL string
	vipDiscount     int
	fetcher         M3U8Fetcher
	auth            AuthStore
	limiter         VoteLimiter
	now             func() time.Time
	resources       *resourceurl.Resolver
}

func (s *ListingService) WithResourceResolver(r *resourceurl.Resolver) *ListingService {
	s.resources = r
	return s
}

type ListingRequest struct {
	Action      string
	PathParams  string
	QueryPage   string
	IsH5Request bool
}

type ErrorReportRequest struct {
	Token      string
	VODID      int
	PlayURL    string
	AppVersion string
	SysVersion string
	Model      string
	Channel    string
	Network    string
	Details    string
	ClientIP   string
}

func NewListingService(store ListingStore, resourceBaseURL string, vipDiscount int) *ListingService {
	if vipDiscount == 0 {
		vipDiscount = 100
	}
	return &ListingService{
		store:           store,
		resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"),
		vipDiscount:     vipDiscount,
		fetcher:         httpM3U8Fetcher{},
		limiter:         newMemoryVoteLimiter(),
		now:             time.Now,
	}
}

func (s *ListingService) ErrorReport(ctx context.Context, req ErrorReportRequest) (int, string, error) {
	if strings.TrimSpace(req.PlayURL) == "" ||
		strings.TrimSpace(req.AppVersion) == "" ||
		strings.TrimSpace(req.SysVersion) == "" ||
		strings.TrimSpace(req.Model) == "" ||
		strings.TrimSpace(req.Network) == "" {
		return -9999, "缺少参数", nil
	}
	vod, err := s.store.VODByID(ctx, req.VODID)
	if err != nil {
		return -1, "提交报错失败", err
	}
	if len(vod) == 0 {
		return -9999, "该视频不存在或者已删除", nil
	}
	uid, err := s.errorReportUID(ctx, req.Token)
	if err != nil {
		return -1, "提交报错失败", err
	}
	row, err := s.store.VODErrorByUID(ctx, uid, req.VODID)
	if err != nil {
		return -1, "提交报错失败", err
	}
	if len(row) > 0 {
		return -9999, "您已提交过该视频报错反馈", nil
	}
	_, err = s.store.SaveVODError(ctx, vodRepo.ErrorReportInput{
		UID:        uid,
		VODID:      req.VODID,
		PlayURL:    strings.TrimSpace(req.PlayURL),
		AppVersion: strings.TrimSpace(req.AppVersion),
		SysVersion: strings.TrimSpace(req.SysVersion),
		Model:      strings.TrimSpace(req.Model),
		Channel:    strings.TrimSpace(req.Channel),
		Network:    strings.TrimSpace(req.Network),
		ClientIP:   strings.TrimSpace(req.ClientIP),
		Details:    strings.TrimSpace(req.Details),
		Now:        s.now().Unix(),
	})
	if err != nil {
		return -1, "提交报错失败", err
	}
	return 0, "", nil
}

func (s *ListingService) errorReportUID(ctx context.Context, token string) (string, error) {
	sid := userRepo.CleanToken(strings.TrimSpace(token))
	if sid == "" {
		return "0", nil
	}
	if s.auth == nil {
		return sid, nil
	}
	user, err := s.auth.UserBySession(ctx, sid)
	if err != nil {
		return "", fmt.Errorf("load user: %w", err)
	}
	if user != nil && atoi(str(user["uid"])) > 0 {
		return str(user["uid"]), nil
	}
	return sid, nil
}

func (s *ListingService) WithAuth(auth AuthStore) *ListingService {
	s.auth = auth
	return s
}

func (s *ListingService) List(ctx context.Context, req ListingRequest) (domain.VODListingData, error) {
	params := parseParams(req.PathParams)
	if atoi(params["page"]) == 0 {
		params["page"] = req.QueryPage
		if params["page"] == "" {
			params["page"] = "0"
		}
	}

	categories, areas, years, servers, err := s.listMetadata(ctx)
	if err != nil {
		return domain.VODListingData{}, err
	}

	categoryIDs := descendantCategoryIDs(categories, atoi(params["cateid"]))
	filter := vodRepo.ListingFilter{
		CateIDs:    categoryIDs,
		AreaID:     atoi(params["areaid"]),
		YearID:     atoi(params["yearid"]),
		Definition: atoi(params["definition"]),
		Duration:   atoi(params["duration"]),
		FreeType:   atoi(params["freetype"]),
		Mosaic:     atoi(params["mosaic"]),
		LangVoice:  atoi(params["langvoice"]),
	}

	pageSize := 16
	page := atoi(params["page"])
	total := 0
	rows := []map[string]interface{}{}
	if req.Action == "recommend" {
		rows, err = s.store.RandomVODs(ctx, pageSize)
	} else {
		orderBy := orderBy(req.Action, atoi(params["orderby"]))
		total, err = s.store.CountVODs(ctx, filter)
		if err == nil {
			rows, err = s.store.ListVODs(ctx, filter, total, page, pageSize, orderBy)
		}
	}
	if err != nil {
		return domain.VODListingData{}, fmt.Errorf("list vods: %w", err)
	}

	tagRows, err := s.store.TagsByNames(ctx, collectTagNames(rows))
	if err != nil {
		return domain.VODListingData{}, fmt.Errorf("list tags: %w", err)
	}

	now := s.now().Unix()
	return domain.VODListingData{
		Now:          now,
		Action:       req.Action,
		SampleParams: sampleParams,
		Params:       params,
		VODRows:      s.processRows(rows, s.withResources(ctx, processEnv(categories, areas, years, servers, tagRows)), req.IsH5Request, now),
		PageInfo:     pageInfo(total, pageSize, page, "/vod/"+req.Action+"-"+buildParams(params, map[string]string{"page": "[?]"})),
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

func (s *ListingService) LikeRows(ctx context.Context, isH5Request bool) (domain.VODLikeRowsData, error) {
	categories, areas, years, servers, err := s.listMetadata(ctx)
	if err != nil {
		return domain.VODLikeRowsData{}, err
	}

	rows, err := s.store.RandomVODs(ctx, 6)
	if err != nil {
		return domain.VODLikeRowsData{}, fmt.Errorf("list random vods: %w", err)
	}
	tagRows, err := s.store.TagsByNames(ctx, collectTagNames(rows))
	if err != nil {
		return domain.VODLikeRowsData{}, fmt.Errorf("list tags: %w", err)
	}

	return domain.VODLikeRowsData{
		LikeRows: s.processRows(rows, s.withResources(ctx, processEnv(categories, areas, years, servers, tagRows)), isH5Request, s.now().Unix()),
	}, nil
}

func (s *ListingService) Search(ctx context.Context, keyword string, freeOnly bool, page int, isH5Request bool) (interface{}, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return s.searchIndex(ctx, freeOnly, isH5Request)
	}
	return s.searchList(ctx, keyword, freeOnly, page, isH5Request)
}

func (s *ListingService) MiniSearch(ctx context.Context, keyword string, page int, isH5Request bool) (interface{}, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return s.miniSearchIndex(ctx, isH5Request)
	}
	return s.miniSearchList(ctx, keyword, page, isH5Request)
}

func (s *ListingService) searchIndex(ctx context.Context, freeOnly bool, isH5Request bool) (domain.SearchIndexData, error) {
	categories, areas, years, servers, err := s.listMetadata(ctx)
	if err != nil {
		return domain.SearchIndexData{}, err
	}

	hotWordsRow, err := s.store.CalldataByUUID(ctx, "search.hotwords")
	if err != nil {
		return domain.SearchIndexData{}, fmt.Errorf("get hotwords: %w", err)
	}
	hotWords := decodeCalldataJSON(hotWordsRow)

	hotVODRow, err := s.store.CalldataByUUID(ctx, "search.hotvods")
	if err != nil {
		return domain.SearchIndexData{}, fmt.Errorf("get hot vods: %w", err)
	}
	ids := splitIDs(str(hotVODRow["content"]))
	hotRows, err := s.store.VODsByIDsLimited(ctx, ids, freeOnly, 20, true)
	if err != nil {
		return domain.SearchIndexData{}, fmt.Errorf("list hot vods: %w", err)
	}

	likeTags, err := s.searchLikeTags(ctx, freeOnly)
	if err != nil {
		return domain.SearchIndexData{}, fmt.Errorf("list search like tags: %w", err)
	}

	tagRows, err := s.store.TagsByNames(ctx, collectFieldTagNames(hotRows, "tags"))
	if err != nil {
		return domain.SearchIndexData{}, fmt.Errorf("list hot tags: %w", err)
	}

	return domain.SearchIndexData{
		HotWords:    hotWords,
		HotRows:     s.processRowsWithDiscount(hotRows, processEnv(categories, areas, years, servers, tagRows), isH5Request, s.now().Unix(), 100),
		YouMayLikes: likeTags,
	}, nil
}

func (s *ListingService) miniSearchIndex(ctx context.Context, isH5Request bool) (domain.SearchIndexData, error) {
	categories, areas, years, servers, err := s.listMetadata(ctx)
	if err != nil {
		return domain.SearchIndexData{}, err
	}

	hotWordsRow, err := s.store.CalldataByUUID(ctx, "search.minihotwords")
	if err != nil {
		return domain.SearchIndexData{}, fmt.Errorf("get mini hotwords: %w", err)
	}
	hotWords := decodeCalldataJSON(hotWordsRow)

	hotVODRow, err := s.store.CalldataByUUID(ctx, "search.minihotvods")
	if err != nil {
		return domain.SearchIndexData{}, fmt.Errorf("get mini hot vods: %w", err)
	}
	ids := splitIDs(str(hotVODRow["content"]))
	hotRows, err := s.store.MiniVODsByIDsLimited(ctx, ids, 20, true)
	if err != nil {
		return domain.SearchIndexData{}, fmt.Errorf("list mini hot vods: %w", err)
	}

	likeTags, err := s.miniSearchLikeTags(ctx)
	if err != nil {
		return domain.SearchIndexData{}, fmt.Errorf("list mini search like tags: %w", err)
	}

	tagRows, err := s.store.TagsByNames(ctx, collectFieldTagNames(hotRows, "tags"))
	if err != nil {
		return domain.SearchIndexData{}, fmt.Errorf("list mini hot tags: %w", err)
	}
	vodRows := s.processMiniRowsWithDiscount(hotRows, s.withResources(ctx, processEnv(categories, areas, years, servers, tagRows)), isH5Request, s.now().Unix(), 100)

	return domain.SearchIndexData{
		HotWords:    hotWords,
		HotRows:     wrapVODRows(vodRows),
		YouMayLikes: likeTags,
	}, nil
}

func (s *ListingService) searchList(ctx context.Context, keyword string, freeOnly bool, page int, isH5Request bool) (domain.SearchListData, error) {
	categories, areas, years, servers, err := s.listMetadata(ctx)
	if err != nil {
		return domain.SearchListData{}, err
	}

	now := s.now().Unix()
	pageSize := 16
	if page < 1 {
		page = 1
	}

	searchLog, err := s.store.SearchLog(ctx, keyword)
	if err != nil {
		return domain.SearchListData{}, fmt.Errorf("get search log: %w", err)
	}

	total := 0
	rows := []map[string]interface{}{}
	if len(searchLog) == 0 || atoi64(str(searchLog["schtime"]))+3600 < now {
		allRows, err := s.store.SearchVODs(ctx, keyword, false, 1000)
		if err != nil {
			return domain.SearchListData{}, fmt.Errorf("search vods: %w", err)
		}
		allIDs := rowIDs(allRows, "vodid")
		if err := s.store.UpsertSearchLog(ctx, keyword, now, len(allIDs), allIDs); err != nil {
			return domain.SearchListData{}, err
		}
		if freeOnly {
			allRows, err = s.store.VODsByIDsLimited(ctx, allIDs, true, 1000, false)
			if err != nil {
				return domain.SearchListData{}, fmt.Errorf("search free vods: %w", err)
			}
		}
		total = len(allRows)
		rows = pageSlice(allRows, page, pageSize)
	} else if atoi(str(searchLog["total"])) > 0 {
		ids := splitIDs(str(searchLog["vodids"]))
		total = len(ids)
		ids = pageSliceInts(ids, page, pageSize)
		rows, err = s.store.VODsByIDsLimited(ctx, ids, freeOnly, 1000, false)
		if err != nil {
			return domain.SearchListData{}, fmt.Errorf("list cached search vods: %w", err)
		}
	}

	pageData := pageInfo(total, pageSize, page, "/search?wd="+url.QueryEscape(keyword)+"&page=[?]")
	if atoi(str(pageData["page"])) == 1 {
		prev := atoi64(str(searchLog["sch_lasttime"]))
		if len(searchLog) == 0 {
			prev = now
		}
		if err := s.store.IncrementSearchLog(ctx, keyword, prev, now); err != nil {
			return domain.SearchListData{}, err
		}
	}

	tagRows, err := s.store.TagsByNames(ctx, collectTagNames(rows))
	if err != nil {
		return domain.SearchListData{}, fmt.Errorf("list search tags: %w", err)
	}

	return domain.SearchListData{
		VODRows:  s.processRowsWithDiscount(rows, processEnv(categories, areas, years, servers, tagRows), isH5Request, now, 100),
		PageInfo: pageData,
	}, nil
}

func (s *ListingService) miniSearchList(ctx context.Context, keyword string, page int, isH5Request bool) (domain.MiniSearchListData, error) {
	categories, areas, years, servers, err := s.listMetadata(ctx)
	if err != nil {
		return domain.MiniSearchListData{}, err
	}

	now := s.now().Unix()
	pageSize := 16
	if page < 1 {
		page = 1
	}

	searchLog, err := s.store.MiniSearchLog(ctx, keyword)
	if err != nil {
		return domain.MiniSearchListData{}, fmt.Errorf("get mini search log: %w", err)
	}

	total := 0
	rows := []map[string]interface{}{}
	if len(searchLog) == 0 || atoi64(str(searchLog["schtime"]))+3600 < now {
		allRows, err := s.store.MiniSearchVODs(ctx, keyword, 1000)
		if err != nil {
			return domain.MiniSearchListData{}, fmt.Errorf("search mini vods: %w", err)
		}
		allIDs := rowIDs(allRows, "vodid")
		if err := s.store.UpsertMiniSearchLog(ctx, keyword, now, len(allIDs), allIDs); err != nil {
			return domain.MiniSearchListData{}, err
		}
		total = len(allRows)
		rows = pageSlice(allRows, page, pageSize)
	} else if atoi(str(searchLog["total"])) > 0 {
		ids := splitIDs(str(searchLog["vodids"]))
		total = len(ids)
		ids = pageSliceInts(ids, page, pageSize)
		rows, err = s.store.MiniVODsByIDsLimited(ctx, ids, 1000, false)
		if err != nil {
			return domain.MiniSearchListData{}, fmt.Errorf("list cached mini search vods: %w", err)
		}
	}

	pageData := pageInfo(total, pageSize, page, "/search?wd="+url.QueryEscape(keyword)+"&page=[?]")
	if atoi(str(pageData["page"])) == 1 {
		prev := atoi64(str(searchLog["sch_lasttime"]))
		if len(searchLog) == 0 {
			prev = now
		}
		if err := s.store.IncrementMiniSearchLog(ctx, keyword, prev, now); err != nil {
			return domain.MiniSearchListData{}, err
		}
	}

	tagRows, err := s.store.TagsByNames(ctx, collectTagNames(rows))
	if err != nil {
		return domain.MiniSearchListData{}, fmt.Errorf("list mini search tags: %w", err)
	}
	vodRows := s.processMiniRowsWithDiscount(rows, s.withResources(ctx, processEnv(categories, areas, years, servers, tagRows)), isH5Request, now, 100)

	return domain.MiniSearchListData{
		Rows:     wrapVODRows(vodRows),
		PageInfo: pageData,
	}, nil
}

func (s *ListingService) Show(ctx context.Context, vodID int, isH5Request bool) (domain.VODShowData, error) {
	categories, areas, years, servers, err := s.listMetadata(ctx)
	if err != nil {
		return domain.VODShowData{}, err
	}

	row, err := s.store.VODByID(ctx, vodID)
	if err != nil {
		return domain.VODShowData{}, fmt.Errorf("get vod: %w", err)
	}
	if len(row) == 0 || atoi(str(row["showtype"])) > 0 {
		return domain.VODShowData{}, ErrVODNotFound
	}

	tagRows, err := s.store.TagsByNames(ctx, collectTagNames([]map[string]interface{}{row}))
	if err != nil {
		return domain.VODShowData{}, fmt.Errorf("list tags: %w", err)
	}

	similarRows, err := s.store.SimilarVODsByTagIDs(ctx, tagIDs(tagRows), vodID, s.now().AddDate(0, 0, -90).Unix(), 11)
	if err != nil {
		return domain.VODShowData{}, fmt.Errorf("list similar vods: %w", err)
	}
	if len(similarRows) > 10 {
		similarRows = similarRows[:10]
	}
	if len(similarRows) < 10 {
		fillRows, err := s.store.RandomVODsExcept(ctx, 10-len(similarRows), vodID, 0)
		if err != nil {
			return domain.VODShowData{}, fmt.Errorf("list similar fallback vods: %w", err)
		}
		similarRows = append(similarRows, fillRows...)
		if len(similarRows) > 10 {
			similarRows = similarRows[:10]
		}
	}

	likeRows, err := s.store.RandomVODsExcept(ctx, 5, vodID, atoi(str(row["cateid"])))
	if err != nil {
		return domain.VODShowData{}, fmt.Errorf("list like vods: %w", err)
	}

	allRows := append([]map[string]interface{}{row}, similarRows...)
	allRows = append(allRows, likeRows...)
	allTagRows, err := s.store.TagsByNames(ctx, collectTagNames(allRows))
	if err != nil {
		return domain.VODShowData{}, fmt.Errorf("list all tags: %w", err)
	}

	env := processEnv(categories, areas, years, servers, allTagRows)
	now := s.now().Unix()
	vodRows := s.processRows([]map[string]interface{}{row}, s.withResources(ctx, env), isH5Request, now)
	vodRow := map[string]interface{}{}
	if len(vodRows) > 0 {
		vodRow = vodRows[0]
	}

	return domain.VODShowData{
		VODRow:      vodRow,
		Categories:  categoryParents(categories, atoi(str(row["cateid"]))),
		SimilarRows: s.processRows(similarRows, s.withResources(ctx, env), isH5Request, now),
		LikeRows:    s.processRows(likeRows, s.withResources(ctx, env), isH5Request, now),
	}, nil
}

func (s *ListingService) Vote(ctx context.Context, token string, vodID int, up bool) (int, string, error) {
	row, err := s.store.VODByID(ctx, vodID)
	if err != nil {
		return -1, "视频操作失败", err
	}
	if len(row) == 0 || atoi(str(row["showtype"])) > 0 {
		return -1, "记录不存在或已被删除", nil
	}
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -1, "视频操作失败", err
	}
	uid := atoi(str(user["uid"]))
	if uid == 0 {
		return s.voteGuest(ctx, str(row["vodid"]), up)
	}
	return s.voteUser(ctx, uid, atoi(str(row["vodid"])), up)
}

func (s *ListingService) ReqPlay(ctx context.Context, token string, vodID int, playIndex int) (map[string]interface{}, int, string, error) {
	return s.reqMedia(ctx, token, vodID, playIndex, true)
}

func (s *ListingService) ReqDown(ctx context.Context, token string, vodID int, playIndex int) (map[string]interface{}, int, string, error) {
	return s.reqMedia(ctx, token, vodID, playIndex, false)
}

func (s *ListingService) reqMedia(ctx context.Context, token string, vodID int, playIndex int, play bool) (map[string]interface{}, int, string, error) {
	row, err := s.store.VODByID(ctx, vodID)
	if err != nil {
		return nil, -1, vodMediaFailMessage(play), err
	}
	if len(row) == 0 || atoi(str(row["showtype"])) > 0 {
		return map[string]interface{}{}, 1, "记录不存在或已被删除", nil
	}
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return nil, -1, vodMediaFailMessage(play), err
	}
	now := s.now().Unix()
	price := atoi(str(row["view_price"]))
	isLimitFree := atoi64(str(row["free_sdate"])) < now && now < atoi64(str(row["free_edate"]))
	if play && atoi(str(row["isvip"])) == 2 && price > 0 && !isLimitFree && !hasVODVIPPerm(user) {
		count, err := s.store.BoughtCount(ctx, atoi(str(user["uid"])), vodID)
		if err != nil {
			return nil, -1, vodMediaFailMessage(play), err
		}
		if count == 0 {
			return map[string]interface{}{"isbought": 0}, 168, "此内容需要付费购买.", nil
		}
	}
	if retcode, errmsg := checkVODPerm(row, user); retcode != 0 {
		return map[string]interface{}{}, retcode, errmsg, nil
	}
	mediaURL, price, serverID := vodMediaSource(row, playIndex, play)
	if mediaURL == "" {
		if play {
			return map[string]interface{}{}, 2, "播放地址不存在", nil
		}
		return map[string]interface{}{}, 2, "下载地址不存在", nil
	}
	if atoi(str(user["uid"])) == 0 && str(user["sid"]) == "" {
		return map[string]interface{}{}, -9999, "客户端游客请先携带信息", nil
	}
	servers, err := s.store.Servers(ctx)
	if err != nil {
		return nil, -1, vodMediaFailMessage(play), err
	}
	httpURL := sanitizeMediaURL(mediaURL)
	if play {
		httpURL = signVODCDNURL(httpURL, row, servers, user, s.now().Unix())
	}
	if !hasURLScheme(httpURL) {
		host := strings.TrimRight(vodServerHost(servers, serverID), "/")
		httpURL = host + "/" + strings.TrimLeft(httpURL, "/")
	}
	data := map[string]interface{}{}
	if play {
		count, err := s.store.FavoriteCount(ctx, atoi(str(user["uid"])), vodID)
		if err != nil {
			return nil, -1, vodMediaFailMessage(play), err
		}
		data["isfavorite"] = boolInt(count > 0)
		data["iszan"] = 0
		if atoi(str(user["uid"])) > 0 {
			updown, err := s.store.UpDownByUser(ctx, atoi(str(user["uid"])), vodID)
			if err != nil {
				return nil, -1, vodMediaFailMessage(play), err
			}
			data["iszan"] = atoi(str(updown["updown"]))
		}
		data["encurl"] = 0
	}
	if price == 0 || (atoi64(str(row["free_sdate"])) < now && now < atoi64(str(row["free_edate"]))) {
		data["httpurl"] = httpURL
		if err := s.recordVODMedia(ctx, user, vodID, playIndex, play, false, now); err != nil {
			return nil, -1, vodMediaFailMessage(play), err
		}
		if play {
			data["httpurls"] = []map[string]interface{}{{"hdtype": "默认", "httpurl": httpURL}}
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
	since := now - 86400*7
	if !play {
		since = now - 86400*365
	}
	var watched int
	if play {
		watched, err = s.store.PlayLogCount(ctx, atoi(str(user["uid"])), str(user["sid"]), vodID, playIndex, since)
	} else {
		watched, err = s.store.DownLogCount(ctx, atoi(str(user["uid"])), str(user["sid"]), vodID, playIndex, since)
	}
	if err != nil {
		return nil, -1, vodMediaFailMessage(play), err
	}
	if watched > 0 {
		data["httpurl"] = httpURL
		if err := s.recordVODMedia(ctx, user, vodID, playIndex, play, false, now); err != nil {
			return nil, -1, vodMediaFailMessage(play), err
		}
		if play {
			data["httpurls"] = []map[string]interface{}{{"hdtype": "默认", "httpurl": httpURL}}
			return data, 0, "本周已观看过继续提供", nil
		}
		return data, 0, "本周已下载过继续提供", nil
	}
	actionDayCount := 0
	dayStart := dayStart(s.now)
	if play {
		actionDayCount, err = s.store.PlayLogCount(ctx, atoi(str(user["uid"])), str(user["sid"]), 0, 0, dayStart)
	} else {
		actionDayCount, err = s.store.DownLogCount(ctx, atoi(str(user["uid"])), str(user["sid"]), 0, 0, dayStart)
	}
	if err != nil {
		return nil, -1, vodMediaFailMessage(play), err
	}
	dayKey := "max.vod.play.daynum"
	if !play {
		dayKey = "max.vod.down.daynum"
	}
	if actionDayCount < getVODPermInt(user["perms"], dayKey) {
		data["httpurl"] = httpURL
		if err := s.recordVODMedia(ctx, user, vodID, playIndex, play, true, now); err != nil {
			return nil, -1, vodMediaFailMessage(play), err
		}
		if play {
			data["httpurls"] = []map[string]interface{}{{"hdtype": "默认", "httpurl": httpURL}}
			if atoi(str(user["uid"])) > 0 {
				return data, 0, "用户权限范围内免费播放", nil
			}
			return data, 0, "游客权限范围内免费播放", nil
		}
		if atoi(str(user["uid"])) > 0 {
			return data, 0, "用户权限范围内免费下载", nil
		}
		return data, 0, "游客权限范围内免费下载", nil
	}
	if play {
		data["httpurl_preview"] = httpURL + "?300"
		if atoi(str(user["uid"])) > 0 {
			return data, 4, "今日观影次数已用完，是否去免费增加次数？", nil
		}
		return data, 3, "今日观看次数已看完，请点击免费注册会员获取更多影片观看次数。", nil
	}
	if atoi(str(user["uid"])) > 0 {
		return data, 4, "今日下载次数已用完，是否去免费增加次数？", nil
	}
	return data, 3, "今日下载次数已用完，请点击免费注册会员获取更多影片下载次数。", nil
}

func (s *ListingService) recordVODMedia(ctx context.Context, user map[string]interface{}, vodID int, playIndex int, play bool, updateTime bool, now int64) error {
	uid := atoi(str(user["uid"]))
	sid := str(user["sid"])
	if play {
		return s.store.RecordVODPlay(ctx, uid, sid, vodID, playIndex, 0, updateTime, now)
	}
	return s.store.RecordVODDown(ctx, uid, sid, vodID, playIndex, 0, updateTime, now)
}

func (s *ListingService) Breaking(ctx context.Context) (map[string]interface{}, int, string, error) {
	now := s.now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	row, err := s.store.BreakingVOD(ctx, 99, dayStart)
	if err != nil {
		return nil, -1, "获取每日爆料失败", err
	}
	if len(row) == 0 || atoi(str(row["showtype"])) > 0 {
		return nil, -1, "记录不存在或已被删除", nil
	}
	return map[string]interface{}{
		"vodid": row["vodid"],
		"title": row["title"],
	}, 0, "ok", nil
}

func (s *ListingService) voteGuest(ctx context.Context, vodID string, up bool) (int, string, error) {
	key := "vod.updown." + vodID + ".guest"
	seen, err := s.limiter.Seen(ctx, key)
	if err != nil {
		return -1, "视频操作失败", err
	}
	if up {
		if seen {
			if err := s.store.IncrementVODCounter(ctx, atoi(vodID), "upnum", -1); err != nil {
				return -1, "视频操作失败", err
			}
			_ = s.store.RecountUpDown(ctx, atoi(vodID))
			_ = s.limiter.Delete(ctx, key)
			return 0, "已取消赞", nil
		}
		if err := s.store.IncrementVODCounter(ctx, atoi(vodID), "upnum", 1); err != nil {
			return -1, "视频操作失败", err
		}
		_ = s.store.RecountUpDown(ctx, atoi(vodID))
		_ = s.limiter.Mark(ctx, key)
		return 0, "已赞", nil
	}
	if seen {
		return -1, "您已经赞/踩过了", nil
	}
	if err := s.store.IncrementVODCounter(ctx, atoi(vodID), "downnum", 1); err != nil {
		return -1, "视频操作失败", err
	}
	_ = s.store.IncrementVODCounter(ctx, atoi(vodID), "upnum", -1)
	_ = s.store.RecountUpDown(ctx, atoi(vodID))
	_ = s.limiter.Mark(ctx, key)
	return 0, "已踩", nil
}

func (s *ListingService) voteUser(ctx context.Context, uid int, vodID int, up bool) (int, string, error) {
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
		return -1, "视频操作失败", err
	}
	if len(item) > 0 {
		if err := s.store.DeleteUpDown(ctx, uid, vodID); err != nil {
			return -1, "视频操作失败", err
		}
		current := atoi(str(item["updown"]))
		if current == target {
			if err := s.store.IncrementVODCounter(ctx, vodID, counter, -1); err != nil {
				return -1, "视频操作失败", err
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
		return -1, "视频操作失败", err
	}
	if id == 0 {
		if up {
			return -1, "您已经赞过了", nil
		}
		return -1, "您已经踩过了", nil
	}
	if err := s.store.IncrementVODCounter(ctx, vodID, counter, 1); err != nil {
		return -1, "视频操作失败", err
	}
	if !up {
		_ = s.store.IncrementVODCounter(ctx, vodID, "upnum", -1)
	}
	_ = s.store.RecountUpDown(ctx, vodID)
	return 0, message, nil
}

func (s *ListingService) userByToken(ctx context.Context, token string) (map[string]interface{}, error) {
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
	if atoi(str(user["uid"])) > 0 && user["perms"] == nil {
		if groupStore, ok := s.auth.(AuthGroupStore); ok {
			groups, err := groupStore.Groups(ctx)
			if err != nil {
				return nil, err
			}
			user["perms"] = initVODPerm(initVODGids(user, s.now), groups)
		}
	}
	return user, nil
}

func checkVODPerm(row map[string]interface{}, user map[string]interface{}) (int, string) {
	perms := user["perms"]
	if atoi(str(row["isvip"])) == 1 && getVODPermInt(perms, "allow.vod.vip") != 1 {
		return 5, "VIP独享内容，请升级"
	}
	if atoi(str(row["islimit"])) > 0 && getVODPermInt(perms, "allow.vod.limit") != 1 {
		return 6, "此内容仅提供给高级别用户，请升级或做任务推广吧"
	}
	if atoi(str(row["islimitv3"])) > 0 && getVODPermInt(perms, "allow.vod.limitv3") != 1 {
		return 6, "此内容仅提供给高级别用户，请升级或做任务推广吧"
	}
	return 0, ""
}

func hasVODVIPPerm(user map[string]interface{}) bool {
	return getVODPermInt(user["perms"], "allow.vod.vip") == 1
}

func vodMediaFailMessage(play bool) string {
	if play {
		return "请求播放地址失败"
	}
	return "请求下载地址失败"
}

func vodMediaSource(row map[string]interface{}, playIndex int, play bool) (string, int, int) {
	if playIndex < 0 {
		playIndex = 0
	}
	serverID := atoi(str(row["play_srvid"]))
	if !play {
		serverID = atoi(str(row["down_srvid"]))
	}
	if playIndex > 0 {
		field := "playlist"
		if !play {
			field = "downlist"
		}
		if item := vodPlaylistItem(str(row[field]), playIndex); len(item) > 0 {
			price := atoi(str(item["view_price"]))
			if price == -1 {
				price = atoi(str(row["view_price"]))
			}
			return str(item["play_url"]), price, serverID
		}
		return "", 0, 0
	}
	if play {
		return str(row["play_url"]), atoi(str(row["view_price"])), serverID
	}
	return str(row["down_url"]), atoi(str(row["view_price"])), serverID
}

func vodPlaylistItem(raw string, index int) map[string]interface{} {
	raw = strings.ReplaceAll(raw, "\r", "")
	raw = strings.TrimSpace(raw)
	if raw == "" || index <= 0 {
		return map[string]interface{}{}
	}
	lines := strings.Split(raw, "\n")
	if index > len(lines) {
		return map[string]interface{}{}
	}
	parts := strings.Split(lines[index-1], "$")
	for len(parts) < 3 {
		parts = append(parts, "")
	}
	return map[string]interface{}{
		"play_name":  parts[0],
		"play_url":   parts[1],
		"view_price": ternary(isNumeric(parts[2]), atoi(parts[2]), -1),
	}
}

func sanitizeMediaURL(value string) string {
	return strings.NewReplacer("\n", "", "\r", "", "\t", "", "'", "", `"`, "").Replace(value)
}

func signVODCDNURL(httpURL string, row map[string]interface{}, servers []map[string]interface{}, user map[string]interface{}, now int64) string {
	playServerID := atoi(str(row["play_srvid"]))
	for _, server := range servers {
		if playServerID != atoi(str(server["srvid"])) || str(server["cdnkey"]) == "" || str(server["cdnparam"]) == "" {
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

func vodServerHost(servers []map[string]interface{}, serverID int) string {
	for _, server := range servers {
		if atoi(str(server["srvid"])) == serverID {
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

func dayStart(now func() time.Time) int64 {
	t := now()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

func getVODPermInt(perms interface{}, key string) int {
	switch typed := perms.(type) {
	case map[string]interface{}:
		return atoi(str(typed[key]))
	case string:
		values := map[string]interface{}{}
		if err := json.Unmarshal([]byte(typed), &values); err == nil {
			return atoi(str(values[key]))
		}
	}
	return 0
}

func initVODGids(user map[string]interface{}, now func() time.Time) []int {
	mainGID := atoi(str(user["gid"]))
	if atoi(str(user["sysgid"])) > 0 {
		mainGID = atoi(str(user["sysgid"]))
	}
	gids := []int{mainGID}
	var extra map[string]interface{}
	switch typed := user["gids"].(type) {
	case map[string]interface{}:
		extra = typed
	case string:
		if typed != "" {
			_ = json.Unmarshal([]byte(typed), &extra)
		}
	}
	ts := now().Unix()
	for gid, exptime := range extra {
		if atoi(str(exptime)) == 0 || atoi64(str(exptime)) > ts {
			gids = append(gids, atoi(str(gid)))
		}
	}
	return uniqueVODInts(gids)
}

func initVODPerm(gids []int, groups []map[string]interface{}) map[string]interface{} {
	selected := make([]map[string]interface{}, 0, len(gids))
	for _, gid := range gids {
		for _, group := range groups {
			if atoi(str(group["scope"])) > 0 || atoi(str(group["gid"])) != gid {
				continue
			}
			selected = append(selected, group)
			break
		}
	}
	sort.SliceStable(selected, func(i, j int) bool {
		return atoi(str(selected[i]["weight"])) > atoi(str(selected[j]["weight"]))
	})
	multiPerms := make([]map[string]interface{}, 0, len(selected))
	for _, group := range selected {
		multiPerms = append(multiPerms, parseVODPermMap(group["perms"]))
	}
	return computeVODPerm(multiPerms)
}

func parseVODPermMap(value interface{}) map[string]interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed
	case string:
		if typed == "" {
			return map[string]interface{}{}
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(typed), &parsed); err != nil {
			return map[string]interface{}{}
		}
		return parsed
	default:
		return map[string]interface{}{}
	}
}

func computeVODPerm(multiPerms []map[string]interface{}) map[string]interface{} {
	keys := make([]string, 0)
	seen := map[string]struct{}{}
	for _, perms := range multiPerms {
		for key := range perms {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			keys = append(keys, key)
		}
	}

	result := make(map[string]interface{}, len(keys))
	for _, key := range keys {
		switch strings.SplitN(key, ".", 2)[0] {
		case "allow", "deny":
			value := 0
			for _, perms := range multiPerms {
				if atoi(str(perms[key])) == 1 {
					value = 1
					break
				}
			}
			result[key] = value
		case "min":
			value := 0
			for _, perms := range multiPerms {
				if _, ok := perms[key]; ok {
					value = minVODInt(value, atoi(str(perms[key])))
				}
			}
			result[key] = value
		case "max":
			value := 0
			for _, perms := range multiPerms {
				if _, ok := perms[key]; ok {
					value = maxVODInt(value, atoi(str(perms[key])))
				}
			}
			result[key] = value
		case "list":
			value := ""
			for _, perms := range multiPerms {
				if str(perms[key]) == "" {
					continue
				}
				if value == "" {
					value = str(perms[key])
				} else {
					value += "," + str(perms[key])
				}
			}
			result[key] = value
		default:
			for _, perms := range multiPerms {
				if _, ok := perms[key]; ok {
					result[key] = perms[key]
					break
				}
			}
		}
	}
	return result
}

func uniqueVODInts(values []int) []int {
	out := make([]int, 0, len(values))
	seen := map[int]struct{}{}
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func minVODInt(a int, b int) int {
	if a == 0 || b < a {
		return b
	}
	return a
}

func maxVODInt(a int, b int) int {
	if b > a {
		return b
	}
	return a
}

func (s *ListingService) ProcessRows(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error) {
	categories, areas, years, servers, err := s.listMetadata(ctx)
	if err != nil {
		return nil, err
	}
	tagRows, err := s.store.TagsByNames(ctx, collectTagNames(rows))
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	return s.processRows(rows, s.withResources(ctx, processEnv(categories, areas, years, servers, tagRows)), isH5Request, s.now().Unix()), nil
}

func (s *ListingService) ProcessRowsFullPrice(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error) {
	categories, areas, years, servers, err := s.listMetadata(ctx)
	if err != nil {
		return nil, err
	}
	tagRows, err := s.store.TagsByNames(ctx, collectTagNames(rows))
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	return s.processRowsWithDiscount(rows, processEnv(categories, areas, years, servers, tagRows), isH5Request, s.now().Unix(), 100), nil
}

func (s *ListingService) ProcessRowsPlain(_ context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error) {
	empty := []map[string]interface{}{}
	return s.processRowsWithDiscount(rows, processEnv(empty, empty, empty, empty, empty), isH5Request, s.now().Unix(), 100), nil
}

func (s *ListingService) ProcessMiniRows(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error) {
	categories, areas, years, servers, err := s.listMetadata(ctx)
	if err != nil {
		return nil, err
	}
	tagRows, err := s.store.TagsByNames(ctx, collectTagNames(rows))
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	return s.processMiniRowsWithDiscount(rows, s.withResources(ctx, processEnv(categories, areas, years, servers, tagRows)), isH5Request, s.now().Unix(), s.vipDiscount), nil
}

func (s *ListingService) ProcessMiniRowsFullPrice(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error) {
	categories, areas, years, servers, err := s.listMetadata(ctx)
	if err != nil {
		return nil, err
	}
	tagRows, err := s.store.TagsByNames(ctx, collectTagNames(rows))
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	return s.processMiniRowsWithDiscount(rows, s.withResources(ctx, processEnv(categories, areas, years, servers, tagRows)), isH5Request, s.now().Unix(), 100), nil
}

func (s *ListingService) listMetadata(ctx context.Context) ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, []map[string]interface{}, error) {
	categories, err := s.store.Categories(ctx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("list categories: %w", err)
	}
	areas, err := s.store.Areas(ctx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("list areas: %w", err)
	}
	years, err := s.store.Years(ctx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("list years: %w", err)
	}
	servers, err := s.store.Servers(ctx)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("list servers: %w", err)
	}
	return categories, areas, years, servers, nil
}

func parseParams(raw string) map[string]string {
	params := map[string]string{}
	defaults := []string{"0", "0", "0", "0", "0", "0", "0", "0", "0", "1"}
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

func orderBy(action string, order int) string {
	switch action {
	case "hot":
		return "playcount_week DESC"
	case "latest":
		return "vodid DESC"
	}
	switch order {
	case 1:
		return "upnum DESC"
	case 2:
		return "playcount_total DESC"
	case 3:
		return "scorenum DESC"
	default:
		return "utimestamp DESC"
	}
}

func processEnv(categories, areas, years, servers, tags []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"categories": indexBy(categories, "cateid"),
		"areas":      indexBy(areas, "areaid"),
		"years":      indexBy(years, "yearid"),
		"servers":    groupBy(servers, "srvtype"),
		"tags":       indexBy(tags, "tagname"),
	}
}

func (s *ListingService) processRows(rows []map[string]interface{}, env map[string]interface{}, isH5 bool, now int64) []map[string]interface{} {
	return s.processRowsWithDiscount(rows, env, isH5, now, s.vipDiscount)
}

func (s *ListingService) processRowsWithDiscount(rows []map[string]interface{}, env map[string]interface{}, isH5 bool, now int64, vipDiscount int) []map[string]interface{} {
	return s.processRowsWithMode(rows, env, isH5, now, vipDiscount, false)
}

func (s *ListingService) processMiniRowsWithDiscount(rows []map[string]interface{}, env map[string]interface{}, isH5 bool, now int64, vipDiscount int) []map[string]interface{} {
	return s.processRowsWithMode(rows, env, isH5, now, vipDiscount, true)
}

func (s *ListingService) processRowsWithMode(rows []map[string]interface{}, env map[string]interface{}, isH5 bool, now int64, vipDiscount int, mini bool) []map[string]interface{} {
	result := []map[string]interface{}{}
	categories := env["categories"].(map[string]map[string]interface{})
	areas := env["areas"].(map[string]map[string]interface{})
	years := env["years"].(map[string]map[string]interface{})
	servers := env["servers"].(map[string][]map[string]interface{})
	tags := env["tags"].(map[string]map[string]interface{})
	resolved, _ := env["resources"].(resourceurl.Resolved)
	if resolved.BaseURL == "" {
		resolved.BaseURL = s.resourceBaseURL
		resolved.Timestamp = now
	}
	previewBase := ""
	if list := servers["preview"]; len(list) > 0 {
		previewBase = fmt.Sprint(list[0]["srvhost"])
	}
	if previewBase == "" {
		previewBase = ""
	}
	for _, row := range rows {
		vodid := str(row["vodid"])
		urlPrefix := "/vod"
		if mini {
			urlPrefix = "/minivod"
		}
		playURL := ""
		downURL := ""
		previewURL := ""
		if str(row["play_url"]) != "" {
			playURL = urlPrefix + "/reqplay/" + vodid
			previewURL = strings.TrimRight(previewBase, "/") + urlPrefix + "/preView/" + vodid + "/index.m3u8"
		}
		if str(row["down_url"]) != "" {
			downURL = urlPrefix + "/reqdown/" + vodid
		}
		coverRaw := cleanCover(str(row["coverpic"]))
		item := map[string]interface{}{
			"vodid":           vodid,
			"title":           str(row["title"]),
			"intro":           str(row["intro"]),
			"coverpic":        s.coverPic(coverRaw, atoi(str(row["cover_srvid"])), servers["cover"], isH5, resolved),
			"coverx":          coverRaw,
			"createtime":      formatTimestamp(atoi64(str(row["ctimestamp"]))),
			"updatetime":      formatTimestamp(atoi64(str(row["utimestamp"]))),
			"vodkey":          str(row["vodkey"]),
			"scorenum":        scoreString(row["scorenum"]),
			"upnum":           str(row["upnum"]),
			"downnum":         str(row["downnum"]),
			"authorid":        str(row["authorid"]),
			"author":          str(row["author"]),
			"play_url":        playURL,
			"down_url":        downURL,
			"preview_url":     previewURL,
			"definition":      str(row["definition"]),
			"duration":        formatDuration(atoi(str(row["duration"]))),
			"yearid":          str(row["yearid"]),
			"yearname":        lookup(years, str(row["yearid"]), "yearname"),
			"mosaic":          str(row["mosaic"]),
			"portrait":        str(row["portrait"]),
			"view_price":      atoi(str(row["view_price"])),
			"vip_price":       float64(atoi(str(row["view_price"]))*vipDiscount) / 100,
			"limit_free":      limitFree(now, atoi64(str(row["free_sdate"])), atoi64(str(row["free_edate"]))),
			"need_buy":        0,
			"isvip":           str(row["isvip"]),
			"islimit":         str(row["islimit"]),
			"islimitv3":       str(row["islimitv3"]),
			"exclusive":       boolInt(str(row["prop4"]) == "1"),
			"commentcount":    str(row["commentcount"]),
			"playcount_total": countValue(atoi(str(row["playcount_total"]))),
			"downcount_total": countValue(atoi(str(row["downcount_total"]))),
			"tags":            resolveTags(str(row["tags"]), tags),
			"actor_tags":      resolveTags(str(row["actor_tags"]), tags),
			"areaid":          str(row["areaid"]),
			"areaname":        lookup(areas, str(row["areaid"]), "areaname"),
			"cateid":          str(row["cateid"]),
			"catename":        lookup(categories, str(row["cateid"]), "catename"),
			"playlist":        playLists(str(row["playlist"])),
			"downlist":        playLists(str(row["downlist"])),
			"episode_total":   str(row["episode_total"]),
			"episode_status":  str(row["episode_status"]),
			"playindex":       0,
		}
		result = append(result, item)
	}
	return result
}

func descendantCategoryIDs(categories []map[string]interface{}, parent int) []int {
	if parent <= 0 {
		return nil
	}
	children := map[int][]int{}
	for _, row := range categories {
		pid := atoi(str(row["parentid"]))
		id := atoi(str(row["cateid"]))
		children[pid] = append(children[pid], id)
	}
	result := []int{parent}
	var walk func(int)
	walk = func(id int) {
		for _, child := range children[id] {
			result = append(result, child)
			walk(child)
		}
	}
	walk(parent)
	return result
}

func collectTagNames(rows []map[string]interface{}) []string {
	names := []string{}
	for _, row := range rows {
		for _, field := range []string{"tags", "actor_tags"} {
			for _, name := range strings.Split(str(row[field]), ",") {
				name = strings.TrimSpace(name)
				if name != "" {
					names = append(names, name)
				}
			}
		}
	}
	return names
}

func collectFieldTagNames(rows []map[string]interface{}, fields ...string) []string {
	names := []string{}
	for _, row := range rows {
		for _, field := range fields {
			for _, name := range strings.Split(str(row[field]), ",") {
				name = strings.TrimSpace(name)
				if name != "" {
					names = append(names, name)
				}
			}
		}
	}
	return names
}

func tagIDs(tags []map[string]interface{}) []int {
	ids := []int{}
	seen := map[int]struct{}{}
	for _, tag := range tags {
		id := atoi(str(tag["tagid"]))
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func categoryParents(categories []map[string]interface{}, cateID int) []map[string]interface{} {
	byID := indexBy(categories, "cateid")
	stack := []map[string]interface{}{}
	for cateID > 0 {
		row, ok := byID[strconv.Itoa(cateID)]
		if !ok {
			break
		}
		stack = append(stack, map[string]interface{}{
			"cateid":    str(row["cateid"]),
			"catename":  str(row["catename"]),
			"itemcount": row["itemcount"],
		})
		cateID = atoi(str(row["parentid"]))
	}
	for left, right := 0, len(stack)-1; left < right; left, right = left+1, right-1 {
		stack[left], stack[right] = stack[right], stack[left]
	}
	return stack
}

func resolveTags(value string, tags map[string]map[string]interface{}) []map[string]interface{} {
	result := []map[string]interface{}{}
	if value == "" {
		return result
	}
	for _, name := range strings.Split(value, ",") {
		if tag, ok := tags[strings.TrimSpace(name)]; ok {
			result = append(result, map[string]interface{}{
				"tagid":     str(tag["tagid"]),
				"tagtype":   str(tag["tagtype"]),
				"tagname":   str(tag["tagname"]),
				"itemcount": str(tag["itemcount"]),
			})
		}
	}
	return result
}

func pageInfo(total int, pageSize int, page int, pageURL string) map[string]interface{} {
	if total < 0 {
		total = 0
	}
	if pageSize < 1 {
		pageSize = 1
	}
	totalPage := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPage < 1 {
		totalPage = 1
	}
	if page < 1 {
		page = 1
	}
	if page > totalPage {
		page = totalPage
	}
	start := 0
	if total > 0 {
		start = (page-1)*pageSize + 1
	}
	end := start + pageSize - 1
	if end > total {
		end = total
	}
	return map[string]interface{}{
		"plist":     plist(page, totalPage, pageURL),
		"pagesize":  pageSize,
		"total":     total,
		"totalpage": totalPage,
		"page":      page,
		"start":     start,
		"end":       end,
		"prev":      ternary(page > 1, page-1, 0),
		"next":      ternary(page < totalPage, page+1, 0),
		"curr_url":  strings.ReplaceAll(pageURL, "[?]", strconv.Itoa(page)),
		"first_url": strings.ReplaceAll(pageURL, "[?]", "1"),
		"prev_url":  strings.ReplaceAll(pageURL, "[?]", strconv.Itoa(ternary(page > 1, page-1, 1))),
		"next_url":  strings.ReplaceAll(pageURL, "[?]", strconv.Itoa(ternary(page < totalPage, page+1, totalPage))),
		"last_url":  strings.ReplaceAll(pageURL, "[?]", strconv.Itoa(totalPage)),
		"page_url":  pageURL,
		"pages":     pageSelector(page, totalPage),
	}
}

func PageInfo(total int, pageSize int, page int, pageURL string) map[string]interface{} {
	return pageInfo(total, pageSize, page, pageURL)
}

func plist(page int, totalPage int, pageURL string) []map[string]interface{} {
	pages := []int{}
	for i := page - 5; i < page; i++ {
		if i > 0 {
			pages = append(pages, i)
		}
	}
	pages = append(pages, page)
	for i := page + 1; i <= totalPage && len(pages) < 10; i++ {
		pages = append(pages, i)
	}
	for i := pages[0] - 1; i > 0 && len(pages) < 10; i-- {
		pages = append([]int{i}, pages...)
	}
	result := []map[string]interface{}{}
	if page-5 > 1 {
		result = append(result, pageLink("first", 1, "FirstPage", pageURL))
		if page-5 > 2 {
			result = append(result, pageLink("more", 0, "...", ""))
		}
	}
	if len(pages) > 0 && pages[0] > 1 {
		result = append(result, pageLink("prev", page-1, "PrevPage", pageURL))
	}
	for _, p := range pages {
		pos := ""
		if p == page {
			pos = "curr"
		}
		result = append(result, pageLink(pos, p, p, pageURL))
	}
	if len(pages) > 0 && pages[len(pages)-1] < totalPage {
		result = append(result, pageLink("next", page+1, "NextPage", pageURL))
	}
	if totalPage-page > 4 {
		if totalPage-page > 5 {
			result = append(result, pageLink("more", 0, "...", ""))
		}
		result = append(result, pageLink("last", totalPage, "LastPage", pageURL))
	}
	return result
}

func pageLink(pos string, page int, text interface{}, pageURL string) map[string]interface{} {
	urlValue := ""
	if pageURL != "" {
		urlValue = strings.ReplaceAll(pageURL, "[?]", strconv.Itoa(page))
	}
	return map[string]interface{}{"pos": pos, "page": page, "text": text, "url": urlValue}
}

func pageSelector(pageNow int, totalPage int) []int {
	showAll := 50
	sliceStart := 5
	sliceEnd := 5
	percent := 20
	rangeSize := 10
	if totalPage < showAll {
		pages := make([]int, 0, totalPage)
		for i := 1; i <= totalPage; i++ {
			pages = append(pages, i)
		}
		return pages
	}
	pages := []int{}
	for i := 1; i <= sliceStart; i++ {
		pages = append(pages, i)
	}
	for i := totalPage - sliceEnd; i <= totalPage; i++ {
		pages = append(pages, i)
	}

	increment := int(math.Floor(float64(totalPage) / float64(percent)))
	if increment < 1 {
		increment = 1
	}
	pageNowMinusRange := pageNow - rangeSize
	pageNowPlusRange := pageNow + rangeSize
	i := sliceStart
	x := totalPage - sliceEnd
	metBoundary := false
	for i <= x {
		if i >= pageNowMinusRange && i <= pageNowPlusRange {
			i++
			metBoundary = true
		} else {
			i += increment
			if i > pageNowMinusRange && !metBoundary {
				i = pageNowMinusRange
			}
		}
		if i > 0 && i <= x {
			pages = append(pages, i)
		}
	}

	i = pageNow
	dist := 1
	for i < x {
		dist *= 2
		i = pageNow + dist
		if i > 0 && i <= x {
			pages = append(pages, i)
		}
	}

	i = pageNow
	dist = 1
	for i > 0 {
		dist *= 2
		i = pageNow - dist
		if i > 0 && i <= x {
			pages = append(pages, i)
		}
	}

	sort.Ints(pages)
	unique := pages[:0]
	var last int
	for idx, page := range pages {
		if idx == 0 || page != last {
			unique = append(unique, page)
			last = page
		}
	}
	return unique
}

func (s *ListingService) coverPic(uri string, srvid int, servers []map[string]interface{}, isH5 bool, values ...resourceurl.Resolved) string {
	resolved := resourceurl.Resolved{BaseURL: s.resourceBaseURL, Timestamp: s.now().Unix()}
	if len(values) > 0 {
		resolved = values[0]
	}
	if uri == "" {
		return ""
	}
	if !isH5 && srvid > 0 {
		if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
			return uri
		}
		for _, server := range servers {
			if atoi(str(server["srvid"])) == srvid {
				return strings.TrimRight(str(server["srvhost"]), "/") + "/" + strings.TrimLeft(uri, "/")
			}
		}
	}
	return resolved.GetRes(uri, "")
}

func (s *ListingService) withResources(ctx context.Context, env map[string]interface{}) map[string]interface{} {
	resolved := resourceurl.Resolved{BaseURL: s.resourceBaseURL, Timestamp: s.now().Unix()}
	if s.resources != nil {
		resolved = resourceurl.Resolved{Timestamp: s.now().Unix()}
		if value, err := s.resources.ResolveContext(ctx); err == nil {
			resolved = value
		}
	}
	env["resources"] = resolved
	return env
}

func playLists(value string) []map[string]interface{} {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.TrimSpace(value)
	if value == "" {
		return []map[string]interface{}{}
	}
	result := []map[string]interface{}{}
	for index, line := range strings.Split(value, "\n") {
		parts := strings.Split(line, "$")
		for len(parts) < 3 {
			parts = append(parts, "")
		}
		result = append(result, map[string]interface{}{
			"playindex":  index + 1,
			"play_name":  parts[0],
			"view_price": ternary(isNumeric(parts[2]), atoi(parts[2]), -1),
		})
	}
	return result
}

func optionRows(items [][2]interface{}) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		rows = append(rows, map[string]interface{}{"keyid": item[0], "value": item[1]})
	}
	return rows
}

func indexBy(rows []map[string]interface{}, key string) map[string]map[string]interface{} {
	result := map[string]map[string]interface{}{}
	for _, row := range rows {
		result[str(row[key])] = row
	}
	return result
}

func groupBy(rows []map[string]interface{}, key string) map[string][]map[string]interface{} {
	result := map[string][]map[string]interface{}{}
	for _, row := range rows {
		group := str(row[key])
		result[group] = append(result[group], row)
	}
	return result
}

func formatTimestamp(ts int64) string {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return time.Unix(ts, 0).In(loc).Format("2006-01-02 15:04:05")
}

func formatDuration(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%02d:%02d", m, s)
	}
	return fmt.Sprintf("%d", s)
}

func countValue(value int) interface{} {
	if value < 1000 {
		return strconv.Itoa(value)
	}
	return int(math.Round(float64(value) / 10000))
}

func cleanCover(value string) string {
	replacer := strings.NewReplacer("\n", "", "\r", "", "\t", "", "'", "", "\"", "")
	return replacer.Replace(value)
}

func lookup(rows map[string]map[string]interface{}, id string, key string) interface{} {
	if row, ok := rows[id]; ok {
		return str(row[key])
	}
	return nil
}

func limitFree(now int64, start int64, end int64) int {
	if now > start && now < end {
		return 1
	}
	return 0
}

func boolInt(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func scoreString(value interface{}) string {
	raw := strings.TrimSpace(str(value))
	if raw == "" {
		return ""
	}
	if strings.Contains(raw, ".") {
		return raw
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return raw
	}
	return strconv.FormatFloat(parsed, 'f', 1, 64)
}

func atoi(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}

func atoi64(value string) int64 {
	parsed, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return parsed
}

func decodeCalldataJSON(row map[string]interface{}) interface{} {
	if str(row["type"]) != "json" {
		return nil
	}
	content := str(row["content"])
	if content == "" {
		return nil
	}
	var decoded interface{}
	if err := json.Unmarshal([]byte(content), &decoded); err != nil {
		return nil
	}
	return decoded
}

func (s *ListingService) searchLikeTags(ctx context.Context, freeOnly bool) (interface{}, error) {
	vodIDs, err := s.store.TopSearchVODIDs(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.store.VODsByIDs(ctx, splitIDs(vodIDs), "vodid DESC")
	if err != nil {
		return nil, err
	}
	if freeOnly {
		filtered := rows[:0]
		for _, row := range rows {
			if str(row["isvip"]) == "0" {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	tags := collectTagNames(rows)
	if len(tags) > 10 {
		return tags[:10], nil
	}
	if len(tags) == 0 {
		return nil, nil
	}
	return nil, nil
}

func (s *ListingService) miniSearchLikeTags(ctx context.Context) (interface{}, error) {
	vodIDs, err := s.store.TopMiniSearchVODIDs(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.store.MiniVODsByIDs(ctx, splitIDs(vodIDs), "vodid DESC")
	if err != nil {
		return nil, err
	}
	tags := collectTagNames(rows)
	if len(tags) > 10 {
		return tags[:10], nil
	}
	if len(tags) == 0 {
		return nil, nil
	}
	return nil, nil
}

func wrapVODRows(vodRows []map[string]interface{}) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(vodRows))
	for _, vodRow := range vodRows {
		rows = append(rows, map[string]interface{}{"vodrow": vodRow})
	}
	return rows
}

func rowIDs(rows []map[string]interface{}, key string) []int {
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		id := atoi(str(row[key]))
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

func pageSlice(rows []map[string]interface{}, page int, pageSize int) []map[string]interface{} {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	totalPage := int(math.Ceil(float64(len(rows)) / float64(pageSize)))
	if totalPage < 1 {
		totalPage = 1
	}
	if page > totalPage {
		return []map[string]interface{}{}
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end]
}

func pageSliceInts(ids []int, page int, pageSize int) []int {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	totalPage := int(math.Ceil(float64(len(ids)) / float64(pageSize)))
	if totalPage < 1 {
		totalPage = 1
	}
	if page > totalPage {
		return []int{}
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(ids) {
		end = len(ids)
	}
	return ids[start:end]
}

func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	_, err := strconv.Atoi(value)
	return err == nil
}

func ternary[T any](ok bool, yes T, no T) T {
	if ok {
		return yes
	}
	return no
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
