package minivod

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"xj_comp/internal/domain"
	minivodRepo "xj_comp/internal/repository/minivod"
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
}

type VODProcessor interface {
	ProcessRowsFullPrice(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error)
	ProcessMiniRowsFullPrice(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error)
}

var (
	ErrVODNotFound    = errors.New("minivod not found")
	ErrAuthorNotFound = errors.New("minivod author not found")
)

type Service struct {
	store           Store
	vodProcessor    VODProcessor
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
	return &Service{store: store, vodProcessor: vodProcessor, now: time.Now, resourceBaseURL: strings.TrimRight(resourceBaseURL, "/")}
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
