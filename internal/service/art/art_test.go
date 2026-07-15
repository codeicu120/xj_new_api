package art

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeStore struct {
	categories []map[string]interface{}
	rows       []map[string]interface{}
	row        map[string]interface{}
}

func (s fakeStore) Categories(context.Context) ([]map[string]interface{}, error) {
	return s.categories, nil
}

func (s fakeStore) CountByCategory(context.Context, int) (int, error) {
	return len(s.rows), nil
}

func (s fakeStore) ListByCategory(context.Context, int, int, int, int) ([]map[string]interface{}, error) {
	return s.rows, nil
}

func (s fakeStore) ArtByID(context.Context, int) (map[string]interface{}, error) {
	return s.row, nil
}

func TestAnnounceBuildsPHPCompatibleRowsAndPageInfo(t *testing.T) {
	service := NewService(fakeStore{
		categories: []map[string]interface{}{{"cateid": "1", "uuid": "announce", "catename": "系统公告"}},
		rows: []map[string]interface{}{{
			"artid":      "2",
			"title":      "公告",
			"subtitle":   "",
			"coverpic":   "/cover.png",
			"ctimestamp": "100",
			"intro":      "intro",
			"cateid":     "1",
		}},
	}, "https://img.example")
	service.now = func() time.Time { return time.Unix(200, 0) }

	data, err := service.Announce(context.Background(), 0)
	if err != nil {
		t.Fatalf("announce: %v", err)
	}
	if got := data.PageInfo["page_url"]; got != "/art/?page=[?]" {
		t.Fatalf("unexpected page url %#v", got)
	}
	row := data.Rows[0]
	if row["coverpic"] != "https://img.example/cover.png" {
		t.Fatalf("unexpected coverpic %#v", row["coverpic"])
	}
	if row["content"] != nil || row["catename"] != "系统公告" {
		t.Fatalf("unexpected row %#v", row)
	}
	if row["addtime"] != "0天前 0小时前 1分钟前 40秒前" {
		t.Fatalf("unexpected addtime %#v", row["addtime"])
	}
}

func TestAnnounceMissingCategory(t *testing.T) {
	service := NewService(fakeStore{categories: []map[string]interface{}{{"cateid": "2", "uuid": "news"}}}, "")

	_, err := service.Announce(context.Background(), 1)
	if !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("expected category error, got %v", err)
	}
}

func TestShowReturnsProcessedRow(t *testing.T) {
	service := NewService(fakeStore{
		categories: []map[string]interface{}{{"cateid": "1", "uuid": "announce", "catename": "系统公告"}},
		row: map[string]interface{}{
			"artid":      "2",
			"title":      "公告",
			"subtitle":   "",
			"coverpic":   "",
			"ctimestamp": "1",
			"intro":      "",
			"content":    "正文",
			"cateid":     "1",
			"showtype":   "0",
		},
	}, "")
	service.now = func() time.Time { return time.Unix(86400*31, 0) }

	data, err := service.Show(context.Background(), 2)
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if data.Row["content"] != "正文" || data.Row["addtime"] != "1969-12-31" {
		t.Fatalf("unexpected show row %#v", data.Row)
	}
}

func TestShowRejectsMissingOrHiddenRows(t *testing.T) {
	for name, row := range map[string]map[string]interface{}{
		"missing": {},
		"hidden":  {"showtype": "1"},
	} {
		t.Run(name, func(t *testing.T) {
			service := NewService(fakeStore{row: row}, "")
			_, err := service.Show(context.Background(), 2)
			if !errors.Is(err, ErrArtNotFound) {
				t.Fatalf("expected not found, got %v", err)
			}
		})
	}
}
