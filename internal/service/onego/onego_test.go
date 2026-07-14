package onego

import (
	"context"
	"errors"
	"testing"
)

type fakeStore struct {
	rules          map[string]interface{}
	rooms          []map[string]interface{}
	room           map[string]interface{}
	currentRecords []map[string]interface{}
	latestRecord   map[string]interface{}
	periodRecords  []map[string]interface{}
	roomRecords    []map[string]interface{}
	user           map[string]interface{}
	bot            map[string]interface{}
	err            error
}

func (s fakeStore) Rules(context.Context) (map[string]interface{}, error) {
	return s.rules, s.err
}

func (s fakeStore) Rooms(context.Context) ([]map[string]interface{}, error) {
	return s.rooms, s.err
}

func (s fakeStore) RoomByID(context.Context, int) (map[string]interface{}, error) {
	return s.room, s.err
}

func (s fakeStore) CurrentRecords(context.Context, int, int64) ([]map[string]interface{}, error) {
	return s.currentRecords, s.err
}

func (s fakeStore) LatestRecord(context.Context) (map[string]interface{}, error) {
	return s.latestRecord, s.err
}

func (s fakeStore) RecordsByRoom(context.Context, int, int, int) ([]map[string]interface{}, error) {
	return s.roomRecords, s.err
}

func (s fakeStore) RecordsByPeriod(context.Context, string, int, int) ([]map[string]interface{}, error) {
	return s.periodRecords, s.err
}

func (s fakeStore) UserByID(context.Context, int) (map[string]interface{}, error) {
	return s.user, s.err
}

func (s fakeStore) BotByID(context.Context, int) (map[string]interface{}, error) {
	return s.bot, s.err
}

func TestRulesReturnsData(t *testing.T) {
	service := NewService(fakeStore{rules: map[string]interface{}{"id": "1", "title": "一元购规则"}})

	data, err := service.Rules(context.Background())
	if err != nil {
		t.Fatalf("rules: %v", err)
	}
	row, ok := data.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected rules map, got %T", data.Data)
	}
	if row["title"] != "一元购规则" {
		t.Fatalf("unexpected rules %#v", row)
	}
}

func TestRulesNotOpen(t *testing.T) {
	service := NewService(fakeStore{})

	_, err := service.Rules(context.Background())
	if !errors.Is(err, ErrNotOpen) {
		t.Fatalf("expected ErrNotOpen, got %v", err)
	}
}

func TestRoomsReturnsData(t *testing.T) {
	service := NewService(fakeStore{rooms: []map[string]interface{}{{"id": "1", "name": "10金币场"}}})

	data, err := service.Rooms(context.Background())
	if err != nil {
		t.Fatalf("rooms: %v", err)
	}
	rows, ok := data.Data.([]map[string]interface{})
	if !ok {
		t.Fatalf("expected room rows, got %T", data.Data)
	}
	if len(rows) != 1 || rows[0]["name"] != "10金币场" {
		t.Fatalf("unexpected rooms %#v", rows)
	}
}

func TestRoomsNotOpen(t *testing.T) {
	service := NewService(fakeStore{})

	_, err := service.Rooms(context.Background())
	if !errors.Is(err, ErrNotOpen) {
		t.Fatalf("expected ErrNotOpen, got %v", err)
	}
}

func TestCurrentReturnsRulesAndRecord(t *testing.T) {
	service := NewService(fakeStore{
		rules:          map[string]interface{}{"id": "1"},
		room:           map[string]interface{}{"id": "2"},
		currentRecords: []map[string]interface{}{sampleRecord("5")},
		user:           map[string]interface{}{"uid": "5", "username": "winner", "avatar": ""},
	})

	data, err := service.Current(context.Background(), 2)
	if err != nil {
		t.Fatalf("current: %v", err)
	}
	payload, ok := data.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected current payload, got %T", data.Data)
	}
	current := payload["current"].(map[string]interface{})
	if current["id"] != 10 || current["winner"].(map[string]interface{})["username"] != "winner" {
		t.Fatalf("unexpected current %#v", current)
	}
}

func TestCurrentRequiresRulesAndRoom(t *testing.T) {
	service := NewService(fakeStore{})
	if _, err := service.Current(context.Background(), 1); !errors.Is(err, ErrNotOpen) {
		t.Fatalf("expected ErrNotOpen, got %v", err)
	}

	service = NewService(fakeStore{rules: map[string]interface{}{"id": "1"}})
	if _, err := service.Current(context.Background(), 1); !errors.Is(err, ErrSelectRoom) {
		t.Fatalf("expected ErrSelectRoom, got %v", err)
	}
}

func TestLastReturnsLatestPeriodRecords(t *testing.T) {
	service := NewService(fakeStore{
		latestRecord:  map[string]interface{}{"period": "2026071401"},
		periodRecords: []map[string]interface{}{sampleRecord("0")},
	})

	data, err := service.Last(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("last: %v", err)
	}
	rows, ok := data.Data.([]map[string]interface{})
	if !ok {
		t.Fatalf("expected rows, got %T", data.Data)
	}
	if len(rows) != 1 || rows[0]["winner"] != 0 || rows[0]["room_id"] != 1 {
		t.Fatalf("unexpected rows %#v", rows)
	}
}

func TestLastNoData(t *testing.T) {
	service := NewService(fakeStore{})

	_, err := service.Last(context.Background(), 0, 1)
	if !errors.Is(err, ErrNoData) {
		t.Fatalf("expected ErrNoData, got %v", err)
	}
}

func TestHashReturnsSHA256AndNumber(t *testing.T) {
	service := NewService(fakeStore{})

	data, err := service.Hash(" abc ")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	payload, ok := data.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected hash payload, got %T", data.Data)
	}
	if payload["hash_code"] != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("unexpected hash_code %v", payload["hash_code"])
	}
	if payload["hash_number"] != "120015" {
		t.Fatalf("unexpected hash_number %v", payload["hash_number"])
	}
}

func TestHashRequiresPlaintext(t *testing.T) {
	service := NewService(fakeStore{})

	_, err := service.Hash(" \t ")
	if !errors.Is(err, ErrMissingPlaintext) {
		t.Fatalf("expected ErrMissingPlaintext, got %v", err)
	}
}

func sampleRecord(winner string) map[string]interface{} {
	return map[string]interface{}{
		"id":          "10",
		"start_time":  "1770000000",
		"end_time":    "1770000600",
		"period":      "2026071401",
		"hash_code":   "abc",
		"hash_period": "123456",
		"room_id":     "1",
		"total_bets":  "3",
		"total_coins": "30",
		"open_no":     "-1",
		"winner":      winner,
		"awards":      "0",
		"win_rate":    "6500",
		"bot":         "0",
		"paid":        "0",
		"open_time":   "0",
	}
}
