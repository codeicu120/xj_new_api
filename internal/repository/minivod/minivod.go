package minivod

import (
	"context"
	"database/sql"
	"encoding/json"
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

const coinTypeMiniVODPlayTask = 25

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

func (r *Repository) LongToShortMapByLongID(ctx context.Context, vodID int) (map[string]interface{}, error) {
	if r.db == nil || vodID <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT start,end FROM vod_map_ls WHERE lvodid=? LIMIT 1", vodID)
	if err != nil {
		return nil, fmt.Errorf("query long to short map: %w", err)
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

func (r *Repository) UserQuota(ctx context.Context, uid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT uid,goldcoin FROM users_quota WHERE uid=?", uid)
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

func (r *Repository) VODsByIDs(ctx context.Context, ids []int, orderByField bool) ([]map[string]interface{}, error) {
	if r.db == nil || len(ids) == 0 {
		return []map[string]interface{}{}, nil
	}
	idList := intListSQL(ids)
	if idList == "NULL" {
		return []map[string]interface{}{}, nil
	}
	orderBy := "vodid DESC"
	if orderByField {
		orderBy = "FIELD(vodid, " + idList + ")"
	}
	return r.queryRows(ctx, "SELECT * FROM vods WHERE vodid IN("+idList+") AND showtype=1 ORDER BY "+orderBy)
}

func (r *Repository) PendingViewLogs(ctx context.Context, uid int, sid string, limit int) ([]map[string]interface{}, error) {
	if r.db == nil || limit <= 0 {
		return []map[string]interface{}{}, nil
	}
	if uid > 0 {
		return r.queryRows(ctx, "SELECT * FROM "+miniUserViewLogTable(uid)+" WHERE uid=? AND showtype=0 ORDER BY logid DESC LIMIT ?", uid, limit)
	}
	if sid == "" {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM "+miniGuestViewLogTable(sid)+" WHERE sid=? AND showtype=0 ORDER BY logid DESC LIMIT ?", sid, limit)
}

func (r *Repository) PullViewLogs(ctx context.Context, uid int, sid string) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	if uid > 0 {
		return r.pullUserViewLogs(ctx, uid)
	}
	if strings.TrimSpace(sid) == "" {
		return 0, nil
	}
	return r.pullGuestViewLogs(ctx, sid)
}

func (r *Repository) MarkViewLogsShown(ctx context.Context, uid int, sid string, logIDs []int, now int64) error {
	if r.db == nil || len(logIDs) == 0 {
		return nil
	}
	cleanIDs := make([]int, 0, len(logIDs))
	for _, id := range logIDs {
		if id > 0 {
			cleanIDs = append(cleanIDs, id)
		}
	}
	if len(cleanIDs) == 0 {
		return nil
	}
	table := miniGuestViewLogTable(sid)
	actorColumn := "sid"
	var actor interface{} = sid
	if uid > 0 {
		table = miniUserViewLogTable(uid)
		actorColumn = "uid"
		actor = uid
	} else if strings.TrimSpace(sid) == "" {
		return nil
	}
	placeholders := make([]string, 0, len(cleanIDs))
	args := []interface{}{now, actor}
	for _, id := range cleanIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	query := "UPDATE " + table + " SET reqtime=?, showtype=1 WHERE " + actorColumn + "=? AND logid IN(" + strings.Join(placeholders, ",") + ")"
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("mark minivod viewlogs shown: %w", err)
	}
	return nil
}

func (r *Repository) MiniVODAdCallRows(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT type,content FROM maintain_calldata WHERE uuid=? AND showtype=0 LIMIT 1", "minivod.ads")
	if err != nil {
		return nil, fmt.Errorf("query minivod ads calldata: %w", err)
	}
	if len(rows) == 0 || str(rows[0]["type"]) != "rows" || strings.TrimSpace(str(rows[0]["content"])) == "" {
		return []map[string]interface{}{}, nil
	}
	var callRows []map[string]interface{}
	if err := json.Unmarshal([]byte(str(rows[0]["content"])), &callRows); err != nil {
		return nil, fmt.Errorf("decode minivod ads calldata: %w", err)
	}
	return callRows, nil
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
		rows, err := r.queryRows(ctx, "SELECT * FROM "+miniUserViewLogTable(uid)+" WHERE uid=? AND vodid=? LIMIT 1", uid, vodID)
		return firstRow(rows, err)
	}
	if sid != "" {
		rows, err := r.queryRows(ctx, "SELECT * FROM "+miniGuestViewLogTable(sid)+" WHERE sid=? AND vodid=? LIMIT 1", sid, vodID)
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
		err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+miniUserViewLogTable(uid)+" WHERE uid=? AND showtype=1 AND "+field+">=?", uid, since).Scan(&total)
		if err != nil {
			return 0, fmt.Errorf("count minivod viewlogs: %w", err)
		}
		return total, nil
	}
	if sid == "" {
		return 0, nil
	}
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+miniGuestViewLogTable(sid)+" WHERE sid=? AND showtype=1 AND "+field+">=?", sid, since).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count minivod guest viewlogs: %w", err)
	}
	return total, nil
}

func (r *Repository) RecordMiniMedia(ctx context.Context, uid int, sid string, vodID int, play bool, deduct int, now int64) error {
	if r.db == nil || vodID <= 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin minivod media log: %w", err)
	}
	defer tx.Rollback()
	prefix := "playcount"
	if !play {
		prefix = "downcount"
	}
	if err := incrementMiniVODMediaCounter(ctx, tx, vodID, prefix, now); err != nil {
		return fmt.Errorf("increment minivod media counter: %w", err)
	}
	if uid > 0 {
		if err := recordMiniViewLog(ctx, tx, miniUserViewLogTable(uid), "uid", uid, vodID, play, deduct, now); err != nil {
			return fmt.Errorf("record minivod user media: %w", err)
		}
	} else if strings.TrimSpace(sid) != "" {
		if err := recordMiniViewLog(ctx, tx, miniGuestViewLogTable(sid), "sid", sid, vodID, play, deduct, now); err != nil {
			return fmt.Errorf("record minivod guest media: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit minivod media log: %w", err)
	}
	return nil
}

func incrementMiniVODMediaCounter(ctx context.Context, tx *sql.Tx, vodID int, prefix string, now int64) error {
	if prefix != "playcount" && prefix != "downcount" {
		return fmt.Errorf("invalid minivod media counter %s", prefix)
	}
	lastColumn := prefix + "_lasttime"
	var previous int64
	if err := tx.QueryRowContext(ctx, "SELECT "+lastColumn+" FROM vods WHERE vodid=?", vodID).Scan(&previous); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	monthValue := prefix + "_month+1"
	if !sameMonth(previous, now) {
		monthValue = "1"
	}
	dayValue := prefix + "_day+1"
	if !sameDay(previous, now) {
		dayValue = "1"
	}
	weekValue := prefix + "_week+1"
	if !sameWeek(previous, now) {
		weekValue = "1"
	}
	_, err := tx.ExecContext(ctx,
		"UPDATE vods SET "+lastColumn+"=?, "+prefix+"_total="+prefix+"_total+1, "+prefix+"_month="+monthValue+", "+prefix+"_day="+dayValue+", "+prefix+"_week="+weekValue+" WHERE vodid=?",
		now,
		vodID,
	)
	return err
}

func recordMiniViewLog[T int | string](ctx context.Context, tx *sql.Tx, table string, actorColumn string, actorID T, vodID int, play bool, deduct int, now int64) error {
	row, err := queryOneTx(ctx, tx, "SELECT * FROM "+table+" WHERE "+actorColumn+"=? AND vodid=? LIMIT 1 FOR UPDATE", actorID, vodID)
	if err != nil {
		return err
	}
	if play {
		if len(row) == 0 {
			_, err = tx.ExecContext(ctx, "INSERT INTO "+table+"("+actorColumn+", vodid, playtime, deduct, reqtime, showtype) VALUES(?, ?, ?, ?, ?, 1)", actorID, vodID, now, deduct, now)
			return err
		}
		if atoi(row["playtime"]) == 0 {
			_, err = tx.ExecContext(ctx, "UPDATE "+table+" SET playtime=?, deduct=? WHERE logid=?", now, deduct, row["logid"])
			return err
		}
		return nil
	}
	if len(row) == 0 {
		_, err = tx.ExecContext(ctx, "INSERT INTO "+table+"("+actorColumn+", vodid, downtime, downdeduct) VALUES(?, ?, ?, ?)", actorID, vodID, now, deduct)
		return err
	}
	if atoi(row["downtime"]) == 0 {
		_, err = tx.ExecContext(ctx, "UPDATE "+table+" SET downtime=?, downdeduct=? WHERE logid=?", now, deduct, row["logid"])
		return err
	}
	return nil
}

func (r *Repository) ReqTaskCoin(ctx context.Context, uid int, sid string, logid int, now int64) (int, string, error) {
	if r.db == nil || logid <= 0 {
		return -1, "记录不存在或已被删除", nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return -1, "", fmt.Errorf("begin minivod reqcoin: %w", err)
	}
	defer tx.Rollback()
	if uid > 0 {
		return r.reqUserTaskCoin(ctx, tx, uid, logid, now)
	}
	return r.reqGuestTaskCoin(ctx, tx, sid, logid, now)
}

func (r *Repository) reqUserTaskCoin(ctx context.Context, tx *sql.Tx, uid int, logid int, now int64) (int, string, error) {
	logrow, err := queryOneTx(ctx, tx, "SELECT * FROM minivod_tasklogs WHERE logid=? FOR UPDATE", logid)
	if err != nil {
		return -1, "", fmt.Errorf("lock minivod task log: %w", err)
	}
	if len(logrow) == 0 {
		return -1, "记录不存在或已被删除", nil
	}
	if atoi(logrow["reqtime"]) > 0 {
		return -1, "您已经领取过金币了", nil
	}
	taskcoin := atoi(logrow["taskcoin"])
	var balance int
	if err := tx.QueryRowContext(ctx, "SELECT goldcoin FROM users_quota WHERE uid=? FOR UPDATE", uid).Scan(&balance); err != nil {
		if err == sql.ErrNoRows {
			return -1, "记录写入失败，请重试", nil
		}
		return -1, "", fmt.Errorf("lock users quota: %w", err)
	}
	newBalance := balance + taskcoin
	if newBalance != balance {
		result, err := tx.ExecContext(ctx, "UPDATE users_quota SET goldcoin=? WHERE uid=?", newBalance, uid)
		if err != nil {
			return -1, "", fmt.Errorf("update users quota: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			return -1, "记录写入失败，请重试", nil
		}
	}
	result, err := tx.ExecContext(ctx, "INSERT INTO user_coinlogs(cointype, uid, coinnum, balance, addtime, remark) VALUES(?, ?, ?, ?, ?, '')", coinTypeMiniVODPlayTask, uid, taskcoin, newBalance, now)
	if err != nil {
		return -1, "", fmt.Errorf("insert user coinlog: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected < 1 {
		return -1, "记录写入失败，请重试", nil
	}
	if err := updateReqTime(ctx, tx, "minivod_tasklogs", logid, now); err != nil {
		return -1, "记录更新失败，请重试", err
	}
	if err := tx.Commit(); err != nil {
		return -1, "", fmt.Errorf("commit minivod reqcoin: %w", err)
	}
	return 0, "领取成功", nil
}

func (r *Repository) reqGuestTaskCoin(ctx context.Context, tx *sql.Tx, sid string, logid int, now int64) (int, string, error) {
	logrow, err := queryOneTx(ctx, tx, "SELECT * FROM minivod_guesttasklogs WHERE logid=? FOR UPDATE", logid)
	if err != nil {
		return -1, "", fmt.Errorf("lock minivod guest task log: %w", err)
	}
	if len(logrow) == 0 {
		return -1, "记录不存在或已被删除", nil
	}
	if atoi(logrow["reqtime"]) > 0 {
		return -1, "您已经领取过金币了", nil
	}
	if _, err := tx.ExecContext(ctx, "UPDATE user_guests SET goldcoin=goldcoin+? WHERE sid=?", atoi(logrow["taskcoin"]), sid); err != nil {
		return -1, "", fmt.Errorf("update guest goldcoin: %w", err)
	}
	if err := updateReqTime(ctx, tx, "minivod_guesttasklogs", logid, now); err != nil {
		return -1, "记录更新失败，请重试", err
	}
	if err := tx.Commit(); err != nil {
		return -1, "", fmt.Errorf("commit minivod guest reqcoin: %w", err)
	}
	return 0, "领取成功", nil
}

func updateReqTime(ctx context.Context, tx *sql.Tx, table string, logid int, now int64) error {
	result, err := tx.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET reqtime=? WHERE logid=?", table), now, logid)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("update reqtime affected %d rows", affected)
	}
	return nil
}

func (r *Repository) pullUserViewLogs(ctx context.Context, uid int) (int, error) {
	row, err := r.ensureSublog(ctx, "minivod_sublogs", "uid", uid)
	if err != nil {
		return 0, err
	}
	return r.pullViewLogsForActor(ctx, miniUserViewLogTable(uid), "uid", uid, row)
}

func (r *Repository) pullGuestViewLogs(ctx context.Context, sid string) (int, error) {
	row, err := r.ensureSublog(ctx, "minivod_guestsublogs", "sid", sid)
	if err != nil {
		return 0, err
	}
	return r.pullViewLogsForActor(ctx, miniGuestViewLogTable(sid), "sid", sid, row)
}

func (r *Repository) ensureSublog(ctx context.Context, table string, actorColumn string, actor interface{}) (map[string]interface{}, error) {
	rows, err := r.queryRows(ctx, "SELECT * FROM "+table+" WHERE "+actorColumn+"=? LIMIT 1", actor)
	if err != nil {
		return nil, fmt.Errorf("query minivod sublog: %w", err)
	}
	if len(rows) > 0 {
		return rows[0], nil
	}
	result, err := r.db.ExecContext(ctx, "INSERT IGNORE INTO "+table+"("+actorColumn+") VALUES(?)", actor)
	if err != nil {
		return nil, fmt.Errorf("create minivod sublog: %w", err)
	}
	logID, _ := result.LastInsertId()
	return map[string]interface{}{"logid": fmt.Sprint(logID), actorColumn: fmt.Sprint(actor)}, nil
}

func (r *Repository) pullViewLogsForActor(ctx context.Context, viewTable string, actorColumn string, actor interface{}, sublog map[string]interface{}) (int, error) {
	oldRows, err := r.queryRows(ctx, "SELECT logid,vodid,playtime,downtime FROM "+viewTable+" WHERE "+actorColumn+"=? ORDER BY logid DESC LIMIT 2000", actor)
	if err != nil {
		return 0, fmt.Errorf("query existing minivod viewlogs: %w", err)
	}
	oldIDs := []int{}
	oldSet := map[int]struct{}{}
	minLogID := 0
	for _, row := range oldRows {
		vodID := atoi(row["vodid"])
		if vodID > 0 {
			oldIDs = append(oldIDs, vodID)
			oldSet[vodID] = struct{}{}
		}
		logID := atoi(row["logid"])
		if logID > 0 && (minLogID == 0 || logID < minLogID) {
			minLogID = logID
		}
	}
	if len(oldRows) >= 2000 && minLogID > 0 {
		daytime := dayStartUnix(time.Now)
		_, err := r.db.ExecContext(ctx, "DELETE FROM "+viewTable+" WHERE "+actorColumn+"=? AND logid<? AND playtime<? AND downtime<?", actor, minLogID, daytime, daytime)
		if err != nil {
			return 0, fmt.Errorf("prune minivod viewlogs: %w", err)
		}
	}

	newIDs := []int{}
	newSet := map[int]struct{}{}
	addNew := func(ids []int) {
		for _, id := range ids {
			if id <= 0 {
				continue
			}
			if _, ok := oldSet[id]; ok {
				continue
			}
			if _, ok := newSet[id]; ok {
				continue
			}
			newSet[id] = struct{}{}
			newIDs = append(newIDs, id)
		}
	}

	tagIDs := append(csvIDs(str(sublog["tagid_news"])), hotIDs(str(sublog["tagid_hots"]))...)
	tagVODIDs, err := r.vodIDsByTags(ctx, tagIDs)
	if err != nil {
		return 0, err
	}
	addNew(tagVODIDs)

	authorIDs := append(csvIDs(str(sublog["authorid_news"])), hotIDs(str(sublog["authorid_hots"]))...)
	authorVODIDs, err := r.vodIDsByAuthors(ctx, authorIDs)
	if err != nil {
		return 0, err
	}
	addNew(authorVODIDs)

	if len(newIDs) < 500 {
		latestIDs, err := r.latestMiniVODIDsAfter(ctx, maxInt(oldIDs))
		if err != nil {
			return 0, err
		}
		addNew(latestIDs)
	}
	if len(newIDs) < 500 {
		fallbackIDs, err := r.miniVODIDsExcept(ctx, oldIDs)
		if err != nil {
			return 0, err
		}
		addNew(fallbackIDs)
	}
	return r.insertViewLogs(ctx, viewTable, actorColumn, actor, newIDs)
}

func (r *Repository) vodIDsByTags(ctx context.Context, tagIDs []int) ([]int, error) {
	if len(tagIDs) == 0 {
		return []int{}, nil
	}
	tagRows, err := r.queryRows(ctx, "SELECT vodid FROM vod_tagmaps WHERE tagid IN("+intListSQL(tagIDs)+") ORDER BY vodid DESC LIMIT 500")
	if err != nil {
		return nil, fmt.Errorf("query minivod tag maps: %w", err)
	}
	candidateIDs := rowIDs(tagRows, "vodid")
	if len(candidateIDs) == 0 {
		return []int{}, nil
	}
	vodRows, err := r.queryRows(ctx, "SELECT vodid FROM vods WHERE vodid IN("+intListSQL(candidateIDs)+") AND showtype=1 AND isvip=0")
	if err != nil {
		return nil, fmt.Errorf("query minivod tag vods: %w", err)
	}
	return rowIDs(vodRows, "vodid"), nil
}

func (r *Repository) vodIDsByAuthors(ctx context.Context, authorIDs []int) ([]int, error) {
	if len(authorIDs) == 0 {
		return []int{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT vodid FROM vods WHERE authorid IN("+intListSQL(authorIDs)+") AND showtype=1 AND isvip=0 ORDER BY vodid DESC LIMIT 500")
	if err != nil {
		return nil, fmt.Errorf("query minivod author vods: %w", err)
	}
	return rowIDs(rows, "vodid"), nil
}

func (r *Repository) latestMiniVODIDsAfter(ctx context.Context, maxVODID int) ([]int, error) {
	rows, err := r.queryRows(ctx, "SELECT vodid FROM vods WHERE vodid>? AND showtype=1 AND isvip=0 ORDER BY vodid ASC LIMIT 500", maxVODID)
	if err != nil {
		return nil, fmt.Errorf("query latest minivods: %w", err)
	}
	return rowIDs(rows, "vodid"), nil
}

func (r *Repository) miniVODIDsExcept(ctx context.Context, oldIDs []int) ([]int, error) {
	where := "showtype=1 AND isvip=0"
	if len(oldIDs) > 0 {
		where += " AND vodid NOT IN(" + intListSQL(oldIDs) + ")"
	}
	rows, err := r.queryRows(ctx, "SELECT vodid FROM vods WHERE "+where+" LIMIT 500")
	if err != nil {
		return nil, fmt.Errorf("query fallback minivods: %w", err)
	}
	return rowIDs(rows, "vodid"), nil
}

func (r *Repository) insertViewLogs(ctx context.Context, table string, actorColumn string, actor interface{}, vodIDs []int) (int, error) {
	if len(vodIDs) == 0 {
		return 0, nil
	}
	placeholders := make([]string, 0, len(vodIDs))
	args := make([]interface{}, 0, len(vodIDs)*2)
	for _, vodID := range vodIDs {
		if vodID <= 0 {
			continue
		}
		placeholders = append(placeholders, "(?,?)")
		args = append(args, actor, vodID)
	}
	if len(placeholders) == 0 {
		return 0, nil
	}
	result, err := r.db.ExecContext(ctx, "INSERT IGNORE INTO "+table+"("+actorColumn+",vodid) VALUES "+strings.Join(placeholders, ","), args...)
	if err != nil {
		return 0, fmt.Errorf("insert minivod viewlogs: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("insert minivod viewlogs affected: %w", err)
	}
	return int(affected), nil
}

func queryOneTx(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (map[string]interface{}, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
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

func miniUserViewLogTable(uid int) string {
	return fmt.Sprintf("minivod_viewlogs_%02d", uid%100)
}

func miniGuestViewLogTable(sid string) string {
	suffix := "0"
	if sid != "" {
		suffix = sid[:1]
	}
	return "minivod_guestviewlogs_" + suffix
}

func rowIDs(rows []map[string]interface{}, key string) []int {
	ids := []int{}
	seen := map[int]struct{}{}
	for _, row := range rows {
		id := atoi(row[key])
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func csvIDs(value string) []int {
	rows := []map[string]interface{}{}
	for _, part := range strings.Split(value, ",") {
		rows = append(rows, map[string]interface{}{"id": strings.TrimSpace(part)})
	}
	return rowIDs(rows, "id")
}

func hotIDs(value string) []int {
	value = strings.TrimSpace(value)
	if value == "" {
		return []int{}
	}
	var rows []map[string]interface{}
	if err := json.Unmarshal([]byte(value), &rows); err != nil {
		return []int{}
	}
	return rowIDs(rows, "id")
}

func maxInt(values []int) int {
	max := 0
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}

func dayStartUnix(now func() time.Time) int64 {
	t := now()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
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
