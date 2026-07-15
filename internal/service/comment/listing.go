package comment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
)

const sampleParams = "$vodid:0-$orderby:0-$page:1"

var paramKeys = []string{"vodid", "orderby", "page"}

var ErrVODNotFound = errors.New("vod not found")

type Store interface {
	VODByID(ctx context.Context, vodID int) (map[string]interface{}, error)
	UserGroups(ctx context.Context) ([]map[string]interface{}, error)
	CountRoots(ctx context.Context, vodID int) (int, error)
	RootComments(ctx context.Context, vodID int, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error)
	CommentByID(ctx context.Context, id int) (map[string]interface{}, error)
	IncrementVote(ctx context.Context, id int, field string) error
	CountByActorSince(ctx context.Context, actor interface{}, since int64, rewarded bool) (int, error)
	RecentCommentsByUID(ctx context.Context, uid int, since int64) ([]map[string]interface{}, error)
	RecentCommentsByIP(ctx context.Context, ip string, since int64) ([]map[string]interface{}, error)
	CreateComment(ctx context.Context, input domain.CommentCreateInput, parent map[string]interface{}) (int, error)
	IncrementVODCommentCount(ctx context.Context, vodID int) error
}

type CommentCreateInput = domain.CommentCreateInput

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type VoteLimiter interface {
	Seen(ctx context.Context, key string) (bool, error)
	Mark(ctx context.Context, key string) error
}

type Service struct {
	store           Store
	auth            AuthStore
	limiter         VoteLimiter
	resourceBaseURL string
	now             func() time.Time
}

type ListingRequest struct {
	PathParams string
	QueryPage  string
}

func NewService(store Store, resourceBaseURL string, opts ...interface{}) *Service {
	var auth AuthStore
	var limiter VoteLimiter
	for _, opt := range opts {
		switch typed := opt.(type) {
		case AuthStore:
			auth = typed
		case VoteLimiter:
			limiter = typed
		}
	}
	if limiter == nil {
		limiter = newMemoryVoteLimiter()
	}
	return &Service{
		store:           store,
		auth:            auth,
		limiter:         limiter,
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

func (s *Service) Vote(ctx context.Context, token string, id int, up bool) (int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -1, "评论操作失败", err
	}
	uid := atoi(str(user["uid"]))
	sid := str(user["sid"])
	if uid == 0 && sid == "" {
		sid = "guest"
	}
	row, err := s.store.CommentByID(ctx, id)
	if err != nil {
		return -1, "评论操作失败", err
	}
	if len(row) == 0 {
		return -1, "记录不存在或已被删除", nil
	}
	actor := sid
	if uid > 0 {
		actor = fmt.Sprint(uid)
	}
	key := fmt.Sprintf("comment.updown.check.%s.%s", actor, str(row["id"]))
	if seen, err := s.limiter.Seen(ctx, key); err != nil {
		return -1, "评论操作失败", err
	} else if seen {
		return -1, "您已经赞/踩过了", nil
	}
	field := "downnum"
	message := "已踩"
	if up {
		field = "upnum"
		message = "已赞"
	}
	if err := s.store.IncrementVote(ctx, atoi(str(row["id"])), field); err != nil {
		return -1, "评论操作失败", err
	}
	if err := s.limiter.Mark(ctx, key); err != nil {
		return -1, "评论操作失败", err
	}
	return 0, message, nil
}

func (s *Service) Post(ctx context.Context, token string, vodID int, parentID int, content string, ip string) (map[string]interface{}, int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return nil, -1, "发表评论失败", err
	}
	uid := atoi(str(user["uid"]))
	if uid == 0 {
		return nil, -9999, "请注册会员并登录APP才可以发表评论噢", nil
	}
	perms, err := s.userPerms(ctx, user)
	if err != nil {
		return nil, -1, "发表评论失败", err
	}
	if getPermInt(perms, "deny.comment.post") == 1 {
		return nil, 1, "您已被禁止评论", nil
	}
	if tooManyByPattern(str(user["nickname"]), `[\pN]`, 5) {
		return nil, 11, "账号异常，请联系管理员", nil
	}
	daytime := dayStart(s.now()).Unix()
	daycount, err := s.store.CountByActorSince(ctx, uid, daytime, false)
	if err != nil {
		return nil, -1, "发表评论失败", err
	}
	if daycount >= getPermInt(perms, "max.comment.post.daynum") {
		return nil, 2, "每日发表评论数已满额", nil
	}
	vodrow, err := s.store.VODByID(ctx, vodID)
	if err != nil {
		return nil, -1, "发表评论失败", err
	}
	if len(vodrow) == 0 || atoi(str(vodrow["showtype"])) > 1 {
		return nil, 3, "记录不存在或已被删除", nil
	}
	content = strings.TrimRight(content, " \t\r\n")
	if runeLen(content) < 1 || runeLen(content) > 30 {
		return nil, 4, "评论允许1-30字之间", nil
	}
	if !commentAllowedChars(content) || tooManyByPattern(content, `[\pP]`, 5) || tooManyByPattern(content, `[\pZ]`, 5) || tooManyByPattern(content, `[\n]`, 5) {
		return nil, 5, "内容含有禁止发布的关键词，请检查！", nil
	}
	parent := map[string]interface{}{}
	if parentID > 0 {
		parent, err = s.store.CommentByID(ctx, parentID)
		if err != nil {
			return nil, -1, "发表评论失败", err
		}
		if len(parent) == 0 || atoi(str(parent["vodid"])) != atoi(str(vodrow["vodid"])) {
			return nil, 7, "回复的评论不正确", nil
		}
		if atoi(str(parent["showtype"])) != 0 {
			return nil, 7, "被回复内容不存在或已删除", nil
		}
	}
	if runeLen(content) > 5 {
		since := s.now().Add(-10 * time.Minute).Unix()
		rows, err := s.store.RecentCommentsByUID(ctx, uid, since)
		if err != nil {
			return nil, -1, "发表评论失败", err
		}
		for _, row := range rows {
			if similarEnough(content, str(row["content"]), 0.80) {
				return nil, 10, "请勿发布重复内容1", nil
			}
		}
		rows, err = s.store.RecentCommentsByIP(ctx, cleanIP(ip), since)
		if err != nil {
			return nil, -1, "发表评论失败", err
		}
		for _, row := range rows {
			if similarEnough(content, str(row["content"]), 0.80) {
				return nil, 10, "请勿发布重复内容2", nil
			}
		}
	}
	input := domain.CommentCreateInput{
		RootID:   parentRootID(parent),
		ParentID: atoi(str(parent["id"])),
		Left:     parentLeft(parent),
		Right:    parentRight(parent),
		Depth:    atoi(str(parent["depth"])) + boolInt(len(parent) > 0),
		VODID:    atoi(str(vodrow["vodid"])),
		UID:      uid,
		SID:      fmt.Sprint(uid),
		Content:  content,
		AddTime:  s.now().Unix(),
		IP:       cleanIP(ip),
		ShowType: 4,
	}
	if input.Depth > 10 {
		return nil, 8, "评论回复深度最深10层", nil
	}
	id, err := s.store.CreateComment(ctx, input, parent)
	if err != nil {
		return nil, -1, "发表评论失败", err
	}
	if id == 0 {
		return nil, 9, "评论发表失败", nil
	}
	if err := s.store.IncrementVODCommentCount(ctx, input.VODID); err != nil {
		return nil, -1, "发表评论失败", err
	}
	return map[string]interface{}{}, 0, "发表成功", nil
}

func (s *Service) userByToken(ctx context.Context, token string) (map[string]interface{}, error) {
	sid := userRepo.CleanToken(token)
	if sid == "" {
		return map[string]interface{}{"uid": "0"}, nil
	}
	if s.auth == nil {
		return map[string]interface{}{"uid": "0", "sid": sid}, nil
	}
	user, err := s.auth.UserBySession(ctx, sid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return map[string]interface{}{"uid": "0", "sid": sid}, nil
	}
	if str(user["sid"]) == "" {
		user["sid"] = sid
	}
	return user, nil
}

func (s *Service) userPerms(ctx context.Context, user map[string]interface{}) (map[string]interface{}, error) {
	if raw, ok := user["perms"].(map[string]interface{}); ok {
		return raw, nil
	}
	if str(user["perms"]) != "" {
		return parsePermMap(user["perms"]), nil
	}
	groups, err := s.store.UserGroups(ctx)
	if err != nil {
		return nil, err
	}
	perms := map[string]interface{}{}
	for _, group := range groups {
		for key, value := range parsePermMap(group["perms"]) {
			if _, exists := perms[key]; !exists {
				perms[key] = value
			}
		}
	}
	return perms, nil
}

func parsePermMap(value interface{}) map[string]interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed
	case string:
		out := map[string]interface{}{}
		_ = json.Unmarshal([]byte(typed), &out)
		return out
	default:
		return map[string]interface{}{}
	}
}

func getPermInt(perms map[string]interface{}, key string) int {
	return atoi(str(perms[key]))
}

func dayStart(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func runeLen(value string) int {
	return len([]rune(value))
}

func commentAllowedChars(content string) bool {
	return regexp.MustCompile(`^[\p{Han}\pL\pP\pZ0-9\r\n\t１２３４５６７８９０]+$`).MatchString(content)
}

func tooManyByPattern(value string, pattern string, maxParts int) bool {
	parts := regexp.MustCompile(pattern).Split(value, -1)
	return len(parts) > maxParts
}

func parentRootID(parent map[string]interface{}) int {
	if len(parent) == 0 {
		return 0
	}
	if atoi(str(parent["rootid"])) == 0 {
		return atoi(str(parent["id"]))
	}
	return atoi(str(parent["rootid"]))
}

func parentLeft(parent map[string]interface{}) int {
	if len(parent) == 0 {
		return 1
	}
	return atoi(str(parent["rgt"]))
}

func parentRight(parent map[string]interface{}) int {
	if len(parent) == 0 {
		return 2
	}
	return atoi(str(parent["rgt"])) + 1
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func cleanIP(ip string) string {
	return regexp.MustCompile(`[^a-zA-Z0-9\.:]`).ReplaceAllString(ip, "")
}

func similarEnough(a string, b string, threshold float64) bool {
	ar := []rune(a)
	br := []rune(b)
	if len(ar) == 0 && len(br) == 0 {
		return true
	}
	dp := make([][]int, len(ar)+1)
	for i := range dp {
		dp[i] = make([]int, len(br)+1)
	}
	for i := 1; i <= len(ar); i++ {
		for j := 1; j <= len(br); j++ {
			if ar[i-1] == br[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] > dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}
	score := float64(dp[len(ar)][len(br)]*2) / float64(len(ar)+len(br))
	return score > threshold
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
