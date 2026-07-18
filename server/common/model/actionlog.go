package model

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ActionLog struct {
	Id        int64  `db:"id"`
	UserId    int64  `db:"user_id"`
	Method    string `db:"method"`
	Endpoint  string `db:"endpoint"`
	Status    int64  `db:"status"`
	LatencyMs int64  `db:"latency_ms"`
	CreatedAt string `db:"created_at"`
}

type ActionLogModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewActionLogModel(conn sqlx.SqlConn) *ActionLogModel {
	return &ActionLogModel{conn: conn, table: "`action_logs`"}
}

const actionLogRows = "id, user_id, method, endpoint, status, latency_ms, created_at"

func (m *ActionLogModel) Insert(ctx context.Context, data *ActionLog) (sql.Result, error) {
	query := "insert into " + m.table + " (user_id, method, endpoint, status, latency_ms) values (?, ?, ?, ?, ?)"
	return m.conn.Exec(query, data.UserId, data.Method, data.Endpoint, data.Status, data.LatencyMs)
}

// FindPage 分页查询操作日志，支持按 endpoint 模糊搜索
func (m *ActionLogModel) FindPage(ctx context.Context, search string, page, pageSize int64) ([]ActionLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	where := ""
	args := []interface{}{}
	if search != "" {
		where = " where endpoint like ? "
		args = append(args, "%"+search+"%")
	}

	var total int64
	if err := m.conn.QueryRow(&total, "select count(*) from "+m.table+where, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := "select " + actionLogRows + " from " + m.table + where + " order by id desc limit ? offset ?"
	listArgs := append(append([]interface{}{}, args...), pageSize, offset)
	var items []ActionLog
	if err := m.conn.QueryRows(&items, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// CountByDay 统计最近 days 天每天的操作量，返回存在数据的日期
func (m *ActionLogModel) CountByDay(ctx context.Context, days int) (map[string]int64, error) {
	query := "select date(created_at) as day, count(*) as value from " + m.table +
		" where created_at >= date_sub(curdate(), interval ? day) group by day"
	var rows []struct {
		Day   string `db:"day"`
		Value int64  `db:"value"`
	}
	if err := m.conn.QueryRows(&rows, query, days); err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, r := range rows {
		out[r.Day] = r.Value
	}
	return out, nil
}
