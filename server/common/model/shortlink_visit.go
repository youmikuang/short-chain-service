package model

import (
	"context"
	"database/sql"
	"time"
)

// ShortLinkVisit 一条短链被访问的明细记录（每次打开 /r/:code 写入一条）。
// 数据落在 ClickHouse 的 click_events 表（不再使用 MySQL）。
type ShortLinkVisit struct {
	Code      string
	LongURL   string
	UserId    int64
	IP        string
	Referer   string
	Status    int64
	Source    string
	LatencyMs int64
	CreatedAt time.Time
}

type ShortLinkVisitModel struct {
	conn *sql.DB
}

func NewShortLinkVisitModel(conn *sql.DB) *ShortLinkVisitModel {
	return &ShortLinkVisitModel{conn: conn}
}

const clickEventTable = "click_events"

// Insert 写入一条访问明细。created_at 由 ClickHouse 的 now() 默认值填充，
// 因此这里不显式传该列。调用方（rpc.Resolve）应以异步方式调用，避免阻塞跳转。
func (m *ShortLinkVisitModel) Insert(ctx context.Context, data *ShortLinkVisit) error {
	_, err := m.conn.ExecContext(ctx,
		"INSERT INTO "+clickEventTable+" (code, long_url, user_id, ip, referer, status, source, latency_ms) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		data.Code, data.LongURL, data.UserId, data.IP, data.Referer, data.Status, data.Source, data.LatencyMs)
	return err
}

// FindPageByUser 用户维度短链访问日志（仅本人短链），支持按 code / long_url 模糊搜索，
// source 可选（"web" / "rpc"），为空时不过滤。
//
// 性能优化：用 count() OVER () 窗口函数在同一条查询里同时拿到「满足条件的总行数」
// 与「当前分页数据」，把原本的 count 查询 + 数据查询两次往返合并为一次，
// 对跨公网的远程 ClickHouse 能显著减少延迟。
func (m *ShortLinkVisitModel) FindPageByUser(ctx context.Context, userId, page, pageSize int64, search, source string) ([]ShortLinkVisit, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	where := " WHERE user_id = ?"
	args := []interface{}{userId}
	if search != "" {
		where += " AND (code LIKE ? OR long_url LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	if source != "" {
		where += " AND source = ?"
		args = append(args, source)
	}

	offset := (page - 1) * pageSize
	// count() OVER () 在 LIMIT 之前计算，total 即为满足 WHERE 的总行数（每一行相同）。
	selCols := "code, long_url, user_id, ip, referer, status, source, latency_ms, created_at, count() OVER () AS total_count"
	query := "SELECT " + selCols + " FROM " + clickEventTable +
		where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	listArgs := append(append([]interface{}{}, args...), pageSize, offset)
	rows, err := m.conn.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]ShortLinkVisit, 0, pageSize)
	var total int64
	for rows.Next() {
		var v ShortLinkVisit
		if err := rows.Scan(&v.Code, &v.LongURL, &v.UserId, &v.IP, &v.Referer, &v.Status, &v.Source, &v.LatencyMs, &v.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}
	return items, total, nil
}

// CountByDay 统计最近 days 天每天的短链访问量（按创建日期分组，全局口径）
func (m *ShortLinkVisitModel) CountByDay(ctx context.Context, days int) (map[string]int64, error) {
	query := "SELECT toDate(created_at) AS day, count() AS value FROM " + clickEventTable +
		" WHERE created_at >= now() - INTERVAL ? DAY GROUP BY day"
	rows, err := m.conn.QueryContext(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]int64)
	for rows.Next() {
		var day time.Time
		var value int64
		if err := rows.Scan(&day, &value); err != nil {
			return nil, err
		}
		out[day.Format("2006-01-02")] = value
	}
	return out, nil
}
