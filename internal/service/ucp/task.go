package ucp

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
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
