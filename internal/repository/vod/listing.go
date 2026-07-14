package vod

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type ListingFilter struct {
	CateIDs    []int
	AreaID     int
	YearID     int
	Definition int
	Duration   int
	FreeType   int
	Mosaic     int
	LangVoice  int
	Recommend  bool
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
	parts = append(parts, "showtype=0")
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
