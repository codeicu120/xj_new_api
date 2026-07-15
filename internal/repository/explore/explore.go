package explore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	coinTypeExploreVODTask  = 23
	coinTypeExploreSignTask = 24
	vipGID                  = 6
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Tabs(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM explore_tabs WHERE 1=1 AND showtype=0 ORDER BY sortnum ASC LIMIT ? OFFSET 0", 100)
}

func (r *Repository) UpdateUserNotificationAll(ctx context.Context, uid int, value string) error {
	if r.db == nil || uid <= 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE users SET notification_all=? WHERE uid=?", value, uid); err != nil {
		return fmt.Errorf("update user notification_all: %w", err)
	}
	return nil
}

func (r *Repository) UpdateGuestNotificationAll(ctx context.Context, sid string, value string) error {
	if r.db == nil || sid == "" {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE user_guests SET notification_all=? WHERE sid=?", value, sid); err != nil {
		return fmt.Errorf("update guest notification_all: %w", err)
	}
	return nil
}

func (r *Repository) SignTask(ctx context.Context, user map[string]interface{}, now int64) (int, int, string, error) {
	if r.db == nil {
		return 0, -1, "记录写入失败，请重试3", nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, -1, "", fmt.Errorf("begin sign task: %w", err)
	}
	defer tx.Rollback()

	uid := atoi(user["uid"])
	var row map[string]interface{}
	if uid > 0 {
		row, err = queryOneTx(ctx, tx, "SELECT * FROM users WHERE uid=? FOR UPDATE", uid)
		if err != nil {
			return 0, -1, "", fmt.Errorf("lock sign user: %w", err)
		}
	} else {
		row, err = queryOneTx(ctx, tx, "SELECT * FROM user_guests WHERE sid=? FOR UPDATE", str(user["sid"]))
		if err != nil {
			return 0, -1, "", fmt.Errorf("lock sign guest: %w", err)
		}
	}
	if len(row) == 0 {
		return 0, -1, "记录写入失败，请重试3", nil
	}
	for key, value := range user {
		if _, ok := row[key]; !ok {
			row[key] = value
		}
	}

	today := dayStartUnix(time.Unix(now, 0))
	yesterday := today - 86400
	if atoi64(row["signed_lasttime"]) >= today {
		return 0, -1, "您今天已经签过到了", nil
	}

	signedContDays := 1
	signedUnitDays := 1
	if atoi64(row["signed_lasttime"]) >= yesterday {
		signedContDays = atoi(row["signed_contdays"]) + 1
		signedUnitDays = atoi(row["signed_unitdays"]) + 1
		if signedUnitDays > 7 {
			signedUnitDays = 1
		}
	}
	signedPeakDays := atoi(row["signed_peakdays"])
	if signedContDays > signedPeakDays {
		signedPeakDays = signedContDays
	}
	coinNum := getPermInt(row["perms"], fmt.Sprintf("max.signtask.coinnum%d", signedUnitDays))

	if uid > 0 {
		if coinNum > 1000 {
			dayLen := coinNum - 1000
			sysgidExpTime := now + int64(dayLen)*86400
			if atoi(row["sysgid"]) == vipGID && atoi64(row["sysgid_exptime"]) > now {
				sysgidExpTime = atoi64(row["sysgid_exptime"]) + int64(dayLen)*86400
			}
			result, err := tx.ExecContext(ctx, "UPDATE users SET sysgid=?, sysgid_exptime=? WHERE uid=?", vipGID, sysgidExpTime, uid)
			if err != nil {
				return 0, -1, "", fmt.Errorf("update sign vip: %w", err)
			}
			if affected, _ := result.RowsAffected(); affected == 0 {
				return 0, -1, "VIP赠送失败，请重试", nil
			}
		} else if coinNum > 0 {
			var balance int
			if err := tx.QueryRowContext(ctx, "SELECT goldcoin FROM users_quota WHERE uid=? FOR UPDATE", uid).Scan(&balance); err != nil {
				if err == sql.ErrNoRows {
					return 0, -1, "记录写入失败，请重试1", nil
				}
				return 0, -1, "", fmt.Errorf("lock sign quota: %w", err)
			}
			newBalance := balance + coinNum
			result, err := tx.ExecContext(ctx, "UPDATE users_quota SET goldcoin=? WHERE uid=?", newBalance, uid)
			if err != nil {
				return 0, -1, "", fmt.Errorf("update sign quota: %w", err)
			}
			if affected, _ := result.RowsAffected(); affected == 0 {
				return 0, -1, "记录写入失败，请重试1", nil
			}
			result, err = tx.ExecContext(ctx, "INSERT INTO user_coinlogs(cointype, uid, coinnum, balance, addtime, remark) VALUES(?, ?, ?, ?, ?, '')", coinTypeExploreSignTask, uid, coinNum, newBalance, now)
			if err != nil {
				return 0, -1, "", fmt.Errorf("insert sign coinlog: %w", err)
			}
			if affected, _ := result.RowsAffected(); affected == 0 {
				return 0, -1, "记录写入失败，请重试1", nil
			}
		}
		result, err := tx.ExecContext(ctx, "INSERT INTO explore_signlogs(uid, signtime) VALUES(?, ?)", uid, now)
		if err != nil {
			return 0, -1, "", fmt.Errorf("insert sign log: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return 0, -1, "记录写入失败，请重试2", nil
		}
		result, err = tx.ExecContext(ctx, "UPDATE users SET signed_peakdays=?, signed_contdays=?, signed_unitdays=?, signed_lasttime=? WHERE uid=?", signedPeakDays, signedContDays, signedUnitDays, now, uid)
		if err != nil {
			return 0, -1, "", fmt.Errorf("update sign user fields: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return 0, -1, "记录写入失败，请重试3", nil
		}
	} else {
		sid := str(row["sid"])
		if coinNum > 0 && coinNum <= 1000 {
			if _, err := tx.ExecContext(ctx, "UPDATE user_guests SET goldcoin=goldcoin+? WHERE sid=?", coinNum, sid); err != nil {
				return 0, -1, "", fmt.Errorf("update sign guest goldcoin: %w", err)
			}
		}
		result, err := tx.ExecContext(ctx, "INSERT INTO explore_guestsignlogs(sid, signtime) VALUES(?, ?)", sid, now)
		if err != nil {
			return 0, -1, "", fmt.Errorf("insert guest sign log: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return 0, -1, "记录写入失败，请重试2", nil
		}
		result, err = tx.ExecContext(ctx, "UPDATE user_guests SET signed_peakdays=?, signed_contdays=?, signed_unitdays=?, signed_lasttime=? WHERE sid=?", signedPeakDays, signedContDays, signedUnitDays, now, sid)
		if err != nil {
			return 0, -1, "", fmt.Errorf("update sign guest fields: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return 0, -1, "记录写入失败，请重试3", nil
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, -1, "", fmt.Errorf("commit sign task: %w", err)
	}
	return coinNum, 0, "签到成功", nil
}

func (r *Repository) VodTaskByID(ctx context.Context, vid int) (map[string]interface{}, error) {
	if r.db == nil || vid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM explore_vods WHERE vid=?", vid)
	if err != nil {
		return nil, fmt.Errorf("query explore vod: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) UserVodTaskLog(ctx context.Context, uid int, today int64, vid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 || vid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM explore_vodlogs WHERE uid=? AND addtime>? AND vid=? LIMIT 1", uid, today, vid)
	if err != nil {
		return nil, fmt.Errorf("query explore vod log: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) GuestVodTaskLog(ctx context.Context, sid string, today int64, vid int) (map[string]interface{}, error) {
	if r.db == nil || sid == "" || vid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM explore_guestvodlogs WHERE sid=? AND addtime>=? AND vid=? LIMIT 1", sid, today, vid)
	if err != nil {
		return nil, fmt.Errorf("query explore guest vod log: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) CreateUserVodTaskLog(ctx context.Context, uid int, vid int, addtime int64, reqcoin int) (int, error) {
	if r.db == nil || uid <= 0 || vid <= 0 {
		return 0, nil
	}
	result, err := r.db.ExecContext(ctx, "INSERT INTO explore_vodlogs(uid, vid, addtime, reqcoin) VALUES(?, ?, ?, ?)", uid, vid, addtime, reqcoin)
	if err != nil {
		return 0, fmt.Errorf("insert explore vod log: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert explore vod log id: %w", err)
	}
	return int(id), nil
}

func (r *Repository) CreateGuestVodTaskLog(ctx context.Context, sid string, vid int, addtime int64, reqcoin int) (int, error) {
	if r.db == nil || sid == "" || vid <= 0 {
		return 0, nil
	}
	result, err := r.db.ExecContext(ctx, "INSERT INTO explore_guestvodlogs(sid, vid, addtime, reqcoin) VALUES(?, ?, ?, ?)", sid, vid, addtime, reqcoin)
	if err != nil {
		return 0, fmt.Errorf("insert explore guest vod log: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert explore guest vod log id: %w", err)
	}
	return int(id), nil
}

func (r *Repository) ReqVodTaskCoin(ctx context.Context, uid int, sid string, logid int, now int64) (int, string, error) {
	if r.db == nil || logid <= 0 {
		return -1, "记录不存在或已被删除", nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return -1, "", fmt.Errorf("begin req vodtask coin: %w", err)
	}
	defer tx.Rollback()
	if uid > 0 {
		return r.reqUserVodTaskCoin(ctx, tx, uid, logid, now)
	}
	return r.reqGuestVodTaskCoin(ctx, tx, sid, logid, now)
}

func (r *Repository) reqUserVodTaskCoin(ctx context.Context, tx *sql.Tx, uid int, logid int, now int64) (int, string, error) {
	logrow, err := queryOneTx(ctx, tx, "SELECT * FROM explore_vodlogs WHERE logid=? FOR UPDATE", logid)
	if err != nil {
		return -1, "", fmt.Errorf("lock explore vod log: %w", err)
	}
	if len(logrow) == 0 || atoi(logrow["uid"]) != uid {
		return -1, "记录不存在或已被删除", nil
	}
	if atoi(logrow["reqtime"]) > 0 {
		return -1, "您已经领取过金币了", nil
	}
	reqcoin := atoi(logrow["reqcoin"])
	var balance int
	if err := tx.QueryRowContext(ctx, "SELECT goldcoin FROM users_quota WHERE uid=? FOR UPDATE", uid).Scan(&balance); err != nil {
		if err == sql.ErrNoRows {
			return -1, "记录写入失败，请重试", nil
		}
		return -1, "", fmt.Errorf("lock user quota: %w", err)
	}
	newBalance := balance + reqcoin
	if newBalance != balance {
		result, err := tx.ExecContext(ctx, "UPDATE users_quota SET goldcoin=? WHERE uid=?", newBalance, uid)
		if err != nil {
			return -1, "", fmt.Errorf("update user quota: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			return -1, "记录写入失败，请重试", nil
		}
	}
	result, err := tx.ExecContext(ctx, "INSERT INTO user_coinlogs(cointype, uid, coinnum, balance, addtime, remark) VALUES(?, ?, ?, ?, ?, '')", coinTypeExploreVODTask, uid, reqcoin, newBalance, now)
	if err != nil {
		return -1, "", fmt.Errorf("insert user coinlog: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected < 1 {
		return -1, "记录写入失败，请重试", nil
	}
	if err := updateReqTime(ctx, tx, "explore_vodlogs", logid, now); err != nil {
		return -1, "记录更新失败，请重试", err
	}
	if err := tx.Commit(); err != nil {
		return -1, "", fmt.Errorf("commit req vodtask coin: %w", err)
	}
	return 0, "领取成功", nil
}

func (r *Repository) reqGuestVodTaskCoin(ctx context.Context, tx *sql.Tx, sid string, logid int, now int64) (int, string, error) {
	logrow, err := queryOneTx(ctx, tx, "SELECT * FROM explore_guestvodlogs WHERE logid=? FOR UPDATE", logid)
	if err != nil {
		return -1, "", fmt.Errorf("lock explore guest vod log: %w", err)
	}
	if len(logrow) == 0 || str(logrow["sid"]) != sid {
		return -1, "记录不存在或已被删除", nil
	}
	if atoi(logrow["reqtime"]) > 0 {
		return -1, "您已经领取过金币了", nil
	}
	if _, err := tx.ExecContext(ctx, "UPDATE user_guests SET goldcoin=goldcoin+? WHERE sid=?", atoi(logrow["reqcoin"]), sid); err != nil {
		return -1, "", fmt.Errorf("update guest goldcoin: %w", err)
	}
	if err := updateReqTime(ctx, tx, "explore_guestvodlogs", logid, now); err != nil {
		return -1, "记录更新失败，请重试", err
	}
	if err := tx.Commit(); err != nil {
		return -1, "", fmt.Errorf("commit req guest vodtask coin: %w", err)
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

func DecodeJSON(value interface{}) interface{} {
	raw := strings.TrimSpace(fmt.Sprint(value))
	if raw == "" {
		return nil
	}
	var decoded interface{}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return nil
	}
	return decoded
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}

func atoi64(value interface{}) int64 {
	var n int64
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}

func dayStartUnix(t time.Time) int64 {
	loc := t.Location()
	y, m, d := t.In(loc).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc).Unix()
}

func getPermInt(perms interface{}, key string) int {
	values := map[string]interface{}{}
	switch v := perms.(type) {
	case string:
		_ = json.Unmarshal([]byte(v), &values)
	case []byte:
		_ = json.Unmarshal(v, &values)
	case map[string]interface{}:
		values = v
	}
	return atoi(values[key])
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
