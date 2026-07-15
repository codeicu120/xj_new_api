package onego

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Rules(ctx context.Context) (map[string]interface{}, error) {
	if r.db == nil {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM one_go LIMIT 1")
	if err != nil {
		return nil, fmt.Errorf("query onego rules: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) Rooms(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM one_go_rooms WHERE 1=1 ORDER BY id ASC LIMIT 10")
}

func (r *Repository) RoomByID(ctx context.Context, roomID int) (map[string]interface{}, error) {
	if r.db == nil || roomID <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM one_go_rooms WHERE id=?", roomID)
	if err != nil {
		return nil, fmt.Errorf("query onego room: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) CurrentRecords(ctx context.Context, roomID int, now int64) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM one_go_records WHERE 1=1 AND open_time=0 AND room_id=? AND start_time<=? AND end_time>=? ORDER BY id DESC LIMIT 10", roomID, now, now)
}

func (r *Repository) LatestRecord(ctx context.Context) (map[string]interface{}, error) {
	if r.db == nil {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM one_go_records WHERE open_time > 0 ORDER BY id DESC LIMIT 1")
	if err != nil {
		return nil, fmt.Errorf("query onego latest record: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) RecordsByRoom(ctx context.Context, roomID int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM one_go_records WHERE 1=1 AND open_time>0 AND room_id=? ORDER BY id DESC LIMIT ? OFFSET ?", roomID, pageSize, offset(page, pageSize))
}

func (r *Repository) RecordsByPeriod(ctx context.Context, period string, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM one_go_records WHERE 1=1 AND period=? ORDER BY id DESC LIMIT ? OFFSET ?", period, pageSize, offset(page, pageSize))
}

func (r *Repository) RankWinCoins(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT SUM(awards) as total_awards, winner FROM one_go_records WHERE awards > 0 GROUP BY winner ORDER BY total_awards DESC")
}

func (r *Repository) UserWins(ctx context.Context, uid int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT COUNT(*) as wins, room_id FROM one_go_records WHERE winner = ? GROUP BY room_id", uid)
}

func (r *Repository) UserOrdersGrouped(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM one_go_orders WHERE 1=1 AND uid=? GROUP BY period, room_id ORDER BY id DESC LIMIT ? OFFSET ?", uid, pageSize, offset(page, pageSize))
}

func (r *Repository) UserOrdersByPeriod(ctx context.Context, period string, roomID int, uid int) ([]map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM one_go_orders WHERE period=? AND room_id=? AND uid=?", period, roomID, uid)
}

func (r *Repository) RecordByPeriod(ctx context.Context, period string, roomID int) (map[string]interface{}, error) {
	if r.db == nil || period == "" || roomID <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM one_go_records WHERE period=? AND room_id=?", period, roomID)
	if err != nil {
		return nil, fmt.Errorf("query onego record by period: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) RankBetCoins(ctx context.Context, period string, roomID int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT uid, room_id, SUM(bet_coins) as total_coins, COUNT(*) as total_bets FROM one_go_orders WHERE period=? AND room_id=? GROUP BY uid ORDER BY total_coins DESC LIMIT ? OFFSET ?", period, roomID, pageSize, offset(page, pageSize))
}

func (r *Repository) UserByID(ctx context.Context, uid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT uid, username, avatar FROM users WHERE uid=?", uid)
	if err != nil {
		return nil, fmt.Errorf("query onego winner user: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) BotByID(ctx context.Context, uid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM bot_users WHERE uid=?", uid)
	if err != nil {
		return nil, fmt.Errorf("query onego winner bot: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func offset(page int, pageSize int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * pageSize
}

func (r *Repository) queryRows(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("read columns: %w", err)
	}
	result := []map[string]interface{}{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		row := make(map[string]interface{}, len(columns))
		for i, column := range columns {
			row[column] = normalizeSQLValue(values[i])
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}
	return result, nil
}

func normalizeSQLValue(value interface{}) interface{} {
	switch v := value.(type) {
	case nil:
		return nil
	case []byte:
		return string(v)
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprint(v)
	}
}
