package hgame

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"xj_comp/internal/domain"
)

type Store interface {
	Count(ctx context.Context, statusOnly bool, excludedShowType int) (int, error)
	List(ctx context.Context, total int, page int, pageSize int, excludedShowType int) ([]map[string]interface{}, error)
}

type Service struct {
	store           Store
	resourceBaseURL string
}

func NewService(store Store, resourceBaseURL string) *Service {
	return &Service{store: store, resourceBaseURL: strings.TrimRight(resourceBaseURL, "/")}
}

func (s *Service) Index(ctx context.Context, page int) (domain.HGameIndexData, int, string, error) {
	const pageSize = 20
	total, err := s.store.Count(ctx, false, 0)
	if err != nil {
		return domain.HGameIndexData{}, -1, "获取游戏列表失败", err
	}
	if total == 0 {
		return domain.HGameIndexData{}, -1, "暂未开放", nil
	}
	listTotal, err := s.store.Count(ctx, true, 1)
	if err != nil {
		return domain.HGameIndexData{}, -1, "获取游戏列表失败", err
	}
	listRows, err := s.store.List(ctx, listTotal, page, pageSize, 1)
	if err != nil {
		return domain.HGameIndexData{}, -1, "获取游戏列表失败", err
	}
	slideTotal, err := s.store.Count(ctx, true, 0)
	if err != nil {
		return domain.HGameIndexData{}, -1, "获取游戏列表失败", err
	}
	slideRows, err := s.store.List(ctx, slideTotal, page, pageSize, 0)
	if err != nil {
		return domain.HGameIndexData{}, -1, "获取游戏列表失败", err
	}
	return domain.HGameIndexData{
		Data: map[string]interface{}{
			"list":  s.processRows(listRows),
			"slide": s.processRows(slideRows),
		},
	}, 0, "", nil
}

func (s *Service) processRows(rows []map[string]interface{}) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, row := range rows {
		result = append(result, map[string]interface{}{
			"id":               atoi(row["id"]),
			"name":             str(row["name"]),
			"image":            resourceURL(s.resourceBaseURL, str(row["image"])),
			"logo":             resourceURL(s.resourceBaseURL, str(row["logo"])),
			"link":             str(row["link"]),
			"download_ios":     str(row["download_ios"]),
			"download_android": str(row["download_android"]),
			"download_win":     str(row["download_win"]),
			"download_mac":     str(row["download_mac"]),
			"brief":            str(row["brief"]),
			"detail":           str(row["detail"]),
			"remark":           decodeRemark(row["remark"]),
			"show_type":        atoi(row["show_type"]),
			"sort":             atoi(row["sort"]),
			"status":           atoi(row["status"]),
		})
	}
	return result
}

func decodeRemark(value interface{}) interface{} {
	raw := str(value)
	if raw == "" {
		return ""
	}
	var decoded interface{}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return raw
	}
	if decoded == nil {
		return raw
	}
	return decoded
}

func resourceURL(baseURL string, path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if baseURL == "" {
		return path
	}
	return baseURL + "/" + strings.TrimLeft(path, "/")
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(str(value), &n)
	return n
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
