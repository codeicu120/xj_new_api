package index

import (
	"context"
	"errors"
	"testing"
)

type fakeCoverFetcher struct {
	calls int
	body  string
	err   error
}

func (f *fakeCoverFetcher) FetchCover(context.Context, string, string) (string, error) {
	f.calls++
	return f.body, f.err
}

func TestCoverServiceEncryptsAndCaches(t *testing.T) {
	store := fakeInitStore{
		settings: map[string]map[string]interface{}{
			"setting": {"value": `a:1:{s:11:"getCoverUrl";s:26:"https://cover.example/api";}`},
		},
	}
	cache := NewMemoryCoverCache()
	fetcher := &fakeCoverFetcher{body: "base64-image-body"}
	service := NewCoverService(store, cache, fetcher)
	pic := "0123456789abcdefghijklmnopqrstuvwxyzABCDEF"

	got, err := service.GetCover(context.Background(), pic)
	if err != nil {
		t.Fatal(err)
	}
	if got == "" || got == fetcher.body {
		t.Fatalf("unexpected encrypted cover %q", got)
	}
	again, err := service.GetCover(context.Background(), pic)
	if err != nil {
		t.Fatal(err)
	}
	if again != got {
		t.Fatalf("cache value mismatch %q != %q", again, got)
	}
	if fetcher.calls != 1 {
		t.Fatalf("fetcher calls = %d", fetcher.calls)
	}
}

func TestCoverServiceNotFound(t *testing.T) {
	store := fakeInitStore{settings: map[string]map[string]interface{}{"setting": {"value": `a:0:{}`}}}
	service := NewCoverService(store, NewMemoryCoverCache(), &fakeCoverFetcher{body: ""})
	if _, err := service.GetCover(context.Background(), ""); !errors.Is(err, ErrCoverNotFound) {
		t.Fatalf("empty pic err = %v", err)
	}
	if _, err := service.GetCover(context.Background(), "short"); !errors.Is(err, ErrCoverNotFound) {
		t.Fatalf("short pic err = %v", err)
	}
}
