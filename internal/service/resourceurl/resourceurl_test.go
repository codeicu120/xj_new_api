package resourceurl

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"testing"
	"time"
)

type fakeSettingsStore struct {
	value string
}

func (s fakeSettingsStore) SettingValue(context.Context) (string, error) { return s.value, nil }

type fakeLocator struct{ parts []string }

func (l fakeLocator) Find(string) ([]string, error) { return l.parts, nil }

func TestResolverMatchesPHPResourceURLSelection(t *testing.T) {
	settings := `a:5:{s:6:"resurl";s:29:"https://app-{rand}.example";s:9:"resurl_h5";s:28:"https://h5-{rand}.example";s:14:"resurl_h5_free";s:30:"https://free-{rand}.example";s:20:"resurl_h5_free_area";s:13:"河北,广东";s:11:"resurl_auth";s:6:"secret";}`
	r := NewResolver(fakeSettingsStore{value: settings}, fakeLocator{parts: []string{"中国", "广东", "深圳"}}, "https://fallback.example")
	r.now = func() time.Time { return time.Date(2026, 7, 22, 3, 4, 5, 0, time.UTC) }

	resolved, err := r.Resolve(context.Background(), Request{HasCookieAuth: true, ClientIP: "1.2.3.4"})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.BaseURL != "https://free-2026072211.example" {
		t.Fatalf("unexpected base URL: %s", resolved.BaseURL)
	}

	sum := md5.Sum([]byte("/images/a.jpg@1784689445@secret"))
	want := "https://free-2026072211.example/images/a.jpg?sign=" + hex.EncodeToString(sum[:]) + "&t=1784689445"
	if got := resolved.GetRes("images/a.jpg", ""); got != want {
		t.Fatalf("unexpected signed URL\n got: %s\nwant: %s", got, want)
	}
}

func TestResolverUsesH5WhenAreaStartsAtPositionZero(t *testing.T) {
	settings := `a:3:{s:9:"resurl_h5";s:10:"https://h5";s:14:"resurl_h5_free";s:12:"https://free";s:20:"resurl_h5_free_area";s:6:"广东";}`
	r := NewResolver(fakeSettingsStore{value: settings}, fakeLocator{parts: []string{"广东", "深圳"}}, "")
	resolved, err := r.Resolve(context.Background(), Request{HasCookieAuth: true})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.BaseURL != "https://h5" {
		t.Fatalf("PHP strpos > 0 behavior not preserved: %s", resolved.BaseURL)
	}
}

func TestResolvedGetResSignsAbsoluteURIAndFallbackIsExplicit(t *testing.T) {
	r := NewResolver(nil, nil, "https://fallback.example")
	r.now = func() time.Time { return time.Unix(100, 0) }
	resolved, err := r.Resolve(context.Background(), Request{})
	if err != nil {
		t.Fatal(err)
	}
	if got := resolved.GetRes("pic/a.jpg", ""); got != "https://fallback.example/pic/a.jpg" {
		t.Fatalf("unexpected fallback URL: %s", got)
	}

	resolved.AuthSecret = "key"
	got := resolved.GetRes("https://cdn.example/a.jpg", "")
	if got == "https://cdn.example/a.jpg" {
		t.Fatal("absolute URI must still receive PHP-compatible auth query")
	}
}
