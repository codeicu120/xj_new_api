package onego

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
)

var ErrNotOpen = errors.New("onego not open")
var ErrSelectRoom = errors.New("onego select room")
var ErrActivityEnded = errors.New("onego activity ended")
var ErrNoData = errors.New("onego no data")
var ErrMissingPlaintext = errors.New("onego missing plaintext")
var ErrHashNumberUnavailable = errors.New("onego hash number unavailable")
var ErrInvalidRoom = errors.New("onego invalid room")
var ErrInvalidPeriod = errors.New("onego invalid period")

type Store interface {
	Rules(ctx context.Context) (map[string]interface{}, error)
	Rooms(ctx context.Context) ([]map[string]interface{}, error)
	RoomByID(ctx context.Context, roomID int) (map[string]interface{}, error)
	CurrentRecords(ctx context.Context, roomID int, now int64) ([]map[string]interface{}, error)
	LatestRecord(ctx context.Context) (map[string]interface{}, error)
	RecordsByRoom(ctx context.Context, roomID int, page int, pageSize int) ([]map[string]interface{}, error)
	RecordsByPeriod(ctx context.Context, period string, page int, pageSize int) ([]map[string]interface{}, error)
	RankWinCoins(ctx context.Context) ([]map[string]interface{}, error)
	UserWins(ctx context.Context, uid int) ([]map[string]interface{}, error)
	UserOrdersGrouped(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error)
	UserOrdersByPeriod(ctx context.Context, period string, roomID int, uid int) ([]map[string]interface{}, error)
	RecordByPeriod(ctx context.Context, period string, roomID int) (map[string]interface{}, error)
	RankBetCoins(ctx context.Context, period string, roomID int, page int, pageSize int) ([]map[string]interface{}, error)
	UserByID(ctx context.Context, uid int) (map[string]interface{}, error)
	BotByID(ctx context.Context, uid int) (map[string]interface{}, error)
	Quota(ctx context.Context, uid int) (map[string]interface{}, error)
}

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type Service struct {
	store Store
	auth  AuthStore
	now   func() time.Time
}

func NewService(store Store, auth ...AuthStore) *Service {
	var authStore AuthStore
	if len(auth) > 0 {
		authStore = auth[0]
	}
	return &Service{store: store, auth: authStore, now: time.Now}
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

func (s *Service) Lucky(ctx context.Context) (domain.OneGoData, error) {
	ranks, err := s.store.RankWinCoins(ctx)
	if err != nil {
		return domain.OneGoData{}, fmt.Errorf("get onego lucky ranks: %w", err)
	}
	for _, row := range ranks {
		winner := atoi(row["winner"])
		row["total_awards"] = atoi(row["total_awards"])
		row["winner"] = winner

		wins, err := s.store.UserWins(ctx, winner)
		if err != nil {
			return domain.OneGoData{}, fmt.Errorf("get onego lucky user wins: %w", err)
		}
		for _, win := range wins {
			win["wins"] = atoi(win["wins"])
			win["room_id"] = atoi(win["room_id"])
		}
		row["wins"] = wins
	}

	ranks, err = s.processRecords(ctx, ranks)
	if err != nil {
		return domain.OneGoData{}, err
	}
	return domain.OneGoData{Data: ranks}, nil
}

func (s *Service) History(ctx context.Context, token string, page int) (domain.OneGoData, int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return domain.OneGoData{}, -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.OneGoData{}, -9999, "您还没有登录", nil
	}
	orders, err := s.store.UserOrdersGrouped(ctx, uid, normalizePage(page), 20)
	if err != nil {
		return domain.OneGoData{}, -1, "获取一元购历史失败", err
	}
	for _, row := range orders {
		row["id"] = atoi(row["id"])
		row["uid"] = atoi(row["uid"])
		row["room_id"] = atoi(row["room_id"])
		row["bet_coins"] = atoi(row["bet_coins"])

		period := str(row["period"])
		roomID := atoi(row["room_id"])
		bets, err := s.store.UserOrdersByPeriod(ctx, period, roomID, uid)
		if err != nil {
			return domain.OneGoData{}, -1, "获取一元购历史失败", err
		}
		open, err := s.store.RecordByPeriod(ctx, period, roomID)
		if err != nil {
			return domain.OneGoData{}, -1, "获取一元购历史失败", err
		}
		betNos := []string{}
		winNo := -1
		winCoins := 0
		openNo := str(open["open_no"])
		if atoi(open["open_no"]) >= 0 {
			winNo = atoi(open["open_no"])
		}
		for _, bet := range bets {
			parts := splitCSV(str(bet["bet_no"]))
			betNos = append(betNos, parts...)
			for _, part := range parts {
				if part == openNo {
					winCoins = atoi(open["awards"])
				}
			}
		}
		row["bet_no"] = betNos
		row["win_no"] = winNo
		row["win_coins"] = winCoins
	}
	return domain.OneGoData{Data: orders}, 0, "", nil
}

func (s *Service) BetEdge(ctx context.Context, token string, period string, roomID int, quantity int) (int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -9999, "您还没有登录", nil
	}
	if quantity < 1 {
		return -1, "押注数量不能为零", nil
	}
	room, err := s.store.RoomByID(ctx, roomID)
	if err != nil {
		return -1, "一元购投注失败", err
	}
	if len(room) == 0 {
		return -1, "无效场次", nil
	}
	record, err := s.store.RecordByPeriod(ctx, period, roomID)
	if err != nil {
		return -1, "一元购投注失败", err
	}
	if len(record) == 0 {
		return -1, "无效的活动期号", nil
	}
	now := s.now().Unix()
	if int64(atoi(record["start_time"])) > now {
		return -1, "活动尚未开始", nil
	}
	if int64(atoi(record["end_time"])) < now {
		return -1, "活动已结束", nil
	}
	quota, err := s.store.Quota(ctx, atoi(user["uid"]))
	if err != nil {
		return -1, "一元购投注失败", err
	}
	if len(quota) == 0 {
		return -1, "未知用户", nil
	}
	if atoi(quota["goldcoin"]) < quantity*atoi(room["coins"]) {
		return -1, "余额不足", nil
	}
	return -1, "一元购投注成功分支暂未迁移", nil
}

func (s *Service) BetRanks(ctx context.Context, period string, roomID int, page int) (domain.OneGoData, error) {
	room, err := s.store.RoomByID(ctx, roomID)
	if err != nil {
		return domain.OneGoData{}, fmt.Errorf("get onego bet ranks room: %w", err)
	}
	if len(room) == 0 {
		return domain.OneGoData{}, ErrInvalidRoom
	}
	record, err := s.store.RecordByPeriod(ctx, period, roomID)
	if err != nil {
		return domain.OneGoData{}, fmt.Errorf("get onego bet ranks record: %w", err)
	}
	if len(record) == 0 {
		return domain.OneGoData{}, ErrInvalidPeriod
	}
	ranks, err := s.store.RankBetCoins(ctx, period, roomID, normalizePage(page), 20)
	if err != nil {
		return domain.OneGoData{}, fmt.Errorf("get onego bet ranks: %w", err)
	}
	ranks, err = s.processOrderRows(ctx, ranks)
	if err != nil {
		return domain.OneGoData{}, err
	}
	return domain.OneGoData{Data: ranks}, nil
}

func (s *Service) Marquee(ctx context.Context) (domain.OneGoData, error) {
	latest, err := s.store.LatestRecord(ctx)
	if err != nil {
		return domain.OneGoData{}, fmt.Errorf("get onego latest record: %w", err)
	}
	if len(latest) == 0 {
		return domain.OneGoData{}, ErrNoData
	}
	rules, err := s.store.Rules(ctx)
	if err != nil {
		return domain.OneGoData{}, fmt.Errorf("get onego rules: %w", err)
	}
	if len(rules) == 0 {
		return domain.OneGoData{}, ErrNotOpen
	}
	records, err := s.store.RecordsByPeriod(ctx, str(latest["period"]), 1, 10)
	if err != nil {
		return domain.OneGoData{}, fmt.Errorf("get onego marquee records: %w", err)
	}
	records, err = s.processRecords(ctx, records)
	if err != nil {
		return domain.OneGoData{}, err
	}

	messages := make([]string, 0, len(records))
	template := str(rules["marquee"])
	for _, row := range records {
		if atoi(row["awards"]) == 0 {
			continue
		}
		room, err := s.store.RoomByID(ctx, atoi(row["room_id"]))
		if err != nil {
			return domain.OneGoData{}, fmt.Errorf("get onego marquee room: %w", err)
		}
		if len(room) == 0 {
			continue
		}
		message := strings.ReplaceAll(template, "{user}", winnerUsername(row["winner"]))
		message = strings.ReplaceAll(message, "{room}", str(room["name"]))
		message = strings.ReplaceAll(message, "{period}", str(row["period"]))
		message = strings.ReplaceAll(message, "{awards}", str(row["awards"]))
		message = strings.ReplaceAll(message, "{win_rate}", formatRate(atoi(row["win_rate"])))
		messages = append(messages, message)
	}
	return domain.OneGoData{Data: messages}, nil
}

func (s *Service) userByToken(ctx context.Context, token string) (map[string]interface{}, error) {
	sid := userRepo.CleanToken(token)
	if sid == "" || s.auth == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	user, err := s.auth.UserBySession(ctx, sid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	return user, nil
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

func (s *Service) processOrderRows(ctx context.Context, rows []map[string]interface{}) ([]map[string]interface{}, error) {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		item := cloneMap(row)
		uid := atoi(item["uid"])
		item["uid"] = uid
		if _, ok := item["total_coins"]; ok {
			item["total_coins"] = atoi(item["total_coins"])
		}
		if _, ok := item["total_bets"]; ok {
			item["total_bets"] = atoi(item["total_bets"])
		}
		roomID := atoi(item["room_id"])
		if roomID > 0 {
			room, err := s.store.RoomByID(ctx, roomID)
			if err != nil {
				return nil, fmt.Errorf("get onego order room: %w", err)
			}
			if len(room) > 0 {
				item["room_name"] = str(room["name"])
				coins := atoi(room["coins"])
				if coins > 0 {
					item["total_bets"] = atoi(item["total_coins"]) / coins
				}
			}
		}
		if uid > 0 {
			user, err := s.store.UserByID(ctx, uid)
			if err != nil {
				return nil, fmt.Errorf("get onego order user: %w", err)
			}
			if len(user) > 0 {
				item["user"] = user
			}
		}
		if uid < 0 {
			bot, err := s.store.BotByID(ctx, -uid)
			if err != nil {
				return nil, fmt.Errorf("get onego order bot: %w", err)
			}
			if len(bot) > 0 {
				item["user"] = bot
			}
		}
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

func splitCSV(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
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

func winnerUsername(value interface{}) string {
	winner, ok := value.(map[string]interface{})
	if !ok {
		return ""
	}
	return str(winner["username"])
}

func formatRate(value int) string {
	return strconv.FormatFloat(float64(value)/100, 'f', -1, 64)
}
