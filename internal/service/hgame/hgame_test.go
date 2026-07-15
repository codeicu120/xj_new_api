package hgame

import (
	"context"
	"testing"
)

type fakeStore struct {
	total int
	rows  map[int][]map[string]interface{}
}

func (s fakeStore) Count(_ context.Context, statusOnly bool, excludedShowType int) (int, error) {
	if !statusOnly {
		return s.total, nil
	}
	return len(s.rows[excludedShowType]), nil
}

func (s fakeStore) List(_ context.Context, _ int, _ int, _ int, excludedShowType int) ([]map[string]interface{}, error) {
	return s.rows[excludedShowType], nil
}

func TestIndexClosed(t *testing.T) {
	service := NewService(fakeStore{}, "")

	_, retcode, errmsg, err := service.Index(context.Background(), 1)
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	if retcode != -1 || errmsg != "暂未开放" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
}

func TestIndexRows(t *testing.T) {
	service := NewService(fakeStore{
		total: 2,
		rows: map[int][]map[string]interface{}{
			1: {{"id": "1", "name": "A", "image": "a.png", "logo": "l.png", "remark": `{"x":1}`, "show_type": "0", "sort": "2", "status": "0"}},
			0: {{"id": "2", "name": "B", "remark": "text", "show_type": "1", "sort": "1", "status": "0"}},
		},
	}, "https://img.test")

	data, retcode, errmsg, err := service.Index(context.Background(), 1)
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("response = %d %q", retcode, errmsg)
	}
	list := data.Data["list"].([]map[string]interface{})
	if len(list) != 1 || list[0]["image"] != "https://img.test/a.png" {
		t.Fatalf("list = %#v", list)
	}
	remark := list[0]["remark"].(map[string]interface{})
	if remark["x"] != float64(1) {
		t.Fatalf("remark = %#v", list[0]["remark"])
	}
}
