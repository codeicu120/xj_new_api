package index

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"xj_comp/internal/service/resourceurl"

	userRepo "xj_comp/internal/repository/user"
)

type InitStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
	Groups(ctx context.Context) ([]map[string]interface{}, error)
	Quota(ctx context.Context, uid int) (map[string]interface{}, error)
	Goldbean(ctx context.Context, uid int) (map[string]interface{}, error)
	SettingByUUID(ctx context.Context, uuid string) (map[string]interface{}, error)
	CalldataByUUID(ctx context.Context, uuid string) (map[string]interface{}, error)
}

type InitService struct {
	store           InitStore
	global          *GlobalService
	resourceBaseURL string
	resources       *resourceurl.Resolver
	now             func() time.Time
}

type InitRequest struct {
	Token         string
	Pkg           string
	Version       string
	XVersion      string
	UserAgent     string
	ClientIP      string
	HasCookieAuth bool
}

func (s *InitService) WithResourceResolver(r *resourceurl.Resolver) *InitService {
	s.resources = r
	return s
}

func NewInitService(store InitStore, global *GlobalService, resourceBaseURL string) *InitService {
	return &InitService{
		store:           store,
		global:          global,
		resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"),
		now:             time.Now,
	}
}

func (s *InitService) Init(ctx context.Context, req InitRequest) (map[string]interface{}, error) {
	resources := resourceurl.Resolved{BaseURL: s.resourceBaseURL}
	var resolveErr error
	if s.resources != nil {
		resources, resolveErr = s.resources.Resolve(ctx, resourceurl.Request{HasCookieAuth: req.HasCookieAuth, ClientIP: req.ClientIP})
		if resolveErr != nil {
			return nil, resolveErr
		}
	}
	user, groups, err := s.user(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	setting, err := s.global.settingArray(ctx, "setting")
	if err != nil {
		return nil, err
	}
	baseset, err := s.global.settingArray(ctx, "baseset")
	if err != nil {
		return nil, err
	}
	globalData, err := s.global.GetGlobalData(ctx, GlobalRequest{
		Pkg:           req.Pkg,
		Version:       req.Version,
		XVersion:      req.XVersion,
		UserAgent:     req.UserAgent,
		HasCookieAuth: req.HasCookieAuth,
		ClientIP:      req.ClientIP,
	})
	if err != nil {
		return nil, err
	}
	appver := globalData["appver"]
	appver, err = s.global.callJSON(ctx, firstExisting("global.appver."+req.Pkg, "global.appver"))
	if err != nil {
		return nil, err
	}
	bonus, err := s.global.settingArray(ctx, "promotion.bonus")
	if err != nil {
		return nil, err
	}
	playHeaders := s.global.mustCallJSON(ctx, "playHeaders", map[string]interface{}{})
	userRow, err := s.processUser(ctx, user, groups, resources)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"globalData":        globalData,
		"invite_bonus":      bonus,
		"user":              userRow,
		"appver":            appver,
		"notification_all":  notificationAll(user["notification_all"]),
		"inviteCodeUrl":     str(baseset["inviteCodeUrl"]),
		"inviteCodeAppid":   str(baseset["inviteCodeAppid"]),
		"playHeaders":       playHeaders,
		"urlHosts":          str(baseset["newHosts"]),
		"csurl":             str(setting["csurl"]),
		"sitelogo":          resources.GetRes(str(setting["sitelogo"]), ""),
		"isclosed":          atoi(setting["isclosed"]),
		"closetips":         str(setting["closetips"]),
		"externalUrlDating": "",
	}
	return data, nil
}

func (s *InitService) user(ctx context.Context, token string) (map[string]interface{}, []map[string]interface{}, error) {
	groups, err := s.store.Groups(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list groups: %w", err)
	}
	sid := userRepo.CleanToken(token)
	if sid == "" {
		return guestUser(""), groups, nil
	}
	user, err := s.store.UserBySession(ctx, sid)
	if err != nil {
		return nil, nil, err
	}
	if len(user) == 0 {
		return guestUser(sid), groups, nil
	}
	return user, groups, nil
}

func (s *InitService) processUser(ctx context.Context, row map[string]interface{}, groups []map[string]interface{}, resources resourceurl.Resolved) (map[string]interface{}, error) {
	uid := atoi(row["uid"])
	if uid > 0 {
		quota, err := s.store.Quota(ctx, uid)
		if err != nil {
			return nil, err
		}
		goldbean, err := s.store.Goldbean(ctx, uid)
		if err != nil {
			return nil, err
		}
		row["goldcoin"] = atoi(quota["goldcoin"])
		row["gold_bean"] = atoi(goldbean["gold_bean"])
	} else {
		row["goldcoin"] = atoi(row["goldcoin"])
		row["gold_bean"] = 0
	}
	now := s.now().Unix()
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
	return map[string]interface{}{
		"uid":             userIDValue(row["uid"]),
		"uniqkey":         strings.ToUpper(strconv.FormatInt(int64(atoi(row["uniqkey"])), 36)),
		"username":        nullableString(row["username"]),
		"nickname":        str(row["nickname"]),
		"mobi":            str(row["mobi"]),
		"email":           str(row["email"]),
		"sysgid":          str(row["sysgid"]),
		"gid":             str(row["gid"]),
		"gids":            nil,
		"gicon":           groupIcon(row, groups),
		"isvip":           vip(row, now),
		"regtime":         formatUnix(atoi64(row["regtime"])),
		"gender":          atoi(row["gender"]),
		"avatar":          str(row["avatar"]),
		"avatar_url":      avatarURL(resources, str(row["avatar"])),
		"newmsg":          str(row["newmsg"]),
		"goldcoin":        atoi(row["goldcoin"]),
		"gold_bean":       atoi(row["gold_bean"]),
		"duetime":         duetime,
		"dueday":          dueday,
		"recommend_total": atoi(row["recommend_total"]),
	}, nil
}

func avatarURL(resources resourceurl.Resolved, avatar string) string {
	if avatar == "" {
		return resources.GetRes("sysavatar/noavatar.png", "")
	}
	if strings.HasPrefix(avatar, "http://") || strings.HasPrefix(avatar, "https://") {
		return avatar
	}
	return resources.GetRes(strings.TrimLeft(avatar, "/"), "")
}

func (s *InitService) resURL(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return s.resourceBaseURL + "/" + strings.TrimLeft(path, "/")
}

func guestUser(sid string) map[string]interface{} {
	return map[string]interface{}{
		"uid":               "0",
		"sid":               sid,
		"uniqkey":           "0",
		"username":          nil,
		"nickname":          "",
		"mobi":              "",
		"email":             "",
		"sysgid":            "0",
		"gid":               "0",
		"sysgid_exptime":    "0",
		"regtime":           "0",
		"gender":            "0",
		"avatar":            "",
		"newmsg":            "0",
		"goldcoin":          "0",
		"recommend_total":   "0",
		"notification_all":  "",
		"signed_lasttime":   "0",
		"signed_contdays":   "0",
		"signed_unitdays":   "0",
		"signed_peakdays":   "0",
		"notification_time": "0",
	}
}

func notificationAll(value interface{}) interface{} {
	raw := strings.TrimSpace(str(value))
	if raw == "" {
		return nil
	}
	var out interface{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	if m, ok := out.(map[string]interface{}); ok && len(m) == 0 {
		return nil
	}
	if list, ok := out.([]interface{}); ok && len(list) == 0 {
		return nil
	}
	return out
}

func nullableString(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	return value
}

func userIDValue(value interface{}) interface{} {
	if atoi(value) == 0 {
		return 0
	}
	return str(value)
}

func groupIcon(row map[string]interface{}, groups []map[string]interface{}) string {
	sysgid := atoi(row["sysgid"])
	gid := atoi(row["gid"])
	for _, group := range groups {
		if atoi(group["gid"]) == sysgid && str(group["gicon"]) != "" {
			return str(group["gicon"])
		}
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
	return time.Unix(ts, 0).In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05")
}

func formatRemain(seconds int64) string {
	days := seconds / 86400
	seconds %= 86400
	hours := seconds / 3600
	seconds %= 3600
	minutes := seconds / 60
	seconds %= 60
	return fmt.Sprintf("%d天后%d小时后%d分钟后%d秒后", days, hours, minutes, seconds)
}

func atoi64(value interface{}) int64 {
	var parsed int64
	_, _ = fmt.Sscan(fmt.Sprint(value), &parsed)
	return parsed
}
