package history

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Kind string

const (
	KindPlay Kind = "play"
	KindDown Kind = "down"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Items(ctx context.Context, kind Kind, uid int, sid string, page int, pageSize int, timeline int, now int64) (int, []map[string]interface{}, error) {
	if r.db == nil {
		return 0, []map[string]interface{}{}, nil
	}
	spec := specFor(kind, uid > 0)
	if spec.table == "" {
		return 0, []map[string]interface{}{}, nil
	}
	where, args := historyWhere(kind, spec, uid, sid, timeline, now)
	total, err := r.count(ctx, spec.table, where, args)
	if err != nil {
		return 0, nil, err
	}
	offset := limitOffset(total, pageSize, page)
	logRows, err := r.queryRows(ctx, "SELECT * FROM "+spec.table+" WHERE "+where+" ORDER BY "+spec.timeField+" DESC LIMIT ? OFFSET ?", append(args, pageSize, offset)...)
	if err != nil {
		return 0, nil, fmt.Errorf("query history logs: %w", err)
	}
	if len(logRows) == 0 {
		return total, []map[string]interface{}{}, nil
	}
	vodRows, err := r.vodsByIDs(ctx, vodIDs(logRows))
	if err != nil {
		return 0, nil, err
	}
	return total, mergeVODRows(logRows, vodRows), nil
}

func (r *Repository) Remove(ctx context.Context, kind Kind, uid int, sid string, vodid int) (int, error) {
	if r.db == nil || vodid <= 0 {
		return 0, nil
	}
	spec := specFor(kind, uid > 0)
	if spec.table == "" {
		return 0, nil
	}
	if kind == KindPlay && uid > 0 {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return 0, fmt.Errorf("begin playlog remove: %w", err)
		}
		count, err := execRemove(ctx, tx, "vod_playlogs", "uid", uid, "", vodid)
		if err == nil {
			_, err = execRemove(ctx, tx, "vod_playlogs_week", "uid", uid, "", vodid)
		}
		if err != nil {
			_ = tx.Rollback()
			return 0, err
		}
		if err := tx.Commit(); err != nil {
			return 0, fmt.Errorf("commit playlog remove: %w", err)
		}
		return count, nil
	}
	result, err := r.db.ExecContext(ctx, "UPDATE "+spec.table+" SET showtype=9 WHERE "+spec.owner+"=? AND vodid=? AND showtype=0", ownerValue(spec.owner, uid, sid), vodid)
	if err != nil {
		return 0, fmt.Errorf("remove history log: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("remove history rows affected: %w", err)
	}
	return int(count), nil
}

type tableSpec struct {
	table     string
	owner     string
	timeField string
}

func specFor(kind Kind, user bool) tableSpec {
	switch kind {
	case KindPlay:
		if user {
			return tableSpec{table: "vod_playlogs", owner: "uid", timeField: "playtime"}
		}
		return tableSpec{table: "vod_guest_playlogs", owner: "sid", timeField: "playtime"}
	case KindDown:
		if user {
			return tableSpec{table: "vod_downlogs", owner: "uid", timeField: "downtime"}
		}
		return tableSpec{table: "vod_guest_downlogs", owner: "sid", timeField: "downtime"}
	default:
		return tableSpec{}
	}
}

func execRemove(ctx context.Context, tx *sql.Tx, table string, owner string, uid int, sid string, vodid int) (int, error) {
	result, err := tx.ExecContext(ctx, "UPDATE "+table+" SET showtype=9 WHERE "+owner+"=? AND vodid=? AND showtype=0", ownerValue(owner, uid, sid), vodid)
	if err != nil {
		return 0, fmt.Errorf("remove history log: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("remove history rows affected: %w", err)
	}
	return int(count), nil
}

func ownerValue(owner string, uid int, sid string) interface{} {
	if owner == "uid" {
		return uid
	}
	return sid
}

func historyWhere(kind Kind, spec tableSpec, uid int, sid string, timeline int, now int64) (string, []interface{}) {
	where := spec.owner + "=? AND showtype=0"
	args := []interface{}{sid}
	if spec.owner == "uid" {
		args[0] = uid
	}
	switch kind {
	case KindPlay:
		switch timeline {
		case 1:
			where += " AND playtime>?"
			args = append(args, now-86400*7)
		case 2:
			where += " AND playtime BETWEEN ? AND ?"
			if spec.owner == "sid" {
				args = append(args, now-86400*7, now-86400*30)
			} else {
				args = append(args, now-86400*30, now-86400*7)
			}
		case 3:
			where += " AND playtime BETWEEN ? AND ?"
			if spec.owner == "sid" {
				args = append(args, now-86400*30, now-86400*90)
			} else {
				args = append(args, now-86400*90, now-86400*30)
			}
		}
	case KindDown:
		switch timeline {
		case 1:
			where += " AND downtime>?"
			args = append(args, now-86400*7)
		case 2:
			where += " AND downtime>?"
			args = append(args, now-86400*30)
		case 3:
			where += " AND downtime<?"
			args = append(args, now-86400*30)
		}
	}
	return where, args
}

func (r *Repository) count(ctx context.Context, table string, where string, args []interface{}) (int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+" WHERE "+where, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count history logs: %w", err)
	}
	return total, nil
}

func (r *Repository) vodsByIDs(ctx context.Context, ids []int) ([]map[string]interface{}, error) {
	if len(ids) == 0 {
		return []map[string]interface{}{}, nil
	}
	placeholders := make([]string, 0, len(ids))
	args := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	if len(args) == 0 {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM vods WHERE vodid IN("+strings.Join(placeholders, ",")+") AND showtype=0", args...)
	if err != nil {
		return nil, fmt.Errorf("query history vods: %w", err)
	}
	return rows, nil
}

func (r *Repository) queryRows(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

func vodIDs(rows []map[string]interface{}) []int {
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, atoi(row["vodid"]))
	}
	return ids
}

func mergeVODRows(logRows []map[string]interface{}, vodRows []map[string]interface{}) []map[string]interface{} {
	byID := map[int]map[string]interface{}{}
	for _, row := range vodRows {
		byID[atoi(row["vodid"])] = row
	}
	merged := []map[string]interface{}{}
	for _, logRow := range logRows {
		vodRow, ok := byID[atoi(logRow["vodid"])]
		if !ok {
			continue
		}
		row := make(map[string]interface{}, len(logRow)+len(vodRow))
		for key, value := range logRow {
			row[key] = value
		}
		for key, value := range vodRow {
			if _, exists := row[key]; !exists {
				row[key] = value
			}
		}
		if _, exists := row["title"]; exists {
			merged = append(merged, row)
		}
	}
	return merged
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

func limitOffset(total int, pageSize int, page int) int {
	if pageSize < 1 {
		pageSize = 1
	}
	totalPage := (total + pageSize - 1) / pageSize
	if totalPage < 1 {
		totalPage = 1
	}
	if page < 1 {
		page = 1
	}
	if page > totalPage {
		page = totalPage
	}
	return (page - 1) * pageSize
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
