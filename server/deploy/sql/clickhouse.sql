-- Short Chain Service — ClickHouse schema (analytics / click events)
--
-- ClickHouse is optional for the core flow. It stores immutable click events
-- for analytics. The short_links.clicks counter in MySQL is the source of
-- truth for the live click count; this table powers heavy analytical queries.
--
-- Apply with:
--   clickhouse-client --host 127.0.0.1 --port 9000 < deploy/sql/clickhouse.sql

CREATE DATABASE IF NOT EXISTS short_chain;
USE short_chain;

CREATE TABLE IF NOT EXISTS click_events (
    code        String,
    user_id     UInt64 DEFAULT 0,
    ip          String DEFAULT '',
    ua          String DEFAULT '',
    referer     String DEFAULT '',
    created_at  DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (code, created_at);
