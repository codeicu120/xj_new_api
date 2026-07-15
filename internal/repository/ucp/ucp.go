package ucp

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"xj_comp/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) RollTitles(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM roll_titles WHERE 1=1 AND status=1 ORDER BY id DESC LIMIT 10")
	if err != nil {
		return nil, fmt.Errorf("query roll titles: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) Posters(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM poster WHERE status=1")
	if err != nil {
		return nil, fmt.Errorf("query posters: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) Taskboxes(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM promotion_taskboxs ORDER BY taskid ASC")
	if err != nil {
		return nil, fmt.Errorf("query taskboxes: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) TaskboxLog(ctx context.Context, uid int, taskID int, dayKey int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	return r.queryOne(ctx, "SELECT * FROM promotion_taskboxlogs WHERE uid=? AND taskid=? AND daykey=?", uid, taskID, dayKey)
}

func (r *Repository) TaskboxCompletedLogs(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT a.*, b.username, b.nickname, b.avatar
FROM promotion_taskboxlogs a
LEFT JOIN users b ON b.uid=a.uid
WHERE 1=1 AND taskstatus=2
ORDER BY a.logid DESC
LIMIT ? OFFSET 0`, limit)
	if err != nil {
		return nil, fmt.Errorf("query taskbox logs: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) CountTaskboxLogs(ctx context.Context, uid int) (int, error) {
	if r.db == nil || uid <= 0 {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM promotion_taskboxlogs WHERE 1=1 AND uid=?", uid).Scan(&total); err != nil {
		return 0, fmt.Errorf("count taskbox logs: %w", err)
	}
	return total, nil
}

func (r *Repository) TaskboxLogs(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return []map[string]interface{}{}, nil
	}
	offset := limitOffset(page, pageSize)
	rows, err := r.db.QueryContext(ctx, `
SELECT a.*, b.username, b.nickname, b.avatar
FROM promotion_taskboxlogs a
LEFT JOIN users b ON b.uid=a.uid
WHERE 1=1 AND a.uid=?
ORDER BY a.addtime DESC
LIMIT ? OFFSET ?`, uid, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query user taskbox logs: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) Bankcards(ctx context.Context, uid int) ([]map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM user_bankcards WHERE uid=? ORDER BY isdef DESC", uid)
	if err != nil {
		return nil, fmt.Errorf("query bankcards: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) Banks(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM banks WHERE showtype=0 ORDER BY sortnum ASC, bankid ASC")
	if err != nil {
		return nil, fmt.Errorf("query banks: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) CountPayments(ctx context.Context, uid int) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM trade_payments WHERE 1=1 AND uid=?", uid).Scan(&total); err != nil {
		return 0, fmt.Errorf("count payments: %w", err)
	}
	return total, nil
}

func (r *Repository) Payments(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM trade_payments WHERE 1=1 AND uid=? ORDER BY createtime DESC LIMIT ? OFFSET ?", uid, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query payments: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) SafePayLogs(ctx context.Context, uid int, since int64, limit int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM trade_payments WHERE 1=1 AND uid=? AND createtime>? AND payway='safepay' ORDER BY payid DESC LIMIT ?", uid, since, limit)
	if err != nil {
		return nil, fmt.Errorf("query safepay logs: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) PaymentsSince(ctx context.Context, uid int, since int64, limit int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM trade_payments WHERE 1=1 AND uid=? AND createtime>? ORDER BY payid DESC LIMIT ?", uid, since, limit)
	if err != nil {
		return nil, fmt.Errorf("query payments since: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) Account(ctx context.Context, uid int) (map[string]interface{}, error) {
	row, err := r.queryOne(ctx, "SELECT uid, balance, frozen, deposit, game_balance, game_frozen FROM users_account WHERE uid=?", uid)
	if err != nil {
		return nil, fmt.Errorf("query account: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	row["available_balance"] = atoi(row["balance"]) - atoi(row["frozen"])
	row["game_available_balance"] = atoi(row["game_balance"]) - atoi(row["game_frozen"])
	return row, nil
}

func (r *Repository) Quota(ctx context.Context, uid int) (map[string]interface{}, error) {
	row, err := r.queryOne(ctx, "SELECT uid, goldcoin, withdraw_freequota FROM users_quota WHERE uid=?", uid)
	if err != nil {
		return nil, fmt.Errorf("query quota: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) Goldbean(ctx context.Context, uid int) (map[string]interface{}, error) {
	row, err := r.queryOne(ctx, "SELECT uid, gold_bean FROM users_goldbean WHERE uid=?", uid)
	if err != nil {
		return nil, fmt.Errorf("query goldbean: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) CountVODPlayLogsSince(ctx context.Context, uid int, since int64) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vod_playlogs_week WHERE uid=? AND playtime>=?", uid, since).Scan(&total); err != nil {
		return 0, fmt.Errorf("count vod play logs: %w", err)
	}
	return total, nil
}

func (r *Repository) CountVODDownLogsSince(ctx context.Context, uid int, since int64) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vod_downlogs WHERE uid=? AND downtime>=?", uid, since).Scan(&total); err != nil {
		return 0, fmt.Errorf("count vod down logs: %w", err)
	}
	return total, nil
}

func (r *Repository) GuestBySID(ctx context.Context, sid string) (map[string]interface{}, error) {
	row, err := r.queryOne(ctx, "SELECT * FROM user_guests WHERE sid=?", sid)
	if err != nil {
		return nil, fmt.Errorf("query guest: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) CountGuestVODPlayLogsSince(ctx context.Context, sid string, since int64) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vod_guest_playlogs WHERE sid=? AND playtime>=?", sid, since).Scan(&total); err != nil {
		return 0, fmt.Errorf("count guest vod play logs: %w", err)
	}
	return total, nil
}

func (r *Repository) CountGuestVODDownLogsSince(ctx context.Context, sid string, since int64) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vod_guest_downlogs WHERE sid=? AND downtime>=?", sid, since).Scan(&total); err != nil {
		return 0, fmt.Errorf("count guest vod down logs: %w", err)
	}
	return total, nil
}

func (r *Repository) CountMiniVODViewLogsSince(ctx context.Context, uid int, since int64, action int) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	table := fmt.Sprintf("minivod_viewlogs_%02d", uid%100)
	column := miniVODTimeColumn(action)
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+" WHERE uid=? AND "+column+">=?", uid, since).Scan(&total)
	if err != nil {
		if isMissingTable(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("count minivod logs: %w", err)
	}
	return total, nil
}

func (r *Repository) CountGuestMiniVODViewLogsSince(ctx context.Context, sid string, since int64, action int) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	table := "minivod_guestviewlogs_" + splitTableBySID(sid)
	column := miniVODTimeColumn(action)
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+" WHERE sid=? AND "+column+">=?", sid, since).Scan(&total)
	if err != nil {
		if isMissingTable(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("count guest minivod logs: %w", err)
	}
	return total, nil
}

func (r *Repository) CountCoinLogsSinceByType(ctx context.Context, uid int, coinType int, since int64) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_coinlogs WHERE uid=? AND cointype=? AND addtime>=?", uid, coinType, since).Scan(&total); err != nil {
		return 0, fmt.Errorf("count coin logs by type: %w", err)
	}
	return total, nil
}

func (r *Repository) CountFeedbacks(ctx context.Context, uid int) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM feedbacks WHERE 1=1 AND uid=?", uid).Scan(&total); err != nil {
		return 0, fmt.Errorf("count feedbacks: %w", err)
	}
	return total, nil
}

func (r *Repository) Feedbacks(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM feedbacks WHERE 1=1 AND uid=? ORDER BY id DESC LIMIT ? OFFSET ?", uid, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query feedbacks: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) CountFeedbacksByType(ctx context.Context, uid int, feedbackType int) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	where, args := feedbackTypeWhere(uid, feedbackType)
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM feedbacks WHERE "+where, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count feedbacks by type: %w", err)
	}
	return total, nil
}

func (r *Repository) FeedbacksByType(ctx context.Context, uid int, feedbackType int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	where, args := feedbackTypeWhere(uid, feedbackType)
	args = append(args, pageSize, offset)
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM feedbacks WHERE "+where+" ORDER BY id DESC LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, fmt.Errorf("query feedbacks by type: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) FeedbackByID(ctx context.Context, id int) (map[string]interface{}, error) {
	row, err := r.queryOne(ctx, "SELECT * FROM feedbacks WHERE id=?", id)
	if err != nil {
		return nil, fmt.Errorf("query feedback by id: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) CountFeedbacksSince(ctx context.Context, uid int, since int64) (int, error) {
	if r.db == nil || uid <= 0 {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM feedbacks WHERE uid=? AND ctimestamp>?", uid, since).Scan(&total); err != nil {
		return 0, fmt.Errorf("count feedbacks since: %w", err)
	}
	return total, nil
}

func (r *Repository) CreateFeedback(ctx context.Context, input domain.FeedbackCreateInput) (int, error) {
	if r.db == nil || input.UID <= 0 {
		return 0, nil
	}
	result, err := r.db.ExecContext(ctx, `INSERT INTO feedbacks
		(uid, cid, content, payid, payname, payaccount, aids, ctimestamp, ip, device, longids, shortids)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.UID, input.CID, input.Content, input.PayID, input.PayName, input.PayAccount, input.AIDs, input.CreatedAt, input.IP, input.Device, input.LongIDs, input.ShortIDs,
	)
	if err != nil {
		return 0, fmt.Errorf("insert feedback: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert feedback id: %w", err)
	}
	return int(id), nil
}

func (r *Repository) PaymentByID(ctx context.Context, payid int) (map[string]interface{}, error) {
	row, err := r.queryOne(ctx, "SELECT * FROM trade_payments WHERE payid=?", payid)
	if err != nil {
		return nil, fmt.Errorf("query payment by id: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) AttachByIDs(ctx context.Context, ids []int) ([]map[string]interface{}, error) {
	if r.db == nil || len(ids) == 0 {
		return []map[string]interface{}{}, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	fieldIDs := make([]string, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
		fieldIDs[i] = strconv.Itoa(id)
	}
	query := "SELECT * FROM attachs WHERE aid IN (" + strings.Join(placeholders, ",") + ") ORDER BY FIELD(aid, " + strings.Join(fieldIDs, ",") + ")"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query attachs by ids: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func feedbackTypeWhere(uid int, feedbackType int) (string, []interface{}) {
	where := "1=1 AND uid=?"
	args := []interface{}{uid}
	switch feedbackType {
	case 1:
		where += " AND cid IN (0,1,2,3,4)"
	case 2:
		where += " AND cid IN (5,6,7)"
	}
	return where, args
}

func (r *Repository) CountMsgConversations(ctx context.Context, uid int) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM msgc WHERE 1=1 AND uid=?", uid).Scan(&total); err != nil {
		return 0, fmt.Errorf("count msg conversations: %w", err)
	}
	return total, nil
}

func (r *Repository) MsgConversations(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `SELECT a.*, b.senderid, b.content, b.sendtime, c.username, c.avatar
FROM msgc a
LEFT JOIN msgs b ON b.msgid=a.last_msgid
LEFT JOIN users c ON c.uid=a.ruid
WHERE 1=1 AND a.uid=?
ORDER BY a.last_sendtime DESC LIMIT ? OFFSET ?`, uid, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query msg conversations: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) MsgConversation(ctx context.Context, uid int, cid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 || cid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM msgc WHERE uid=? AND cid=?", uid, cid)
	if err != nil {
		return nil, fmt.Errorf("query msg conversation: %w", err)
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

func (r *Repository) CountMessages(ctx context.Context, uid int, cid int) (int, error) {
	if r.db == nil || uid <= 0 || cid <= 0 {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM msg_maps WHERE 1=1 AND uid=? AND cid=?", uid, cid).Scan(&total); err != nil {
		return 0, fmt.Errorf("count messages: %w", err)
	}
	return total, nil
}

func (r *Repository) Messages(ctx context.Context, uid int, cid int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil || uid <= 0 || cid <= 0 {
		return []map[string]interface{}{}, nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `SELECT a.*, b.senderid, b.content, b.sendtime, c.username, c.avatar
FROM msg_maps a
LEFT JOIN msgs b ON b.msgid=a.msgid
LEFT JOIN users c ON c.uid=b.senderid
WHERE 1=1 AND a.uid=? AND cid=?
ORDER BY a.sendtime ASC LIMIT ? OFFSET ?`, uid, cid, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) SetMsgRead(ctx context.Context, uid int, cid int) error {
	if r.db == nil || uid <= 0 || cid <= 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin set msg read: %w", err)
	}
	defer tx.Rollback()
	var ruid, newmsg int
	if err := tx.QueryRowContext(ctx, "SELECT ruid, newmsg FROM msgc WHERE uid=? AND cid=?", uid, cid).Scan(&ruid, &newmsg); err != nil {
		if err == sql.ErrNoRows {
			return tx.Commit()
		}
		return fmt.Errorf("query msgc read state: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE msgc SET newmsg=0 WHERE uid=? AND cid=?", uid, cid); err != nil {
		return fmt.Errorf("clear msgc newmsg: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE msgc SET risread=1 WHERE uid=? AND cid=?", ruid, cid); err != nil {
		return fmt.Errorf("set peer risread: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE users SET newmsg=newmsg-? WHERE uid=?", newmsg, uid); err != nil {
		return fmt.Errorf("update user newmsg: %w", err)
	}
	return tx.Commit()
}

func (r *Repository) CleanMsgRead(ctx context.Context, uid int) error {
	if r.db == nil || uid <= 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE users SET newmsg=0 WHERE uid=?", uid); err != nil {
		return fmt.Errorf("clean user newmsg: %w", err)
	}
	return nil
}

func (r *Repository) DeleteMsgConversations(ctx context.Context, uid int, cids []int) error {
	if r.db == nil || uid <= 0 || len(cids) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete msg conversations: %w", err)
	}
	defer tx.Rollback()
	for _, cid := range cids {
		if cid <= 0 {
			continue
		}
		var newmsg int
		if err := tx.QueryRowContext(ctx, "SELECT newmsg FROM msgc WHERE uid=? AND cid=?", uid, cid).Scan(&newmsg); err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("query msgc before delete: %w", err)
		}
		rows, err := tx.QueryContext(ctx, "SELECT msgid FROM msg_maps WHERE uid=? AND cid=?", uid, cid)
		if err != nil {
			return fmt.Errorf("query msg maps before delete: %w", err)
		}
		msgIDs := []int{}
		for rows.Next() {
			var msgID int
			if err := rows.Scan(&msgID); err != nil {
				rows.Close()
				return fmt.Errorf("scan msg map id: %w", err)
			}
			msgIDs = append(msgIDs, msgID)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return fmt.Errorf("iterate msg map ids: %w", err)
		}
		rows.Close()
		if _, err := tx.ExecContext(ctx, "DELETE FROM msgc WHERE uid=? AND cid=?", uid, cid); err != nil {
			return fmt.Errorf("delete msgc: %w", err)
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM msg_maps WHERE uid=? AND cid=?", uid, cid); err != nil {
			return fmt.Errorf("delete msg maps: %w", err)
		}
		if _, err := tx.ExecContext(ctx, "UPDATE users SET newmsg=newmsg-? WHERE uid=?", newmsg, uid); err != nil {
			return fmt.Errorf("update user newmsg after delete: %w", err)
		}
		for _, msgID := range msgIDs {
			if _, err := tx.ExecContext(ctx, "UPDATE msgs SET refcount=refcount-1 WHERE msgid=?", msgID); err != nil {
				return fmt.Errorf("decrement msg refcount: %w", err)
			}
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM msgs WHERE refcount=0"); err != nil {
		return fmt.Errorf("delete unreferenced msgs: %w", err)
	}
	return tx.Commit()
}

func (r *Repository) BalanceLogs(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM user_balancelogs WHERE 1=1 AND uid=? ORDER BY trxtime DESC LIMIT ? OFFSET ?", uid, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query balance logs: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) CountBalanceLogs(ctx context.Context, uid int) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_balancelogs WHERE 1=1 AND uid=?", uid).Scan(&total); err != nil {
		return 0, fmt.Errorf("count balance logs: %w", err)
	}
	return total, nil
}

func (r *Repository) CoinLogs(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	return r.CoinLogsByTypes(ctx, uid, nil, page, pageSize, "logid DESC")
}

func (r *Repository) CountCoinLogsByTypes(ctx context.Context, uid int, coinTypes []int) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	where, args := coinTypesWhere(uid, coinTypes)
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_coinlogs WHERE "+where, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count coin logs: %w", err)
	}
	return total, nil
}

func (r *Repository) CoinLogsByTypes(ctx context.Context, uid int, coinTypes []int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	if orderBy != "addtime DESC" {
		orderBy = "logid DESC"
	}
	where, args := coinTypesWhere(uid, coinTypes)
	args = append(args, pageSize, offset)
	rows, err := r.db.QueryContext(ctx, `SELECT user_coinlogs.*, users.mobi
FROM user_coinlogs
LEFT JOIN users ON users.uid=user_coinlogs.invited_uid
WHERE `+where+`
ORDER BY `+orderBy+` LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, fmt.Errorf("query coin logs: %w", err)
	}
	defer rows.Close()
	items, err := scanRows(rows)
	if err != nil {
		return nil, err
	}
	if err := r.populateInvitedUsernames(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

func coinTypesWhere(uid int, coinTypes []int) (string, []interface{}) {
	where := "1=1 AND user_coinlogs.uid=?"
	args := []interface{}{uid}
	if len(coinTypes) == 0 {
		return where, args
	}
	placeholders := make([]string, len(coinTypes))
	for i, coinType := range coinTypes {
		placeholders[i] = "?"
		args = append(args, coinType)
	}
	return where + " AND cointype IN (" + strings.Join(placeholders, ",") + ")", args
}

func (r *Repository) CoinBonusStats(ctx context.Context, uid int) (map[string]interface{}, error) {
	if r.db == nil {
		return map[string]interface{}{
			"inviteTotal": 0,
			"activeTotal": 0,
			"bonusTotal":  0,
		}, nil
	}
	inviteTotal, err := r.countDistinctInvitedUID(ctx, uid, 11)
	if err != nil {
		return nil, err
	}
	activeTotal, err := r.countDistinctInvitedUID(ctx, uid, 15)
	if err != nil {
		return nil, err
	}
	bonusTotal, err := r.sumCoinBonus(ctx, uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"inviteTotal": inviteTotal,
		"activeTotal": activeTotal,
		"bonusTotal":  bonusTotal,
	}, nil
}

func (r *Repository) countDistinctInvitedUID(ctx context.Context, uid int, coinType int) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT invited_uid) FROM user_coinlogs WHERE uid=? AND cointype=?", uid, coinType).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count coin invited users: %w", err)
	}
	return total, nil
}

func (r *Repository) sumCoinBonus(ctx context.Context, uid int) (int, error) {
	sumTypes := []int{0, 1, 9, 2, 3, 4, 5, 6, 7, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	placeholders := make([]string, len(sumTypes))
	args := make([]interface{}, 0, len(sumTypes)+1)
	args = append(args, uid)
	for i, coinType := range sumTypes {
		placeholders[i] = "?"
		args = append(args, coinType)
	}
	var total sql.NullInt64
	query := "SELECT SUM(coinnum) FROM user_coinlogs WHERE uid=? AND cointype IN (" + strings.Join(placeholders, ",") + ")"
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("sum coin bonus: %w", err)
	}
	if !total.Valid {
		return 0, nil
	}
	return int(total.Int64), nil
}

func (r *Repository) populateInvitedUsernames(ctx context.Context, rows []map[string]interface{}) error {
	if len(rows) == 0 || r.db == nil {
		return nil
	}
	ids := make([]int, 0, len(rows))
	seen := map[int]struct{}{}
	for _, row := range rows {
		uid := atoi(row["invited_uid"])
		if uid <= 0 {
			row["invited_username"] = ""
			continue
		}
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		ids = append(ids, uid)
	}
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, uid := range ids {
		placeholders[i] = "?"
		args[i] = uid
	}
	query := "SELECT uid, username, nickname FROM users WHERE uid IN (" + strings.Join(placeholders, ",") + ")"
	userRows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query invited users: %w", err)
	}
	defer userRows.Close()
	users, err := scanRows(userRows)
	if err != nil {
		return err
	}
	names := make(map[int]string, len(users))
	for _, user := range users {
		name := str(user["nickname"])
		if name == "" {
			name = str(user["username"])
		}
		names[atoi(user["uid"])] = name
	}
	for _, row := range rows {
		uid := atoi(row["invited_uid"])
		if uid <= 0 {
			row["invited_username"] = ""
			continue
		}
		row["invited_username"] = names[uid]
	}
	return nil
}

func (r *Repository) SettingExRate(ctx context.Context) (int, error) {
	row, err := r.queryOne(ctx, "SELECT value FROM settings WHERE uuid='setting'")
	if err != nil {
		return 0, fmt.Errorf("query setting: %w", err)
	}
	return parseSerializedExRate(fmt.Sprint(row["value"])), nil
}

func (r *Repository) queryOne(ctx context.Context, query string, args ...interface{}) (map[string]interface{}, error) {
	if r.db == nil {
		return nil, nil
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
		return nil, nil
	}
	return items[0], nil
}

var serializedExRatePattern = regexp.MustCompile(`s:6:"exrate";i:(\d+);`)

func parseSerializedExRate(value string) int {
	matches := serializedExRatePattern.FindStringSubmatch(value)
	if len(matches) != 2 {
		return 0
	}
	parsed, _ := strconv.Atoi(matches[1])
	return parsed
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

func atoi(value interface{}) int {
	parsed, _ := strconv.Atoi(fmt.Sprint(value))
	return parsed
}

func limitOffset(page int, pageSize int) int {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	return (page - 1) * pageSize
}

func str(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func miniVODTimeColumn(action int) string {
	if action == 2 {
		return "downtime"
	}
	return "playtime"
}

func splitTableBySID(sid string) string {
	if sid == "" {
		return "0"
	}
	first := strings.ToLower(sid[:1])
	if first >= "0" && first <= "9" || first >= "a" && first <= "f" {
		return first
	}
	return "0"
}

func isMissingTable(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "error 1146") || strings.Contains(message, "doesn't exist")
}
