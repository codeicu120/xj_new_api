package stats

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"strings"
	"time"

	userRepo "xj_comp/internal/repository/user"
)

type Store interface {
	GuestExists(ctx context.Context, sid string) (bool, error)
	ShortcutCreatedByIP(ctx context.Context, ip string) (map[string]interface{}, error)
	CreateShortcut(ctx context.Context, ip string, now int64) (int64, error)
	ShortcutStatsByDate(ctx context.Context, statsDate int64) (map[string]interface{}, error)
	CreateShortcutStats(ctx context.Context, statsDate int64) error
	UpdateShortcutStatsCount(ctx context.Context, id int, count int) error
	AdStatBySID(ctx context.Context, sid string, title string, url string) (map[string]interface{}, error)
	CreateAdStat(ctx context.Context, sid string, title string, url string, pos int, click int, install int) error
	UpdateAdStat(ctx context.Context, id int, click int, install int) error
	PlayStatBySID(ctx context.Context, sid string, vid int) (map[string]interface{}, error)
	CreatePlayStat(ctx context.Context, sid string, vid int, mini int, duration int, played int) error
	UpdatePlayStatPlayed(ctx context.Context, id int, played int) error
	CreateGuest(ctx context.Context, sid string, now int64) error
}

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type Service struct {
	store Store
	auth  AuthStore
	now   func() time.Time
}

type AdInput struct {
	Title   string
	URL     string
	Pos     int
	Click   int
	Install int
}

type PlayInput struct {
	VID      int
	Mini     int
	Duration int
	Played   int
}

func NewService(store Store, auth AuthStore) *Service {
	return &Service{store: store, auth: auth, now: time.Now}
}

func (s *Service) ShortcutAdd(ctx context.Context, ip string) error {
	today := startOfDay(s.now()).Unix()
	row, err := s.store.ShortcutCreatedByIP(ctx, ip)
	if err != nil {
		return err
	}
	if len(row) != 0 {
		return nil
	}
	id, err := s.store.CreateShortcut(ctx, ip, s.now().Unix())
	if err != nil {
		return err
	}
	if id <= 0 {
		return nil
	}
	stats, err := s.store.ShortcutStatsByDate(ctx, today)
	if err != nil {
		return err
	}
	if len(stats) == 0 {
		return s.store.CreateShortcutStats(ctx, today)
	}
	return s.store.UpdateShortcutStatsCount(ctx, atoi(stats["id"]), atoi(stats["count"])+1)
}

func (s *Service) AdAdd(ctx context.Context, token string, ip string, input AdInput) (int, string, error) {
	sid, retcode, errmsg, err := s.sid(ctx, token, ip)
	if err != nil || retcode != 0 {
		return retcode, errmsg, err
	}
	input.Title = strings.TrimSpace(input.Title)
	input.URL = strings.TrimSpace(input.URL)
	if input.Title == "" || input.URL == "" {
		return -9999, "缺少参数", nil
	}
	if input.Click <= 0 || input.Pos <= 0 {
		return -9999, "无效参数", nil
	}
	if input.Install >= 1 {
		input.Install = 1
		input.Click = 1
	}
	row, err := s.store.AdStatBySID(ctx, sid, input.Title, input.URL)
	if err != nil {
		return -1, "保存广告统计失败", err
	}
	if len(row) == 0 {
		err = s.store.CreateAdStat(ctx, sid, input.Title, input.URL, input.Pos, input.Click, input.Install)
	} else {
		err = s.store.UpdateAdStat(ctx, atoi(row["id"]), input.Click, input.Install)
	}
	if err != nil {
		return -1, "保存广告统计失败", err
	}
	return 0, "", nil
}

func (s *Service) PlayAdd(ctx context.Context, token string, ip string, input PlayInput) (int, string, error) {
	sid, retcode, errmsg, err := s.sid(ctx, token, ip)
	if err != nil || retcode != 0 {
		return retcode, errmsg, err
	}
	if input.VID <= 0 || input.Duration <= 0 {
		return -9999, "无效参数", nil
	}
	row, err := s.store.PlayStatBySID(ctx, sid, input.VID)
	if err != nil {
		return -1, "保存播放统计失败", err
	}
	if len(row) == 0 {
		err = s.store.CreatePlayStat(ctx, sid, input.VID, input.Mini, input.Duration, input.Played)
	} else if input.Played > atoi(row["played"]) {
		err = s.store.UpdatePlayStatPlayed(ctx, atoi(row["id"]), input.Played)
	}
	if err != nil {
		return -1, "保存播放统计失败", err
	}
	return 0, "", nil
}

func (s *Service) sid(ctx context.Context, token string, ip string) (string, int, string, error) {
	clean := userRepo.CleanToken(token)
	if s.auth != nil {
		user, err := s.auth.UserBySession(ctx, clean)
		if err != nil {
			return "", -1, "获取用户失败", err
		}
		if user != nil && atoi(user["uid"]) > 0 {
			return fmt.Sprint(user["uid"]), 0, "", nil
		}
	}
	if clean == "" && ip != "" {
		clean = guestSID(ip)
		if err := s.store.CreateGuest(ctx, clean, s.now().Unix()); err != nil {
			return "", -1, "获取游客失败", err
		}
	}
	exists, err := s.store.GuestExists(ctx, clean)
	if err != nil {
		return "", -1, "获取游客失败", err
	}
	if !exists {
		return "", -9999, "请登录后操作，客户端游客请先携带信息", nil
	}
	if clean == "" {
		return "", -9999, "无法获取用户唯一标识", nil
	}
	return clean, 0, "", nil
}

func guestSID(ip string) string {
	crc := crc32.ChecksumIEEE([]byte(ip))
	sum := md5.Sum([]byte(fmt.Sprintf("%x", crc)))
	return hex.EncodeToString(sum[:])
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
