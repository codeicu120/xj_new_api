package stats

import (
	"context"
	"testing"
	"time"
)

type fakeStore struct {
	guestExists bool
	shortcut    map[string]interface{}
	stats       map[string]interface{}
	ad          map[string]interface{}
	play        map[string]interface{}
	createdIP   string
	adCreated   AdInput
	adUpdated   AdInput
	playCreated PlayInput
	playUpdated int
}

func (s *fakeStore) GuestExists(context.Context, string) (bool, error) { return s.guestExists, nil }
func (s *fakeStore) ShortcutCreatedByIP(context.Context, string) (map[string]interface{}, error) {
	return s.shortcut, nil
}
func (s *fakeStore) CreateShortcut(_ context.Context, ip string, _ int64) (int64, error) {
	s.createdIP = ip
	return 1, nil
}
func (s *fakeStore) ShortcutStatsByDate(context.Context, int64) (map[string]interface{}, error) {
	return s.stats, nil
}
func (s *fakeStore) CreateShortcutStats(context.Context, int64) error { return nil }
func (s *fakeStore) UpdateShortcutStatsCount(context.Context, int, int) error {
	return nil
}
func (s *fakeStore) AdStatBySID(context.Context, string, string, string) (map[string]interface{}, error) {
	return s.ad, nil
}
func (s *fakeStore) CreateAdStat(_ context.Context, _ string, title string, url string, pos int, click int, install int) error {
	s.adCreated = AdInput{Title: title, URL: url, Pos: pos, Click: click, Install: install}
	return nil
}
func (s *fakeStore) UpdateAdStat(_ context.Context, _ int, click int, install int) error {
	s.adUpdated = AdInput{Click: click, Install: install}
	return nil
}
func (s *fakeStore) PlayStatBySID(context.Context, string, int) (map[string]interface{}, error) {
	return s.play, nil
}
func (s *fakeStore) CreatePlayStat(_ context.Context, _ string, vid int, mini int, duration int, played int) error {
	s.playCreated = PlayInput{VID: vid, Mini: mini, Duration: duration, Played: played}
	return nil
}
func (s *fakeStore) UpdatePlayStatPlayed(_ context.Context, _ int, played int) error {
	s.playUpdated = played
	return nil
}
func (s *fakeStore) CreateGuest(context.Context, string, int64) error {
	s.guestExists = true
	return nil
}

type fakeAuth struct {
	user map[string]interface{}
}

func (a fakeAuth) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return a.user, nil
}

func TestShortcutAddCreatesOnlyForNewIP(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, nil)
	service.now = func() time.Time { return time.Unix(1710460800+3600, 0) }

	if err := service.ShortcutAdd(context.Background(), "127.0.0.1"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if store.createdIP != "127.0.0.1" {
		t.Fatalf("expected created ip, got %q", store.createdIP)
	}

	store.createdIP = ""
	store.shortcut = map[string]interface{}{"id": "1"}
	if err := service.ShortcutAdd(context.Background(), "127.0.0.1"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if store.createdIP != "" {
		t.Fatalf("expected duplicate ip not to create, got %q", store.createdIP)
	}
}

func TestAdAddValidationAndInstallNormalization(t *testing.T) {
	service := NewService(&fakeStore{}, nil)
	retcode, errmsg, err := service.AdAdd(context.Background(), "", "", AdInput{Title: "a", URL: "b", Pos: 1, Click: 1})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if retcode != -9999 || errmsg != "请登录后操作，客户端游客请先携带信息" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}

	store := &fakeStore{guestExists: true}
	service = NewService(store, nil)
	retcode, errmsg, err = service.AdAdd(context.Background(), "12345678901234567890123456789012", "", AdInput{Title: " t ", URL: " u ", Pos: 2, Click: 3, Install: 2})
	if err != nil || retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q err=%v", retcode, errmsg, err)
	}
	if store.adCreated.Title != "t" || store.adCreated.URL != "u" || store.adCreated.Click != 1 || store.adCreated.Install != 1 {
		t.Fatalf("unexpected create %+v", store.adCreated)
	}
}

func TestPlayAddUpdatesOnlyWhenPlayedIncreases(t *testing.T) {
	store := &fakeStore{guestExists: true, play: map[string]interface{}{"id": "9", "played": "30"}}
	service := NewService(store, nil)

	retcode, errmsg, err := service.PlayAdd(context.Background(), "12345678901234567890123456789012", "", PlayInput{VID: 7, Duration: 100, Played: 20})
	if err != nil || retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q err=%v", retcode, errmsg, err)
	}
	if store.playUpdated != 0 {
		t.Fatalf("expected no update, got %d", store.playUpdated)
	}

	retcode, errmsg, err = service.PlayAdd(context.Background(), "12345678901234567890123456789012", "", PlayInput{VID: 7, Duration: 100, Played: 40})
	if err != nil || retcode != 0 || errmsg != "" {
		t.Fatalf("unexpected response %d %q err=%v", retcode, errmsg, err)
	}
	if store.playUpdated != 40 {
		t.Fatalf("expected played update, got %d", store.playUpdated)
	}
}
