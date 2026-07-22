package amazing

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"xj_comp/internal/domain"
	amazingRepo "xj_comp/internal/repository/amazing"
	"xj_comp/internal/service/resourceurl"
)

const listingSampleParams = "$category_id:0-$orderby:0-$page:1"

type SoftwareStore interface {
	CountActive(ctx context.Context, filter amazingRepo.SoftwareFilter) (int, error)
	ListActive(ctx context.Context, filter amazingRepo.SoftwareFilter, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error)
}

type ListingService struct {
	store           SoftwareStore
	resourceBaseURL string
	now             func() time.Time
	resources       *resourceurl.Resolver
}

func (s *ListingService) WithResourceResolver(r *resourceurl.Resolver) *ListingService {
	s.resources = r
	return s
}

type ListingRequest struct {
	Action     string
	PathParams string
	QueryPage  string
}

func NewListingService(store SoftwareStore, resourceBaseURL string) *ListingService {
	return &ListingService{
		store:           store,
		resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"),
		now:             time.Now,
	}
}

func (s *ListingService) List(ctx context.Context, req ListingRequest) (domain.AmazingListingData, error) {
	params := parseListingParams(req.PathParams)
	if atoi(params["page"]) == 0 {
		params["page"] = req.QueryPage
		if params["page"] == "" {
			params["page"] = "0"
		}
	}

	filter := amazingRepo.SoftwareFilter{
		CategoryID:  atoi(params["category_id"]),
		IsRecommend: req.Action == "recommend",
	}
	pageSize := 20
	page := atoi(params["page"])
	total, err := s.store.CountActive(ctx, filter)
	if err != nil {
		return domain.AmazingListingData{}, fmt.Errorf("count amazing listing: %w", err)
	}
	rows, err := s.store.ListActive(ctx, filter, total, page, pageSize, orderBy(req.Action))
	if err != nil {
		return domain.AmazingListingData{}, fmt.Errorf("list amazing listing: %w", err)
	}
	resolved := resourceurl.Resolved{BaseURL: s.resourceBaseURL, Timestamp: s.now().Unix()}
	if s.resources != nil {
		resolved, err = s.resources.ResolveContext(ctx)
		if err != nil {
			return domain.AmazingListingData{}, err
		}
	}
	s.prefixResourceRows(rows, resolved)
	return domain.AmazingListingData{
		Now:      s.now().Unix(),
		Rows:     rows,
		PageInfo: pageInfo(total, pageSize, page, "/amazing/"+req.Action+"-"+buildListingParams(params, map[string]string{"page": "[?]"})),
	}, nil
}

func parseListingParams(raw string) map[string]string {
	keys := []string{"category_id", "orderby", "page"}
	defaults := []string{"0", "0", "1"}
	values := []string{}
	if raw != "" {
		values = strings.Split(raw, "-")
	}
	params := map[string]string{}
	for i, key := range keys {
		value := defaults[i]
		if i < len(values) && values[i] != "" {
			value = values[i]
		}
		params[key] = value
	}
	return params
}

func buildListingParams(params map[string]string, replace map[string]string) string {
	keys := []string{"category_id", "orderby", "page"}
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		value := params[key]
		if next, ok := replace[key]; ok {
			value = next
		}
		values = append(values, value)
	}
	return strings.Join(values, "-")
}

func orderBy(action string) string {
	switch action {
	case "hot":
		return "dl_count DESC"
	case "latest":
		return "id DESC"
	default:
		return "id DESC"
	}
}

func (s *ListingService) prefixResourceRows(rows []map[string]interface{}, resolved resourceurl.Resolved) {
	for _, row := range rows {
		for _, key := range []string{"icon", "image"} {
			value := fmt.Sprint(row[key])
			row[key] = resolved.GetRes(value, "")
		}
	}
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
		"plist":     []map[string]interface{}{pageLink("curr", page, page, pageURL)},
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
		"pages":     pageSelector(totalPage),
	}
}

func pageSelector(totalPage int) []int {
	pages := make([]int, 0, totalPage)
	for i := 1; i <= totalPage; i++ {
		pages = append(pages, i)
	}
	return pages
}

func pageLink(pos string, page int, text interface{}, pageURL string) map[string]interface{} {
	return map[string]interface{}{
		"pos":  pos,
		"page": page,
		"text": text,
		"url":  strings.ReplaceAll(pageURL, "[?]", strconv.Itoa(page)),
	}
}

func atoi(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}

func ternary[T any](ok bool, yes T, no T) T {
	if ok {
		return yes
	}
	return no
}
