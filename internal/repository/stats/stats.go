package stats

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

func (r *Repository) GuestExists(ctx context.Context, sid string) (bool, error) {
	if r.db == nil || sid == "" {
		return false, nil
	}
	var exists int
	if err := r.db.QueryRowContext(ctx, "SELECT 1 FROM user_guests WHERE sid=? LIMIT 1", sid).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("query guest: %w", err)
	}
	return exists == 1, nil
}

func (r *Repository) ShortcutCreatedByIP(ctx context.Context, ip string) (map[string]interface{}, error) {
	if r.db == nil || ip == "" {
		return map[string]interface{}{}, nil
	}
	return r.queryOne(ctx, "SELECT * FROM shortcut_created WHERE user_ip=?", ip)
}

func (r *Repository) CreateShortcut(ctx context.Context, ip string, now int64) (int64, error) {
	if r.db == nil {
		return 0, nil
	}
	result, err := r.db.ExecContext(ctx, "INSERT INTO shortcut_created(user_ip, create_time) VALUES(?, ?) ON DUPLICATE KEY UPDATE create_time=?", ip, now, now)
	if err != nil {
		return 0, fmt.Errorf("insert shortcut_created: %w", err)
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func (r *Repository) ShortcutStatsByDate(ctx context.Context, statsDate int64) (map[string]interface{}, error) {
	if r.db == nil || statsDate <= 0 {
		return map[string]interface{}{}, nil
	}
	return r.queryOne(ctx, "SELECT * FROM shortcut_stats WHERE stats_date=?", statsDate)
}

func (r *Repository) CreateShortcutStats(ctx context.Context, statsDate int64) error {
	if r.db == nil {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "INSERT INTO shortcut_stats(stats_date, count) VALUES(?, 1)", statsDate); err != nil {
		return fmt.Errorf("insert shortcut_stats: %w", err)
	}
	return nil
}

func (r *Repository) UpdateShortcutStatsCount(ctx context.Context, id int, count int) error {
	if r.db == nil {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE shortcut_stats SET count=? WHERE id=?", count, id); err != nil {
		return fmt.Errorf("update shortcut_stats: %w", err)
	}
	return nil
}

func (r *Repository) AdStatBySID(ctx context.Context, sid string, title string, url string) (map[string]interface{}, error) {
	if r.db == nil || sid == "" || title == "" || url == "" {
		return map[string]interface{}{}, nil
	}
	return r.queryOne(ctx, "SELECT * FROM ad_stats WHERE sid=? AND title=? AND url=?", sid, title, url)
}

func (r *Repository) CreateAdStat(ctx context.Context, sid string, title string, url string, pos int, click int, install int) error {
	if r.db == nil {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "INSERT INTO ad_stats(sid, title, url, pos, click, install) VALUES(?, ?, ?, ?, ?, ?)", sid, title, url, pos, click, install); err != nil {
		return fmt.Errorf("insert ad_stats: %w", err)
	}
	return nil
}

func (r *Repository) UpdateAdStat(ctx context.Context, id int, click int, install int) error {
	if r.db == nil {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE ad_stats SET click=?, install=? WHERE id=?", click, install, id); err != nil {
		return fmt.Errorf("update ad_stats: %w", err)
	}
	return nil
}

func (r *Repository) PlayStatBySID(ctx context.Context, sid string, vid int) (map[string]interface{}, error) {
	if r.db == nil || sid == "" || vid <= 0 {
		return map[string]interface{}{}, nil
	}
	return r.queryOne(ctx, "SELECT * FROM play_stats WHERE sid=? AND vid=?", sid, vid)
}

func (r *Repository) CreatePlayStat(ctx context.Context, sid string, vid int, mini int, duration int, played int) error {
	if r.db == nil {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "INSERT INTO play_stats(sid, vid, mini, duration, played) VALUES(?, ?, ?, ?, ?)", sid, vid, mini, duration, played); err != nil {
		return fmt.Errorf("insert play_stats: %w", err)
	}
	return nil
}

func (r *Repository) CreateGuest(ctx context.Context, sid string, now int64) error {
	if r.db == nil || sid == "" {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "INSERT IGNORE INTO user_guests(sid, goldcoin, timestamp) VALUES(?, 0, ?)", sid, now); err != nil {
		return fmt.Errorf("insert guest: %w", err)
	}
	return nil
}

func (r *Repository) UpdatePlayStatPlayed(ctx context.Context, id int, played int) error {
	if r.db == nil {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE play_stats SET played=? WHERE id=?", played, id); err != nil {
		return fmt.Errorf("update play_stats: %w", err)
	}
	return nil
}

func (r *Repository) queryOne(ctx context.Context, query string, args ...interface{}) (map[string]interface{}, error) {
	if r.db == nil {
		return map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return map[string]interface{}{}, nil
	}
	return items[0], nil
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
