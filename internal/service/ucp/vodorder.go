package ucp

import (
	"context"
	"time"
)

func (s *Service) VODOrderIndex(ctx context.Context, token string, page int) (map[string]interface{}, int, string, error) {
	user, groups, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	latestIssue, err := s.store.LatestVODIssue(ctx)
	if err != nil {
		return nil, -9999, "暂无求片记录", err
	}
	if len(latestIssue) == 0 {
		return nil, -9999, "暂无求片记录", nil
	}
	settingRow, err := s.store.SettingByUUID(ctx, "setting")
	if err != nil {
		return nil, -9999, "系统配置错误", err
	}
	setting := parseTaskPHPSerializedMap(str(settingRow["value"]))
	periodDays := atoi(setting["vod_order_period"])
	if periodDays == 0 {
		return nil, -9999, "系统配置错误", nil
	}

	start, end, issue := vodOrderIssueRange(s.now(), atoi64(latestIssue["issue"]), periodDays)
	const pageSize = 20
	if _, err := s.store.CountVODOrdersByCreateTime(ctx, start, end); err != nil {
		return nil, -9999, "暂无求片记录", err
	}
	rows, err := s.store.VODOrdersByCreateTime(ctx, start, end, page, pageSize)
	if err != nil {
		return nil, -9999, "暂无求片记录", err
	}
	if len(rows) == 0 {
		return nil, -9999, "暂无求片记录", nil
	}
	processed, err := s.processVODOrderIndexRows(ctx, rows, uid, groups)
	if err != nil {
		return nil, -1, "获取求片榜单失败", err
	}
	return map[string]interface{}{
		"data":  processed,
		"issue": formatVODOrderIssue(issue),
	}, 0, "", nil
}

func (s *Service) VODOrderMyOrders(ctx context.Context, token string, page int) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	const pageSize = 20
	total, err := s.store.CountVODOrders(ctx, uid, nil)
	if err != nil {
		return nil, -1, "获取求片记录失败", err
	}
	rows, err := s.store.VODOrders(ctx, uid, nil, page, pageSize, "id DESC")
	if err != nil {
		return nil, -1, "获取求片记录失败", err
	}
	cost, err := s.store.SumVODOrderCoins(ctx, uid, 1)
	if err != nil {
		return nil, -1, "获取求片记录失败", err
	}
	frozen, err := s.store.SumVODOrderCoins(ctx, uid, 0)
	if err != nil {
		return nil, -1, "获取求片记录失败", err
	}
	supportCost, err := s.store.SumVODSupportCoins(ctx, uid, false)
	if err != nil {
		return nil, -1, "获取求片记录失败", err
	}
	supportFrozen, err := s.store.SumVODSupportCoins(ctx, uid, true)
	if err != nil {
		return nil, -1, "获取求片记录失败", err
	}
	_ = total
	return map[string]interface{}{
		"data":           rows,
		"total_cost":     cost + supportCost,
		"current_frozen": frozen + supportFrozen,
	}, 0, "", nil
}

func (s *Service) VODOrderMySupports(ctx context.Context, token string, page int) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	const pageSize = 20
	if _, err := s.store.CountVODSupports(ctx, uid); err != nil {
		return nil, -1, "获取助力求片记录失败", err
	}
	rows, err := s.store.VODSupports(ctx, uid, page, pageSize)
	if err != nil {
		return nil, -1, "获取助力求片记录失败", err
	}
	return map[string]interface{}{"data": rows}, 0, "", nil
}

func (s *Service) processVODOrderIndexRows(ctx context.Context, rows []map[string]interface{}, uid int, groups []map[string]interface{}) ([]map[string]interface{}, error) {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		item := copyMap(row)
		orderID := atoi(item["id"])
		top, err := s.vodOrderTopUser(ctx, item, orderID, groups)
		if err != nil {
			return nil, err
		}
		item["top"] = top
		mySupport, err := s.store.MyVODSupportCoins(ctx, orderID, uid)
		if err != nil {
			return nil, err
		}
		if atoi(item["uid"]) == uid {
			mySupport = atoi(item["coins"])
		}
		item["my_support"] = mySupport
		out = append(out, item)
	}
	return out, nil
}

func (s *Service) vodOrderTopUser(ctx context.Context, row map[string]interface{}, orderID int, groups []map[string]interface{}) (map[string]interface{}, error) {
	top, err := s.store.MaxVODSupport(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if len(top) == 0 {
		top = map[string]interface{}{"uid": row["uid"], "total_coins": row["coins"]}
	}
	topUserID := atoi(top["uid"])
	result := map[string]interface{}{
		"uid":         top["uid"],
		"total_coins": top["total_coins"],
	}
	if topUserID > 0 {
		user, err := s.store.UserByID(ctx, topUserID)
		if err != nil {
			return nil, err
		}
		if len(user) != 0 {
			processed := singleUser(s.processUsers(ctx, []map[string]interface{}{user}, groups))
			result["username"] = processed["username"]
			result["nickname"] = processed["username"]
			result["avatar"] = processed["avatar_url"]
		}
		return result, nil
	}
	bot, err := s.store.BotByID(ctx, -topUserID)
	if err != nil {
		return nil, err
	}
	if len(bot) != 0 {
		processed := s.processVODOrderBot(ctx, bot)
		result["username"] = processed["username"]
		result["nickname"] = processed["nickname"]
		result["avatar"] = processed["avatar_url"]
	}
	return result, nil
}

func (s *Service) processVODOrderBot(ctx context.Context, row map[string]interface{}) map[string]interface{} {
	avatar := str(row["avatar"])
	avatarURL := avatar
	if avatar == "" || !isNumericString(avatar) {
		avatarURL = s.avatarURL(ctx, avatar)
	}
	return map[string]interface{}{
		"uid":        row["uid"],
		"username":   str(row["username"]),
		"nickname":   str(row["username"]),
		"avatar":     avatar,
		"avatar_url": avatarURL,
	}
}

func vodOrderIssueRange(now time.Time, latestIssue int64, periodDays int) (int64, int64, int64) {
	currentIssue := dayStartUnix(now)
	periodSeconds := int64(periodDays) * 86400
	intervalDays := (currentIssue - latestIssue) / 86400
	currentIssue = latestIssue + (intervalDays/int64(periodDays))*periodSeconds
	nextIssue := currentIssue + periodSeconds
	if now.Unix() < nextIssue {
		previousIssue := currentIssue - periodSeconds
		return previousIssue, currentIssue, previousIssue
	}
	return currentIssue, nextIssue, currentIssue
}

func formatVODOrderIssue(issue int64) string {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return time.Unix(issue, 0).In(loc).Format("060102")
}

func copyMap(row map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(row))
	for key, value := range row {
		out[key] = value
	}
	return out
}

func isNumericString(value string) bool {
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

func (s *Service) VODOrderHistoryOrders(ctx context.Context, token string, page int) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	const pageSize = 20
	status := 1
	if _, err := s.store.CountVODOrders(ctx, 0, &status); err != nil {
		return nil, -1, "获取历史求片失败", err
	}
	rows, err := s.store.VODOrders(ctx, 0, &status, page, pageSize, "")
	if err != nil {
		return nil, -1, "获取历史求片失败", err
	}
	return map[string]interface{}{"data": rows}, 0, "", nil
}
