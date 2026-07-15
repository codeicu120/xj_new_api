package minivod

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Filter struct {
	CateIDs    []int
	AreaID     int
	YearID     int
	TagIDs     []int
	Definition int
	Duration   int
	FreeOnly   bool
	Mosaic     int
	LangVoice  int
	Recommend  bool
	TopIDs     []int
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Categories(ctx context.Context) ([]map[string]interface{}, error) {
	return r.queryRows(ctx, "SELECT cateid,parentid,uuid,catename FROM vod_categories ORDER BY sort ASC, cateid ASC")
}

func (r *Repository) Areas(ctx context.Context) ([]map[string]interface{}, error) {
	return r.queryRows(ctx, "SELECT * FROM vod_areas ORDER BY sortnum ASC, areaid ASC")
}

func (r *Repository) Years(ctx context.Context) ([]map[string]interface{}, error) {
	return r.queryRows(ctx, "SELECT * FROM vod_years ORDER BY sortnum ASC, yearid ASC")
}

func (r *Repository) Servers(ctx context.Context) ([]map[string]interface{}, error) {
	return r.queryRows(ctx, "SELECT * FROM vod_servers ORDER BY sortnum ASC, srvid ASC")
}

func (r *Repository) TagsByNames(ctx context.Context, names []string) ([]map[string]interface{}, error) {
	if r.db == nil || len(names) == 0 {
		return []map[string]interface{}{}, nil
	}
	seen := map[string]struct{}{}
	args := []interface{}{}
	holders := []string{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		args = append(args, name)
		holders = append(holders, "?")
	}
	if len(args) == 0 {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM vod_tags WHERE tagname IN("+strings.Join(holders, ",")+")", args...)
}

func (r *Repository) Count(ctx context.Context, filter Filter, now int64) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	where, args := buildWhere(filter, now)
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vods WHERE 1=1 "+where, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count minivods: %w", err)
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, filter Filter, total int, page int, pageSize int, orderBy string, now int64) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	where, args := buildWhere(filter, now)
	offset := limitOffset(total, pageSize, page)
	args = append(args, pageSize, offset)
	return r.queryRows(ctx, "SELECT * FROM vods WHERE 1=1 "+where+" ORDER BY "+orderBy+" LIMIT ? OFFSET ?", args...)
}

func (r *Repository) CountByAuthor(ctx context.Context, authorID int) (int, error) {
	if r.db == nil || authorID <= 0 {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vods WHERE authorid=? AND showtype=1", authorID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count author minivods: %w", err)
	}
	return total, nil
}

func (r *Repository) ListByAuthor(ctx context.Context, authorID int, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error) {
	if r.db == nil || authorID <= 0 {
		return []map[string]interface{}{}, nil
	}
	offset := limitOffset(total, pageSize, page)
	return r.queryRows(ctx, "SELECT * FROM vods WHERE authorid=? AND showtype=1 ORDER BY "+orderBy+" LIMIT ? OFFSET ?", authorID, pageSize, offset)
}

func (r *Repository) Random(ctx context.Context, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM vods WHERE showtype=1 ORDER BY RAND() LIMIT ?", pageSize)
}

func (r *Repository) VODByID(ctx context.Context, vodID int) (map[string]interface{}, error) {
	if r.db == nil || vodID <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM vods WHERE vodid=?", vodID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) UserByID(ctx context.Context, uid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM users WHERE uid=?", uid)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) SimilarVODsByTagIDs(ctx context.Context, tagIDs []int, excludeID int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil || len(tagIDs) == 0 || pageSize <= 0 {
		return []map[string]interface{}{}, nil
	}
	ids := intListSQL(tagIDs)
	if ids == "NULL" {
		return []map[string]interface{}{}, nil
	}
	query := "SELECT * FROM vods WHERE vodid IN(SELECT vodid FROM vod_tagmaps WHERE tagid IN(" + ids + ")) AND showtype=0"
	if excludeID > 0 {
		query += fmt.Sprintf(" AND vodid<>%d", excludeID)
	}
	query += " ORDER BY utimestamp DESC LIMIT ?"
	return r.queryRows(ctx, query, pageSize)
}

func (r *Repository) RandomVODsExcept(ctx context.Context, pageSize int, excludeID int, cateID int) ([]map[string]interface{}, error) {
	if r.db == nil || pageSize <= 0 {
		return []map[string]interface{}{}, nil
	}
	where := "showtype=0"
	args := []interface{}{}
	if excludeID > 0 {
		where += " AND vodid<>?"
		args = append(args, excludeID)
	}
	if cateID > 0 {
		where += " AND cateid=?"
		args = append(args, cateID)
	}
	args = append(args, pageSize)
	return r.queryRows(ctx, "SELECT * FROM vods WHERE "+where+" ORDER BY RAND() LIMIT ?", args...)
}

func (r *Repository) Setting(ctx context.Context, key string) (string, error) {
	if r.db == nil || strings.TrimSpace(key) == "" {
		return "", nil
	}
	var value sql.NullString
	var typ sql.NullString
	if err := r.db.QueryRowContext(ctx, "SELECT value,type FROM settings WHERE uuid=?", key).Scan(&value, &typ); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("query setting %s: %w", key, err)
	}
	if typ.String == "array" {
		return strings.Join(serializedIntStrings(value.String), ","), nil
	}
	return value.String, nil
}

func (r *Repository) UsersByIDs(ctx context.Context, ids []int) ([]map[string]interface{}, error) {
	if r.db == nil || len(ids) == 0 {
		return []map[string]interface{}{}, nil
	}
	idList := intListSQL(ids)
	if idList == "NULL" {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM users WHERE uid IN("+idList+")")
}

func (r *Repository) UpDownByUser(ctx context.Context, uid int, vodID int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 || vodID <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM vod_updowns WHERE uid=? AND vodid=? LIMIT 1", uid, vodID)
	if err != nil {
		return nil, fmt.Errorf("query updown: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) DeleteUpDown(ctx context.Context, uid int, vodID int) error {
	if r.db == nil || uid <= 0 || vodID <= 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "DELETE FROM vod_updowns WHERE uid=? AND vodid=?", uid, vodID); err != nil {
		return fmt.Errorf("delete updown: %w", err)
	}
	return nil
}

func (r *Repository) SaveUpDown(ctx context.Context, uid int, vodID int, updown int, now int64) (int, error) {
	if r.db == nil || uid <= 0 || vodID <= 0 || (updown != 1 && updown != 2) {
		return 0, nil
	}
	result, err := r.db.ExecContext(ctx, "INSERT IGNORE INTO vod_updowns(vodid, uid, updown, addtime) VALUES(?, ?, ?, ?)", vodID, uid, updown, now)
	if err != nil {
		return 0, fmt.Errorf("save updown: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("save updown rows affected: %w", err)
	}
	return int(affected), nil
}

func (r *Repository) IncrementVODCounter(ctx context.Context, vodID int, field string, delta int) error {
	if r.db == nil || vodID <= 0 {
		return nil
	}
	if field != "upnum" && field != "downnum" {
		return fmt.Errorf("unsupported vod counter %q", field)
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE vods SET "+field+"=GREATEST("+field+"+?,0) WHERE vodid=?", delta, vodID); err != nil {
		return fmt.Errorf("increment vod %s: %w", field, err)
	}
	return nil
}

func (r *Repository) RecountUpDown(ctx context.Context, vodID int) error {
	if r.db == nil || vodID <= 0 {
		return nil
	}
	query := `UPDATE vods SET
		upnum=(SELECT COUNT(*) FROM vod_updowns WHERE vodid=? AND updown=1),
		downnum=(SELECT COUNT(*) FROM vod_updowns WHERE vodid=? AND updown=2)
		WHERE vodid=?`
	if _, err := r.db.ExecContext(ctx, query, vodID, vodID, vodID); err != nil {
		return fmt.Errorf("recount updown: %w", err)
	}
	return nil
}

func (r *Repository) FavoriteCount(ctx context.Context, uid int, vodID int) (int, error) {
	if r.db == nil || uid <= 0 || vodID <= 0 {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM minivod_favorites WHERE uid=? AND vodid=?", uid, vodID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count minivod favorite: %w", err)
	}
	return total, nil
}

func (r *Repository) MiniViewLog(ctx context.Context, uid int, sid string, vodID int) (map[string]interface{}, error) {
	if r.db == nil || vodID <= 0 {
		return map[string]interface{}{}, nil
	}
	if uid > 0 {
		rows, err := r.queryRows(ctx, "SELECT * FROM minivod_viewlogs WHERE uid=? AND vodid=? LIMIT 1", uid, vodID)
		return firstRow(rows, err)
	}
	if sid != "" {
		rows, err := r.queryRows(ctx, "SELECT * FROM minivod_guestviewlogs WHERE sid=? AND vodid=? LIMIT 1", sid, vodID)
		return firstRow(rows, err)
	}
	return map[string]interface{}{}, nil
}

func (r *Repository) CountMiniViewLogsSince(ctx context.Context, uid int, sid string, since int64, action int) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	field := "playtime"
	if action == 2 {
		field = "downtime"
	}
	var total int
	if uid > 0 {
		err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM minivod_viewlogs WHERE uid=? AND showtype=1 AND "+field+">=?", uid, since).Scan(&total)
		if err != nil {
			return 0, fmt.Errorf("count minivod viewlogs: %w", err)
		}
		return total, nil
	}
	if sid == "" {
		return 0, nil
	}
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM minivod_guestviewlogs WHERE sid=? AND showtype=1 AND "+field+">=?", sid, since).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count minivod guest viewlogs: %w", err)
	}
	return total, nil
}

func firstRow(rows []map[string]interface{}, err error) (map[string]interface{}, error) {
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func buildWhere(filter Filter, now int64) (string, []interface{}) {
	where := " AND showtype=1"
	args := []interface{}{}
	if filter.Recommend {
		where += " AND prop3=1"
	}
	if len(filter.TopIDs) > 0 {
		where += " AND vodid IN(" + intListSQL(filter.TopIDs) + ")"
	}
	if len(filter.TagIDs) > 0 {
		where += " AND vodid IN(SELECT vodid FROM vod_tagmaps WHERE tagid IN(" + intListSQL(filter.TagIDs) + "))"
	}
	if len(filter.CateIDs) > 0 {
		where += " AND cateid IN(" + intListSQL(filter.CateIDs) + ")"
	}
	if filter.AreaID > 0 {
		where += " AND areaid=?"
		args = append(args, filter.AreaID)
	}
	if filter.YearID > 0 {
		where += " AND yearid=?"
		args = append(args, filter.YearID)
	}
	if filter.Definition > 0 {
		where += " AND definition=?"
		args = append(args, filter.Definition)
	}
	if filter.Duration == 1 {
		where += " AND duration>1800"
	} else if filter.Duration == 2 {
		where += " AND duration<=1800"
	}
	if filter.FreeOnly {
		where += " AND (view_price=0 OR (free_sdate<? AND free_edate>?))"
		args = append(args, now, now)
	}
	if filter.Mosaic > 0 {
		where += " AND mosaic=?"
		args = append(args, filter.Mosaic)
	}
	if filter.LangVoice > 0 {
		where += " AND langvoice=?"
		args = append(args, filter.LangVoice)
	}
	return where, args
}

func (r *Repository) queryRows(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
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

func intListSQL(ids []int) string {
	parts := []string{}
	for _, id := range ids {
		if id > 0 {
			parts = append(parts, fmt.Sprint(id))
		}
	}
	if len(parts) == 0 {
		return "NULL"
	}
	return strings.Join(parts, ",")
}

func serializedIntStrings(value string) []string {
	matches := regexp.MustCompile(`s:\d+:"(\d+)"`).FindAllStringSubmatch(value, -1)
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			out = append(out, match[1])
		}
	}
	return out
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
