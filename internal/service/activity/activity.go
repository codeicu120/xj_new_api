package activity

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	userRepo "xj_comp/internal/repository/user"
)

type Store interface {
	CurrentActivities(ctx context.Context, now int64, page int) ([]map[string]interface{}, error)
	ActivityByID(ctx context.Context, id int) (map[string]interface{}, error)
	PrizesByActivityID(ctx context.Context, id int) ([]map[string]interface{}, error)
	PrizeLogs(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error)
	CountActivityRecords(ctx context.Context, aid int) (int, error)
	ActivityRecords(ctx context.Context, aid int, page int, pageSize int) ([]map[string]interface{}, error)
	ActivityRanking(ctx context.Context, aid int, uid int) (map[string]interface{}, error)
	BotByID(ctx context.Context, uid int) (map[string]interface{}, error)
	CountRecommendedUsers(ctx context.Context, recommenderUID int, start int64, end int64) (int, error)
	RecommendedUsers(ctx context.Context, recommenderUID int, start int64, end int64, page int, pageSize int) ([]map[string]interface{}, error)
	UserGroups(ctx context.Context) ([]map[string]interface{}, error)
}

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type Service struct {
	store           Store
	auth            AuthStore
	resourceBaseURL string
	now             func() time.Time
}

func NewService(store Store, auth AuthStore, resourceBaseURL string) *Service {
	return &Service{store: store, auth: auth, resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"), now: time.Now}
}

func (s *Service) LuckyPrizes() map[string]interface{} {
	return map[string]interface{}{
		"data": []map[string]interface{}{
			{"keyid": "prize.vip.365", "prizename": "365天VIP"},
			{"keyid": "prize.vip.180", "prizename": "180天VIP"},
			{"keyid": "prize.vip.90", "prizename": "90天VIP"},
			{"keyid": "prize.vip.60", "prizename": "60天VIP"},
			{"keyid": "prize.vip.30", "prizename": "30天VIP"},
		},
	}
}

func (s *Service) NewYear2020() (int, string) {
	if s.now().After(time.Date(2020, 2, 8, 23, 59, 59, 0, time.Local)) {
		return -1, "抽奖活动已结束，谢谢支持"
	}
	return 0, ""
}

func (s *Service) LuckyDraw() (int, string) {
	if s.now().After(time.Date(2023, 2, 28, 23, 59, 59, 0, time.Local)) {
		return -1, "抽奖活动已结束，谢谢支持"
	}
	return 0, ""
}

func (s *Service) Index(ctx context.Context, page int) (map[string]interface{}, int, string, error) {
	rows, err := s.store.CurrentActivities(ctx, s.now().Unix(), page)
	if err != nil {
		return nil, -1, "获取活动信息失败", err
	}
	if len(rows) == 0 {
		return nil, -9999, "当前没有进行中的活动", nil
	}
	return map[string]interface{}{"data": rows}, 0, "", nil
}

func (s *Service) Details(ctx context.Context, aid int) (map[string]interface{}, int, string, error) {
	activity, err := s.store.ActivityByID(ctx, aid)
	if err != nil {
		return nil, -1, "获取活动信息失败", err
	}
	if len(activity) == 0 {
		return nil, -9999, "获取活动信息失败", nil
	}
	prizes, err := s.store.PrizesByActivityID(ctx, aid)
	if err != nil {
		return nil, -1, "获取活动信息失败", err
	}
	return map[string]interface{}{
		"data": map[string]interface{}{
			"activity":       activity,
			"activity_prize": processPrizes(prizes),
		},
	}, 0, "", nil
}

func (s *Service) LuckyDrawHistory(ctx context.Context, token string, page int) (map[string]interface{}, int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return nil, -1, "获取抽奖历史失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "请登录后操作", nil
	}
	rows, err := s.store.PrizeLogs(ctx, uid, page, 20)
	if err != nil {
		return nil, -1, "获取抽奖历史失败", err
	}
	names := map[string]string{
		"prize.vip.365": "365天VIP",
		"prize.vip.180": "180天VIP",
		"prize.vip.90":  "90天VIP",
		"prize.vip.60":  "60天VIP",
		"prize.vip.30":  "30天VIP",
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		item := clone(row)
		item["prizename"] = names[fmt.Sprint(item["keyid"])]
		out = append(out, item)
	}
	return map[string]interface{}{"data": out}, 0, "", nil
}

func (s *Service) Ranking(ctx context.Context, token string, aid int, page int) (map[string]interface{}, int, string, error) {
	if _, retcode, errmsg, err := s.loggedInUser(ctx, token); retcode != 0 || err != nil {
		return nil, retcode, errmsg, err
	}
	activity, prizes, retcode, errmsg, err := s.activityAndPrizes(ctx, aid)
	if retcode != 0 || err != nil {
		return nil, retcode, errmsg, err
	}
	total, err := s.store.CountActivityRecords(ctx, aid)
	if err != nil {
		return nil, -1, "获取活动排名失败", err
	}
	prizeUsers := atoi(activity["prize_users"])
	if total > prizeUsers {
		total = prizeUsers
	}
	rows, err := s.store.ActivityRecords(ctx, aid, page, 20)
	if err != nil {
		return nil, -1, "获取活动排名失败", err
	}
	if len(rows) > total {
		rows = rows[:total]
	}
	processed, err := s.processActivityRecords(ctx, rows, prizes)
	if err != nil {
		return nil, -1, "获取活动排名失败", err
	}
	return map[string]interface{}{"data": processed}, 0, "", nil
}

func (s *Service) Receive(ctx context.Context, token string, aid int) (map[string]interface{}, int, string, error) {
	user, retcode, errmsg, err := s.loggedInUser(ctx, token)
	if retcode != 0 || err != nil {
		return nil, retcode, errmsg, err
	}
	activity, prizes, retcode, errmsg, err := s.activityAndPrizes(ctx, aid)
	if retcode != 0 || err != nil {
		return nil, retcode, errmsg, err
	}
	if s.now().Unix() > int64(atoi(activity["reward_expire_time"])) {
		return nil, -9999, "超过该活动领奖截止日期", nil
	}
	rank, err := s.store.ActivityRanking(ctx, aid, atoi(user["uid"]))
	if err != nil {
		return nil, -1, "获取排名信息失败", err
	}
	if len(rank) == 0 {
		return nil, -9999, "获取排名信息失败", nil
	}
	if atoi(rank["ranking"]) > atoi(activity["prize_users"]) {
		return nil, -9999, "很遗憾，您未中奖", nil
	}
	prizeLevel := ""
	prizeMoney := interface{}(0)
	for _, prize := range prizes {
		if atoi(rank["ranking"]) <= atoi(prize["ranking"]) {
			prizeLevel = fmt.Sprint(prize["level"])
			prizeMoney = prize["prize"]
		}
	}
	out := clone(rank)
	out["prize_level"] = prizeLevel
	out["prize_money"] = prizeMoney
	return map[string]interface{}{"data": out}, 0, "", nil
}

func (s *Service) Recommends(ctx context.Context, token string, aid int, page int) (map[string]interface{}, int, string, error) {
	user, retcode, errmsg, err := s.loggedInUser(ctx, token)
	if retcode != 0 || err != nil {
		return nil, retcode, errmsg, err
	}
	activity, err := s.store.ActivityByID(ctx, aid)
	if err != nil {
		return nil, -1, "获取活动信息失败", err
	}
	if len(activity) == 0 {
		return nil, -9999, "获取活动信息失败", nil
	}
	total, err := s.store.CountRecommendedUsers(ctx, atoi(user["uid"]), int64(atoi(activity["effect_time"])), int64(atoi(activity["expire_time"])))
	if err != nil {
		return nil, -1, "获取邀请记录失败", err
	}
	rows, err := s.store.RecommendedUsers(ctx, atoi(user["uid"]), int64(atoi(activity["effect_time"])), int64(atoi(activity["expire_time"])), page, 20)
	if err != nil {
		return nil, -1, "获取邀请记录失败", err
	}
	groups, err := s.store.UserGroups(ctx)
	if err != nil {
		return nil, -1, "获取邀请记录失败", err
	}
	return map[string]interface{}{
		"data":  s.processUserRows(rows, groups),
		"total": total,
	}, 0, "", nil
}

func (s *Service) loggedInUser(ctx context.Context, token string) (map[string]interface{}, int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return nil, -1, "获取用户失败", err
	}
	if atoi(user["uid"]) == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	return user, 0, "", nil
}

func (s *Service) activityAndPrizes(ctx context.Context, aid int) (map[string]interface{}, []map[string]interface{}, int, string, error) {
	activity, err := s.store.ActivityByID(ctx, aid)
	if err != nil {
		return nil, nil, -1, "获取活动信息失败", err
	}
	if len(activity) == 0 {
		return nil, nil, -9999, "获取活动信息失败", nil
	}
	prizes, err := s.store.PrizesByActivityID(ctx, aid)
	if err != nil {
		return nil, nil, -1, "获取活动信息失败", err
	}
	return activity, prizes, 0, "", nil
}

func (s *Service) processActivityRecords(ctx context.Context, rows []map[string]interface{}, prizes []map[string]interface{}) ([]map[string]interface{}, error) {
	out := make([]map[string]interface{}, 0, len(rows))
	for index, row := range rows {
		prizeLevel := ""
		prizeMoney := interface{}(0)
		rank := index + 1
		for _, prize := range prizes {
			if rank <= atoi(prize["ranking"]) {
				prizeLevel = fmt.Sprint(prize["level"])
				prizeMoney = prize["prize"]
				break
			}
		}
		username := row["username"]
		avatar := row["avatar"]
		if atoi(row["uid"]) < 0 {
			bot, err := s.store.BotByID(ctx, -atoi(row["uid"]))
			if err != nil {
				return nil, err
			}
			username = bot["username"]
			avatar = bot["avatar"]
		}
		out = append(out, map[string]interface{}{
			"id":          row["id"],
			"aid":         row["aid"],
			"uid":         row["uid"],
			"username":    username,
			"avatar":      avatar,
			"score":       row["score"],
			"received":    row["received"],
			"prize_level": prizeLevel,
			"prize_money": prizeMoney,
			"create_time": row["create_time"],
			"update_time": row["update_time"],
		})
	}
	return out, nil
}

func (s *Service) processUserRows(rows []map[string]interface{}, groups []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	now := s.now().Unix()
	for _, row := range rows {
		sysgidExptime := int64(atoi(row["sysgid_exptime"]))
		duetime := ""
		dueday := ""
		if sysgidExptime > 0 {
			duetime = time.Unix(sysgidExptime, 0).Format("2006-01-02 15:04:05")
			if sysgidExptime > now {
				dueday = "过期"
			} else {
				dueday = "已过期"
			}
		}
		out = append(out, map[string]interface{}{
			"uid":             row["uid"],
			"uniqkey":         strings.ToUpper(base36(atoi(row["uniqkey"]))),
			"username":        row["username"],
			"nickname":        row["nickname"],
			"mobi":            row["mobi"],
			"email":           row["email"],
			"sysgid":          row["sysgid"],
			"gid":             row["gid"],
			"gids":            nil,
			"gicon":           groupIcon(row["sysgid"], row["gid"], groups),
			"isvip":           vipFlag(row, now),
			"regtime":         unixDateTime(row["regtime"]),
			"gender":          atoi(row["gender"]),
			"avatar":          row["avatar"],
			"avatar_url":      s.avatarURL(fmt.Sprint(row["avatar"])),
			"newmsg":          row["newmsg"],
			"goldcoin":        atoi(row["goldcoin"]),
			"gold_bean":       atoi(row["gold_bean"]),
			"duetime":         duetime,
			"dueday":          dueday,
			"recommend_total": atoi(row["recommend_total"]),
		})
	}
	return out
}

func groupIcon(sysgidValue interface{}, gidValue interface{}, groups []map[string]interface{}) string {
	gid := atoi(gidValue)
	if atoi(sysgidValue) > 0 {
		gid = atoi(sysgidValue)
	}
	for _, group := range groups {
		if atoi(group["gid"]) == gid {
			return fmt.Sprint(group["gicon"])
		}
	}
	return ""
}

func vipFlag(row map[string]interface{}, now int64) int {
	if atoi(row["sysgid"]) == 6 && int64(atoi(row["sysgid_exptime"])) > now {
		return 1
	}
	return 0
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

func unixDateTime(value interface{}) string {
	n := int64(atoi(value))
	if n <= 0 {
		return "1970-01-01 00:00:00"
	}
	return time.Unix(n, 0).Format("2006-01-02 15:04:05")
}

func base36(value int) string {
	if value <= 0 {
		return "0"
	}
	return strconv.FormatInt(int64(value), 36)
}

func (s *Service) userByToken(ctx context.Context, token string) (map[string]interface{}, error) {
	sid := userRepo.CleanToken(token)
	if sid == "" || s.auth == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	user, err := s.auth.UserBySession(ctx, sid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	return user, nil
}

func processPrizes(rows []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	prevRank := 1
	for _, row := range rows {
		rankingValue := atoi(row["ranking"])
		prizeUsers := rankingValue
		var ranking interface{} = row["ranking"]
		if rankingValue > prevRank {
			prizeUsers = rankingValue - prevRank
			ranking = fmt.Sprintf("%d-%d", prevRank+1, rankingValue)
			prevRank = rankingValue
		}
		out = append(out, map[string]interface{}{
			"id":          row["id"],
			"aid":         row["aid"],
			"level":       row["level"],
			"ranking":     ranking,
			"prize":       row["prize"],
			"prize_users": fmt.Sprint(prizeUsers),
		})
	}
	return out
}

func clone(row map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(row)+1)
	for key, value := range row {
		out[key] = value
	}
	return out
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
