-- Short Chain Service — ClickHouse schema (analytics / click events)
--
-- click_events 存储短链被访问的明细记录，由 rpc.Resolve 在每次跳转时异步写入，
-- 供 /api/logs（按用户分页）与 /api/usage-trends（按天聚合）查询。
-- MySQL 中的 short_link.clicks 仍是实时点击数的权威来源；本表用于分析型查询。
--
-- Apply with:
--   clickhouse-client --host 127.0.0.1 --port 9000 < deploy/sql/clickhouse.sql

CREATE DATABASE IF NOT EXISTS short_chain;
USE short_chain;

CREATE TABLE IF NOT EXISTS click_events (
    code        String,
    long_url    String,
    user_id     UInt64 DEFAULT 0,
    ip          String DEFAULT '',
    referer     String DEFAULT '',
    status      UInt32 DEFAULT 200,
    created_at  DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (user_id, created_at);
