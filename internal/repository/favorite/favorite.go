package favorite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Kind string

const (
	KindVOD  Kind = "vod"
	KindMini Kind = "mini"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Items(ctx context.Context, kind Kind, uid int, page int, pageSize int, keyword string) (int, []map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return 0, []map[string]interface{}{}, nil
	}
	table, showtype := tableSpec(kind)
	if table == "" {
		return 0, []map[string]interface{}{}, nil
	}
	where := "uid=?"
	args := []interface{}{uid}
	if strings.TrimSpace(keyword) != "" {
		where += " AND title LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	countQuery := "SELECT COUNT(*) FROM " + table
	if strings.TrimSpace(keyword) != "" {
		countQuery += " f LEFT JOIN vods v ON f.vodid=v.vodid WHERE " + where
	} else {
		countQuery += " WHERE " + where
	}
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return 0, nil, fmt.Errorf("count favorites: %w", err)
	}
	offset := limitOffset(total, pageSize, page)
	listWhere := "a.uid=? AND b.showtype=?"
	listArgs := []interface{}{uid, showtype}
	if strings.TrimSpace(keyword) != "" {
		listWhere += " AND title LIKE ?"
		listArgs = append(listArgs, "%"+keyword+"%")
	}
	listArgs = append(listArgs, pageSize, offset)
	rows, err := r.queryRows(ctx, "SELECT a.*, b.* FROM "+table+" a LEFT JOIN vods b ON b.vodid=a.vodid WHERE "+listWhere+" ORDER BY a.favtime DESC LIMIT ? OFFSET ?", listArgs...)
	if err != nil {
		return 0, nil, fmt.Errorf("query favorites: %w", err)
	}
	return total, rows, nil
}

func (r *Repository) Remove(ctx context.Context, kind Kind, uid int, vodid int) (int, error) {
	if r.db == nil || uid <= 0 || vodid <= 0 {
		return 0, nil
	}
	table, _ := tableSpec(kind)
	if table == "" {
		return 0, nil
	}
	result, err := r.db.ExecContext(ctx, "DELETE FROM "+table+" WHERE uid=? AND vodid=?", uid, vodid)
	if err != nil {
		return 0, fmt.Errorf("delete favorite: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("favorite rows affected: %w", err)
	}
	return int(count), nil
}

func (r *Repository) VODByID(ctx context.Context, vodid int) (map[string]interface{}, error) {
	if r.db == nil || vodid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM vods WHERE vodid=?", vodid)
	if err != nil {
		return nil, fmt.Errorf("query favorite vod: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) Count(ctx context.Context, kind Kind, uid int, vodid int, since int64) (int, error) {
	if r.db == nil || uid <= 0 {
		return 0, nil
	}
	table, _ := tableSpec(kind)
	if table == "" {
		return 0, nil
	}
	where := "uid=?"
	args := []interface{}{uid}
	if vodid > 0 {
		where += " AND vodid=?"
		args = append(args, vodid)
	}
	if since > 0 {
		where += " AND favtime>=?"
		args = append(args, since)
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+" WHERE "+where, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count favorite rows: %w", err)
	}
	return total, nil
}

func (r *Repository) Add(ctx context.Context, kind Kind, uid int, vodid int, now int64) error {
	if r.db == nil || uid <= 0 || vodid <= 0 {
		return nil
	}
	table, _ := tableSpec(kind)
	if table == "" {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "INSERT INTO "+table+"(uid, vodid, favtime) VALUES(?, ?, ?)", uid, vodid, now); err != nil {
		return fmt.Errorf("insert favorite: %w", err)
	}
	return nil
}

func tableSpec(kind Kind) (string, int) {
	if kind == KindMini {
		return "minivod_favorites", 1
	}
	if kind == KindVOD {
		return "vod_favorites", 0
	}
	return "", 0
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
