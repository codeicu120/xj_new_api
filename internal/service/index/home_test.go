package index

import (
	"context"
	"reflect"
	"testing"
	"time"

	vodRepo "xj_comp/internal/repository/vod"
)

type fakeHomeStore struct {
	calls       map[string]map[string]interface{}
	listCalls   []homeListCall
	randomRows  []map[string]interface{}
	idsRows     []map[string]interface{}
	defaultRows []map[string]interface{}
}

type homeListCall struct {
	filter   vodRepo.ListingFilter
	pageSize int
	orderBy  string
}

func (s *fakeHomeStore) CalldataByUUID(_ context.Context, uuid string) (map[string]interface{}, error) {
	if s.calls == nil {
		return map[string]interface{}{}, nil
	}
	if row, ok := s.calls[uuid]; ok {
		return row, nil
	}
	return map[string]interface{}{}, nil
}

func (s *fakeHomeStore) ListVODs(_ context.Context, filter vodRepo.ListingFilter, _ int, _ int, pageSize int, orderBy string) ([]map[string]interface{}, error) {
	s.listCalls = append(s.listCalls, homeListCall{filter: filter, pageSize: pageSize, orderBy: orderBy})
	return cloneHomeRows(s.defaultRows), nil
}

func (s *fakeHomeStore) RandomVODs(_ context.Context, _ int) ([]map[string]interface{}, error) {
	return cloneHomeRows(s.randomRows), nil
}

func (s *fakeHomeStore) VODsByIDsLimited(_ context.Context, _ []int, _ bool, _ int, _ bool) ([]map[string]interface{}, error) {
	return cloneHomeRows(s.idsRows), nil
}

type fakeHomeProcessor struct {
	calls int
}

func (p *fakeHomeProcessor) ProcessRows(_ context.Context, rows []map[string]interface{}, _ bool) ([]map[string]interface{}, error) {
	p.calls++
	out := cloneHomeRows(rows)
	for _, row := range out {
		row["processed"] = true
	}
	return out, nil
}

func TestHomeIndexShape(t *testing.T) {
	store := &fakeHomeStore{
		calls: map[string]map[string]interface{}{
			"index.slide":          {"type": "rows", "content": `[{"title0":"scene","url0":"vod.show","title1":"vodid","url1":"10","title2":"title","url2":"A","title3":"newpic","url3":"new.jpg","pic":"slide.jpg","showweight":"1"}]`},
			"index.slide.v2":       {"type": "rows", "content": `[]`},
			"index.slide.pc":       {"type": "rows", "content": `[]`},
			"index.slide.mb":       {"type": "rows", "content": `[]`},
			"index.recommend.vods": {"type": "code", "content": "8,9"},
		},
		idsRows:     []map[string]interface{}{{"vodid": "8"}},
		randomRows:  []map[string]interface{}{{"vodid": "7"}},
		defaultRows: []map[string]interface{}{{"vodid": "6"}},
	}
	processor := &fakeHomeProcessor{}
	service := NewHomeService(store, processor, "https://res.example")
	service.now = func() time.Time { return time.Unix(1_700_000_000, 0) }

	data, err := service.Index(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}

	expectedKeys := []string{"sliderows", "v2sliderows", "pcsliderows", "mbsliderows", "dayrows", "latestrows", "likerows", "a_vodrows", "b_vodrows", "c_vodrows", "d_vodrows", "tagvodrows", "hotrows"}
	for _, key := range expectedKeys {
		if _, ok := data[key]; !ok {
			t.Fatalf("missing key %s in %#v", key, data)
		}
	}
	slides := data["sliderows"].([]map[string]interface{})
	if slides[0]["pic"] != "https://res.example/slide.jpg" || slides[0]["newpic"] != "new.jpg" {
		t.Fatalf("slides = %#v", slides)
	}
	if processor.calls != 9 {
		t.Fatalf("processor calls = %d", processor.calls)
	}
	if len(store.listCalls) != 7 {
		t.Fatalf("list calls = %#v", store.listCalls)
	}
	if !reflect.DeepEqual(store.listCalls[1].filter.CateIDs, []int{4}) {
		t.Fatalf("a_vod filter = %#v", store.listCalls[1].filter)
	}
	if store.listCalls[4].filter.LangVoice != 1 {
		t.Fatalf("d_vod filter = %#v", store.listCalls[4].filter)
	}
	if store.listCalls[6].filter.CTimeAfter == 0 || store.listCalls[6].filter.CTimeBefore == 0 {
		t.Fatalf("hot filter = %#v", store.listCalls[6].filter)
	}
}

func cloneHomeRows(rows []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		clone := map[string]interface{}{}
		for key, value := range row {
			clone[key] = value
		}
		out = append(out, clone)
	}
	return out
}
