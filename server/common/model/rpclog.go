package model

import (
	"context"
	"database/sql"
	"time"
)

// RpcLog 一条短链核心 gRPC 调用的记录（每次 RPC 方法被调用写入一条）。
// 数据落在 ClickHouse 的 rpc_logs 表（不再使用 MySQL）。
type RpcLog struct {
	Method    string
	UserId    int64
	Code      string
	Status    int64
	LatencyMs int64
	Error     string
	CreatedAt time.Time
}

type RpcLogModel struct {
	conn *sql.DB
}

func NewRpcLogModel(conn *sql.DB) *RpcLogModel {
	return &RpcLogModel{conn: conn}
}

const rpcLogTable = "rpc_logs"

// Insert 写入一条 RPC 调用日志。created_at 由 ClickHouse 的 now() 默认值填充，
// 因此这里不显式传该列。调用方（rpc 拦截器）应以异步方式调用，避免阻塞请求。
func (m *RpcLogModel) Insert(ctx context.Context, data *RpcLog) error {
	_, err := m.conn.ExecContext(ctx,
		"INSERT INTO "+rpcLogTable+" (method, user_id, code, status, latency_ms, error) VALUES (?, ?, ?, ?, ?, ?)",
		data.Method, data.UserId, data.Code, data.Status, data.LatencyMs, data.Error)
	return err
}

// CountByDay 统计最近 days 天每天生成的 RPC 调用日志数量（按创建日期分组）
func (m *RpcLogModel) CountByDay(ctx context.Context, days int) (map[string]int64, error) {
	query := "SELECT toDate(created_at) AS day, count() AS value FROM " + rpcLogTable +
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
