package explore

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
)

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
	Groups(ctx context.Context) ([]map[string]interface{}, error)
}

type Store interface {
	Tabs(ctx context.Context) ([]map[string]interface{}, error)
	UpdateUserNotificationAll(ctx context.Context, uid int, value string) error
	UpdateGuestNotificationAll(ctx context.Context, sid string, value string) error
}

type Service struct {
	auth            AuthStore
	store           Store
	resourceBaseURL string
	now             func() time.Time
}

func NewService(auth AuthStore, store Store, resourceBaseURL string) *Service {
	return &Service{
		auth:            auth,
		store:           store,
		resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"),
		now:             time.Now,
	}
}

func (s *Service) Index(ctx context.Context, token string) (domain.ExploreIndexData, error) {
	user, err := s.userWithPerms(ctx, token)
	if err != nil {
		return domain.ExploreIndexData{}, err
	}
	tabs, err := s.store.Tabs(ctx)
	if err != nil {
		return domain.ExploreIndexData{}, fmt.Errorf("list explore tabs: %w", err)
	}
	today := dayStartUnix(s.now())
	signedUnitDays := atoi(user["signed_unitdays"])
	if int64(atoi(user["signed_lasttime"])) < today {
		signedUnitDays++
	}
	dayRows := make([]map[string]interface{}, 0, 7)
	for i := 0; i < 7; i++ {
		current := signedUnitDays + i
		if current > 7 {
			current -= 7
		}
		dayRows = append(dayRows, map[string]interface{}{
			"day":     time.Unix(today+int64(i)*86400, 0).In(chinaLocation()).Format("01-02"),
			"coinnum": getPermInt(user["perms"], fmt.Sprintf("max.signtask.coinnum%d", current)),
		})
	}
	return domain.ExploreIndexData{
		TabRows: processTabs(tabs, s.resourceBaseURL),
		DayRows: dayRows,
		SignData: map[string]interface{}{
			"signed_today":    boolInt(int64(atoi(user["signed_lasttime"])) >= today),
			"signed_peakdays": atoi(user["signed_peakdays"]),
			"signed_contdays": atoi(user["signed_contdays"]),
			"signed_unitdays": atoi(user["signed_unitdays"]),
		},
	}, nil
}

func (s *Service) CleanNotification(ctx context.Context, token string, tabKey string) (map[string]interface{}, int, string, error) {
	tabKey = strings.TrimSpace(tabKey)
	if tabKey == "" {
		return nil, -1, "请提供频道键名", nil
	}
	user, err := s.userWithPerms(ctx, token)
	if err != nil {
		return nil, -1, "清除红点失败", err
	}
	notificationAll := parseStringMap(user["notification_all"])
	if tabKey == "all" {
		notificationAll = nil
	} else {
		if notificationAll == nil {
			return nil, -1, "指定的频道键名不存在", nil
		}
		if _, ok := notificationAll[tabKey]; !ok {
			return nil, -1, "指定的频道键名不存在", nil
		}
		notificationAll[tabKey] = float64(0)
	}
	raw := "null"
	if notificationAll != nil {
		encoded, _ := json.Marshal(notificationAll)
		raw = string(encoded)
	}
	if uid := atoi(user["uid"]); uid > 0 {
		err = s.store.UpdateUserNotificationAll(ctx, uid, raw)
	} else {
		err = s.store.UpdateGuestNotificationAll(ctx, str(user["sid"]), raw)
	}
	if err != nil {
		return nil, -1, "清除红点失败", err
	}
	var dataValue interface{}
	if notificationAll != nil {
		dataValue = notificationAll
	}
	return map[string]interface{}{"notification_all": dataValue}, 0, "", nil
}

func (s *Service) userWithPerms(ctx context.Context, token string) (map[string]interface{}, error) {
	groups, err := s.auth.Groups(ctx)
	if err != nil {
		return nil, fmt.Errorf("list user groups: %w", err)
	}
	sid := userRepo.CleanToken(token)
	user, err := s.auth.UserBySession(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("load user by session: %w", err)
	}
	if user == nil {
		user = map[string]interface{}{"uid": "0", "sid": sid}
	}
	if atoi(user["uid"]) > 0 {
		user["perms"] = initPerm(initGids(user, s.now), groups)
	} else {
		user["perms"] = initPerm([]int{0}, groups)
	}
	return user, nil
}

func processTabs(rows []map[string]interface{}, resourceBaseURL string) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, row := range rows {
		result = append(result, map[string]interface{}{
			"tabkey":    str(row["tabkey"]),
			"tabname":   str(row["tabname"]),
			"intro":     str(row["intro"]),
			"coverpic":  resourceURL(resourceBaseURL, str(row["coverpic"])),
			"coverpic2": resourceURL(resourceBaseURL, str(row["coverpic2"])),
			"extjson":   decodeJSON(row["extjson"]),
		})
	}
	return result
}

func initGids(user map[string]interface{}, now func() time.Time) []int {
	gids := []int{}
	gid := atoi(user["gid"])
	if sysgid := atoi(user["sysgid"]); sysgid > 0 && int64(atoi(user["sysgid_exptime"])) > now().Unix() {
		gid = sysgid
	}
	gids = append(gids, gid)
	for _, part := range strings.Split(str(user["gids"]), ",") {
		if id := atoi(part); id > 0 {
			gids = append(gids, id)
		}
	}
	return gids
}

func initPerm(gids []int, groups []map[string]interface{}) map[string]interface{} {
	selected := []map[string]interface{}{}
	seen := map[int]struct{}{}
	for _, gid := range gids {
		if _, ok := seen[gid]; ok {
			continue
		}
		seen[gid] = struct{}{}
		for _, group := range groups {
			if atoi(group["gid"]) == gid {
				selected = append(selected, group)
				break
			}
		}
	}
	sort.SliceStable(selected, func(i, j int) bool {
		return atoi(selected[i]["weight"]) > atoi(selected[j]["weight"])
	})
	multiPerms := make([]map[string]interface{}, 0, len(selected))
	for _, group := range selected {
		multiPerms = append(multiPerms, parsePermMap(group["perms"]))
	}
	return computePerm(multiPerms)
}

func computePerm(multiPerms []map[string]interface{}) map[string]interface{} {
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
	result := map[string]interface{}{}
	for _, key := range keys {
		permType := key
		if index := strings.Index(key, "."); index >= 0 {
			permType = key[:index]
		}
		switch permType {
		case "allow", "deny":
			value := 0
			for _, perms := range multiPerms {
				if atoi(perms[key]) == 1 {
					value = 1
					break
				}
			}
			result[key] = value
		case "min":
			set := false
			value := 0
			for _, perms := range multiPerms {
				if _, ok := perms[key]; !ok {
					continue
				}
				if !set || atoi(perms[key]) < value {
					value = atoi(perms[key])
					set = true
				}
			}
			result[key] = value
		case "max":
			value := 0
			for _, perms := range multiPerms {
				if _, ok := perms[key]; ok && atoi(perms[key]) > value {
					value = atoi(perms[key])
				}
			}
			result[key] = value
		default:
			for _, perms := range multiPerms {
				if value, ok := perms[key]; ok {
					result[key] = value
					break
				}
			}
		}
	}
	return result
}

func parsePermMap(value interface{}) map[string]interface{} {
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

func parseStringMap(value interface{}) map[string]interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed
	case string:
		if strings.TrimSpace(typed) == "" {
			return nil
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(typed), &parsed); err != nil {
			return nil
		}
		return parsed
	default:
		return nil
	}
}

func getPermInt(perms interface{}, key string) int {
	values, ok := perms.(map[string]interface{})
	if !ok {
		values = parsePermMap(perms)
	}
	return atoi(values[key])
}

func decodeJSON(value interface{}) interface{} {
	raw := strings.TrimSpace(str(value))
	if raw == "" {
		return nil
	}
	var decoded interface{}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return nil
	}
	return decoded
}

func resourceURL(baseURL string, path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if baseURL == "" {
		return path
	}
	return baseURL + "/" + strings.TrimLeft(path, "/")
}

func dayStartUnix(now time.Time) int64 {
	loc := chinaLocation()
	local := now.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc).Unix()
}

func chinaLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(str(value), &n)
	return n
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
