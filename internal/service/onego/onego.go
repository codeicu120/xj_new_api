package onego

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"xj_comp/internal/domain"
)

var ErrNotOpen = errors.New("onego not open")
var ErrSelectRoom = errors.New("onego select room")
var ErrActivityEnded = errors.New("onego activity ended")
var ErrNoData = errors.New("onego no data")
var ErrMissingPlaintext = errors.New("onego missing plaintext")
var ErrHashNumberUnavailable = errors.New("onego hash number unavailable")

type Store interface {
	Rules(ctx context.Context) (map[string]interface{}, error)
	Rooms(ctx context.Context) ([]map[string]interface{}, error)
	RoomByID(ctx context.Context, roomID int) (map[string]interface{}, error)
	CurrentRecords(ctx context.Context, roomID int, now int64) ([]map[string]interface{}, error)
	LatestRecord(ctx context.Context) (map[string]interface{}, error)
	RecordsByRoom(ctx context.Context, roomID int, page int, pageSize int) ([]map[string]interface{}, error)
	RecordsByPeriod(ctx context.Context, period string, page int, pageSize int) ([]map[string]interface{}, error)
	UserByID(ctx context.Context, uid int) (map[string]interface{}, error)
	BotByID(ctx context.Context, uid int) (map[string]interface{}, error)
}

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: time.Now}
}

func (s *Service) Rules(ctx context.Context) (domain.OneGoData, error) {
	row, err := s.store.Rules(ctx)
	if err != nil {
		return domain.OneGoData{}, fmt.Errorf("get onego rules: %w", err)
	}
	if len(row) == 0 {
		return domain.OneGoData{}, ErrNotOpen
	}
	return domain.OneGoData{Data: row}, nil
}

func (s *Service) Rooms(ctx context.Context) (domain.OneGoData, error) {
	rows, err := s.store.Rooms(ctx)
	if err != nil {
		return domain.OneGoData{}, fmt.Errorf("get onego rooms: %w", err)
	}
	if len(rows) == 0 {
		return domain.OneGoData{}, ErrNotOpen
	}
	return domain.OneGoData{Data: rows}, nil
}

func (s *Service) Current(ctx context.Context, roomID int) (domain.OneGoData, error) {
	rules, err := s.store.Rules(ctx)
	if err != nil {
		return domain.OneGoData{}, fmt.Errorf("get onego rules: %w", err)
	}
	if len(rules) == 0 {
		return domain.OneGoData{}, ErrNotOpen
	}
	room, err := s.store.RoomByID(ctx, roomID)
	if err != nil {
		return domain.OneGoData{}, fmt.Errorf("get onego room: %w", err)
	}
	if len(room) == 0 {
		return domain.OneGoData{}, ErrSelectRoom
	}
	records, err := s.store.CurrentRecords(ctx, roomID, s.now().Unix())
	if err != nil {
		return domain.OneGoData{}, fmt.Errorf("get onego current records: %w", err)
	}
	if len(records) == 0 {
		return domain.OneGoData{}, ErrActivityEnded
	}
	records, err = s.processRecords(ctx, records)
	if err != nil {
		return domain.OneGoData{}, err
	}
	return domain.OneGoData{Data: map[string]interface{}{"rules": rules, "current": records[0]}}, nil
}

func (s *Service) Last(ctx context.Context, roomID int, page int) (domain.OneGoData, error) {
	var records []map[string]interface{}
	var err error
	if roomID > 0 {
		room, err := s.store.RoomByID(ctx, roomID)
		if err != nil {
			return domain.OneGoData{}, fmt.Errorf("get onego room: %w", err)
		}
		if len(room) == 0 {
			return domain.OneGoData{}, ErrSelectRoom
		}
		records, err = s.store.RecordsByRoom(ctx, roomID, normalizePage(page), 20)
	} else {
		latest, err := s.store.LatestRecord(ctx)
		if err != nil {
			return domain.OneGoData{}, fmt.Errorf("get onego latest record: %w", err)
		}
		if len(latest) == 0 {
			return domain.OneGoData{}, ErrNoData
		}
		records, err = s.store.RecordsByPeriod(ctx, str(latest["period"]), normalizePage(page), 20)
	}
	if err != nil {
		return domain.OneGoData{}, fmt.Errorf("get onego records: %w", err)
	}
	records, err = s.processRecords(ctx, records)
	if err != nil {
		return domain.OneGoData{}, err
	}
	return domain.OneGoData{Data: records}, nil
}

func (s *Service) Hash(plaintext string) (domain.OneGoData, error) {
	plaintext = strings.TrimSpace(plaintext)
	if plaintext == "" {
		return domain.OneGoData{}, ErrMissingPlaintext
	}

	sum := sha256.Sum256([]byte(plaintext))
	hashCode := hex.EncodeToString(sum[:])
	digits := digitsOnly(hashCode)
	if len(digits) < 6 {
		return domain.OneGoData{}, ErrHashNumberUnavailable
	}

	needLength := 6
	hashNumber := digits[len(digits)-needLength:]
	for hashNumber[0] == '0' {
		needLength++
		if len(digits) < needLength {
			return domain.OneGoData{}, ErrHashNumberUnavailable
		}
		hashNumber = digits[len(digits)-needLength:]
	}

	return domain.OneGoData{Data: map[string]interface{}{
		"hash_code":   hashCode,
		"hash_number": hashNumber,
	}}, nil
}

func (s *Service) processRecords(ctx context.Context, rows []map[string]interface{}) ([]map[string]interface{}, error) {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		item := cloneMap(row)
		for _, key := range []string{"id", "start_time", "end_time", "hash_period", "room_id", "total_bets", "total_coins", "open_no", "awards", "win_rate", "open_time"} {
			item[key] = atoi(item[key])
		}
		winner := atoi(item["winner"])
		item["winner"] = winner
		if winner > 0 {
			user, err := s.store.UserByID(ctx, winner)
			if err != nil {
				return nil, fmt.Errorf("get onego winner user: %w", err)
			}
			if len(user) > 0 {
				item["winner"] = user
			}
		}
		if winner < 0 {
			bot, err := s.store.BotByID(ctx, -winner)
			if err != nil {
				return nil, fmt.Errorf("get onego winner bot: %w", err)
			}
			if len(bot) > 0 {
				item["winner"] = bot
			}
		}
		delete(item, "bot")
		out = append(out, item)
	}
	return out, nil
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func cloneMap(row map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(row))
	for key, value := range row {
		out[key] = value
	}
	return out
}

func atoi(value interface{}) int {
	var parsed int
	_, _ = fmt.Sscan(fmt.Sprint(value), &parsed)
	return parsed
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func digitsOnly(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range value {
		if r >= '0' && r <= '9' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
