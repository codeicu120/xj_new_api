package index

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	vodRepo "xj_comp/internal/repository/vod"
)

type HomeStore interface {
	CalldataByUUID(ctx context.Context, uuid string) (map[string]interface{}, error)
	ListVODs(ctx context.Context, filter vodRepo.ListingFilter, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error)
	RandomVODs(ctx context.Context, pageSize int) ([]map[string]interface{}, error)
	VODsByIDsLimited(ctx context.Context, ids []int, freeOnly bool, limit int, orderByField bool) ([]map[string]interface{}, error)
}

type VODProcessor interface {
	ProcessRows(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error)
}

type HomeService struct {
	store           HomeStore
	vodProcessor    VODProcessor
	resourceBaseURL string
	now             func() time.Time
}

func NewHomeService(store HomeStore, vodProcessor VODProcessor, resourceBaseURL string) *HomeService {
	return &HomeService{
		store:           store,
		vodProcessor:    vodProcessor,
		resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"),
		now:             time.Now,
	}
}

func (s *HomeService) Index(ctx context.Context, isH5Request bool) (map[string]interface{}, error) {
	dayRows, err := s.recommendRows(ctx, 6, isH5Request)
	if err != nil {
		return nil, err
	}
	latestRows, err := s.vodRows(ctx, vodRepo.ListingFilter{}, 16, "utimestamp DESC", isH5Request)
	if err != nil {
		return nil, err
	}
	likeRows, err := s.randomRows(ctx, 6, isH5Request)
	if err != nil {
		return nil, err
	}
	aRows, err := s.vodRows(ctx, vodRepo.ListingFilter{CateIDs: []int{4}}, 6, "utimestamp DESC", isH5Request)
	if err != nil {
		return nil, err
	}
	bRows, err := s.vodRows(ctx, vodRepo.ListingFilter{CateIDs: []int{11}}, 6, "utimestamp DESC", isH5Request)
	if err != nil {
		return nil, err
	}
	cRows, err := s.vodRows(ctx, vodRepo.ListingFilter{CateIDs: []int{14}}, 6, "utimestamp DESC", isH5Request)
	if err != nil {
		return nil, err
	}
	dRows, err := s.vodRows(ctx, vodRepo.ListingFilter{LangVoice: 1}, 6, "utimestamp DESC", isH5Request)
	if err != nil {
		return nil, err
	}
	tagRows, err := s.vodRows(ctx, vodRepo.ListingFilter{}, 6, "utimestamp DESC", isH5Request)
	if err != nil {
		return nil, err
	}
	now := s.now().Unix()
	hotRows, err := s.vodRows(ctx, vodRepo.ListingFilter{CTimeAfter: now - 14*86400, CTimeBefore: now - 7*86400}, 6, "playcount_week DESC", isH5Request)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"sliderows":   s.slideRows(ctx, "index.slide"),
		"v2sliderows": s.slideRows(ctx, "index.slide.v2"),
		"pcsliderows": s.slideRows(ctx, "index.slide.pc"),
		"mbsliderows": s.slideRows(ctx, "index.slide.mb"),
		"dayrows":     dayRows,
		"latestrows":  latestRows,
		"likerows":    likeRows,
		"a_vodrows":   aRows,
		"b_vodrows":   bRows,
		"c_vodrows":   cRows,
		"d_vodrows":   dRows,
		"tagvodrows":  tagRows,
		"hotrows":     hotRows,
	}, nil
}

func (s *HomeService) slideRows(ctx context.Context, uuid string) []map[string]interface{} {
	row, err := s.store.CalldataByUUID(ctx, uuid)
	if err != nil {
		return []map[string]interface{}{}
	}
	raw := mustJSONRows(row)
	out := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		slide := map[string]interface{}{}
		for i := 0; i <= 2; i++ {
			title := str(itemMap[fmt.Sprintf("title%d", i)])
			if title != "" {
				slide[title] = str(itemMap[fmt.Sprintf("url%d", i)])
			}
		}
		slide["pic"] = s.resURL(str(itemMap["pic"]))
		slide["showweight"] = str(itemMap["showweight"])
		if str(itemMap["title3"]) == "newpic" && str(itemMap["url3"]) != "" {
			slide["newpic"] = str(itemMap["url3"])
		}
		out = append(out, slide)
	}
	return out
}

func (s *HomeService) recommendRows(ctx context.Context, pageSize int, isH5Request bool) ([]map[string]interface{}, error) {
	row, err := s.store.CalldataByUUID(ctx, "index.recommend.vods")
	if err != nil {
		return nil, err
	}
	ids := splitInts(str(row["content"]))
	if len(ids) == 0 {
		return s.vodRows(ctx, vodRepo.ListingFilter{}, pageSize, "vodid DESC", isH5Request)
	}
	rows, err := s.store.VODsByIDsLimited(ctx, ids, false, pageSize, true)
	if err != nil {
		return nil, err
	}
	return s.processRows(ctx, rows, isH5Request)
}

func (s *HomeService) randomRows(ctx context.Context, pageSize int, isH5Request bool) ([]map[string]interface{}, error) {
	rows, err := s.store.RandomVODs(ctx, pageSize)
	if err != nil {
		return nil, err
	}
	return s.processRows(ctx, rows, isH5Request)
}

func (s *HomeService) vodRows(ctx context.Context, filter vodRepo.ListingFilter, pageSize int, orderBy string, isH5Request bool) ([]map[string]interface{}, error) {
	rows, err := s.store.ListVODs(ctx, filter, 0, 1, pageSize, orderBy)
	if err != nil {
		return nil, err
	}
	return s.processRows(ctx, rows, isH5Request)
}

func (s *HomeService) processRows(ctx context.Context, rows []map[string]interface{}, isH5Request bool) ([]map[string]interface{}, error) {
	if s.vodProcessor == nil {
		return rows, nil
	}
	return s.vodProcessor.ProcessRows(ctx, rows, isH5Request)
}

func (s *HomeService) resURL(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return s.resourceBaseURL + "/" + strings.TrimLeft(path, "/")
}

func mustJSONRows(row map[string]interface{}) []interface{} {
	if str(row["type"]) != "rows" && str(row["type"]) != "json" {
		return []interface{}{}
	}
	var rows []interface{}
	if err := json.Unmarshal([]byte(str(row["content"])), &rows); err != nil {
		return []interface{}{}
	}
	return rows
}

func splitInts(value string) []int {
	parts := strings.Split(value, ",")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.Atoi(part)
		if err != nil || id <= 0 {
			continue
		}
		out = append(out, id)
	}
	return out
}
