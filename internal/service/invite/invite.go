package invite

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
)

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type Store interface {
	RecordRecommend(ctx context.Context, uid int) (map[string]interface{}, error)
	UserByInviteKey(ctx context.Context, inviteCode string) (map[string]interface{}, error)
	Groups(ctx context.Context) ([]map[string]interface{}, error)
	SettingByUUID(ctx context.Context, uuid string) (map[string]interface{}, error)
	DeletedUserTag(ctx context.Context, mobi string) (bool, error)
	BindInvite(ctx context.Context, input domain.InviteBindInput) (bool, error)
}

type Service struct {
	auth  AuthStore
	store Store
}

func NewService(auth AuthStore, store Store) *Service {
	return &Service{auth: auth, store: store}
}

func (s *Service) Info(ctx context.Context, token string) (map[string]interface{}, int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return nil, -1, "获取邀请码失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	row, err := s.store.RecordRecommend(ctx, uid)
	if err != nil {
		return nil, -1, "获取邀请码失败", err
	}
	if len(row) == 0 {
		return map[string]interface{}{"data": nil}, 0, "", nil
	}
	key := ""
	if uniqkey := atoi(row["uniqkey"]); uniqkey > 0 {
		key = strconv.FormatInt(int64(uniqkey), 36)
	}
	return map[string]interface{}{"data": key}, 0, "", nil
}

func (s *Service) Bind(ctx context.Context, token string, inviteCode string) (map[string]interface{}, int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return nil, -1, "绑定邀请码失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	row, err := s.store.RecordRecommend(ctx, uid)
	if err != nil {
		return nil, -1, "绑定邀请码失败", err
	}
	if len(row) > 0 {
		key := ""
		if uniqkey := atoi(row["uniqkey"]); uniqkey > 0 {
			key = strconv.FormatInt(int64(uniqkey), 36)
		}
		return nil, -1, "您已经绑定了邀请码:" + key, nil
	}
	inviteCode = strings.TrimSpace(inviteCode)
	if inviteCode == "" {
		return nil, -1, "请输入邀请码", nil
	}
	inviter, err := s.store.UserByInviteKey(ctx, inviteCode)
	if err != nil {
		return nil, -1, "绑定邀请码失败", err
	}
	inviterUID := atoi(inviter["uid"])
	if inviterUID == 0 {
		return nil, -1, "无效邀请码", nil
	}
	if inviterUID == uid {
		return nil, -1, "无法绑定自己", nil
	}
	noReward, err := s.store.DeletedUserTag(ctx, strings.TrimSpace(fmt.Sprint(user["mobi"])))
	if err != nil {
		return nil, -1, "绑定邀请码失败", err
	}
	groups, err := s.store.Groups(ctx)
	if err != nil {
		return nil, -1, "绑定邀请码失败", err
	}
	bonus, err := s.promotionBonus(ctx)
	if err != nil {
		return nil, -1, "绑定邀请码失败", err
	}
	ok, err := s.store.BindInvite(ctx, domain.InviteBindInput{
		UID:        uid,
		InviterUID: inviterUID,
		InviteCode: inviteCode,
		Now:        time.Now().Unix(),
		NoReward:   noReward,
		Bonus:      bonus,
		Groups:     groups,
	})
	if err != nil {
		return nil, -1, "绑定邀请码失败", err
	}
	if !ok {
		return nil, -1, "绑定失败，请重试", nil
	}
	return map[string]interface{}{"data": inviteCode}, 0, "", nil
}

func (s *Service) BindEdge(ctx context.Context, token string, inviteCode string) (int, string, error) {
	_, retcode, errmsg, err := s.Bind(ctx, token, inviteCode)
	return retcode, errmsg, err
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

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}

func inviteBase10(inviteCode string) int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(inviteCode), 36, 64)
	if err != nil {
		return 0
	}
	return value
}

func (s *Service) promotionBonus(ctx context.Context) (map[string]int, error) {
	row, err := s.store.SettingByUUID(ctx, "promotion.bonus")
	if err != nil {
		return nil, err
	}
	raw := fmt.Sprint(row["value"])
	out := map[string]int{}
	re := regexp.MustCompile(`s:\d+:"([^"]+)";(?:i:(-?\d+)|d:([0-9.]+)|s:\d+:"([^"]*)")`)
	for _, match := range re.FindAllStringSubmatch(raw, -1) {
		value := 0
		switch {
		case match[2] != "":
			value = atoi(match[2])
		case match[3] != "":
			value = atoi(match[3])
		case match[4] != "":
			value = atoi(match[4])
		}
		out[match[1]] = value
	}
	return out, nil
}
