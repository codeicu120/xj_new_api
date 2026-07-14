package ucp

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
)

type UserStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
	Groups(ctx context.Context) ([]map[string]interface{}, error)
	CountRecommended(ctx context.Context, uid int) (int, error)
	RecommendedUsers(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error)
}

type Service struct {
	store           UserStore
	resourceBaseURL string
	now             func() time.Time
}

func NewService(store UserStore, resourceBaseURL string) *Service {
	return &Service{
		store:           store,
		resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"),
		now:             time.Now,
	}
}

func (s *Service) MyAff(ctx context.Context, token string, page int) (domain.UCPMyAffData, int, string, error) {
	user, groups, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return domain.UCPMyAffData{}, -1, "获取用户失败", err
	}
	if atoi(user["uid"]) == 0 {
		return domain.UCPMyAffData{}, -9999, "请登录后操作", nil
	}

	pageSize := 20
	total, err := s.store.CountRecommended(ctx, atoi(user["uid"]))
	if err != nil {
		return domain.UCPMyAffData{}, -1, "获取推广列表失败", err
	}
	page = normalizePage(total, pageSize, page)
	rows, err := s.store.RecommendedUsers(ctx, atoi(user["uid"]), page, pageSize)
	if err != nil {
		return domain.UCPMyAffData{}, -1, "获取推广列表失败", err
	}

	return domain.UCPMyAffData{
		Rows:     s.processUsers(rows, groups),
		PageInfo: pageInfo(total, pageSize, page, "/ucp/myaff?page=[?]"),
	}, 0, "", nil
}

func (s *Service) authenticatedUser(ctx context.Context, token string) (map[string]interface{}, []map[string]interface{}, error) {
	groups, err := s.store.Groups(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list user groups: %w", err)
	}
	sid := userRepo.CleanToken(token)
	user, err := s.store.UserBySession(ctx, sid)
	if err != nil {
		return nil, nil, fmt.Errorf("load user by session: %w", err)
	}
	if user == nil {
		user = map[string]interface{}{"uid": "0", "sid": sid}
	}
	return user, groups, nil
}

func (s *Service) processUsers(rows []map[string]interface{}, groups []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	now := s.now().Unix()
	for _, row := range rows {
		sysgidExptime := atoi64(row["sysgid_exptime"])
		duetime := ""
		dueday := ""
		if sysgidExptime > 0 {
			duetime = formatUnix(sysgidExptime)
			if remaining := sysgidExptime - now; remaining > 0 {
				dueday = formatRemain(remaining) + "过期"
			} else {
				dueday = "已过期"
			}
		}
		out = append(out, map[string]interface{}{
			"uid":             str(row["uid"]),
			"uniqkey":         strings.ToUpper(strconv.FormatInt(int64(atoi(row["uniqkey"])), 36)),
			"username":        str(row["username"]),
			"nickname":        str(row["nickname"]),
			"mobi":            str(row["mobi"]),
			"email":           str(row["email"]),
			"sysgid":          str(row["sysgid"]),
			"gid":             str(row["gid"]),
			"gids":            nil,
			"gicon":           gicon(row, groups),
			"isvip":           vip(row, now),
			"regtime":         formatUnix(atoi64(row["regtime"])),
			"gender":          atoi(row["gender"]),
			"avatar":          str(row["avatar"]),
			"avatar_url":      s.avatarURL(str(row["avatar"])),
			"newmsg":          str(row["newmsg"]),
			"goldcoin":        atoi(row["goldcoin"]),
			"gold_bean":       atoi(row["gold_bean"]),
			"duetime":         duetime,
			"dueday":          dueday,
			"recommend_total": atoi(row["recommend_total"]),
		})
	}
	return out
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

func pageInfo(total int, pageSize int, page int, url string) map[string]interface{} {
	page = normalizePage(total, pageSize, page)
	totalPage := totalPages(total, pageSize)
	if pageSize < 1 {
		pageSize = 1
	}
	start := 0
	if total > 0 {
		start = (page-1)*pageSize + 1
	}
	end := start + pageSize - 1
	if end > total {
		end = total
	}
	pages := make([]int, totalPage)
	for i := range pages {
		pages[i] = i + 1
	}
	currURL := strings.ReplaceAll(url, "[?]", strconv.Itoa(page))
	firstURL := strings.ReplaceAll(url, "[?]", "1")
	prevPage := 0
	if page > 1 {
		prevPage = page - 1
	}
	nextPage := 0
	if page < totalPage {
		nextPage = page + 1
	}
	prevURLPage := 1
	if prevPage > 0 {
		prevURLPage = prevPage
	}
	nextURLPage := totalPage
	if nextPage > 0 {
		nextURLPage = nextPage
	}
	return map[string]interface{}{
		"plist": []map[string]interface{}{
			{"pos": "curr", "page": page, "text": page, "url": currURL},
		},
		"pagesize":  pageSize,
		"total":     total,
		"totalpage": totalPage,
		"page":      page,
		"start":     start,
		"end":       end,
		"prev":      prevPage,
		"next":      nextPage,
		"curr_url":  currURL,
		"first_url": firstURL,
		"prev_url":  strings.ReplaceAll(url, "[?]", strconv.Itoa(prevURLPage)),
		"next_url":  strings.ReplaceAll(url, "[?]", strconv.Itoa(nextURLPage)),
		"last_url":  strings.ReplaceAll(url, "[?]", strconv.Itoa(totalPage)),
		"page_url":  url,
		"pages":     pages,
	}
}

func normalizePage(total int, pageSize int, page int) int {
	totalPage := totalPages(total, pageSize)
	if page < 1 {
		page = 1
	}
	if page > totalPage {
		page = totalPage
	}
	return page
}

func totalPages(total int, pageSize int) int {
	if pageSize < 1 {
		pageSize = 1
	}
	totalPage := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPage < 1 {
		totalPage = 1
	}
	return totalPage
}

func gicon(row map[string]interface{}, groups []map[string]interface{}) string {
	gid := atoi(row["gid"])
	if atoi(row["sysgid"]) > 0 {
		gid = atoi(row["sysgid"])
	}
	for _, group := range groups {
		if atoi(group["gid"]) == gid {
			return str(group["gicon"])
		}
	}
	return ""
}

func vip(row map[string]interface{}, now int64) int {
	if atoi(row["sysgid"]) == 6 && atoi64(row["sysgid_exptime"]) > now {
		return 1
	}
	return 0
}

func formatUnix(ts int64) string {
	if ts <= 0 {
		return "1970-01-01 08:00:00"
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return time.Unix(ts, 0).In(loc).Format("2006-01-02 15:04:05")
}

func formatRemain(seconds int64) string {
	days := seconds / 86400
	seconds %= 86400
	hours := seconds / 3600
	seconds %= 3600
	minutes := seconds / 60
	seconds %= 60
	if days > 0 {
		return fmt.Sprintf("%d天后%d小时后%d分钟后%d秒后", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%d小时后%d分钟后%d秒后", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d分钟后%d秒后", minutes, seconds)
	}
	return fmt.Sprintf("%d秒后", seconds)
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func atoi(value interface{}) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(str(value)))
	return parsed
}

func atoi64(value interface{}) int64 {
	parsed, _ := strconv.ParseInt(strings.TrimSpace(str(value)), 10, 64)
	return parsed
}
