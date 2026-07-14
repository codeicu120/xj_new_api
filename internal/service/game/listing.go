package game

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"xj_comp/internal/domain"
)

type GameStore interface {
	ListActive(ctx context.Context, platformID int, categoryID int) ([]map[string]interface{}, error)
}

type BroadcastStore interface {
	ListActive(ctx context.Context) ([]map[string]interface{}, error)
}

type ListingService struct {
	store           GameStore
	resourceBaseURL string
}

func NewListingService(store GameStore, resourceBaseURL string) *ListingService {
	return &ListingService{
		store:           store,
		resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"),
	}
}

func (s *ListingService) List(ctx context.Context, platformID int, categoryID int) (domain.GamesData, error) {
	rows, err := s.store.ListActive(ctx, platformID, categoryID)
	if err != nil {
		return domain.GamesData{}, fmt.Errorf("list games: %w", err)
	}
	for _, row := range rows {
		prefixResource(row, "icon", s.resourceBaseURL)
		prefixResource(row, "image", s.resourceBaseURL)
		prefixResource(row, "cover", s.resourceBaseURL)
	}
	return domain.GamesData{Data: rows}, nil
}

type BroadcastService struct {
	store BroadcastStore
	rand  *rand.Rand
}

func NewBroadcastService(store BroadcastStore) *BroadcastService {
	return &BroadcastService{
		store: store,
		rand:  rand.New(rand.NewSource(1)),
	}
}

func (s *BroadcastService) List(ctx context.Context) (domain.GameBroadcastsData, error) {
	rows, err := s.store.ListActive(ctx)
	if err != nil {
		return domain.GameBroadcastsData{}, fmt.Errorf("list broadcasts: %w", err)
	}

	msgs := []string{}
	for _, row := range rows {
		account := fmt.Sprintf("%d*******%d", 13+s.rand.Intn(7), 10+s.rand.Intn(90))
		minValue := atoi(fmt.Sprint(row["min_value"]))
		maxValue := atoi(fmt.Sprint(row["max_value"]))
		amount := minValue
		if maxValue > minValue {
			amount = minValue + s.rand.Intn(maxValue-minValue+1)
		}
		msg := strings.ReplaceAll(fmt.Sprint(row["msg"]), "{user}", account)
		msg = strings.ReplaceAll(msg, "{amount}", fmt.Sprint(amount))
		msgs = append(msgs, msg)
	}
	return domain.GameBroadcastsData{Data: msgs}, nil
}

func atoi(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}
