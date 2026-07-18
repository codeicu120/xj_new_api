package invite

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"xj_comp/internal/domain"
)

type Repository struct {
	db *sql.DB
}

const (
	vipGID            = 6
	coinTypeInvite    = 32
	coinTypeInviteDay = 201
)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) RecordRecommend(ctx context.Context, uid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM user_recommend AS r LEFT JOIN users AS u ON r.recommend_uid=u.uid WHERE r.uid=?", uid)
	if err != nil {
		return nil, fmt.Errorf("query recommend record: %w", err)
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

func (r *Repository) UserByInviteKey(ctx context.Context, inviteCode string) (map[string]interface{}, error) {
	if r.db == nil {
		return map[string]interface{}{}, nil
	}
	uniqkey, err := strconv.ParseInt(strings.TrimSpace(inviteCode), 36, 64)
	if err != nil || uniqkey <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM users WHERE uniqkey=?", uniqkey)
	if err != nil {
		return nil, fmt.Errorf("query invite user by key: %w", err)
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

func (r *Repository) UserByID(ctx context.Context, uid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM users WHERE uid=?", uid)
	if err != nil {
		return nil, fmt.Errorf("query user by id: %w", err)
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

func (r *Repository) Groups(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM user_groups ORDER BY gid ASC")
	if err != nil {
		return nil, fmt.Errorf("query user groups: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *Repository) SettingByUUID(ctx context.Context, uuid string) (map[string]interface{}, error) {
	if r.db == nil || strings.TrimSpace(uuid) == "" {
		return map[string]interface{}{}, nil
	}
	rows, err := r.db.QueryContext(ctx, "SELECT uuid,value,type FROM settings WHERE uuid=?", uuid)
	if err != nil {
		return nil, fmt.Errorf("query setting %s: %w", uuid, err)
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

func (r *Repository) DeletedUserTag(context.Context, string) (bool, error) {
	return false, nil
}

func (r *Repository) BindInvite(ctx context.Context, input domain.InviteBindInput) (bool, error) {
	if r.db == nil {
		return true, nil
	}
	if input.UID <= 0 || input.InviterUID <= 0 {
		return false, nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin invite bind: %w", err)
	}
	defer tx.Rollback()

	existing, err := queryOneTx(ctx, tx, "SELECT uid FROM user_recommend WHERE uid=? FOR UPDATE", input.UID)
	if err != nil {
		return false, fmt.Errorf("lock recommend record: %w", err)
	}
	if len(existing) > 0 {
		return false, nil
	}
	current, err := queryOneTx(ctx, tx, "SELECT * FROM users WHERE uid=? FOR UPDATE", input.UID)
	if err != nil {
		return false, fmt.Errorf("lock invite current user: %w", err)
	}
	if len(current) == 0 {
		return false, nil
	}
	inviter, err := queryOneTx(ctx, tx, "SELECT * FROM users WHERE uid=? FOR UPDATE", input.InviterUID)
	if err != nil {
		return false, fmt.Errorf("lock invite inviter: %w", err)
	}
	if len(inviter) == 0 || atoi(inviter["uid"]) == input.UID {
		return false, nil
	}
	result, err := tx.ExecContext(ctx, "INSERT IGNORE INTO user_recommend(uid,recommend_uid,regtime) VALUES(?, ?, ?)", input.UID, input.InviterUID, input.Now)
	if err != nil {
		return false, fmt.Errorf("insert user recommend: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected < 1 {
		return false, nil
	}
	if !input.NoReward {
		if ok, err := r.applyInviteRewards(ctx, tx, input, current, inviter); err != nil || !ok {
			return ok, err
		}
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit invite bind: %w", err)
	}
	return true, nil
}

func (r *Repository) applyInviteRewards(ctx context.Context, tx *sql.Tx, input domain.InviteBindInput, current map[string]interface{}, inviter map[string]interface{}) (bool, error) {
	recommendTotal := atoi(inviter["recommend_total"]) + 1
	group := inviteGroup(input.Groups, recommendTotal)
	result, err := tx.ExecContext(ctx, "UPDATE users SET recommend_total=?, gid=? WHERE uid=?", recommendTotal, atoi(group["gid"]), input.InviterUID)
	if err != nil {
		return false, fmt.Errorf("update inviter recommend total: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return false, nil
	}

	incrDays := input.Bonus["invite3"]
	if recommendTotal == 1 {
		incrDays = input.Bonus["invite1"]
	} else if recommendTotal == 2 {
		incrDays = input.Bonus["invite2"]
	}
	if incrDays > 0 {
		expireAt := input.Now + int64(incrDays)*86400
		if atoi(inviter["sysgid"]) == vipGID && atoi64(inviter["sysgid_exptime"]) > input.Now {
			expireAt = atoi64(inviter["sysgid_exptime"]) + int64(incrDays)*86400
		}
		result, err = tx.ExecContext(ctx, "UPDATE users SET sysgid=?, sysgid_exptime=? WHERE uid=?", vipGID, expireAt, input.InviterUID)
		if err != nil {
			return false, fmt.Errorf("update inviter vip days: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			return false, nil
		}
		remark := fmt.Sprintf("邀请好友累计%d个,赠送vip%d天", recommendTotal, incrDays)
		if _, err := tx.ExecContext(ctx, "INSERT INTO user_coinlogs(cointype, uid, coinnum, addtime, balance, remark, invited_uid) VALUES(?, ?, 0, ?, ?, ?, ?)", coinTypeInviteDay, input.InviterUID, input.Now, incrDays, remark, input.UID); err != nil {
			return false, fmt.Errorf("insert invite vip days coinlog: %w", err)
		}
	}

	user2 := map[string]interface{}{}
	user3 := map[string]interface{}{}
	if uid2 := atoi(inviter["recommend_uid"]); uid2 > 0 {
		var err error
		user2, err = queryOneTx(ctx, tx, "SELECT * FROM users WHERE uid=?", uid2)
		if err != nil {
			return false, fmt.Errorf("query invite second level user: %w", err)
		}
	}
	if uid3 := atoi(user2["recommend_uid"]); uid3 > 0 {
		var err error
		user3, err = queryOneTx(ctx, tx, "SELECT * FROM users WHERE uid=?", uid3)
		if err != nil {
			return false, fmt.Errorf("query invite third level user: %w", err)
		}
	}
	if input.Bonus["reg1"] > 0 {
		if ok, err := rewardInviteCoin(ctx, tx, input.InviterUID, input.Bonus["reg1"], "推广得金币奖励: "+fmt.Sprint(current["username"]), input.UID, input.Now); err != nil || !ok {
			return ok, err
		}
	}
	if input.Bonus["reg2"] > 0 && atoi(user2["uid"]) > 0 {
		if ok, err := rewardInviteCoin(ctx, tx, atoi(user2["uid"]), input.Bonus["reg2"], "推广得金币奖励: "+fmt.Sprint(inviter["username"]), input.InviterUID, input.Now); err != nil || !ok {
			return ok, err
		}
	}
	if input.Bonus["reg3"] > 0 && atoi(user3["uid"]) > 0 {
		if ok, err := rewardInviteCoin(ctx, tx, atoi(user3["uid"]), input.Bonus["reg3"], "推广得金币奖励: "+fmt.Sprint(user2["username"]), atoi(user2["uid"]), input.Now); err != nil || !ok {
			return ok, err
		}
	}

	currentExpireAt := input.Now + 3*86400
	if atoi64(current["sysgid_exptime"]) > input.Now {
		currentExpireAt = atoi64(current["sysgid_exptime"]) + 3*86400
	}
	result, err = tx.ExecContext(ctx, "UPDATE users SET sysgid_exptime=?, sysgid=? WHERE uid=?", currentExpireAt, vipGID, input.UID)
	if err != nil {
		return false, fmt.Errorf("update current invite vip: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return false, nil
	}
	return true, nil
}

func rewardInviteCoin(ctx context.Context, tx *sql.Tx, uid int, coins int, remark string, invitedUID int, now int64) (bool, error) {
	var balance int
	if err := tx.QueryRowContext(ctx, "SELECT goldcoin FROM users_quota WHERE uid=? FOR UPDATE", uid).Scan(&balance); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("lock invite reward quota: %w", err)
	}
	newBalance := balance + coins
	if newBalance != balance {
		result, err := tx.ExecContext(ctx, "UPDATE users_quota SET goldcoin=? WHERE uid=?", newBalance, uid)
		if err != nil {
			return false, fmt.Errorf("update invite reward quota: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			return false, nil
		}
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO user_coinlogs(cointype, uid, coinnum, balance, addtime, remark, invited_uid) VALUES(?, ?, ?, ?, ?, ?, ?)", coinTypeInvite, uid, coins, newBalance, now, remark, invitedUID); err != nil {
		return false, fmt.Errorf("insert invite reward coinlog: %w", err)
	}
	return true, nil
}

func inviteGroup(groups []map[string]interface{}, minup int) map[string]interface{} {
	defGroup := map[string]interface{}{}
	candidates := make([]map[string]interface{}, 0, len(groups))
	for _, group := range groups {
		if atoi(group["gid"]) == 0 {
			defGroup = group
			continue
		}
		candidates = append(candidates, group)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return atoi(candidates[i]["minup"]) < atoi(candidates[j]["minup"])
	})
	prevIndex := -1
	for i, group := range candidates {
		if minup < atoi(group["minup"]) {
			prevIndex = i - 1
			break
		}
	}
	if prevIndex >= 0 && prevIndex < len(candidates) {
		return candidates[prevIndex]
	}
	return defGroup
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
