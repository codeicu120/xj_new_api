package ucp

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

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

func (s *Service) TaskboxQRLink(ctx context.Context, token string, pid string) (map[string]interface{}, int, string, error) {
	return s.taskQRLink(ctx, token, pid, "taskbox.qrcode.link")
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

func taskBase36(value int) string {
	if value <= 0 {
		return "0"
	}
	return strconv.FormatInt(int64(value), 36)
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
