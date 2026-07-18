package user

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Repository struct {
	db           *sql.DB
	deletionList RedisHashStore
}

type RedisHashStore interface {
	HExists(ctx context.Context, key string, field string) (bool, error)
	HSet(ctx context.Context, key string, field string, value interface{}) error
}

type RedisHashDeleteStore interface {
	HDel(ctx context.Context, key string, field string) error
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithDeletionList(store RedisHashStore) *Repository {
	r.deletionList = store
	return r
}

func (r *Repository) UserBySession(ctx context.Context, sid string) (map[string]interface{}, error) {
	if r.db == nil || !validSID(sid) {
		return nil, nil
	}

	session, err := r.queryOne(ctx, "SELECT sid,token,uid,type,timestamp FROM sessions WHERE sid=?", sid)
	if err != nil {
		return nil, fmt.Errorf("query session: %w", err)
	}
	if len(session) == 0 {
		return nil, nil
	}
	if fmt.Sprint(session["type"]) != "0" {
		return nil, nil
	}

	user, err := r.queryOne(ctx, "SELECT * FROM users WHERE uid=?", session["uid"])
	if err != nil {
		return nil, fmt.Errorf("query session user: %w", err)
	}
	if len(user) == 0 {
		return nil, nil
	}
	expected := md5.Sum([]byte(fmt.Sprintf("%s_%s", user["password"], user["salt"])))
	if fmt.Sprint(session["token"]) != hex.EncodeToString(expected[:]) {
		return nil, nil
	}
	user["sid"] = sid
	return user, nil
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

func (r *Repository) UserByID(ctx context.Context, uid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	row, err := r.queryOne(ctx, "SELECT * FROM users WHERE uid=?", uid)
	if err != nil {
		return nil, fmt.Errorf("query user by id: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) UpdateUserProfile(ctx context.Context, uid int, gender int, nickname *string) error {
	if r.db == nil || uid <= 0 {
		return nil
	}
	if nickname != nil {
		if _, err := r.db.ExecContext(ctx, "UPDATE users SET gender=?, nickname=? WHERE uid=?", gender, *nickname, uid); err != nil {
			return fmt.Errorf("update user profile: %w", err)
		}
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE users SET gender=? WHERE uid=?", gender, uid); err != nil {
		return fmt.Errorf("update user profile gender: %w", err)
	}
	return nil
}

func (r *Repository) ResetPassword(ctx context.Context, uid int, passwordHash string, salt string) error {
	if r.db == nil || uid <= 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE users SET password=?, salt=? WHERE uid=?", passwordHash, salt, uid); err != nil {
		return fmt.Errorf("reset user password: %w", err)
	}
	return nil
}

func (r *Repository) UserByMobi(ctx context.Context, mobi string) (map[string]interface{}, error) {
	if r.db == nil || strings.TrimSpace(mobi) == "" {
		return map[string]interface{}{}, nil
	}
	row, err := r.queryOne(ctx, "SELECT * FROM users WHERE mobi=?", mobi)
	if err != nil {
		return nil, fmt.Errorf("query user by mobi: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) UserByEmail(ctx context.Context, email string) (map[string]interface{}, error) {
	if r.db == nil || strings.TrimSpace(email) == "" {
		return map[string]interface{}{}, nil
	}
	row, err := r.queryOne(ctx, "SELECT * FROM users WHERE email=?", email)
	if err != nil {
		return nil, fmt.Errorf("query user by email: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) UserByUsername(ctx context.Context, username string) (map[string]interface{}, error) {
	if r.db == nil || strings.TrimSpace(username) == "" {
		return map[string]interface{}{}, nil
	}
	row, err := r.queryOne(ctx, "SELECT * FROM users WHERE username=?", username)
	if err != nil {
		return nil, fmt.Errorf("query user by username: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) SettingByUUID(ctx context.Context, uuid string) (map[string]interface{}, error) {
	if r.db == nil || strings.TrimSpace(uuid) == "" {
		return map[string]interface{}{}, nil
	}
	row, err := r.queryOne(ctx, "SELECT uuid,value,type FROM settings WHERE uuid=?", uuid)
	if err != nil {
		return nil, fmt.Errorf("query setting %s: %w", uuid, err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) KeylimitCountSince(ctx context.Context, key string, since int64) (int, error) {
	if r.db == nil || strings.TrimSpace(key) == "" {
		return 0, nil
	}
	var total sql.NullInt64
	if err := r.db.QueryRowContext(ctx, "SELECT SUM(keynum) FROM keylimits WHERE keyid=? AND ctimestamp>?", md5Hex(key), since).Scan(&total); err != nil {
		return 0, fmt.Errorf("query keylimit %s: %w", key, err)
	}
	if !total.Valid {
		return 0, nil
	}
	return int(total.Int64), nil
}

func (r *Repository) AccountDeletionExists(ctx context.Context, uid int) (bool, error) {
	if r.deletionList == nil || uid <= 0 {
		return false, nil
	}
	exists, err := r.deletionList.HExists(ctx, "delAccountList", strconv.Itoa(uid))
	if err != nil {
		return false, fmt.Errorf("query account deletion list: %w", err)
	}
	return exists, nil
}

func (r *Repository) RequestAccountDeletion(ctx context.Context, uid int, sid string, now int64) error {
	if uid <= 0 {
		return nil
	}
	if r.deletionList != nil {
		if err := r.deletionList.HSet(ctx, "delAccountList", strconv.Itoa(uid), now); err != nil {
			return fmt.Errorf("write account deletion list: %w", err)
		}
	}
	return r.Logout(ctx, sid)
}

func (r *Repository) ChangePhone(ctx context.Context, uid int, mobi string) (bool, string, error) {
	if r.db == nil || uid <= 0 || strings.TrimSpace(mobi) == "" {
		return true, "", nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, "", fmt.Errorf("begin change phone: %w", err)
	}
	defer tx.Rollback()
	row, err := queryOneTx(ctx, tx, "SELECT uid FROM users WHERE mobi=? FOR UPDATE", mobi)
	if err != nil {
		return false, "", fmt.Errorf("lock user by mobi: %w", err)
	}
	if len(row) > 0 {
		return false, "手机号已经存在", nil
	}
	result, err := tx.ExecContext(ctx, "UPDATE users SET mobi=? WHERE uid=?", mobi, uid)
	if err != nil {
		return false, "", fmt.Errorf("update user mobi: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return false, "手机号更换失败,请重试", nil
	}
	if err := tx.Commit(); err != nil {
		return false, "", fmt.Errorf("commit change phone: %w", err)
	}
	return true, "", nil
}

func (r *Repository) CreateLoginSession(ctx context.Context, uid int, passwordHash string, salt string, now int64) (string, error) {
	if r.db == nil || uid <= 0 {
		return "", nil
	}
	sid, err := randomSID()
	if err != nil {
		return "", err
	}
	token := md5Hex(passwordHash + "_" + salt)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin login session: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, "REPLACE INTO sessions(sid, token, uid, type, timestamp) VALUES(?, ?, ?, ?, ?)", sid, token, uid, 0, now)
	if err != nil {
		return "", fmt.Errorf("replace login session: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return "", nil
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE uid=? AND type=0 AND sid<>?", uid, sid); err != nil {
		return "", fmt.Errorf("delete old sessions: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit login session: %w", err)
	}
	return sid, nil
}

func (r *Repository) Quota(ctx context.Context, uid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	row, err := r.queryOne(ctx, "SELECT * FROM users_quota WHERE uid=?", uid)
	if err != nil {
		return nil, fmt.Errorf("query user quota: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) Goldbean(ctx context.Context, uid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	row, err := r.queryOne(ctx, "SELECT * FROM users_goldbean WHERE uid=?", uid)
	if err != nil {
		return nil, fmt.Errorf("query user goldbean: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) ClearAccountDeletion(ctx context.Context, uid int) error {
	if r.deletionList == nil || uid <= 0 {
		return nil
	}
	deleter, ok := r.deletionList.(RedisHashDeleteStore)
	if !ok || deleter == nil {
		return nil
	}
	return deleter.HDel(ctx, "delAccountList", strconv.Itoa(uid))
}

func (r *Repository) BotByID(ctx context.Context, uid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	row, err := r.queryOne(ctx, "SELECT * FROM bot_users WHERE uid=?", uid)
	if err != nil {
		return nil, fmt.Errorf("query bot user by id: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
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

func (r *Repository) CountRecommended(ctx context.Context, uid int) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users a WHERE 1=1 AND a.recommend_uid=?", uid).Scan(&total); err != nil {
		return 0, fmt.Errorf("count recommended users: %w", err)
	}
	return total, nil
}

func (r *Repository) UpdateAvatar(ctx context.Context, uid int, avatarID string) error {
	if r.db == nil {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE users SET avatar=? WHERE uid=?", avatarID, uid); err != nil {
		return fmt.Errorf("update user avatar: %w", err)
	}
	return nil
}

func (r *Repository) Logout(ctx context.Context, sid string) error {
	if r.db == nil || !validSID(sid) {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE sid=? AND type=0", sid); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (r *Repository) RecommendedUsers(ctx context.Context, uid int, page int, pageSize int) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `
SELECT a.*, b.gname AS sysgname, c.gname
FROM users a
LEFT JOIN user_groups b ON b.gid=a.sysgid
LEFT JOIN user_groups c ON c.gid=a.gid
WHERE 1=1 AND a.recommend_uid=?
ORDER BY a.uid DESC
LIMIT ? OFFSET ?`, uid, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query recommended users: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
}

func validSID(sid string) bool {
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

func md5Hex(value string) string {
	sum := md5.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

func randomSID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
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

func CleanToken(token string) string {
	token = strings.TrimSpace(token)
	decoded, err := hex.DecodeString(token)
	if err == nil && len(decoded) == 32 {
		return string(decoded)
	}
	return token
}
