package ucp

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

const (
	coinTypeVODShare        = 2
	coinTypeVODComment      = 3
	coinTypeVODFavorite     = 4
	coinTypeVODPlay10       = 5
	coinTypeSaveQRCode      = 6
	coinTypeAdViewClick     = 7
	coinTypeMiniVODDownTask = 22
)

type QRCodeRenderer interface {
	PNG(content string) ([]byte, error)
}

type goQRCodeRenderer struct{}

func (goQRCodeRenderer) PNG(content string) ([]byte, error) {
	qr, err := qrcode.Encode(content, qrcode.Low, 390)
	if err != nil {
		return nil, err
	}
	src, err := png.Decode(bytes.NewReader(qr))
	if err != nil {
		return nil, err
	}
	dst := image.NewRGBA(image.Rect(0, 0, 400, 400))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(5, 5, 395, 395), src, src.Bounds().Min, draw.Src)
	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *Service) TaskIndex(ctx context.Context, token string) (map[string]interface{}, int, string, error) {
	user, groups, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -1, "获取任务中心失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	daytime := dayStartUnix(s.now())
	perms := user["perms"]

	share, err := s.coinTaskStat(ctx, uid, daytime, coinTypeVODShare, "max.goldcoin.share.num", "max.goldcoin.share.limit", perms)
	if err != nil {
		return nil, -1, "获取任务中心失败", err
	}
	comment, err := s.commentTaskStat(ctx, uid, daytime, perms)
	if err != nil {
		return nil, -1, "获取任务中心失败", err
	}
	favorite, err := s.favoriteTaskStat(ctx, uid, daytime, perms)
	if err != nil {
		return nil, -1, "获取任务中心失败", err
	}
	play10, err := s.playTaskStat(ctx, uid, daytime, perms)
	if err != nil {
		return nil, -1, "获取任务中心失败", err
	}
	saveqrcode, err := s.coinTaskStat(ctx, uid, daytime, coinTypeSaveQRCode, "max.goldcoin.saveqrcode.num", "max.goldcoin.saveqrcode.num", perms)
	if err != nil {
		return nil, -1, "获取任务中心失败", err
	}
	adviewclick, err := s.coinTaskStat(ctx, uid, daytime, coinTypeAdViewClick, "max.goldcoin.adviewclick.num", "max.goldcoin.adviewclick.num", perms)
	if err != nil {
		return nil, -1, "获取任务中心失败", err
	}
	minivoddown, err := s.coinTaskStat(ctx, uid, daytime, coinTypeMiniVODDownTask, "max.goldcoin.minivod.down.coinnum", "max.goldcoin.minivod.down.limit", perms)
	if err != nil {
		return nil, -1, "获取任务中心失败", err
	}

	return map[string]interface{}{
		"user":        singleUser(s.processUsers([]map[string]interface{}{user}, groups)),
		"share":       share,
		"comment":     comment,
		"favorite":    favorite,
		"play10":      play10,
		"saveqrcode":  saveqrcode,
		"adviewclick": adviewclick,
		"minivoddown": minivoddown,
	}, 0, "", nil
}

func (s *Service) coinTaskStat(ctx context.Context, uid int, daytime int64, coinType int, dayKey string, limitKey string, perms interface{}) (map[string]interface{}, error) {
	coinnum, err := s.store.SumCoinLogsSinceByType(ctx, uid, coinType, daytime)
	if err != nil {
		return nil, err
	}
	donenum, err := s.store.CountCoinLogsSinceByType(ctx, uid, coinType, daytime)
	if err != nil {
		return nil, err
	}
	return taskStat(getPermInt(perms, dayKey), getPermInt(perms, limitKey), coinnum, donenum), nil
}

func (s *Service) commentTaskStat(ctx context.Context, uid int, daytime int64, perms interface{}) (map[string]interface{}, error) {
	coinnum, err := s.store.SumCoinLogsSinceByType(ctx, uid, coinTypeVODComment, daytime)
	if err != nil {
		return nil, err
	}
	donenum, err := s.store.CountVODCommentsSince(ctx, uid, daytime, true)
	if err != nil {
		return nil, err
	}
	return taskStat(getPermInt(perms, "max.goldcoin.comment.num"), getPermInt(perms, "max.goldcoin.comment.limit"), coinnum, donenum), nil
}

func (s *Service) favoriteTaskStat(ctx context.Context, uid int, daytime int64, perms interface{}) (map[string]interface{}, error) {
	coinnum, err := s.store.SumCoinLogsSinceByType(ctx, uid, coinTypeVODFavorite, daytime)
	if err != nil {
		return nil, err
	}
	donenum, err := s.store.CountVODFavoritesSince(ctx, uid, daytime)
	if err != nil {
		return nil, err
	}
	return taskStat(getPermInt(perms, "max.goldcoin.favorite.num"), getPermInt(perms, "max.goldcoin.favorite.limit"), coinnum, donenum), nil
}

func (s *Service) playTaskStat(ctx context.Context, uid int, daytime int64, perms interface{}) (map[string]interface{}, error) {
	coinnum, err := s.store.SumCoinLogsSinceByType(ctx, uid, coinTypeVODPlay10, daytime)
	if err != nil {
		return nil, err
	}
	donenum, err := s.store.CountVODPlayLogsSince(ctx, uid, daytime)
	if err != nil {
		return nil, err
	}
	return taskStat(getPermInt(perms, "max.goldcoin.play10.num"), getPermInt(perms, "max.goldcoin.play10.limit"), coinnum, donenum), nil
}

func taskStat(daynum int, limit int, coinnum int, donenum int) map[string]interface{} {
	return map[string]interface{}{
		"daynum":  daynum,
		"limit":   limit,
		"coinnum": coinnum,
		"donenum": donenum,
	}
}

func (s *Service) TaskSharePic(ctx context.Context) (map[string]interface{}, error) {
	rows, err := s.store.Posters(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]interface{}{"data": []interface{}{}}, nil
	}
	return map[string]interface{}{"data": rows[rand.Intn(len(rows))]}, nil
}

func (s *Service) TaskQRLink(ctx context.Context, token string, pid string) (map[string]interface{}, int, string, error) {
	return s.taskQRLink(ctx, token, pid, "global.qrcode.link")
}

func (s *Service) TaskInvite(ctx context.Context, token string) (int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	return 0, "", nil
}

func (s *Service) TaskShare(ctx context.Context, token string, pid string) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -1, "获取分享文案失败", err
	}
	pid = sanitizeTaskPID(pid)
	sharetext, err := s.taskCallHTML(ctx, pid, "global.share.text")
	if err != nil {
		return nil, -1, "获取分享文案失败", err
	}
	inviteCode := randomInviteCode(4)
	data := map[string]interface{}{}
	uid := atoi(user["uid"])
	if uid > 0 {
		inviteCode = strings.ToUpper(taskBase36(atoi(user["uniqkey"])))
		addCoin := getPermInt(user["perms"], "max.goldcoin.share.num")
		maxCoin := getPermInt(user["perms"], "max.goldcoin.share.limit")
		sentCoin, err := s.store.SumCoinLogsSinceByType(ctx, uid, coinTypeVODShare, dayStartUnix(s.now()))
		if err != nil {
			return nil, -1, "获取分享文案失败", err
		}
		if sentCoin+addCoin >= maxCoin {
			addCoin = maxCoin - sentCoin
		}
		if addCoin > 0 {
			if err := s.store.AwardCoins(ctx, uid, coinTypeVODShare, addCoin, s.now().Unix(), ""); err != nil {
				return nil, -1, "获取分享文案失败", err
			}
			data["taskdone"] = addCoin
		}
	}
	inviteURL, err := s.taskInviteURL(ctx)
	if err != nil {
		return nil, -1, "获取分享文案失败", err
	}
	sharetext = strings.ReplaceAll(sharetext, "{inviteUrl}", inviteURL)
	sharetext = strings.ReplaceAll(sharetext, "{inviteCode}", inviteCode)
	data["sharetext"] = sharetext
	return data, 0, "", nil
}

func (s *Service) TaskQRCode(ctx context.Context, token string, pid string) ([]byte, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	key := "task.qrcode." + fmt.Sprint(uid) + "." + taskYMD(s.now())
	if err := s.store.SetKeylimit(ctx, key, 1, "", s.now().Unix()); err != nil {
		return nil, -1, "生成二维码失败", err
	}
	data, retcode, errmsg, err := s.TaskQRLink(ctx, token, pid)
	if err != nil || retcode != 0 {
		return nil, retcode, errmsg, err
	}
	body, err := s.qrRenderer.PNG(str(data["qrlink"]))
	if err != nil {
		return nil, -1, "生成二维码失败", err
	}
	return body, 0, "", nil
}

func (s *Service) TaskboxQRLink(ctx context.Context, token string, pid string) (map[string]interface{}, int, string, error) {
	return s.taskQRLink(ctx, token, pid, "taskbox.qrcode.link")
}

func (s *Service) TaskboxQRCode(ctx context.Context, token string, pid string) ([]byte, int, string, error) {
	data, retcode, errmsg, err := s.TaskboxQRLink(ctx, token, pid)
	if err != nil || retcode != 0 {
		return nil, retcode, errmsg, err
	}
	link := str(data["qrlink"])
	body, err := s.qrRenderer.PNG(link)
	if err != nil {
		return nil, -1, "生成任务宝箱二维码失败", err
	}
	return body, 0, "", nil
}

func (s *Service) TaskboxShare(ctx context.Context, token string, pid string) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -1, "获取任务宝箱分享文案失败", err
	}
	pid = sanitizeTaskPID(pid)
	sharetext, err := s.taskCallHTML(ctx, pid, "taskbox.share.text")
	if err != nil {
		return nil, -1, "获取任务宝箱分享文案失败", err
	}
	inviteURL, err := s.taskInviteURL(ctx)
	if err != nil {
		return nil, -1, "获取任务宝箱分享文案失败", err
	}
	inviteCode := randomInviteCode(4)
	if atoi(user["uid"]) > 0 {
		inviteCode = strings.ToUpper(taskBase36(atoi(user["uniqkey"])))
	}
	sharetext = strings.ReplaceAll(sharetext, "{inviteUrl}", inviteURL)
	sharetext = strings.ReplaceAll(sharetext, "{inviteCode}", inviteCode)
	return map[string]interface{}{"sharetext": sharetext}, 0, "", nil
}

func (s *Service) taskQRLink(ctx context.Context, token string, pid string, uuid string) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	pid = sanitizeTaskPID(pid)
	url, err := s.taskCallCode(ctx, pid, uuid)
	if err != nil {
		return nil, -1, "获取二维码链接失败", err
	}
	inviteURL, err := s.taskInviteURL(ctx)
	if err != nil {
		return nil, -1, "获取二维码链接失败", err
	}
	url = strings.ReplaceAll(url, "{inviteUrl}", inviteURL)
	url = strings.ReplaceAll(url, "{inviteCode}", strings.ToUpper(taskBase36(atoi(user["uniqkey"]))))
	return map[string]interface{}{"qrlink": url}, 0, "", nil
}

func (s *Service) TaskboxIndex(ctx context.Context, token string) (map[string]interface{}, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, err
	}
	uid := atoi(user["uid"])
	now := s.now()
	nowUnix := now.Unix()
	dayKeyDaily, dayKeyWeekly, weekday, startTime := taskboxTimes(now)
	recommendTotal := 0
	if uid > 0 {
		recommendTotal = atoi(user["recommend_total"])
	}
	taskRows, err := s.store.Taskboxes(ctx)
	if err != nil {
		return nil, err
	}
	processedTasks := s.processTaskboxRows(taskRows)
	for _, task := range processedTasks {
		taskID := atoi(task["taskid"])
		dayKey := 0
		ready := false
		switch taskID {
		case 1022:
			dayKey = dayKeyDaily
			ready = nowUnix >= startTime && nowUnix < startTime+300
		case 1622:
			dayKey = dayKeyWeekly
			ready = nowUnix >= startTime && nowUnix < startTime+300 && weekday == 6
		default:
			ready = recommendTotal >= taskID
		}
		logRow, err := s.store.TaskboxLog(ctx, uid, taskID, dayKey)
		if err != nil {
			return nil, err
		}
		if len(logRow) != 0 {
			task["taskstatus"] = 2
		} else if ready {
			task["taskstatus"] = 1
		} else {
			task["taskstatus"] = 0
		}
	}
	logRows, err := s.store.TaskboxCompletedLogs(ctx, 30)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"taskrows": processedTasks,
		"logrows":  s.processTaskboxLogRows(logRows),
	}, nil
}

func sanitizeTaskPID(pid string) string {
	pid = strings.TrimSpace(pid)
	if pid == "" {
		return ""
	}
	if regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(pid) {
		return pid
	}
	return ""
}

func (s *Service) taskCallCode(ctx context.Context, pid string, uuid string) (string, error) {
	row, err := s.store.CalldataByUUID(ctx, taskUUID(pid, uuid))
	if err != nil {
		return "", err
	}
	if pid != "" && len(row) == 0 {
		row, err = s.store.CalldataByUUID(ctx, uuid)
		if err != nil {
			return "", err
		}
	}
	if str(row["type"]) != "code" {
		return "", nil
	}
	return strings.TrimSpace(str(row["content"])), nil
}

func (s *Service) taskCallHTML(ctx context.Context, pid string, uuid string) (string, error) {
	row, err := s.store.CalldataByUUID(ctx, taskUUID(pid, uuid))
	if err != nil {
		return "", err
	}
	if pid != "" && len(row) == 0 {
		row, err = s.store.CalldataByUUID(ctx, uuid)
		if err != nil {
			return "", err
		}
	}
	if str(row["type"]) != "html" {
		return "", nil
	}
	return strings.TrimSpace(str(row["content"])), nil
}

func (s *Service) taskInviteURL(ctx context.Context) (string, error) {
	row, err := s.store.SettingByUUID(ctx, "baseset")
	if err != nil {
		return "", err
	}
	setting := parseTaskPHPSerializedMap(str(row["value"]))
	groups := dailyInviteURLGroups(str(setting["inviteUrls"]))
	if len(groups) == 0 {
		return "", nil
	}
	day := s.now().Day() - 1
	selected := groups[0]
	if day >= 0 && day < len(groups) && len(groups[day]) > 0 {
		selected = groups[day]
	}
	if len(selected) == 0 {
		return "", nil
	}
	return selected[rand.Intn(len(selected))], nil
}

func taskUUID(pid string, uuid string) string {
	if pid == "" {
		return uuid
	}
	return pid + "." + uuid
}

func dailyInviteURLGroups(raw string) [][]string {
	groups := [][]string{{}}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "-") {
			groups = append(groups, []string{})
			continue
		}
		groups[len(groups)-1] = append(groups[len(groups)-1], line)
	}
	out := groups[:0]
	for _, group := range groups {
		if len(group) > 0 {
			out = append(out, group)
		}
	}
	return out
}

func parseTaskPHPSerializedMap(value string) map[string]interface{} {
	out := map[string]interface{}{}
	for i := 0; i < len(value); {
		key, next, ok := parsePHPSerializedString(value, i)
		if !ok {
			i++
			continue
		}
		i = next
		switch {
		case strings.HasPrefix(value[i:], "s:"):
			val, n, ok := parsePHPSerializedString(value, i)
			if !ok {
				continue
			}
			out[key] = val
			i = n
		case strings.HasPrefix(value[i:], "i:"):
			end := strings.IndexByte(value[i:], ';')
			if end < 0 {
				continue
			}
			out[key] = atoi(value[i+2 : i+end])
			i += end + 1
		case strings.HasPrefix(value[i:], "d:"):
			end := strings.IndexByte(value[i:], ';')
			if end < 0 {
				continue
			}
			out[key] = value[i+2 : i+end]
			i += end + 1
		case strings.HasPrefix(value[i:], "N;"):
			out[key] = nil
			i += 2
		}
	}
	return out
}

func parsePHPSerializedString(value string, start int) (string, int, bool) {
	if start < 0 || start >= len(value) || !strings.HasPrefix(value[start:], "s:") {
		return "", start, false
	}
	lengthStart := start + 2
	lengthEnd := strings.IndexByte(value[lengthStart:], ':')
	if lengthEnd < 0 {
		return "", start, false
	}
	lengthEnd += lengthStart
	size := atoi(value[lengthStart:lengthEnd])
	dataStart := lengthEnd + 2
	if lengthEnd+1 >= len(value) || value[lengthEnd+1] != '"' || dataStart+size > len(value) {
		return "", start, false
	}
	dataEnd := dataStart + size
	if dataEnd+1 >= len(value) || value[dataEnd] != '"' || value[dataEnd+1] != ';' {
		end := strings.Index(value[dataStart:], `";`)
		if end < 0 {
			return "", start, false
		}
		dataEnd = dataStart + end
	}
	return value[dataStart:dataEnd], dataEnd + 2, true
}

func taskBase36(value int) string {
	if value <= 0 {
		return "0"
	}
	return strconv.FormatInt(int64(value), 36)
}

func randomInviteCode(size int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if size <= 0 {
		return ""
	}
	out := make([]byte, size)
	for i := range out {
		out[i] = chars[rand.Intn(len(chars))]
	}
	return string(out)
}

func (s *Service) TaskboxLogListing(ctx context.Context, token string, page int) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	const pageSize = 20
	total, err := s.store.CountTaskboxLogs(ctx, uid)
	if err != nil {
		return nil, -1, "获取任务宝箱日志失败", err
	}
	rows, err := s.store.TaskboxLogs(ctx, uid, page, pageSize)
	if err != nil {
		return nil, -1, "获取任务宝箱日志失败", err
	}
	return map[string]interface{}{
		"logrows":  s.processTaskboxLogRows(rows),
		"pageinfo": pageInfo(total, pageSize, page, "/ucp/taskbox/taskboxlog?page=[?]"),
	}, 0, "", nil
}

func (s *Service) processTaskboxRows(rows []map[string]interface{}) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, row := range rows {
		taskID := atoi(row["taskid"])
		result = append(result, map[string]interface{}{
			"taskid":   str(row["taskid"]),
			"taskname": str(row["taskname"]),
			"showtype": str(row["showtype"]),
			"tasktype": taskboxType(taskID),
		})
	}
	return result
}

func (s *Service) processTaskboxLogRows(rows []map[string]interface{}) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, row := range rows {
		result = append(result, map[string]interface{}{
			"logid":      str(row["logid"]),
			"username":   nullableValue(row["username"]),
			"nickname":   nullableValue(row["nickname"]),
			"avatar_url": s.avatarURL(str(row["avatar"])),
			"addtime":    formatUnix(atoi64(row["addtime"])),
			"tasktype":   taskboxType(atoi(row["taskid"])),
			"addcoin":    str(row["addcoin"]),
			"prize":      str(row["prize"]),
			"taskstatus": taskboxStatus(atoi(row["taskstatus"])),
		})
	}
	return result
}

func nullableValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	return value
}

func taskboxTimes(now time.Time) (int, int, int, int64) {
	loc := now.Location()
	dayKeyDaily, _ := strconv.Atoi(now.Format("060102"))
	_, week := now.ISOWeek()
	dayKeyWeekly, _ := strconv.Atoi(now.Format("06") + fmt.Sprintf("%02d", week))
	start := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, loc).Unix()
	return dayKeyDaily, dayKeyWeekly, int(now.Weekday()+6)%7 + 1, start
}

func taskboxType(taskID int) string {
	values := map[int]string{
		1:    "推广1人",
		3:    "推广3人",
		5:    "推广5人",
		15:   "推广15人",
		30:   "推广30人",
		50:   "推广50人",
		1022: "每日神秘宝箱",
		1622: "每周神秘宝箱",
	}
	return values[taskID]
}

func taskboxStatus(status int) string {
	values := map[int]string{0: "无效", 1: "待发放", 2: "已发放"}
	return values[status]
}
