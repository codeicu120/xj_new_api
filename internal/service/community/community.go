package community

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"xj_comp/internal/domain"
	communityRepo "xj_comp/internal/repository/community"
	userRepo "xj_comp/internal/repository/user"
	vodService "xj_comp/internal/service/vod"
)

const listingSampleParams = "$category_id:0-$type:0-$orderby:0-$page:1"
const commentsSampleParams = "$orderby:0-$page:1"

var listingParamKeys = []string{"category_id", "type", "orderby", "page"}
var commentsParamKeys = []string{"orderby", "page"}

var ErrLoginRequired = errors.New("community login required")
var ErrTopicNotFound = errors.New("community topic not found")

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type Store interface {
	CountTopics(ctx context.Context, filter communityRepo.TopicFilter) (int, error)
	ListTopics(ctx context.Context, filter communityRepo.TopicFilter, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error)
	Servers(ctx context.Context) ([]map[string]interface{}, error)
	ImagesByTIDs(ctx context.Context, tids []int) (map[int][]map[string]interface{}, error)
	VideosByTIDs(ctx context.Context, tids []int) (map[int][]map[string]interface{}, error)
	FavoriteTopicIDs(ctx context.Context, uid int, tids []int) (map[int]int, error)
	UpTopicIDs(ctx context.Context, uid int, tids []int) (map[int]int, error)
	TopicByID(ctx context.Context, tid int) (map[string]interface{}, error)
	IncrementTopicVisit(ctx context.Context, tid int) error
	SetTopicUp(ctx context.Context, tid int, uid int, up bool, now int64) error
	IncrementTopicUp(ctx context.Context, tid int, delta int) error
	CountComments(ctx context.Context, tid int) (int, error)
	ListComments(ctx context.Context, tid int, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error)
	UpCommentIDs(ctx context.Context, uid int, ids []int) (map[int]int, error)
	CommentByID(ctx context.Context, cid int) (map[string]interface{}, error)
	SetCommentUp(ctx context.Context, cid int, uid int, up bool, now int64) error
	IncrementCommentUp(ctx context.Context, cid int, delta int) error
}

type Service struct {
	auth            AuthStore
	store           Store
	resourceBaseURL string
	now             func() time.Time
}

type ListingRequest struct {
	Action      string
	PathParams  string
	QueryPage   string
	IsH5Request bool
	Token       string
}

type CommentListingRequest struct {
	PathParams string
	QueryPage  string
	QueryOrder string
	TID        int
	Token      string
}

type ShowRequest struct {
	TID        int
	QueryOrder string
	Token      string
}

func NewService(auth AuthStore, store Store, resourceBaseURL string) *Service {
	return &Service{auth: auth, store: store, resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"), now: time.Now}
}

func (s *Service) Listing(ctx context.Context, req ListingRequest) (domain.CommunityListingData, error) {
	params := parseParams(req.PathParams, listingParamKeys, []string{"0", "0", "0", "1"})
	if atoi(params["page"]) == 0 {
		params["page"] = req.QueryPage
		if params["page"] == "" {
			params["page"] = "0"
		}
	}
	user, err := s.userByToken(ctx, req.Token)
	if err != nil {
		return domain.CommunityListingData{}, err
	}
	uid := atoi(user["uid"])
	if req.Action == "favorite" && uid == 0 {
		return domain.CommunityListingData{}, ErrLoginRequired
	}
	filter := communityRepo.TopicFilter{Action: req.Action, CategoryID: atoi(params["category_id"]), Type: atoi(params["type"])}
	if req.Action == "favorite" {
		filter.FavoriteUID = uid
	}
	orderBy := "tid DESC"
	switch req.Action {
	case "hot":
		orderBy = "visit_count DESC"
	case "latest":
		orderBy = "tid DESC"
	default:
		switch atoi(params["orderby"]) {
		case 1:
			orderBy = "upnum DESC"
		case 2:
			orderBy = "visit_count DESC"
		}
	}
	const pageSize = 20
	total, err := s.store.CountTopics(ctx, filter)
	if err != nil {
		return domain.CommunityListingData{}, err
	}
	rows, err := s.store.ListTopics(ctx, filter, total, atoi(params["page"]), pageSize, orderBy)
	if err != nil {
		return domain.CommunityListingData{}, err
	}
	rows, err = s.processTopics(ctx, rows, uid)
	if err != nil {
		return domain.CommunityListingData{}, err
	}
	pageURL := "/community/" + req.Action + "-" + buildParams(params, listingParamKeys, map[string]string{"page": "[?]"})
	return domain.CommunityListingData{
		Now:          s.now().Unix(),
		Action:       req.Action,
		SampleParams: listingSampleParams,
		Params:       params,
		Rows:         rows,
		PageInfo:     vodService.PageInfo(total, pageSize, atoi(params["page"]), pageURL),
	}, nil
}

func (s *Service) CommentListing(ctx context.Context, req CommentListingRequest) (domain.CommunityCommentListingData, error) {
	params := parseParams(req.PathParams, commentsParamKeys, []string{"0", "1"})
	if atoi(params["orderby"]) == 0 && req.QueryOrder != "" {
		params["orderby"] = req.QueryOrder
	}
	if req.QueryPage != "" {
		params["page"] = req.QueryPage
	}
	topic, err := s.store.TopicByID(ctx, req.TID)
	if err != nil {
		return domain.CommunityCommentListingData{}, err
	}
	if len(topic) == 0 {
		return domain.CommunityCommentListingData{}, ErrTopicNotFound
	}
	orderBy := "a.addtime DESC"
	if atoi(params["orderby"]) == 1 {
		orderBy = "a.upnum DESC"
	}
	user, err := s.userByToken(ctx, req.Token)
	if err != nil {
		return domain.CommunityCommentListingData{}, err
	}
	const pageSize = 20
	total, err := s.store.CountComments(ctx, req.TID)
	if err != nil {
		return domain.CommunityCommentListingData{}, err
	}
	rows, err := s.store.ListComments(ctx, req.TID, total, atoi(params["page"]), pageSize, orderBy)
	if err != nil {
		return domain.CommunityCommentListingData{}, err
	}
	rows = s.processComments(rows)
	uid := atoi(user["uid"])
	if uid > 0 {
		upIDs, err := s.store.UpCommentIDs(ctx, uid, commentIDs(rows))
		if err != nil {
			return domain.CommunityCommentListingData{}, err
		}
		for _, row := range rows {
			row["is_up"] = upIDs[atoi(row["id"])]
		}
	}
	pageURL := "/community/clisting-" + buildParams(params, commentsParamKeys, map[string]string{"page": "[?]"})
	return domain.CommunityCommentListingData{
		Now:          s.now().Unix(),
		SampleParams: commentsSampleParams,
		Params:       commentParamsForJSON(params),
		Rows:         rows,
		PageInfo:     vodService.PageInfo(total, pageSize, atoi(params["page"]), pageURL),
	}, nil
}

func (s *Service) Show(ctx context.Context, req ShowRequest) (map[string]interface{}, error) {
	topic, err := s.store.TopicByID(ctx, req.TID)
	if err != nil {
		return nil, err
	}
	if len(topic) == 0 {
		return nil, ErrTopicNotFound
	}
	if err := s.store.IncrementTopicVisit(ctx, atoi(topic["tid"])); err != nil {
		return nil, err
	}
	user, err := s.userByToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	processed, err := s.processTopics(ctx, []map[string]interface{}{topic}, atoi(user["uid"]))
	if err != nil {
		return nil, err
	}
	orderBy := "a.addtime DESC"
	if atoi(req.QueryOrder) == 1 {
		orderBy = "a.upnum DESC"
	}
	total, err := s.store.CountComments(ctx, req.TID)
	if err != nil {
		return nil, err
	}
	comments, err := s.store.ListComments(ctx, req.TID, total, 1, 100, orderBy)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"row":               processed[0],
		"totalCommentCount": total,
		"comments":          s.processComments(comments),
	}, nil
}

func (s *Service) UpTopic(ctx context.Context, token string, tid int) (int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -1, "社区点赞失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "请登录后操作", nil
	}
	topic, err := s.store.TopicByID(ctx, tid)
	if err != nil {
		return -1, "社区点赞失败", err
	}
	if len(topic) == 0 {
		return -1, "记录不存在或已删除", nil
	}
	upIDs, err := s.store.UpTopicIDs(ctx, uid, []int{tid})
	if err != nil {
		return -1, "社区点赞失败", err
	}
	if upIDs[tid] > 0 {
		if err := s.store.SetTopicUp(ctx, tid, uid, false, s.now().Unix()); err != nil {
			return -1, "社区点赞失败", err
		}
		if err := s.store.IncrementTopicUp(ctx, tid, -1); err != nil {
			return -1, "社区点赞失败", err
		}
		return 0, "取消赞成功", nil
	}
	if err := s.store.SetTopicUp(ctx, tid, uid, true, s.now().Unix()); err != nil {
		return -1, "社区点赞失败", err
	}
	if err := s.store.IncrementTopicUp(ctx, tid, 1); err != nil {
		return -1, "社区点赞失败", err
	}
	return 0, "已赞", nil
}

func (s *Service) UpComment(ctx context.Context, token string, cid int) (int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -1, "社区评论点赞失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -9999, "请登录后操作", nil
	}
	comment, err := s.store.CommentByID(ctx, cid)
	if err != nil {
		return -1, "社区评论点赞失败", err
	}
	if len(comment) == 0 {
		return -1, "记录不存在或已删除", nil
	}
	upIDs, err := s.store.UpCommentIDs(ctx, uid, []int{cid})
	if err != nil {
		return -1, "社区评论点赞失败", err
	}
	if upIDs[cid] > 0 {
		if err := s.store.SetCommentUp(ctx, cid, uid, false, s.now().Unix()); err != nil {
			return -1, "社区评论点赞失败", err
		}
		if err := s.store.IncrementCommentUp(ctx, cid, -1); err != nil {
			return -1, "社区评论点赞失败", err
		}
		return 0, "取消赞成功", nil
	}
	if err := s.store.SetCommentUp(ctx, cid, uid, true, s.now().Unix()); err != nil {
		return -1, "社区评论点赞失败", err
	}
	if err := s.store.IncrementCommentUp(ctx, cid, 1); err != nil {
		return -1, "社区评论点赞失败", err
	}
	return 0, "已赞", nil
}

func (s *Service) processTopics(ctx context.Context, rows []map[string]interface{}, uid int) ([]map[string]interface{}, error) {
	if len(rows) == 0 {
		return []map[string]interface{}{}, nil
	}
	servers, err := s.store.Servers(ctx)
	if err != nil {
		return nil, err
	}
	tids := topicIDs(rows)
	images, err := s.store.ImagesByTIDs(ctx, tids)
	if err != nil {
		return nil, err
	}
	videos, err := s.store.VideosByTIDs(ctx, tids)
	if err != nil {
		return nil, err
	}
	favs, ups := map[int]int{}, map[int]int{}
	if uid > 0 {
		favs, err = s.store.FavoriteTopicIDs(ctx, uid, tids)
		if err != nil {
			return nil, err
		}
		ups, err = s.store.UpTopicIDs(ctx, uid, tids)
		if err != nil {
			return nil, err
		}
	}
	for _, row := range rows {
		tid := atoi(row["tid"])
		row["content"] = s.displayContent(str(row["content"]), atoi(row["image_srvid"]), atoi(row["video_srvid"]), servers)
		row["images"] = s.processImages(images[tid], atoi(row["image_srvid"]), servers)
		media := extractMediaRows(str(row["content"]))
		row["content_images"] = media["images"]
		row["videos"] = media["videos"]
		if len(row["videos"].([]map[string]interface{})) == 0 {
			row["videos"] = videos[tid]
			if row["videos"] == nil {
				row["videos"] = []map[string]interface{}{}
			}
		}
		if uid > 0 {
			row["is_favorite"] = favs[tid]
			row["is_up"] = ups[tid]
		}
	}
	return rows, nil
}

func (s *Service) processImages(rows []map[string]interface{}, srvid int, servers []map[string]interface{}) []map[string]interface{} {
	out := []map[string]interface{}{}
	for _, row := range rows {
		cp := clone(row)
		cp["image_path"] = s.coverURL(str(cp["image_path"]), srvid, servers)
		out = append(out, cp)
	}
	return out
}

func (s *Service) processComments(rows []map[string]interface{}) []map[string]interface{} {
	out := []map[string]interface{}{}
	for _, row := range rows {
		item := s.processComment(row)
		subrows := []map[string]interface{}{}
		if raw, ok := row["subrows"].([]map[string]interface{}); ok {
			for _, subrow := range raw {
				subrows = append(subrows, s.processComment(subrow))
			}
		}
		item["subrows"] = subrows
		out = append(out, item)
	}
	return out
}

func (s *Service) processComment(row map[string]interface{}) map[string]interface{} {
	now := s.now().Unix()
	return map[string]interface{}{
		"id":           str(row["id"]),
		"rootid":       str(row["rootid"]),
		"parentid":     str(row["parentid"]),
		"lft":          str(row["lft"]),
		"rgt":          str(row["rgt"]),
		"depth":        str(row["depth"]),
		"tid":          str(row["tid"]),
		"uid":          str(row["uid"]),
		"sid":          str(row["sid"]),
		"username":     str(row["username"]),
		"nickname":     str(row["nickname"]),
		"gender":       atoi(row["gender"]),
		"gicon":        "",
		"isvip":        0,
		"content":      str(row["content"]),
		"upnum":        str(row["upnum"]),
		"downnum":      str(row["downnum"]),
		"avatar_url":   s.avatarURL(str(row["avatar"])),
		"addtime":      legacyRelativeTime(atoi64(row["addtime"]), now),
		"__closenum__": row["__closenum__"],
	}
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

func (s *Service) displayContent(html string, imageSrvid int, videoSrvid int, servers []map[string]interface{}) string {
	imgRe := regexp.MustCompile(`<img([^>]*?)src=["']([^"']+)["']([^>]*)>`)
	html = imgRe.ReplaceAllStringFunc(html, func(tag string) string {
		m := imgRe.FindStringSubmatch(tag)
		if len(m) < 4 {
			return tag
		}
		src := s.coverURL(m[2], imageSrvid, servers)
		if strings.Contains(tag, "data-uri=") {
			return strings.Replace(tag, m[2], src, 1)
		}
		return `<img` + m[1] + `src="` + src + `" data-uri="` + m[2] + `"` + m[3] + `>`
	})
	videoRe := regexp.MustCompile(`<(video|source)([^>]*?)src=["']([^"']+)["']([^>]*)>`)
	return videoRe.ReplaceAllStringFunc(html, func(tag string) string {
		m := videoRe.FindStringSubmatch(tag)
		if len(m) < 5 {
			return tag
		}
		src := s.coverURL(m[3], videoSrvid, servers)
		if strings.Contains(tag, "data-uri=") {
			return strings.Replace(tag, m[3], src, 1)
		}
		return `<` + m[1] + m[2] + `src="` + src + `" data-uri="` + m[3] + `"` + m[4] + `>`
	})
}

func extractMediaRows(html string) map[string][]map[string]interface{} {
	out := map[string][]map[string]interface{}{"images": {}, "videos": {}}
	imgRe := regexp.MustCompile(`<img[^>]*(?:src|data-uri)=["']([^"']+)["'][^>]*>`)
	for _, m := range imgRe.FindAllStringSubmatch(html, -1) {
		out["images"] = append(out["images"], map[string]interface{}{"url": m[1], "raw_url": m[1], "srvid": 0})
	}
	videoRe := regexp.MustCompile(`<(?:video|source)[^>]*(?:src|data-uri)=["']([^"']+)["'][^>]*>`)
	for _, m := range videoRe.FindAllStringSubmatch(html, -1) {
		out["videos"] = append(out["videos"], map[string]interface{}{"url": m[1], "raw_url": m[1], "srvid": 0})
	}
	return out
}

func (s *Service) coverURL(uri string, srvid int, servers []map[string]interface{}) string {
	if uri == "" || strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return uri
	}
	if srvid > 0 {
		for _, server := range servers {
			if str(server["srvtype"]) == "cover" && atoi(server["srvid"]) == srvid {
				return strings.TrimRight(str(server["srvhost"]), "/") + "/" + strings.TrimLeft(uri, "/")
			}
		}
	}
	return s.resourceBaseURL + "/" + strings.TrimLeft(uri, "/")
}

func (s *Service) avatarURL(avatar string) string {
	if avatar == "" {
		return s.resourceBaseURL + "/sysavatar/noavatar.png"
	}
	if strings.HasPrefix(avatar, "http://") || strings.HasPrefix(avatar, "https://") {
		return avatar
	}
	if isDigits(avatar) {
		return avatar
	}
	if strings.HasPrefix(avatar, "avatar/") {
		return s.resourceBaseURL + "/C1/" + strings.TrimLeft(avatar, "/")
	}
	return s.resourceBaseURL + "/C1/avatar/" + strings.TrimLeft(avatar, "/")
}

func parseParams(raw string, keys []string, defaults []string) map[string]string {
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

func buildParams(params map[string]string, keys []string, replace map[string]string) string {
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		value := params[key]
		if next, ok := replace[key]; ok {
			value = next
		}
		values = append(values, value)
	}
	return strings.Join(values, "-")
}

func commentParamsForJSON(params map[string]string) map[string]interface{} {
	return map[string]interface{}{
		"orderby": atoi(params["orderby"]),
		"page":    params["page"],
	}
}

func topicIDs(rows []map[string]interface{}) []int {
	ids := []int{}
	for _, row := range rows {
		id := atoi(row["tid"])
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

func commentIDs(rows []map[string]interface{}) []int {
	ids := []int{}
	for _, row := range rows {
		id := atoi(row["id"])
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

func clone(row map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for key, value := range row {
		out[key] = value
	}
	return out
}

func legacyRelativeTime(ts int64, now int64) string {
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

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
