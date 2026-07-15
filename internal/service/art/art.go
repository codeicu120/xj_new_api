package art

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"xj_comp/internal/domain"
)

var (
	ErrCategoryNotFound = errors.New("art category not found")
	ErrArtNotFound      = errors.New("art not found")
)

type Store interface {
	Categories(ctx context.Context) ([]map[string]interface{}, error)
	CountByCategory(ctx context.Context, cateID int) (int, error)
	ListByCategory(ctx context.Context, cateID int, total int, page int, pageSize int) ([]map[string]interface{}, error)
	ArtByID(ctx context.Context, artID int) (map[string]interface{}, error)
}

type Service struct {
	store           Store
	resourceBaseURL string
	now             func() time.Time
}

func NewService(store Store, resourceBaseURL string) *Service {
	return &Service{store: store, resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"), now: time.Now}
}

func (s *Service) Announce(ctx context.Context, page int) (domain.ArtListingData, error) {
	categories, category, err := s.category(ctx, "announce")
	if err != nil {
		return domain.ArtListingData{}, err
	}
	cateID := atoi(category["cateid"])
	pageSize := 20
	total, err := s.store.CountByCategory(ctx, cateID)
	if err != nil {
		return domain.ArtListingData{}, err
	}
	rows, err := s.store.ListByCategory(ctx, cateID, total, page, pageSize)
	if err != nil {
		return domain.ArtListingData{}, err
	}
	return domain.ArtListingData{
		Rows:     s.processRows(rows, categories),
		PageInfo: pageInfo(total, pageSize, page, "/art/?page=[?]"),
	}, nil
}

func (s *Service) Show(ctx context.Context, artID int) (domain.ArtShowData, error) {
	row, err := s.store.ArtByID(ctx, artID)
	if err != nil {
		return domain.ArtShowData{}, err
	}
	if len(row) == 0 || atoi(row["showtype"]) != 0 {
		return domain.ArtShowData{}, ErrArtNotFound
	}
	categories, err := s.store.Categories(ctx)
	if err != nil {
		return domain.ArtShowData{}, err
	}
	processed := s.processRows([]map[string]interface{}{row}, categories)
	if len(processed) == 0 {
		return domain.ArtShowData{}, ErrArtNotFound
	}
	return domain.ArtShowData{Row: processed[0]}, nil
}

func (s *Service) category(ctx context.Context, uuid string) ([]map[string]interface{}, map[string]interface{}, error) {
	categories, err := s.store.Categories(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, row := range categories {
		if fmt.Sprint(row["uuid"]) == uuid {
			return categories, row, nil
		}
	}
	return categories, nil, ErrCategoryNotFound
}

func (s *Service) processRows(rows []map[string]interface{}, categories []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	cats := indexBy(categories, "cateid")
	now := s.now().Unix()
	for _, row := range rows {
		cateID := fmt.Sprint(row["cateid"])
		content := row["content"]
		if _, ok := row["content"]; !ok {
			content = nil
		}
		out = append(out, map[string]interface{}{
			"artid":    fmt.Sprint(row["artid"]),
			"title":    fmt.Sprint(row["title"]),
			"subtitle": fmt.Sprint(row["subtitle"]),
			"coverpic": s.resourceURL(fmt.Sprint(row["coverpic"])),
			"addtime":  artAddTime(now, atoi64(row["ctimestamp"])),
			"intro":    fmt.Sprint(row["intro"]),
			"content":  content,
			"cateid":   cateID,
			"catename": lookup(cats, cateID, "catename"),
		})
	}
	return out
}

func (s *Service) resourceURL(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if s.resourceBaseURL == "" {
		return path
	}
	return s.resourceBaseURL + "/" + strings.TrimLeft(path, "/")
}

func artAddTime(now int64, created int64) string {
	if created <= 0 {
		return "1970-01-01"
	}
	diff := now - created
	if diff >= 0 && diff < 86400*30 {
		days := diff / 86400
		hours := (diff % 86400) / 3600
		mins := (diff % 3600) / 60
		secs := diff % 60
		return fmt.Sprintf("%d天前 %d小时前 %d分钟前 %d秒前", days, hours, mins, secs)
	}
	return time.Unix(created, 0).Format("2006-01-02")
}

func pageInfo(total int, pageSize int, page int, pageURL string) map[string]interface{} {
	if total < 0 {
		total = 0
	}
	if pageSize < 1 {
		pageSize = 1
	}
	totalPage := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPage < 1 {
		totalPage = 1
	}
	if page < 1 {
		page = 1
	}
	if page > totalPage {
		page = totalPage
	}
	start := 0
	if total > 0 {
		start = (page-1)*pageSize + 1
	}
	end := start + pageSize - 1
	if end > total {
		end = total
	}
	return map[string]interface{}{
		"plist":     plist(page, totalPage, pageURL),
		"pagesize":  pageSize,
		"total":     total,
		"totalpage": totalPage,
		"page":      page,
		"start":     start,
		"end":       end,
		"prev":      ternary(page > 1, page-1, 0),
		"next":      ternary(page < totalPage, page+1, 0),
		"curr_url":  strings.ReplaceAll(pageURL, "[?]", strconv.Itoa(page)),
		"first_url": strings.ReplaceAll(pageURL, "[?]", "1"),
		"prev_url":  strings.ReplaceAll(pageURL, "[?]", strconv.Itoa(ternary(page > 1, page-1, 1))),
		"next_url":  strings.ReplaceAll(pageURL, "[?]", strconv.Itoa(ternary(page < totalPage, page+1, totalPage))),
		"last_url":  strings.ReplaceAll(pageURL, "[?]", strconv.Itoa(totalPage)),
		"page_url":  pageURL,
		"pages":     pageSelector(page, totalPage),
	}
}

func plist(page int, totalPage int, pageURL string) []map[string]interface{} {
	pages := []int{}
	for i := page - 5; i < page; i++ {
		if i > 0 {
			pages = append(pages, i)
		}
	}
	pages = append(pages, page)
	for i := page + 1; i <= totalPage && len(pages) < 10; i++ {
		pages = append(pages, i)
	}
	for i := pages[0] - 1; i > 0 && len(pages) < 10; i-- {
		pages = append([]int{i}, pages...)
	}
	result := []map[string]interface{}{}
	for _, p := range pages {
		pos := ""
		if p == page {
			pos = "curr"
		}
		result = append(result, pageLink(pos, p, p, pageURL))
	}
	return result
}

func pageLink(pos string, page int, text interface{}, pageURL string) map[string]interface{} {
	return map[string]interface{}{"pos": pos, "page": page, "text": text, "url": strings.ReplaceAll(pageURL, "[?]", strconv.Itoa(page))}
}

func pageSelector(pageNow int, totalPage int) []int {
	pages := make([]int, 0, totalPage)
	for i := 1; i <= totalPage; i++ {
		pages = append(pages, i)
	}
	sort.Ints(pages)
	return pages
}

func indexBy(rows []map[string]interface{}, key string) map[string]map[string]interface{} {
	out := map[string]map[string]interface{}{}
	for _, row := range rows {
		out[fmt.Sprint(row[key])] = row
	}
	return out
}

func lookup(rows map[string]map[string]interface{}, id string, key string) interface{} {
	if row, ok := rows[id]; ok {
		return fmt.Sprint(row[key])
	}
	return nil
}

func atoi(value interface{}) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(fmt.Sprint(value)))
	return parsed
}

func atoi64(value interface{}) int64 {
	parsed, _ := strconv.ParseInt(strings.TrimSpace(fmt.Sprint(value)), 10, 64)
	return parsed
}

func ternary[T any](ok bool, yes T, no T) T {
	if ok {
		return yes
	}
	return no
}
