package vod

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
	vodRepo "xj_comp/internal/repository/vod"
)

const specialSampleParams = "$sptype:0-$orderby:0-$page:1"

var specialParamKeys = []string{"sptype", "orderby", "page"}

var ErrSpecialNotFound = errors.New("special not found")

type SpecialStore interface {
	Categories(ctx context.Context) ([]map[string]interface{}, error)
	Areas(ctx context.Context) ([]map[string]interface{}, error)
	Years(ctx context.Context) ([]map[string]interface{}, error)
	Servers(ctx context.Context) ([]map[string]interface{}, error)
	TagsByNames(ctx context.Context, names []string) ([]map[string]interface{}, error)
	CountSpecials(ctx context.Context, filter vodRepo.SpecialFilter) (int, error)
	ListSpecials(ctx context.Context, filter vodRepo.SpecialFilter, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error)
	ListActorSpecials(ctx context.Context, pageSize int) ([]map[string]interface{}, error)
	SpecialByID(ctx context.Context, spid int) (map[string]interface{}, error)
	VODsByIDs(ctx context.Context, ids []int, orderBy string) ([]map[string]interface{}, error)
	UpdateSpecialRand(ctx context.Context, minSPID int) error
	IncrementSpecialViews(ctx context.Context, spid int, viewsLastTime int64, now int64) error
	GuestBySID(ctx context.Context, sid string) (map[string]interface{}, error)
	KeylimitCount(ctx context.Context, key string) (int, error)
	SetKeylimit(ctx context.Context, key string, keynum int, keydata string, now int64) error
	IncrementSpecialVote(ctx context.Context, spid int, field string) error
}

type SpecialAuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type SpecialService struct {
	store     SpecialStore
	authStore SpecialAuthStore
	processor *ListingService
}

type SpecialListingRequest struct {
	PathParams  string
	QueryPage   string
	IsH5Request bool
}

type SpecialDetailRequest struct {
	SPID        int
	PathParams  string
	IsH5Request bool
}

func NewSpecialService(store SpecialStore, authStore SpecialAuthStore, resourceBaseURL string, vipDiscount int) *SpecialService {
	return &SpecialService{
		store:     store,
		authStore: authStore,
		processor: NewListingService(nil, resourceBaseURL, vipDiscount),
	}
}

func (s *SpecialService) Listing(ctx context.Context, req SpecialListingRequest) (domain.SpecialListingData, error) {
	params := parseSpecialParams(req.PathParams)
	pageFromInput := false
	if atoi(params["page"]) == 0 {
		pageFromInput = true
		params["page"] = req.QueryPage
		if params["page"] == "" {
			params["page"] = "0"
		}
	}

	filter := vodRepo.SpecialFilter{SPType: atoi(params["sptype"])}
	pageSize := 16
	page := atoi(params["page"])
	total, err := s.store.CountSpecials(ctx, filter)
	if err != nil {
		return domain.SpecialListingData{}, fmt.Errorf("count specials: %w", err)
	}
	rows, err := s.store.ListSpecials(ctx, filter, total, page, pageSize, specialOrderBy(atoi(params["orderby"])))
	if err != nil {
		return domain.SpecialListingData{}, fmt.Errorf("list specials: %w", err)
	}

	rowsWithVODs, err := s.attachListingVODRows(ctx, rows, req.IsH5Request)
	if err != nil {
		return domain.SpecialListingData{}, err
	}
	specialRows := s.processSpecialRows(rowsWithVODs)

	pageinfo := pageInfo(total, pageSize, page, "/special/listing-"+buildSpecialParams(params, map[string]string{"page": "[?]"}))
	actorRows := []map[string]interface{}{}
	if atoi(fmt.Sprint(pageinfo["page"])) == 1 {
		actors, err := s.store.ListActorSpecials(ctx, 100)
		if err != nil {
			return domain.SpecialListingData{}, fmt.Errorf("list actor specials: %w", err)
		}
		actorRows = s.processSpecialRows(ensureSpecialVODRows(actors))
	}
	if atoi(params["orderby"]) == 3 {
		if err := s.store.UpdateSpecialRand(ctx, minSpecialID(rows)); err != nil {
			return domain.SpecialListingData{}, err
		}
	}

	return domain.SpecialListingData{
		Rows:         specialRows,
		PageInfo:     pageinfo,
		SampleParams: specialSampleParams,
		Params:       specialParamsData(params, pageFromInput),
		ActorRows:    actorRows,
	}, nil
}

func (s *SpecialService) Detail(ctx context.Context, req SpecialDetailRequest) (domain.SpecialDetailData, error) {
	row, err := s.store.SpecialByID(ctx, req.SPID)
	if err != nil {
		return domain.SpecialDetailData{}, fmt.Errorf("get special: %w", err)
	}
	if len(row) == 0 || atoi(fmt.Sprint(row["showtype"])) != 0 {
		return domain.SpecialDetailData{}, ErrSpecialNotFound
	}

	params := parseSpecialDetailParams(req.PathParams)
	order := atoi(params["orderby"])
	ids := splitIDs(fmt.Sprint(row["vodids"]))
	vodRows, err := s.specialVODRows(ctx, ids, specialDetailOrderBy(order), req.IsH5Request)
	if err != nil {
		return domain.SpecialDetailData{}, err
	}
	if order == 0 {
		vodRows = orderVODRowsByIDs(vodRows, ids)
	}
	if err := s.store.IncrementSpecialViews(ctx, atoi(fmt.Sprint(row["spid"])), atoi64(fmt.Sprint(row["views_lasttime"])), s.processor.now().Unix()); err != nil {
		return domain.SpecialDetailData{}, err
	}

	row = cloneMap(row)
	return domain.SpecialDetailData{
		Row:     s.processSpecialRows([]map[string]interface{}{row})[0],
		VODRows: vodRows,
	}, nil
}

func (s *SpecialService) Vote(ctx context.Context, token string, spid int, action string) (int, string, error) {
	row, err := s.store.SpecialByID(ctx, spid)
	if err != nil {
		return -1, "记录不存在或已被删除", fmt.Errorf("get special: %w", err)
	}
	if len(row) == 0 || atoi(fmt.Sprint(row["showtype"])) != 0 {
		return -1, "记录不存在或已被删除", ErrSpecialNotFound
	}

	user, sid, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -1, "获取用户失败", err
	}
	uid := atoi(fmt.Sprint(user["uid"]))
	voter := fmt.Sprint(uid)
	if uid == 0 {
		guest, err := s.store.GuestBySID(ctx, sid)
		if err != nil {
			return -1, "获取游客失败", err
		}
		if len(guest) == 0 {
			return -9999, "请登录后操作，客户端游客请先携带信息", nil
		}
		voter = sid
	}

	key := fmt.Sprintf("special.updown.%v.%s", row["spid"], voter)
	count, err := s.store.KeylimitCount(ctx, key)
	if err != nil {
		return -1, "检查点赞状态失败", err
	}
	if count > 0 {
		return -1, "您已经赞/踩过了", nil
	}

	field := "upnum"
	if action == "down" {
		field = "downnum"
	}
	if err := s.store.IncrementSpecialVote(ctx, spid, field); err != nil {
		return -1, "点赞失败", err
	}
	if err := s.store.SetKeylimit(ctx, key, 1, "", s.processor.now().Unix()); err != nil {
		return -1, "点赞失败", err
	}
	return 0, "已赞", nil
}

func (s *SpecialService) authenticatedUser(ctx context.Context, token string) (map[string]interface{}, string, error) {
	sid := userRepo.CleanToken(token)
	if s.authStore == nil {
		return map[string]interface{}{"uid": "0", "sid": sid}, sid, nil
	}
	user, err := s.authStore.UserBySession(ctx, sid)
	if err != nil {
		return nil, sid, err
	}
	if user == nil {
		return map[string]interface{}{"uid": "0", "sid": sid}, sid, nil
	}
	return user, fmt.Sprint(user["sid"]), nil
}

func (s *SpecialService) attachListingVODRows(ctx context.Context, rows []map[string]interface{}, isH5 bool) ([]map[string]interface{}, error) {
	out := make([]map[string]interface{}, 0, len(rows))
	allIDs := []int{}
	rowIDs := make([][]int, 0, len(rows))
	for _, row := range rows {
		ids := splitIDs(fmt.Sprint(row["vodids"]))
		if len(ids) > 4 {
			ids = ids[:4]
		}
		rowIDs = append(rowIDs, ids)
		allIDs = append(allIDs, ids...)
	}
	vodRows, err := s.specialVODRows(ctx, allIDs, "", isH5)
	if err != nil {
		return nil, err
	}
	vodByID := map[string]map[string]interface{}{}
	for _, row := range vodRows {
		vodByID[fmt.Sprint(row["vodid"])] = row
	}
	for index, row := range rows {
		item := cloneMap(row)
		item["vodrows"] = []map[string]interface{}{}
		for _, id := range rowIDs[index] {
			if vod, ok := vodByID[fmt.Sprint(id)]; ok {
				item["vodrows"] = append(item["vodrows"].([]map[string]interface{}), vod)
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *SpecialService) specialVODRows(ctx context.Context, ids []int, orderBy string, isH5 bool) ([]map[string]interface{}, error) {
	ids = uniquePositive(ids)
	if len(ids) == 0 {
		return []map[string]interface{}{}, nil
	}
	categories, areas, years, servers, err := s.listMetadata(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.store.VODsByIDs(ctx, ids, orderBy)
	if err != nil {
		return nil, fmt.Errorf("list special vods: %w", err)
	}
	tagRows, err := s.store.TagsByNames(ctx, collectTagNames(rows))
	if err != nil {
		return nil, fmt.Errorf("list special tags: %w", err)
	}
	return s.processor.processRows(rows, processEnv(categories, areas, years, servers, tagRows), isH5, s.processor.now().Unix()), nil
}

func (s *SpecialService) listMetadata(ctx context.Context) ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, []map[string]interface{}, error) {
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

func (s *SpecialService) processSpecialRows(rows []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"spid":       fmt.Sprint(row["spid"]),
			"sptype":     fmt.Sprint(row["sptype"]),
			"spname":     fmt.Sprint(row["spname"]),
			"intro":      fmt.Sprint(row["intro"]),
			"coverpic":   s.resourceURL(fmt.Sprint(row["coverpic"])),
			"coverx":     fmt.Sprint(row["coverpic"]),
			"avatar":     s.resourceURL(fmt.Sprint(row["avatar"])),
			"avatarx":    fmt.Sprint(row["avatar"]),
			"cup":        fmt.Sprint(row["cup"]),
			"age":        fmt.Sprint(row["age"]),
			"upnum":      fmt.Sprint(row["upnum"]),
			"downnum":    fmt.Sprint(row["downnum"]),
			"addtime":    formatTimestamp(atoi64(fmt.Sprint(row["addtime"]))),
			"updatetime": formatTimestamp(atoi64(fmt.Sprint(row["updatetime"]))),
			"itemcount":  fmt.Sprint(row["itemcount"]),
			"vodrows":    specialVODRowsValue(row["vodrows"]),
		})
	}
	return out
}

func (s *SpecialService) resourceURL(uri string) string {
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return ""
	}
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return uri
	}
	return strings.TrimRight(s.processor.resourceBaseURL, "/") + "/" + strings.TrimLeft(uri, "/")
}

func parseSpecialParams(raw string) map[string]string {
	return parseFixedParams(raw, specialParamKeys, []string{"0", "0", "1"})
}

func parseSpecialDetailParams(raw string) map[string]string {
	return parseFixedParams(raw, []string{"orderby"}, []string{"0"})
}

func parseFixedParams(raw string, keys []string, defaults []string) map[string]string {
	params := map[string]string{}
	values := []string{}
	if raw != "" {
		values = strings.Split(raw, "-")
	}
	for i, key := range keys {
		value := defaults[i]
		if i < len(values) && values[i] != "" {
			value = values[i]
		}
		params[key] = value
	}
	return params
}

func buildSpecialParams(params map[string]string, replace map[string]string) string {
	values := make([]string, 0, len(specialParamKeys))
	for _, key := range specialParamKeys {
		value := params[key]
		if next, ok := replace[key]; ok {
			value = next
		}
		values = append(values, value)
	}
	return strings.Join(values, "-")
}

func specialParamsData(params map[string]string, pageFromInput bool) map[string]interface{} {
	var page interface{} = params["page"]
	if pageFromInput {
		page = atoi(params["page"])
	}
	return map[string]interface{}{
		"sptype":  params["sptype"],
		"orderby": params["orderby"],
		"page":    page,
	}
}

func specialOrderBy(order int) string {
	switch order {
	case 1:
		return "addtime DESC"
	case 2:
		return "addtime ASC"
	case 3:
		return "randnum ASC"
	default:
		return "updatetime DESC"
	}
}

func specialDetailOrderBy(order int) string {
	switch order {
	case 1:
		return "playcount_total DESC"
	case 2:
		return "playcount_total ASC"
	case 3:
		return "vodid ASC"
	default:
		return "vodid DESC"
	}
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

func uniquePositive(ids []int) []int {
	seen := map[int]struct{}{}
	out := []int{}
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func orderVODRowsByIDs(rows []map[string]interface{}, ids []int) []map[string]interface{} {
	positions := map[string]int{}
	for index, id := range ids {
		if _, ok := positions[fmt.Sprint(id)]; !ok {
			positions[fmt.Sprint(id)] = index
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return positions[fmt.Sprint(rows[i]["vodid"])] < positions[fmt.Sprint(rows[j]["vodid"])]
	})
	return rows
}

func ensureSpecialVODRows(rows []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, cloneMap(row))
	}
	return out
}

func minSpecialID(rows []map[string]interface{}) int {
	minID := 0
	for _, row := range rows {
		id := atoi(fmt.Sprint(row["spid"]))
		if id < minID {
			minID = id
		}
	}
	return minID
}

func cloneMap(row map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(row))
	for key, value := range row {
		out[key] = value
	}
	return out
}

func specialVODRowsValue(value interface{}) []map[string]interface{} {
	if rows, ok := value.([]map[string]interface{}); ok {
		return rows
	}
	return nil
}
