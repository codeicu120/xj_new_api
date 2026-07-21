package comment

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"xj_comp/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) VODByID(ctx context.Context, vodID int) (map[string]interface{}, error) {
	if r.db == nil || vodID <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT vodid,showtype FROM vods WHERE vodid=?", vodID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) UserGroups(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT gid,gicon,perms,weight,scope FROM user_groups ORDER BY gid ASC")
}

func (r *Repository) CountRoots(ctx context.Context, vodID int) (int, error) {
	if r.db == nil || vodID <= 0 {
		return 0, nil
	}
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vod_comments a WHERE 1=1 AND a.vodid=? AND a.rootid=0 AND a.showtype=0", vodID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count comments: %w", err)
	}
	return total, nil
}

func (r *Repository) RootComments(ctx context.Context, vodID int, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error) {
	if r.db == nil || vodID <= 0 {
		return []map[string]interface{}{}, nil
	}
	offset := limitOffset(total, pageSize, page)
	rows, err := r.queryRows(ctx, `
SELECT a.*, b.username, b.nickname, b.avatar, b.gender, b.sysgid, b.sysgid_exptime, b.gid, b.gids
FROM vod_comments a
LEFT JOIN users b ON b.uid=a.uid
WHERE 1=1 AND a.vodid=? AND a.rootid=0 AND a.showtype=0
ORDER BY `+orderBy+`
LIMIT ? OFFSET ?`, vodID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		subs, err := r.queryRows(ctx, `
SELECT a.*, b.username, b.nickname, b.avatar, b.gender, b.sysgid, b.sysgid_exptime, b.gid, b.gids
FROM vod_comments a
LEFT JOIN users b ON b.uid=a.uid
WHERE a.rootid=? AND a.showtype=0
ORDER BY lft ASC
LIMIT 1000`, row["id"])
		if err != nil {
			return nil, err
		}
		prevDepth := 0
		for _, sub := range subs {
			depth := atoi(sub["depth"])
			if prevDepth > 0 && prevDepth >= depth {
				sub["__closenum__"] = fmt.Sprint(prevDepth - depth + 1)
			} else {
				sub["__closenum__"] = "0"
			}
			prevDepth = depth
		}
		row["subrows"] = subs
		row["__closenum__"] = fmt.Sprint(prevDepth)
	}
	return rows, nil
}

func (r *Repository) CommentByID(ctx context.Context, id int) (map[string]interface{}, error) {
	if r.db == nil || id <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM vod_comments WHERE id=?", id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) IncrementVote(ctx context.Context, id int, field string) error {
	if r.db == nil || id <= 0 {
		return nil
	}
	if field != "upnum" && field != "downnum" {
		return fmt.Errorf("invalid vote field %s", field)
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE vod_comments SET "+field+"="+field+"+1 WHERE id=?", id); err != nil {
		return fmt.Errorf("increment comment vote: %w", err)
	}
	return nil
}

func (r *Repository) CountByActorSince(ctx context.Context, actor interface{}, since int64, rewarded bool) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	query := "SELECT COUNT(*) FROM vod_comments WHERE sid=? AND addtime>=?"
	args := []interface{}{fmt.Sprint(actor), since}
	if rewarded {
		query += " AND showtype=0"
	}
	var total int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count comments by actor: %w", err)
	}
	return total, nil
}

func (r *Repository) RecentCommentsByUID(ctx context.Context, uid int, since int64) ([]map[string]interface{}, error) {
	if r.db == nil || uid <= 0 {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT content FROM vod_comments WHERE uid=? AND addtime>?", uid, since)
}

func (r *Repository) RecentCommentsByIP(ctx context.Context, ip string, since int64) ([]map[string]interface{}, error) {
	if r.db == nil || ip == "" {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT content FROM vod_comments WHERE addtime>? AND ip=?", since, ip)
}

func (r *Repository) CreateComment(ctx context.Context, input domain.CommentCreateInput, parent map[string]interface{}) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin comment tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if len(parent) > 0 {
		rootID := input.RootID
		rgt := atoi(parent["rgt"])
		if _, err := tx.ExecContext(ctx, "UPDATE vod_comments SET rgt=rgt+2 WHERE rootid=? AND rgt>=?", rootID, rgt); err != nil {
			return 0, fmt.Errorf("shift comment rgt: %w", err)
		}
		if _, err := tx.ExecContext(ctx, "UPDATE vod_comments SET lft=lft+2 WHERE rootid=? AND lft>?", rootID, rgt); err != nil {
			return 0, fmt.Errorf("shift comment lft: %w", err)
		}
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO vod_comments
		(rootid,parentid,lft,rgt,depth,vodid,uid,sid,content,addtime,ip,showtype)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		input.RootID, input.ParentID, input.Left, input.Right, input.Depth, input.VODID, input.UID, input.SID, input.Content, input.AddTime, input.IP, input.ShowType,
	)
	if err != nil {
		return 0, fmt.Errorf("insert comment: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert comment id: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit comment tx: %w", err)
	}
	committed = true
	return int(id), nil
}

func (r *Repository) IncrementVODCommentCount(ctx context.Context, vodID int) error {
	if r.db == nil || vodID <= 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE vods SET commentcount=commentcount+1 WHERE vodid=?", vodID); err != nil {
		return fmt.Errorf("increment vod comment count: %w", err)
	}
	return nil
}

func (r *Repository) queryRows(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query comments: %w", err)
	}
	defer rows.Close()
	return scanRows(rows)
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

func atoi(value interface{}) int {
	var parsed int
	_, _ = fmt.Sscan(fmt.Sprint(value), &parsed)
	return parsed
}
