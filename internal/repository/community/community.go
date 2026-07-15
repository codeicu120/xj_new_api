package community

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type TopicFilter struct {
	Action      string
	CategoryID  int
	Type        int
	FavoriteUID int
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CountTopics(ctx context.Context, filter TopicFilter) (int, error) {
	if r.db == nil {
		return 0, nil
	}
	where, args := topicWhere(filter)
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM topics WHERE 1=1 "+where, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count topics: %w", err)
	}
	return total, nil
}

func (r *Repository) ListTopics(ctx context.Context, filter TopicFilter, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	where, args := topicWhere(filter)
	args = append(args, pageSize, limitOffset(total, pageSize, page))
	return r.queryRows(ctx, "SELECT * FROM topics WHERE 1=1 "+where+" ORDER BY "+orderBy+" LIMIT ? OFFSET ?", args...)
}

func (r *Repository) Servers(ctx context.Context) ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	return r.queryRows(ctx, "SELECT * FROM vod_servers ORDER BY sortnum ASC, srvid ASC")
}

func (r *Repository) ImagesByTIDs(ctx context.Context, tids []int) (map[int][]map[string]interface{}, error) {
	out := map[int][]map[string]interface{}{}
	if r.db == nil || len(tids) == 0 {
		return out, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM topic_images WHERE tid IN("+intListSQL(tids)+")")
	if err != nil {
		return nil, fmt.Errorf("query topic images: %w", err)
	}
	for _, row := range rows {
		tid := atoi(row["tid"])
		out[tid] = append(out[tid], row)
	}
	return out, nil
}

func (r *Repository) VideosByTIDs(ctx context.Context, tids []int) (map[int][]map[string]interface{}, error) {
	out := map[int][]map[string]interface{}{}
	if r.db == nil || len(tids) == 0 {
		return out, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM topic_videos WHERE tid IN("+intListSQL(tids)+")")
	if err != nil {
		return nil, fmt.Errorf("query topic videos: %w", err)
	}
	for _, row := range rows {
		tid := atoi(row["tid"])
		out[tid] = append(out[tid], row)
	}
	return out, nil
}

func (r *Repository) FavoriteTopicIDs(ctx context.Context, uid int, tids []int) (map[int]int, error) {
	return r.flagTopicIDs(ctx, "topic_favorites", "tid", "uid", uid, tids)
}

func (r *Repository) UpTopicIDs(ctx context.Context, uid int, tids []int) (map[int]int, error) {
	return r.flagTopicIDs(ctx, "topic_ups", "tid", "uid", uid, tids)
}

func (r *Repository) TopicByID(ctx context.Context, tid int) (map[string]interface{}, error) {
	if r.db == nil || tid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM topics WHERE tid=?", tid)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) IncrementTopicVisit(ctx context.Context, tid int) error {
	if r.db == nil || tid <= 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE topics SET visit_count=visit_count+1 WHERE tid=?", tid); err != nil {
		return fmt.Errorf("increment topic visit: %w", err)
	}
	return nil
}

func (r *Repository) SetTopicUp(ctx context.Context, tid int, uid int, up bool, now int64) error {
	if r.db == nil || tid <= 0 || uid <= 0 {
		return nil
	}
	if !up {
		if _, err := r.db.ExecContext(ctx, "DELETE FROM topic_ups WHERE tid=? AND uid=?", tid, uid); err != nil {
			return fmt.Errorf("delete topic up: %w", err)
		}
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "INSERT IGNORE INTO topic_ups(uid,tid,created_at) VALUES(?,?,?)", uid, tid, now); err != nil {
		return fmt.Errorf("insert topic up: %w", err)
	}
	return nil
}

func (r *Repository) IncrementTopicUp(ctx context.Context, tid int, delta int) error {
	if r.db == nil || tid <= 0 || delta == 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE topics SET upnum=upnum+? WHERE tid=?", delta, tid); err != nil {
		return fmt.Errorf("increment topic up: %w", err)
	}
	return nil
}

func (r *Repository) CountComments(ctx context.Context, tid int) (int, error) {
	if r.db == nil || tid <= 0 {
		return 0, nil
	}
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM topic_comments a WHERE 1=1 AND a.tid=? AND a.rootid=0 AND a.showtype=0", tid).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count topic comments: %w", err)
	}
	return total, nil
}

func (r *Repository) ListComments(ctx context.Context, tid int, total int, page int, pageSize int, orderBy string) ([]map[string]interface{}, error) {
	if r.db == nil || tid <= 0 {
		return []map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT a.*, b.username, b.nickname, b.avatar, b.gender, b.sysgid, b.sysgid_exptime, b.gid, b.gids FROM topic_comments a LEFT JOIN users b ON b.uid=a.uid WHERE 1=1 AND a.tid=? AND a.rootid=0 AND a.showtype=0 ORDER BY "+orderBy+" LIMIT ? OFFSET ?", tid, pageSize, limitOffset(total, pageSize, page))
	if err != nil {
		return nil, fmt.Errorf("query topic comments: %w", err)
	}
	for _, row := range rows {
		subrows, err := r.queryRows(ctx, "SELECT a.*, b.username, b.nickname, b.avatar, b.gender, b.sysgid, b.sysgid_exptime, b.gid, b.gids FROM topic_comments a LEFT JOIN users b ON b.uid=a.uid WHERE a.rootid=? AND a.showtype=0 ORDER BY lft ASC LIMIT 1000", row["id"])
		if err != nil {
			return nil, fmt.Errorf("query topic subcomments: %w", err)
		}
		prevDepth := 0
		for _, subrow := range subrows {
			depth := atoi(subrow["depth"])
			if prevDepth > 0 && prevDepth >= depth {
				subrow["__closenum__"] = prevDepth - depth + 1
			} else {
				subrow["__closenum__"] = 0
			}
			prevDepth = depth
		}
		row["subrows"] = subrows
		row["__closenum__"] = prevDepth
	}
	return rows, nil
}

func (r *Repository) CommentByID(ctx context.Context, cid int) (map[string]interface{}, error) {
	if r.db == nil || cid <= 0 {
		return map[string]interface{}{}, nil
	}
	rows, err := r.queryRows(ctx, "SELECT * FROM topic_comments WHERE id=?", cid)
	if err != nil {
		return nil, fmt.Errorf("query topic comment by id: %w", err)
	}
	if len(rows) == 0 {
		return map[string]interface{}{}, nil
	}
	return rows[0], nil
}

func (r *Repository) SetCommentUp(ctx context.Context, cid int, uid int, up bool, now int64) error {
	if r.db == nil || cid <= 0 || uid <= 0 {
		return nil
	}
	if !up {
		if _, err := r.db.ExecContext(ctx, "DELETE FROM topic_comments_ups WHERE cid=? AND uid=?", cid, uid); err != nil {
			return fmt.Errorf("delete topic comment up: %w", err)
		}
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "INSERT IGNORE INTO topic_comments_ups(uid,cid,created_at) VALUES(?,?,?)", uid, cid, now); err != nil {
		return fmt.Errorf("insert topic comment up: %w", err)
	}
	return nil
}

func (r *Repository) IncrementCommentUp(ctx context.Context, cid int, delta int) error {
	if r.db == nil || cid <= 0 || delta == 0 {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE topic_comments SET upnum=upnum+? WHERE id=?", delta, cid); err != nil {
		return fmt.Errorf("increment topic comment up: %w", err)
	}
	return nil
}

func (r *Repository) UpCommentIDs(ctx context.Context, uid int, ids []int) (map[int]int, error) {
	return r.flagTopicIDs(ctx, "topic_comments_ups", "cid", "uid", uid, ids)
}

func (r *Repository) flagTopicIDs(ctx context.Context, table string, idColumn string, ownerColumn string, uid int, ids []int) (map[int]int, error) {
	out := map[int]int{}
	if r.db == nil || uid <= 0 || len(ids) == 0 {
		return out, nil
	}
	rows, err := r.queryRows(ctx, "SELECT "+idColumn+" FROM "+table+" WHERE "+ownerColumn+"=? AND "+idColumn+" IN("+intListSQL(ids)+")", uid)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[atoi(row[idColumn])] = 1
	}
	return out, nil
}

func topicWhere(filter TopicFilter) (string, []interface{}) {
	where := ""
	args := []interface{}{}
	switch filter.Action {
	case "recommend":
		where += " AND is_recommend=1"
	}
	if filter.Type > 0 {
		where += " AND type=?"
		args = append(args, filter.Type)
	}
	if filter.CategoryID > 0 {
		where += " AND category_id=?"
		args = append(args, filter.CategoryID)
	}
	if filter.FavoriteUID > 0 {
		where += " AND tid IN(SELECT tid FROM topic_favorites WHERE uid=?)"
		args = append(args, filter.FavoriteUID)
	}
	return where, args
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
	return result, rows.Err()
}

func normalizeSQLValue(value interface{}) interface{} {
	switch v := value.(type) {
	case nil:
		return ""
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}

func limitOffset(total int, pageSize int, page int) int {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		return 0
	}
	offset := (page - 1) * pageSize
	if total > 0 && offset >= total {
		last := (total - 1) / pageSize
		offset = last * pageSize
	}
	return offset
}

func intListSQL(ids []int) string {
	parts := []string{}
	seen := map[int]struct{}{}
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		parts = append(parts, fmt.Sprint(id))
	}
	if len(parts) == 0 {
		return "NULL"
	}
	return strings.Join(parts, ",")
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
