-- Short Chain Service — ClickHouse schema (analytics / click events)
--
-- click_events 存储短链被访问的明细记录，由 rpc.Resolve 在每次跳转时异步写入，
-- 供 /api/logs（按用户分页）与 /api/usage-trends（按天聚合）查询。
-- rpc_logs 存储短链核心 gRPC 服务的调用日志，由 rpc 拦截器异步写入，仅供运维排查。
-- MySQL 中的 short_link.clicks 仍是实时点击数的权威来源；本表用于分析型查询。
--
-- Apply with:
--   clickhouse-client --host 127.0.0.1 --port 9000 < deploy/sql/clickhouse.sql
--
-- 说明：
--   * 列备注使用 ClickHouse 的 COLUMN COMMENT 语法。
--   * 索引说明：ClickHouse 的 ORDER BY 即主键（稀疏索引），user_id 已置于 ORDER BY 首列，
--     天然支持按用户等值/范围过滤；其余高频过滤字段（code / long_url / method / status）
--     通过跳数索引（skipping index）加速。PARTITION BY 按月分区，配合 created_at 范围查询做剪枝。
--   * 跳数索引（INDEX ...）必须写在列定义的小括号内部、ENGINE 之前。
--   * 若表已存在（IF NOT EXISTS 不会改写），需手动 ALTER TABLE ... ADD INDEX 或重建表。

CREATE DATABASE IF NOT EXISTS short_chain;
USE short_chain;

CREATE TABLE IF NOT EXISTS click_events (
    code        String      COMMENT '短码（短链唯一标识，如 gh28a）',
    long_url    String      COMMENT '跳转目标长链',
    user_id     UInt64      DEFAULT 0 COMMENT '短链所属用户ID，用于按用户隔离',
    ip          String      DEFAULT '' COMMENT '访问者 IP',
    referer     String      DEFAULT '' COMMENT '来源 Referer',
    status      UInt32      DEFAULT 200 COMMENT 'HTTP 状态码（200 正常，其余为异常）',
    source      String      DEFAULT 'web' COMMENT '生成来源：web（网页 JWT）/ rpc（第三方 API Key 经核心服务）',
    latency_ms  UInt32      DEFAULT 0 COMMENT '解析耗时（毫秒），用于访问性能统计',
    created_at  DateTime    DEFAULT now() COMMENT '访问时间',
    -- 跳数索引：加速 /api/logs 中 code / long_url 的模糊搜索（LIKE '%x%'）
    INDEX idx_code      code      TYPE ngrambf_v1(3, 1024, 2, 0) GRANULARITY 1,
    INDEX idx_long_url  long_url  TYPE ngrambf_v1(3, 1024, 2, 0) GRANULARITY 1
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (user_id, created_at);

-- 兼容老表：若 click_events 在 source 列之前已创建，自动补列（幂等，可重复执行）。
-- 这样 logs 与 url（short_links）一致，都带有生成来源 source 字段。
ALTER TABLE IF EXISTS click_events
    ADD COLUMN IF NOT EXISTS source String DEFAULT 'web' COMMENT '生成来源：web（网页 JWT）/ rpc（第三方 API Key 经核心服务）';

-- 存量迁移：早期写入的 click_events.source='api' 统一改名为 'rpc'，
-- 与 short_links 及新建数据的命名保持一致（幂等：仅影响仍为 'api' 的行）。
-- ClickHouse 的 UPDATE 是异步 mutation，大表可能耗时，可在低峰期执行。
ALTER TABLE IF EXISTS click_events UPDATE source = 'rpc' WHERE source = 'api';

-- 兼容老表：补充解析耗时列（用于访问性能统计）。
ALTER TABLE IF EXISTS click_events
    ADD COLUMN IF NOT EXISTS latency_ms UInt32 DEFAULT 0 COMMENT '解析耗时（毫秒）';

-- ---------------------------------------------------------------------------
-- rpc_logs: 短链核心 gRPC 服务的调用日志（由 rpc 拦截器异步写入，仅供运维排查）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS rpc_logs (
    method      String      COMMENT 'RPC 方法全名，如 /slink.slink/Resolve',
    user_id     UInt64      DEFAULT 0 COMMENT '用户ID（内部服务调用暂记 0）',
    code        String      DEFAULT '' COMMENT '相关短码（如 Resolve 时携带，可选）',
    status      UInt32      DEFAULT 0 COMMENT '状态码（0=成功，非 0=errorx 错误码）',
    latency_ms  UInt32      DEFAULT 0 COMMENT '调用耗时（毫秒）',
    error       String      DEFAULT '' COMMENT '错误信息（成功时为空）',
    created_at  DateTime    DEFAULT now() COMMENT '调用时间',
    -- 跳数索引：加速按方法名等值过滤与按状态码过滤
    INDEX idx_method  method  TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_status  status  TYPE minmax GRANULARITY 1
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (user_id, created_at);
