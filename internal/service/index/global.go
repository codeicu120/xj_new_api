package index

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"xj_comp/internal/service/resourceurl"
)

type GlobalStore interface {
	CalldataByUUID(ctx context.Context, uuid string) (map[string]interface{}, error)
	SettingByUUID(ctx context.Context, uuid string) (map[string]interface{}, error)
}

type GlobalService struct {
	store           GlobalStore
	resourceBaseURL string
	resources       *resourceurl.Resolver
}

type GlobalRequest struct {
	Pkg           string
	Version       string
	XVersion      string
	UserAgent     string
	HasCookieAuth bool
	ClientIP      string
}

func (s *GlobalService) WithResourceResolver(r *resourceurl.Resolver) *GlobalService {
	s.resources = r
	return s
}

func NewGlobalService(store GlobalStore, resourceBaseURL string) *GlobalService {
	return &GlobalService{store: store, resourceBaseURL: strings.TrimRight(resourceBaseURL, "/")}
}

func (s *GlobalService) GetGlobalData(ctx context.Context, req GlobalRequest) (map[string]interface{}, error) {
	resources := resourceurl.Resolved{BaseURL: s.resourceBaseURL}
	var err error
	if s.resources != nil {
		resources, err = s.resources.Resolve(ctx, resourceurl.Request{HasCookieAuth: req.HasCookieAuth, ClientIP: req.ClientIP})
		if err != nil {
			return nil, err
		}
	}
	pkg := strings.TrimSpace(req.Pkg)
	dotpkg := ""
	if pkg != "" {
		dotpkg = "." + pkg
	}
	if strings.TrimSpace(req.XVersion) == "5.2.0" {
		dotpkg = ".5.2.0"
	}
	appver, err := s.callJSON(ctx, firstExisting("global.appver"+dotpkg, "global.appver"))
	if err != nil {
		return nil, err
	}
	if strings.Contains(strings.ToLower(req.UserAgent), "wallpaper") {
		appverMap, ok := appver.(map[string]interface{})
		if !ok {
			appverMap = map[string]interface{}{}
			appver = appverMap
		}
		appverMap["iOSVer"] = "3.5.2"
	}
	if req.Version != "" {
		appverMap, ok := appver.(map[string]interface{})
		if !ok {
			appverMap = map[string]interface{}{}
			appver = appverMap
		}
		appverMap["AndroidVer"] = req.Version
		appverMap["iOSVer"] = req.Version
	}

	setting, err := s.settingArray(ctx, "setting")
	if err != nil {
		return nil, err
	}
	baseset, err := s.settingArray(ctx, "baseset")
	if err != nil {
		return nil, err
	}
	regopt, err := s.settingArray(ctx, "user.regopt")
	if err != nil {
		return nil, err
	}

	inviteCode := "0000"
	inviteURL := chooseLine(str(baseset["inviteUrls"]))
	qrlink := strings.ReplaceAll(s.callCodeOrEmpty(ctx, "global.qrcode.link"), "{inviteUrl}", inviteURL)
	qrlink = strings.ReplaceAll(qrlink, "{inviteCode}", inviteCode)
	sharetext := strings.ReplaceAll(s.callHTMLOrEmpty(ctx, "global.share.text"), "{inviteUrl}", inviteURL)
	sharetext = strings.ReplaceAll(sharetext, "{inviteCode}", inviteCode)

	data := map[string]interface{}{
		"webreg":                       atoi(regopt["webreg"]),
		"appver":                       appver,
		"hotwords":                     s.mustCallJSON(ctx, "search.hotwords", []interface{}{}),
		"hottags":                      s.mustCallJSON(ctx, "global.hottags", []interface{}{}),
		"hotcategories":                s.mustCallJSON(ctx, "global.hotcategories", []interface{}{}),
		"appdownurl":                   s.callCodeOrEmpty(ctx, prefixed(pkg, "global.appdownurl")),
		"appdownurl2":                  s.callCodeOrEmpty(ctx, prefixed(pkg, "global.appdownurl2")),
		"appdownurl3":                  s.callCodeOrEmpty(ctx, prefixed(pkg, "global.appdownurl3")),
		"adrows":                       s.adRows(ctx, "global.ads", resources),
		"popuptext":                    s.popup(ctx, "global.popup", dotpkg),
		"popuptext_v2":                 s.popup(ctx, "global.popup.v2", dotpkg),
		"popuptext_iOS":                s.popup(ctx, "global.popup.iOS", dotpkg),
		"popuptext_Android":            s.popup(ctx, "global.popup.Android", dotpkg),
		"popuptextnew_iOS":             s.popup(ctx, "global.popupnew.iOS", dotpkg),
		"popuptextlist_iOS":            s.popup(ctx, "global.popuplist.iOS", dotpkg),
		"popuptextlist_Android":        s.popup(ctx, "global.popuplist.Android", dotpkg),
		"popuptext_Android_purify":     s.popup(ctx, "global.popup.Android.purify", dotpkg),
		"popuptextlist_Android_purify": s.popup(ctx, "global.popuplist.Android.purify", dotpkg),
		"popuptextslide_iOS":           s.popup(ctx, "global.popupslide.iOS", dotpkg),
		"popuptextslide_Android":       s.popup(ctx, "global.popupslide.Android", dotpkg),
		"gameslide_iOS":                s.popup(ctx, "global.gameslide.iOS", dotpkg),
		"gameslide_Android":            s.popup(ctx, "global.gameslide.Android", dotpkg),
		"gameslide_AndroidV2":          s.popup(ctx, "global.gameslide.AndroidV2", dotpkg),
		"gameactivity_iOS":             s.popup(ctx, "global.gameactivity.iOS", dotpkg),
		"gameactivity_Android":         s.popup(ctx, "global.gameactivity.Android", dotpkg),
		"qrlink":                       qrlink,
		"newurl":                       chooseLine(str(baseset["newUrls"])),
		"sharetext":                    sharetext,
		"adgroups":                     s.adGroups(ctx, prefixed(pkg, "global.adgroup.all"), "global.adgroup.all", pkg+"_", "", resources),
		"iOS_adgroups":                 s.adGroups(ctx, prefixed(pkg, "iOS.global.adgroup.all"), "iOS.global.adgroup.all", pkg+"_", "iOS.", resources),
		"Android_adgroups":             s.adGroups(ctx, prefixed(pkg, "Android.global.adgroup.all"), "Android.global.adgroup.all", pkg+"_", "Android.", resources),
		"app_launch_times_adshow":      s.callInt(ctx, "global.app.launch.times.adshow", 0),
		"promotion_earn_dscr":          s.mustCallJSON(ctx, "promotion.earn.dscr", nil),
		"app_launch_type_adshow":       s.mustCallJSON(ctx, "global.app.launch.type.adshow", nil),
		"potatolink":                   s.callCodeOrEmpty(ctx, "global.potatolink"),
		"appstorelink":                 s.callCodeOrEmpty(ctx, "global.appstorelink"),
		"appintervaltime":              s.callInt(ctx, "global.app.interval_time", 300),
		"gameDisabled":                 atoi(setting["gameDisabled"]),
		"vodOrderStatus":               atoi(setting["vodOrderStatus"]),
		"videoReportDisabled":          atoi(setting["videoReportDisabled"]),
		"liveStatus":                   atoi(setting["liveStatus"]),
		"onegoStatus":                  atoi(setting["onegoStatus"]),
		"inviteCodeStatus":             atoi(setting["inviteCodeStatus"]),
		"skipAds":                      atoi(setting["skipAds"]),
		"csurl":                        str(setting["csurl"]),
		"sitelogo":                     resources.GetRes(str(setting["sitelogo"]), ""),
		"splashimage":                  resources.GetRes(str(setting["splashimage"]), ""),
		"umengDeduct":                  atoi(setting["umengDeduct"]),
		"h5old":                        str(setting["h5old"]),
		"aiundress":                    str(setting["aiundress"]),
		"aiundressvideo":               str(setting["aiundressvideo"]),
		"aiundressvideox":              str(setting["aiundressvideox"]),
		"aiUndressStatus":              atoi(setting["aiUndressStatus"]),
		"smscaptcha":                   atoi(setting["smscaptcha"]),
		"externalUrlDating":            "",
	}
	return data, nil
}

func (s *GlobalService) callJSON(ctx context.Context, uuid string) (interface{}, error) {
	row, err := s.store.CalldataByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	if str(row["type"]) != "json" && str(row["type"]) != "rows" {
		return nil, nil
	}
	var out interface{}
	if err := json.Unmarshal([]byte(str(row["content"])), &out); err != nil {
		return nil, nil
	}
	return out, nil
}

func (s *GlobalService) mustCallJSON(ctx context.Context, uuid string, fallback interface{}) interface{} {
	out, err := s.callJSON(ctx, uuid)
	if err != nil || out == nil {
		return fallback
	}
	return out
}

func (s *GlobalService) callCodeOrEmpty(ctx context.Context, uuid string) string {
	row, err := s.store.CalldataByUUID(ctx, uuid)
	if err != nil || str(row["type"]) != "code" {
		return ""
	}
	return strings.TrimSpace(str(row["content"]))
}

func (s *GlobalService) callHTMLOrEmpty(ctx context.Context, uuid string) string {
	row, err := s.store.CalldataByUUID(ctx, uuid)
	if err != nil || str(row["type"]) != "html" {
		return ""
	}
	return strings.TrimSpace(str(row["content"]))
}

func (s *GlobalService) callInt(ctx context.Context, uuid string, fallback int) int {
	value := s.callCodeOrEmpty(ctx, uuid)
	if value == "" {
		return fallback
	}
	return atoi(value)
}

func (s *GlobalService) popup(ctx context.Context, base string, suffix string) interface{} {
	if suffix != "" {
		if out := s.mustCallJSON(ctx, base+suffix, nil); out != nil {
			return out
		}
	}
	if out := s.mustCallJSON(ctx, base, nil); out != nil {
		return out
	}
	return s.callHTMLOrEmpty(ctx, base)
}

func (s *GlobalService) adRows(ctx context.Context, uuid string, resources resourceurl.Resolved) []map[string]interface{} {
	raw := s.mustCallJSON(ctx, uuid, []interface{}{})
	list, _ := raw.([]interface{})
	rows := []map[string]interface{}{}
	for _, item := range list {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		out := adRow(row)
		if pic := str(row["pic"]); pic != "" {
			out["pic"] = resources.GetRes(pic, "")
		}
		rows = append(rows, out)
	}
	if len(rows) == 0 {
		return rows
	}
	return []map[string]interface{}{rows[rand.Intn(len(rows))]}
}

func (s *GlobalService) adGroups(ctx context.Context, uuid string, fallbackUUID string, channelPrefix string, platformPrefix string, resources resourceurl.Resolved) interface{} {
	uuids, found := s.adGroupUUIDs(ctx, uuid)
	if !found && fallbackUUID != uuid {
		uuids, _ = s.adGroupUUIDs(ctx, fallbackUUID)
	}
	if len(uuids) == 0 {
		return nil
	}
	out := map[string]interface{}{}
	for _, id := range uuids {
		rows := s.adRows(ctx, id, resources)
		if len(rows) == 0 {
			continue
		}
		key := strings.TrimPrefix(id, channelPrefix)
		key = strings.TrimPrefix(key, platformPrefix)
		out[key] = rows
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (s *GlobalService) adGroupUUIDs(ctx context.Context, uuid string) ([]string, bool) {
	row, err := s.store.CalldataByUUID(ctx, uuid)
	if err != nil || len(row) == 0 {
		return nil, false
	}
	if str(row["type"]) != "code" {
		return nil, true
	}
	return splitCSV(strings.TrimSpace(str(row["content"]))), true
}

func adRow(row map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{"showweight": row["showweight"]}
	for i := 0; i <= 3; i++ {
		title := str(row[fmt.Sprintf("title%d", i)])
		if title != "" {
			out[title] = row[fmt.Sprintf("url%d", i)]
		}
	}
	return out
}

func (s *GlobalService) settingArray(ctx context.Context, uuid string) (map[string]interface{}, error) {
	row, err := s.store.SettingByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return parsePHPSerializedMap(str(row["value"])), nil
}

func parsePHPSerializedMap(value string) map[string]interface{} {
	out := map[string]interface{}{}
	re := regexp.MustCompile(`s:\d+:"([^"]+)";(?:s:\d+:"([^"]*)"|i:(-?\d+)|d:([0-9.]+)|N;)`)
	for _, match := range re.FindAllStringSubmatch(value, -1) {
		key := match[1]
		switch {
		case match[2] != "":
			out[key] = match[2]
		case match[3] != "":
			out[key] = atoi(match[3])
		case match[4] != "":
			out[key] = match[4]
		default:
			out[key] = nil
		}
	}
	return out
}

func firstExisting(primary string, fallback string) string {
	if primary != "" && !strings.HasSuffix(primary, ".") {
		return primary
	}
	return fallback
}

func prefixed(prefix string, uuid string) string {
	if prefix == "" {
		return uuid
	}
	return prefix + "." + uuid
}

func chooseLine(value string) string {
	lines := []string{}
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "-") {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return lines[rand.Intn(len(lines))]
}

func splitCSV(value string) []string {
	out := []string{}
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func (s *GlobalService) resURL(uri string) string {
	if uri == "" || strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return uri
	}
	return s.resourceBaseURL + "/" + strings.TrimLeft(uri, "/")
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
