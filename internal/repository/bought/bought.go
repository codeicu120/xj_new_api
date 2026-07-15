package bought

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const coinTypeBuyVOD = 113

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Items(ctx context.Context, uid int, page int, pageSize int) (int, []map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return 0, []map[string]interface{}{}, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_bought WHERE uid=?", uid).Scan(&total); err != nil {
		return 0, nil, fmt.Errorf("count bought vods: %w", err)
	}
	offset := limitOffset(total, pageSize, page)
	rows, err := r.queryRows(
		ctx,
		"SELECT a.*, b.* FROM user_bought a LEFT JOIN vods b ON b.vodid=a.vodid WHERE a.uid=? AND b.showtype=0 ORDER BY a.buytime DESC LIMIT ? OFFSET ?",
		uid,
		pageSize,
		offset,
	)
	if err != nil {
		return 0, nil, fmt.Errorf("list bought vods: %w", err)
	}
	return total, rows, nil
}

func (r *Repository) Delete(ctx context.Context, uid int, vodid int) error {
	if r.db == nil || uid <= 0 || vodid <= 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "DELETE FROM user_bought WHERE uid=? AND vodid=?", uid, vodid); err != nil {
		return fmt.Errorf("delete bought vod: %w", err)
	}
	return nil
}

func (r *Repository) VODByID(ctx context.Context, vodid int) (map[string]interface{}, error) {
	if r.db == nil || vodid <= 0 {
		return map[string]interface{}{}, nil
	}
	row, err := r.queryOne(ctx, "SELECT * FROM vods WHERE vodid=?", vodid)
	if err != nil {
		return nil, fmt.Errorf("query vod: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) BoughtCount(ctx context.Context, uid int, vodid int) (int, error) {
	if r.db == nil || uid <= 0 || vodid <= 0 {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_bought WHERE uid=? AND vodid=?", uid, vodid).Scan(&total); err != nil {
		return 0, fmt.Errorf("count bought vod: %w", err)
	}
	return total, nil
}

func (r *Repository) Goldbean(ctx context.Context, uid int) (map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return map[string]interface{}{}, nil
	}
	row, err := r.queryOne(ctx, "SELECT uid, gold_bean FROM users_goldbean WHERE uid=?", uid)
	if err != nil {
		return nil, fmt.Errorf("query goldbean: %w", err)
	}
	if row == nil {
		return map[string]interface{}{}, nil
	}
	return row, nil
}

func (r *Repository) BuyVOD(ctx context.Context, uid int, vodid int, price int) error {
	if r.db == nil || uid <= 0 || vodid <= 0 || price <= 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin buy vod: %w", err)
	}
	defer tx.Rollback()

	var balance int
	err = tx.QueryRowContext(ctx, "SELECT gold_bean FROM users_goldbean WHERE uid=? FOR UPDATE", uid).Scan(&balance)
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("lock goldbean: %w", err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO users_goldbean(uid, gold_bean) VALUES(?, 0)", uid); err != nil {
			return fmt.Errorf("create goldbean: %w", err)
		}
		balance = 0
	}
	newBalance := balance - price
	if newBalance < 0 {
		return fmt.Errorf("goldbean insufficient")
	}
	result, err := tx.ExecContext(ctx, "UPDATE users_goldbean SET gold_bean=? WHERE uid=?", newBalance, uid)
	if err != nil {
		return fmt.Errorf("update goldbean: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("update goldbean affected %d rows", affected)
	}
	now := time.Now().Unix()
	result, err = tx.ExecContext(ctx, "INSERT INTO user_beanlogs(bean_type, uid, bean_num, balance, add_time, remark) VALUES(?, ?, ?, ?, ?, ?)", coinTypeBuyVOD, uid, -price, newBalance, now, vodid)
	if err != nil {
		return fmt.Errorf("insert beanlog: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected < 1 {
		return fmt.Errorf("insert beanlog affected %d rows", affected)
	}
	result, err = tx.ExecContext(ctx, "REPLACE INTO user_bought(uid, vodid, buytime) VALUES(?, ?, ?)", uid, vodid, now)
	if err != nil {
		return fmt.Errorf("save bought vod: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected < 1 {
		return fmt.Errorf("save bought vod affected %d rows", affected)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit buy vod: %w", err)
	}
	return nil
}

func (r *Repository) queryOne(ctx context.Context, query string, args ...interface{}) (map[string]interface{}, error) {
	rows, err := r.queryRows(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
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
