package comment

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"xj_comp/internal/domain"
)

const sampleParams = "$vodid:0-$orderby:0-$page:1"

var paramKeys = []string{"vodid", "orderby", "page"}

var ErrVODNotFound = errors.New("vod not found")

type Store interface {
	VODByID(ctx context.Context, vodID int) (map[string]interface{}, error)
	UserGroups(ctx context.Context) ([]map[string]interface{}, error)
	CountRoots(ctx context.Context, vodID int) (int, error)
	RootComments(ctx context.Context, vodID int, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error)
}

type Service struct {
	store           Store
	resourceBaseURL string
	now             func() time.Time
}

type ListingRequest struct {
	PathParams string
	QueryPage  string
}

func NewService(store Store, resourceBaseURL string) *Service {
	return &Service{
		store:           store,
		resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"),
		now:             time.Now,
	}
}

func (s *Service) Listing(ctx context.Context, req ListingRequest) (domain.CommentListingData, error) {
	params := parseParams(req.PathParams)
	if atoi(params["page"]) == 0 {
		params["page"] = req.QueryPage
	}
	vodID := atoi(params["vodid"])
	vod, err := s.store.VODByID(ctx, vodID)
	if err != nil {
		return domain.CommentListingData{}, fmt.Errorf("get vod: %w", err)
	}
	if len(vod) == 0 || atoi(str(vod["showtype"])) > 1 {
		return domain.CommentListingData{}, ErrVODNotFound
	}

	pageSize := 20
	page := atoi(params["page"])
	orderBy := "a.addtime DESC"
	if atoi(params["orderby"]) == 1 {
		orderBy = "a.upnum DESC"
	}
	total, err := s.store.CountRoots(ctx, vodID)
	if err != nil {
		return domain.CommentListingData{}, fmt.Errorf("count comments: %w", err)
	}
	rows, err := s.store.RootComments(ctx, vodID, total, page, pageSize, orderBy)
	if err != nil {
		return domain.CommentListingData{}, fmt.Errorf("list comments: %w", err)
	}
	groups, err := s.store.UserGroups(ctx)
	if err != nil {
		return domain.CommentListingData{}, fmt.Errorf("list user groups: %w", err)
	}

	pageURL := "/comment/listing-" + buildParams(params, map[string]string{"page": "[?]"})
	return domain.CommentListingData{
		Rows:     s.processRows(rows, groups),
		PageInfo: pageInfo(total, pageSize, page, pageURL),
	}, nil
}

func (s *Service) processRows(rows []map[string]interface{}, groups []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		item := s.processRow(row, groups)
		subrows, _ := row["subrows"].([]map[string]interface{})
		item["subrows"] = []map[string]interface{}{}
		for _, subrow := range subrows {
			item["subrows"] = append(item["subrows"].([]map[string]interface{}), s.processRow(subrow, groups))
		}
		out = append(out, item)
	}
	return out
}

func (s *Service) processRow(row map[string]interface{}, groups []map[string]interface{}) map[string]interface{} {
	now := s.now().Unix()
	item := map[string]interface{}{
		"id":           str(row["id"]),
		"rootid":       str(row["rootid"]),
		"parentid":     str(row["parentid"]),
		"lft":          str(row["lft"]),
		"rgt":          str(row["rgt"]),
		"depth":        str(row["depth"]),
		"vodid":        str(row["vodid"]),
		"uid":          str(row["uid"]),
		"sid":          str(row["sid"]),
		"username":     str(row["username"]),
		"nickname":     str(row["nickname"]),
		"gender":       atoi(str(row["gender"])),
		"gicon":        gicon(row, groups),
		"isvip":        vip(row, now),
		"content":      str(row["content"]),
		"upnum":        str(row["upnum"]),
		"downnum":      str(row["downnum"]),
		"avatar_url":   s.avatarURL(str(row["avatar"])),
		"addtime":      commentTime(atoi64(str(row["addtime"])), now),
		"__closenum__": atoi(str(row["__closenum__"])),
	}
	if atoi(str(row["showtype"])) != 0 {
		item["username"] = "???"
		item["nickname"] = "???"
		item["content"] = "评论审核中..."
		item["avatar_url"] = s.avatarURL("")
	}
	return item
}

func (s *Service) avatarURL(avatar string) string {
	if avatar == "" {
		return s.resourceBaseURL + "/sysavatar/noavatar.png"
	}
	if strings.HasPrefix(avatar, "http://") || strings.HasPrefix(avatar, "https://") {
		return avatar
	}
	if strings.HasPrefix(avatar, "sysavatar/") {
		return s.resourceBaseURL + "/" + strings.TrimLeft(avatar, "/")
	}
	return s.resourceBaseURL + "/C1/avatar/" + strings.TrimLeft(avatar, "/")
}

func parseParams(raw string) map[string]string {
	params := map[string]string{}
	defaults := []string{"0", "0", "1"}
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

func pageInfo(total int, pageSize int, page int, pageURL string) map[string]interface{} {
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
	if totalPage < 50 {
		pages := make([]int, 0, totalPage)
		for i := 1; i <= totalPage; i++ {
			pages = append(pages, i)
		}
		return pages
	}
	pages := []int{}
	for i := 1; i <= 5; i++ {
		pages = append(pages, i)
	}
	for i := totalPage - 5; i <= totalPage; i++ {
		pages = append(pages, i)
	}
	increment := int(math.Floor(float64(totalPage) / 20))
	if increment < 1 {
		increment = 1
	}
	minRange := pageNow - 10
	maxRange := pageNow + 10
	i := 5
	x := totalPage - 5
	metBoundary := false
	for i <= x {
		if i >= minRange && i <= maxRange {
			i++
			metBoundary = true
		} else {
			i += increment
			if i > minRange && !metBoundary {
				i = minRange
			}
		}
		if i > 0 && i <= x {
			pages = append(pages, i)
		}
	}
	sort.Ints(pages)
	unique := pages[:0]
	var last int
	for idx, p := range pages {
		if idx == 0 || p != last {
			unique = append(unique, p)
			last = p
		}
	}
	return unique
}

func commentTime(addTime int64, now int64) string {
	if addTime > now-86400*30 {
		return relativeTime(now - addTime)
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return time.Unix(addTime, 0).In(loc).Format("2006-01-02")
}

func relativeTime(seconds int64) string {
	if seconds < 0 {
		seconds = 0
	}
	d := seconds / 86400
	seconds -= d * 86400
	h := seconds / 3600
	seconds -= h * 3600
	m := seconds / 60
	seconds -= m * 60
	switch {
	case d > 0:
		return fmt.Sprintf("%d天前 ", d)
	case h > 0:
		return fmt.Sprintf("%d小时前 ", h)
	case m > 0:
		return fmt.Sprintf("%d分钟前 ", m)
	default:
		return fmt.Sprintf("%d秒前", seconds)
	}
}

func gicon(row map[string]interface{}, groups []map[string]interface{}) string {
	gid := atoi(str(row["gid"]))
	if atoi(str(row["sysgid"])) > 0 {
		gid = atoi(str(row["sysgid"]))
	}
	for _, group := range groups {
		if atoi(str(group["gid"])) == gid {
			return str(group["gicon"])
		}
	}
	return ""
}

func vip(row map[string]interface{}, now int64) int {
	if atoi(str(row["sysgid"])) == 6 && atoi64(str(row["sysgid_exptime"])) > now {
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

func atoi(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}

func atoi64(value string) int64 {
	parsed, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return parsed
}

func ternary[T any](ok bool, yes T, no T) T {
	if ok {
		return yes
	}
	return no
}
