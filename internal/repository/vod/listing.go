package vod

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type ListingFilter struct {
	CateIDs     []int
	AreaID      int
	YearID      int
	Definition  int
	Duration    int
	FreeType    int
	Mosaic      int
	LangVoice   int
	Recommend   bool
	CTimeAfter  int64
	CTimeBefore int64
}

type SpecialFilter struct {
	SPType int
}

type ErrorReportInput struct {
	UID        string
	VODID      int
	PlayURL    string
	AppVersion string
	SysVersion string
	Model      string
	Channel    string
	Network    string
	ClientIP   string
	Details    string
	Now        int64
}

type ListingRepository struct {
	db *sql.DB
}

func NewListingRepository(db *sql.DB) *ListingRepository {
	return &ListingRepository{db: db}
}

func (r *ListingRepository) Categories(ctx context.Context) ([]map[string]interface{}, error) {
	return r.queryRows(ctx, "SELECT cateid,parentid,uuid,catename FROM vod_categories ORDER BY sort ASC, cateid ASC")
}

func (r *ListingRepository) Areas(ctx context.Context) ([]map[string]interface{}, error) {
	return r.queryRows(ctx, "SELECT * FROM vod_areas ORDER BY sortnum ASC, areaid ASC")
}

func (r *ListingRepository) Years(ctx context.Context) ([]map[string]interface{}, error) {
	return r.queryRows(ctx, "SELECT * FROM vod_years ORDER BY sortnum ASC, yearid ASC")
}

func (r *ListingRepository) Servers(ctx context.Context) ([]map[string]interface{}, error) {
	return r.queryRows(ctx, "SELECT * FROM vod_servers ORDER BY sortnum ASC, srvid ASC")
}

func (r *ListingRepository) CountVODs(ctx context.Context, filter ListingFilter) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	where, args := buildWhere(filter)
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vods WHERE 1=1 "+where, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count vods: %w", err)
	}
	return total, nil
}

func (r *ListingRepository) ListVODs(ctx context.Context, filter ListingFilter, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	where, args := buildWhere(filter)
	offset := limitOffset(total, pageSize, page)
	query := "SELECT * FROM vods WHERE 1=1 " + where + " ORDER BY " + orderBy + " LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	return r.queryRows(ctx, query, args...)
}

func (r *ListingRepository) RandomVODs(ctx context.Context, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM vods WHERE showtype=0 ORDER BY RAND() LIMIT ?", pageSize)
}

func (r *ListingRepository) RandomVODsExcept(ctx context.Context, pageSize int, excludeID int, cateID int) ([]map[string]interface{}, error) {
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

func (r *ListingRepository) VODByID(ctx context.Context, vodID int) (map[string]interface{}, error) {
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

func (r *ListingRepository) VODErrorByUID(ctx context.Context, uid string, vodID int) (map[string]interface{}, error) {
	if r.db == nil || strings.TrimSpace(uid) == "" || vodID <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM vod_errors WHERE uid=? AND vodid=?", uid, vodID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *ListingRepository) SaveVODError(ctx context.Context, input ErrorReportInput) (int, error) {
	if r.db == nil || strings.TrimSpace(input.UID) == "" || input.VODID <= 0 {
		return 0, nil
	}
	result, err := r.db.ExecContext(ctx, `
INSERT INTO vod_errors(uid, vodid, play_url, app_ver, sys_ver, model, channel, network, client_ip, details, status, create_time, update_time)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		input.UID,
		input.VODID,
		input.PlayURL,
		input.AppVersion,
		input.SysVersion,
		input.Model,
		input.Channel,
		input.Network,
		input.ClientIP,
		input.Details,
		input.Now,
		input.Now,
	)
	if err != nil {
		return 0, fmt.Errorf("insert vod error report: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert vod error report id: %w", err)
	}
	return int(id), nil
}

func (r *ListingRepository) SimilarVODsByTagIDs(ctx context.Context, tagIDs []int, excludeID int, since int64, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil || len(tagIDs) == 0 || pageSize <= 0 {
		return []map[string]interface{}{}, nil
	}
	holders := make([]string, 0, len(tagIDs))
	args := make([]interface{}, 0, len(tagIDs)+3)
	for _, id := range tagIDs {
		if id <= 0 {
			continue
		}
		holders = append(holders, "?")
		args = append(args, id)
	}
	if len(args) == 0 {
		return []map[string]interface{}{}, nil
	}
	args = append(args, since)
	query := "SELECT * FROM vods WHERE vodid IN(SELECT vodid FROM vod_tagmaps WHERE tagid IN(" + strings.Join(holders, ",") + ")) AND showtype=0 AND utimestamp>?"
	if excludeID > 0 {
		query += " AND vodid<>?"
		args = append(args, excludeID)
	}
	query += " ORDER BY RAND() LIMIT ?"
	args = append(args, pageSize)
	return r.queryRows(ctx, query, args...)
}

func (r *ListingRepository) TagsByNames(ctx context.Context, names []string) ([]map[string]interface{}, error) {
	if r.db == nil || len(names) == 0 {
		return []map[string]interface{}{}, nil
	}
	unique := map[string]struct{}{}
	args := []interface{}{}
	placeholders := []string{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := unique[name]; ok {
			continue
		}
		unique[name] = struct{}{}
		args = append(args, name)
		placeholders = append(placeholders, "?")
	}
	if len(args) == 0 {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM vod_tags WHERE tagname IN("+strings.Join(placeholders, ",")+")", args...)
}

func (r *ListingRepository) CalldataByUUID(ctx context.Context, uuid string) (map[string]interface{}, error) {
	if r.db == nil || strings.TrimSpace(uuid) == "" {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM maintain_calldata WHERE uuid=?", uuid)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *ListingRepository) VODsByIDsLimited(ctx context.Context, ids []int, freeOnly bool, limit int, orderByField bool) ([]map[string]interface{}, error) {
	if r.db == nil || len(ids) == 0 || limit <= 0 {
		return []map[string]interface{}{}, nil
	}
	idList := intListSQL(ids)
	if idList == "NULL" {
		return []map[string]interface{}{}, nil
	}
	where := "vodid IN(" + idList + ") AND showtype=0"
	if freeOnly {
		where += " AND isvip=0"
	}
	orderBy := "upnum DESC"
	if orderByField {
		orderBy = "FIELD(vodid, " + idList + ")"
	}
	return r.queryRows(ctx, "SELECT * FROM vods WHERE "+where+" ORDER BY "+orderBy+" LIMIT ?", limit)
}

func (r *ListingRepository) SearchVODs(ctx context.Context, keyword string, freeOnly bool, limit int) ([]map[string]interface{}, error) {
	if r.db == nil || strings.TrimSpace(keyword) == "" || limit <= 0 {
		return []map[string]interface{}{}, nil
	}
	where := "showtype=0 AND (title LIKE ? OR tags LIKE ? OR actor_tags LIKE ? OR vodkey LIKE ?)"
	if freeOnly {
		where += " AND isvip=0"
	}
	like := "%" + keyword + "%"
	return r.queryRows(ctx, "SELECT * FROM vods WHERE "+where+" ORDER BY upnum DESC LIMIT ?", like, like, like, like, limit)
}

func (r *ListingRepository) SearchLog(ctx context.Context, keyword string) (map[string]interface{}, error) {
	if r.db == nil || strings.TrimSpace(keyword) == "" {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM vod_schlogs WHERE schwd=?", keyword)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *ListingRepository) UpsertSearchLog(ctx context.Context, keyword string, now int64, total int, vodIDs []int) error {
	if r.db == nil || strings.TrimSpace(keyword) == "" {
		return nil
	}
	if err := r.updateOrInsertSearchLog(ctx, "vod_schlogs", keyword, now, total, vodIDs); err != nil {
		return fmt.Errorf("upsert search log: %w", err)
	}
	return nil
}

func (r *ListingRepository) updateOrInsertSearchLog(ctx context.Context, table string, keyword string, now int64, total int, vodIDs []int) error {
	result, err := r.db.ExecContext(
		ctx,
		"UPDATE "+table+" SET schtime=?, total=?, vodids=? WHERE schwd=?",
		now,
		total,
		joinInts(vodIDs),
		keyword,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	_, err = r.db.ExecContext(
		ctx,
		"INSERT INTO "+table+"(schwd,schtime,total,vodids) VALUES(?,?,?,?)",
		keyword,
		now,
		total,
		joinInts(vodIDs),
	)
	return err
}

func (r *ListingRepository) IncrementSearchLog(ctx context.Context, keyword string, previous int64, now int64) error {
	if r.db == nil || strings.TrimSpace(keyword) == "" {
		return nil
	}
	monthValue := "sch_month+1"
	if !sameMonth(previous, now) {
		monthValue = "1"
	}
	dayValue := "sch_day+1"
	if !sameDay(previous, now) {
		dayValue = "1"
	}
	weekValue := "sch_week+1"
	if !sameWeek(previous, now) {
		weekValue = "1"
	}
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE vod_schlogs SET sch_lasttime=?, sch_total=sch_total+1, sch_month="+monthValue+", sch_day="+dayValue+", sch_week="+weekValue+" WHERE schwd=?",
		now,
		keyword,
	)
	if err != nil {
		return fmt.Errorf("increment search log: %w", err)
	}
	return nil
}

func (r *ListingRepository) TopSearchVODIDs(ctx context.Context) (string, error) {
	if r.db == nil {
		return "", nil
	}
	var value sql.NullString
	if err := r.db.QueryRowContext(ctx, "SELECT vodids FROM vod_schlogs WHERE 1=1 ORDER BY total DESC LIMIT 1").Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("top search vodids: %w", err)
	}
	return value.String, nil
}

func (r *ListingRepository) MiniVODsByIDsLimited(ctx context.Context, ids []int, limit int, orderByField bool) ([]map[string]interface{}, error) {
	if r.db == nil || len(ids) == 0 || limit <= 0 {
		return []map[string]interface{}{}, nil
	}
	idList := intListSQL(ids)
	if idList == "NULL" {
		return []map[string]interface{}{}, nil
	}
	orderBy := "upnum DESC"
	if orderByField {
		orderBy = "FIELD(vodid, " + idList + ")"
	}
	return r.queryRows(ctx, "SELECT * FROM vods WHERE vodid IN("+idList+") AND showtype=1 ORDER BY "+orderBy+" LIMIT ?", limit)
}

func (r *ListingRepository) MiniVODsByIDs(ctx context.Context, ids []int, orderBy string) ([]map[string]interface{}, error) {
	if r.db == nil || len(ids) == 0 {
		return []map[string]interface{}{}, nil
	}
	idList := intListSQL(ids)
	if idList == "NULL" {
		return []map[string]interface{}{}, nil
	}
	if strings.TrimSpace(orderBy) == "" {
		orderBy = "vodid DESC"
	}
	return r.queryRows(ctx, "SELECT * FROM vods WHERE vodid IN("+idList+") AND showtype=1 ORDER BY "+orderBy)
}

func (r *ListingRepository) BreakingVOD(ctx context.Context, cateID int, since int64) (map[string]interface{}, error) {
	if r.db == nil {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM vods WHERE cateid=? AND utimestamp>=? LIMIT 1", cateID, since)
	if err != nil {
		return nil, fmt.Errorf("query breaking vod: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *ListingRepository) MiniSearchVODs(ctx context.Context, keyword string, limit int) ([]map[string]interface{}, error) {
	if r.db == nil || strings.TrimSpace(keyword) == "" || limit <= 0 {
		return []map[string]interface{}{}, nil
	}
	like := "%" + keyword + "%"
	return r.queryRows(ctx, "SELECT * FROM vods WHERE showtype=1 AND (title LIKE ? OR tags LIKE ? OR actor_tags LIKE ? OR vodkey LIKE ?) ORDER BY upnum DESC LIMIT ?", like, like, like, like, limit)
}

func (r *ListingRepository) MiniSearchLog(ctx context.Context, keyword string) (map[string]interface{}, error) {
	if r.db == nil || strings.TrimSpace(keyword) == "" {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM minivod_schlogs WHERE schwd=?", keyword)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *ListingRepository) UpsertMiniSearchLog(ctx context.Context, keyword string, now int64, total int, vodIDs []int) error {
	if r.db == nil || strings.TrimSpace(keyword) == "" {
		return nil
	}
	if err := r.updateOrInsertSearchLog(ctx, "minivod_schlogs", keyword, now, total, vodIDs); err != nil {
		return fmt.Errorf("upsert mini search log: %w", err)
	}
	return nil
}

func (r *ListingRepository) UpDownByUser(ctx context.Context, uid int, vodID int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 || vodID <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM vod_updowns WHERE uid=? AND vodid=?", uid, vodID)
	if err != nil {
		return nil, fmt.Errorf("query vod updown: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *ListingRepository) DeleteUpDown(ctx context.Context, uid int, vodID int) error {
	if r.db == nil || uid <= 0 || vodID <= 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "DELETE FROM vod_updowns WHERE uid=? AND vodid=?", uid, vodID); err != nil {
		return fmt.Errorf("delete vod updown: %w", err)
	}
	return nil
}

func (r *ListingRepository) SaveUpDown(ctx context.Context, uid int, vodID int, updown int, now int64) (int, error) {
	if r.db == nil || uid <= 0 || vodID <= 0 {
		return 0, nil
	}
	result, err := r.db.ExecContext(ctx, "INSERT IGNORE INTO vod_updowns(vodid, uid, updown, addtime) VALUES(?, ?, ?, ?)", vodID, uid, updown, now)
	if err != nil {
		return 0, fmt.Errorf("insert vod updown: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("vod updown rows affected: %w", err)
	}
	if affected == 0 {
		return 0, nil
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

func (r *ListingRepository) IncrementVODCounter(ctx context.Context, vodID int, field string, delta int) error {
	if r.db == nil || vodID <= 0 || delta == 0 {
		return nil
	}
	if field != "upnum" && field != "downnum" {
		return fmt.Errorf("invalid vod counter %s", field)
	}
	expr := field + "+?"
	if delta < 0 {
		expr = "GREATEST(" + field + "+?,0)"
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE vods SET "+field+"="+expr+" WHERE vodid=?", delta, vodID); err != nil {
		return fmt.Errorf("increment vod counter: %w", err)
	}
	return nil
}

func (r *ListingRepository) RecountUpDown(ctx context.Context, vodID int) error {
	if r.db == nil || vodID <= 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, `
UPDATE vods SET
	upnum=(SELECT COUNT(*) FROM vod_updowns WHERE vodid=? AND updown=1),
	downnum=(SELECT COUNT(*) FROM vod_updowns WHERE vodid=? AND updown=2)
WHERE vodid=?`, vodID, vodID, vodID); err != nil {
		return fmt.Errorf("recount vod updown: %w", err)
	}
	return nil
}

func (r *ListingRepository) IncrementMiniSearchLog(ctx context.Context, keyword string, previous int64, now int64) error {
	if r.db == nil || strings.TrimSpace(keyword) == "" {
		return nil
	}
	monthValue := "sch_month+1"
	if !sameMonth(previous, now) {
		monthValue = "1"
	}
	dayValue := "sch_day+1"
	if !sameDay(previous, now) {
		dayValue = "1"
	}
	weekValue := "sch_week+1"
	if !sameWeek(previous, now) {
		weekValue = "1"
	}
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE minivod_schlogs SET sch_lasttime=?, sch_total=sch_total+1, sch_month="+monthValue+", sch_day="+dayValue+", sch_week="+weekValue+" WHERE schwd=?",
		now,
		keyword,
	)
	if err != nil {
		return fmt.Errorf("increment mini search log: %w", err)
	}
	return nil
}

func (r *ListingRepository) TopMiniSearchVODIDs(ctx context.Context) (string, error) {
	if r.db == nil {
		return "", nil
	}
	var value sql.NullString
	if err := r.db.QueryRowContext(ctx, "SELECT vodids FROM minivod_schlogs WHERE 1=1 ORDER BY total DESC LIMIT 1").Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("top mini search vodids: %w", err)
	}
	return value.String, nil
}

func (r *ListingRepository) CountSpecials(ctx context.Context, filter SpecialFilter) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	where, args := buildSpecialWhere(filter)
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vod_specials WHERE 1=1 "+where, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count specials: %w", err)
	}
	return total, nil
}

func (r *ListingRepository) ListSpecials(ctx context.Context, filter SpecialFilter, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	where, args := buildSpecialWhere(filter)
	offset := limitOffset(total, pageSize, page)
	args = append(args, pageSize, offset)
	return r.queryRows(ctx, "SELECT * FROM vod_specials WHERE 1=1 "+where+" ORDER BY "+orderBy+" LIMIT ? OFFSET ?", args...)
}

func (r *ListingRepository) ListActorSpecials(ctx context.Context, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM vod_specials WHERE 1=1 AND sptype=1 ORDER BY updatetime DESC LIMIT ? OFFSET 0", pageSize)
}

func (r *ListingRepository) SpecialByID(ctx context.Context, spid int) (map[string]interface{}, error) {
	if r.db == nil || spid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM vod_specials WHERE spid=?", spid)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *ListingRepository) VODsByIDs(ctx context.Context, ids []int, orderBy string) ([]map[string]interface{}, error) {
	if r.db == nil || len(ids) == 0 {
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
	if strings.TrimSpace(orderBy) == "" {
		orderBy = "vodid DESC"
	}
	return r.queryRows(ctx, "SELECT * FROM vods WHERE vodid IN("+strings.Join(placeholders, ",")+") AND showtype=0 ORDER BY "+orderBy, args...)
}

func (r *ListingRepository) UpdateSpecialRand(ctx context.Context, minSPID int) error {
	if r.db == nil {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE vod_specials SET randnum=FLOOR(1+RAND()*65534) WHERE spid>? LIMIT 10", minSPID); err != nil {
		return fmt.Errorf("update special rand: %w", err)
	}
	return nil
}

func (r *ListingRepository) IncrementSpecialViews(ctx context.Context, spid int, viewsLastTime int64, now int64) error {
	if r.db == nil || spid <= 0 {
		return nil
	}
	updates := []string{"views_lasttime=?", "views_total=views_total+1"}
	args := []interface{}{now}
	if sameMonth(viewsLastTime, now) {
		updates = append(updates, "views_month=views_month+1")
	} else {
		updates = append(updates, "views_month=1")
	}
	if sameDay(viewsLastTime, now) {
		updates = append(updates, "views_day=views_day+1")
	} else {
		updates = append(updates, "views_day=1")
	}
	if sameWeek(viewsLastTime, now) {
		updates = append(updates, "views_week=views_week+1")
	} else {
		updates = append(updates, "views_week=1")
	}
	args = append(args, spid)
	if _, err := r.db.ExecContext(ctx, "UPDATE vod_specials SET "+strings.Join(updates, ",")+" WHERE spid=?", args...); err != nil {
		return fmt.Errorf("increment special views: %w", err)
	}
	return nil
}

func (r *ListingRepository) GuestBySID(ctx context.Context, sid string) (map[string]interface{}, error) {
	if r.db == nil || !validGuestSID(sid) {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM user_guests WHERE sid=?", sid)
	if err != nil {
		return nil, fmt.Errorf("query guest: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *ListingRepository) KeylimitCount(ctx context.Context, key string) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var total sql.NullInt64
	if err := r.db.QueryRowContext(ctx, "SELECT SUM(keynum) FROM keylimits WHERE keyid=?", md5Hex(key)).Scan(&total); err != nil {
		return 0, fmt.Errorf("query keylimit: %w", err)
	}
	if !total.Valid {
		return 0, nil
	}
	return int(total.Int64), nil
}

func (r *ListingRepository) SetKeylimit(ctx context.Context, key string, keynum int, keydata string, now int64) error {
	if r.db == nil {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "INSERT INTO keylimits(keyid, keynum, keydata, ctimestamp) VALUES(?, ?, ?, ?)", md5Hex(key), keynum, keydata, now); err != nil {
		return fmt.Errorf("insert keylimit: %w", err)
	}
	return nil
}

func (r *ListingRepository) IncrementSpecialVote(ctx context.Context, spid int, field string) error {
	if r.db == nil || spid <= 0 {
		return nil
	}
	if field != "upnum" && field != "downnum" {
		return fmt.Errorf("invalid special vote field %q", field)
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE vod_specials SET "+field+"="+field+"+1 WHERE spid=?", spid); err != nil {
		return fmt.Errorf("increment special vote: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE vod_specials SET scorenum=IF(upnum+downnum=0, 9.9, ROUND(upnum/(upnum+downnum)*10, 1)) WHERE spid=?", spid); err != nil {
		return fmt.Errorf("recount special score: %w", err)
	}
	return nil
}

func (r *ListingRepository) queryRows(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query vod rows: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func buildWhere(filter ListingFilter) (string, []interface{}) {
	parts := []string{}
	args := []interface{}{}
	if filter.Recommend {
		parts = append(parts, "prop3=1")
	}
	if len(filter.CateIDs) > 0 {
		holders := make([]string, 0, len(filter.CateIDs))
		for _, id := range filter.CateIDs {
			holders = append(holders, "?")
			args = append(args, id)
		}
		parts = append(parts, "cateid IN("+strings.Join(holders, ",")+")")
	}
	if filter.AreaID > 0 {
		parts = append(parts, "areaid=?")
		args = append(args, filter.AreaID)
	}
	if filter.YearID > 0 {
		parts = append(parts, "yearid=?")
		args = append(args, filter.YearID)
	}
	if filter.Definition > 0 {
		parts = append(parts, "definition=?")
		args = append(args, filter.Definition)
	}
	if filter.Duration > 0 {
		if filter.Duration == 1 {
			parts = append(parts, "duration>1800")
		} else {
			parts = append(parts, "duration<=1800")
		}
	}
	if filter.FreeType > 0 {
		now := time.Now().Unix()
		parts = append(parts, "(view_price=0 OR (free_sdate<? AND free_edate>?))")
		args = append(args, now, now)
	}
	if filter.Mosaic > 0 {
		parts = append(parts, "mosaic=?")
		args = append(args, filter.Mosaic)
	}
	if filter.LangVoice > 0 {
		parts = append(parts, "langvoice=?")
		args = append(args, filter.LangVoice)
	}
	if filter.CTimeAfter > 0 {
		parts = append(parts, "ctimestamp>?")
		args = append(args, filter.CTimeAfter)
	}
	if filter.CTimeBefore > 0 {
		parts = append(parts, "ctimestamp<?")
		args = append(args, filter.CTimeBefore)
	}
	parts = append(parts, "showtype=0")
	return " AND " + strings.Join(parts, " AND "), args
}

func buildSpecialWhere(filter SpecialFilter) (string, []interface{}) {
	parts := []string{}
	args := []interface{}{}
	if filter.SPType > 0 {
		parts = append(parts, "sptype=?")
		args = append(args, filter.SPType)
	}
	parts = append(parts, "showtype=0", "itemcount>=4")
	return " AND " + strings.Join(parts, " AND "), args
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

func sameMonth(a int64, b int64) bool {
	at := time.Unix(a, 0)
	bt := time.Unix(b, 0)
	return at.Year() == bt.Year() && at.Month() == bt.Month()
}

func sameDay(a int64, b int64) bool {
	at := time.Unix(a, 0)
	bt := time.Unix(b, 0)
	return at.Year() == bt.Year() && at.YearDay() == bt.YearDay()
}

func sameWeek(a int64, b int64) bool {
	at := time.Unix(a, 0)
	bt := time.Unix(b, 0)
	ay, aw := at.ISOWeek()
	by, bw := bt.ISOWeek()
	return ay == by && aw == bw
}

func intListSQL(ids []int) string {
	parts := make([]string, 0, len(ids))
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

func joinInts(ids []int) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			parts = append(parts, fmt.Sprint(id))
		}
	}
	return strings.Join(parts, ",")
}

func md5Hex(value string) string {
	sum := md5.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

func validGuestSID(sid string) bool {
	if len(sid) != 32 {
		return false
	}
	for _, ch := range sid {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') {
			return false
		}
	}
	return true
}
