# MySQL 表结构

> 短链服务（go-zero）业务库表设计
> 更新时间：2026-07-15
> 字符集统一：`utf8mb4` / 排序规则 `utf8mb4_general_ci`
> 引擎统一：`InnoDB`

业务持久化全部落在 MySQL；Redis 仅做短链缓存 / 黑名单 Set / 点击计数，ClickHouse 仅做访问分析（见文末）。

可执行建表脚本见 `server/deploy/sql/schema.sql`（含 `short_chain` 库与全部 6 张表），ClickHouse DDL 见 `server/deploy/sql/clickhouse.sql`。

---

## 1. users —— 用户表

用户体系（邮箱注册 / GitHub OAuth）的本地账户，JWT 签发与密码校验都基于此表。

```sql
CREATE TABLE `users` (
  `id`            BIGINT       NOT NULL AUTO_INCREMENT,
  `email`         VARCHAR(255) NOT NULL,
  `password_hash` VARCHAR(255) NOT NULL DEFAULT '',
  `nickname`      VARCHAR(128) NOT NULL DEFAULT '',
  `github_id`     VARCHAR(64)  NOT NULL DEFAULT '',
  `avatar`        VARCHAR(512) NOT NULL DEFAULT '',
  `status`        TINYINT      NOT NULL DEFAULT 1 COMMENT '1=正常 0=禁用',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_email` (`email`),
  UNIQUE KEY `uk_github_id` (`github_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
```

对应接口：`POST /api/auth/register`、`POST /api/auth/login`、`GET /api/auth/github/callback`、`GET /api/profile`、`POST /api/profile`、`POST /api/profile/password`。

---

## 2. api_keys —— API Key 表

开放 API 调用凭证。明文 key 仅创建时返回一次（`slk_` 前缀 + 32 位随机串），库存 `key_hash`（SHA-256）；`prefix` 为展示用前缀。`X-API-Key` 中间件据此校验。

```sql
CREATE TABLE `api_keys` (
  `id`         BIGINT       NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT       NOT NULL,
  `name`       VARCHAR(128) NOT NULL DEFAULT '',
  `key_hash`   VARCHAR(255) NOT NULL,
  `prefix`     VARCHAR(16)  NOT NULL DEFAULT '',
  `status`     TINYINT      NOT NULL DEFAULT 1 COMMENT '1=启用 0=吊销',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key_hash` (`key_hash`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API Key 表';
```

对应接口：`POST /api/keys`、`GET /api/keys`、`DELETE /api/keys/:id`。

---

## 3. short_links —— 短链表

短码生成（Snowflake + Base62）后落库；跳转时先查 Redis 缓存，未命中回源此表。`clicks` 为累计点击（Redis 实时 incr 并同步回写 MySQL）。

```sql
CREATE TABLE `short_links` (
  `id`         BIGINT       NOT NULL AUTO_INCREMENT,
  `code`       VARCHAR(32)  NOT NULL COMMENT '短码（Base62）',
  `long_url`   TEXT         NOT NULL,
  `user_id`    BIGINT       NOT NULL DEFAULT 0,
  `status`     TINYINT      NOT NULL DEFAULT 1 COMMENT '1=正常 0=禁用',
  `clicks`     BIGINT       NOT NULL DEFAULT 0,
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='短链表';
```

对应接口：`POST /api/short-links`、`GET /api/short-links/:code`、`GET /r/:code`、`GET /admin/api/links`、`DELETE`（rpc 核心）。

---

## 4. domain_blacklist —— 域名黑名单表

创建/跳转前校验长链域名，命中则拦截（不跳转）。MySQL 为来源，Redis Set（`domain:blacklist`）为热数据副本。

```sql
CREATE TABLE `domain_blacklist` (
  `id`         BIGINT       NOT NULL AUTO_INCREMENT,
  `domain`     VARCHAR(255) NOT NULL,
  `reason`     VARCHAR(255) NOT NULL DEFAULT '',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_domain` (`domain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='域名黑名单表';
```

对应接口：`POST /admin/api/blacklist`。

---

## 5. user_settings —— 用户偏好表

支撑 Web 端 `GET/PUT /api/settings`（通知偏好），以 `user_id` 为主键，写入采用 `INSERT ... ON DUPLICATE KEY UPDATE`。

```sql
CREATE TABLE `user_settings` (
  `user_id`         BIGINT   NOT NULL,
  `email_notif`     TINYINT  NOT NULL DEFAULT 1 COMMENT '邮件周报',
  `security_alerts` TINYINT  NOT NULL DEFAULT 1 COMMENT '安全告警',
  `marketing_comm`  TINYINT  NOT NULL DEFAULT 0 COMMENT '营销通信',
  `created_at`      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户偏好表';
```

对应接口：`GET /api/settings`、`PUT /api/settings`。

---

## 6. access_logs —— 访问日志表

HTTP 网关全局中间件记录每次请求，驱动 Web 端 `GET /api/logs` 与 `GET /api/usage-trends`（近 30 天每日请求量）。`/api/logs`、`/api/usage-trends` 自身不写日志，避免自递归。

```sql
CREATE TABLE `access_logs` (
  `id`         BIGINT       NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT       NOT NULL DEFAULT 0,
  `method`     VARCHAR(16)  NOT NULL DEFAULT '',
  `endpoint`   VARCHAR(255) NOT NULL DEFAULT '' COMMENT '请求路径',
  `status`     INT          NOT NULL DEFAULT 0 COMMENT 'HTTP 状态码',
  `latency_ms` INT          NOT NULL DEFAULT 0 COMMENT '耗时（毫秒）',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_endpoint` (`endpoint`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='访问日志表';
```

对应接口：`GET /api/logs`、`GET /api/usage-trends`。

> 说明：高吞吐场景下访问日志建议走 Kafka → ClickHouse，MySQL 仅保留管理后台的轻量查询；此处表结构用于管理后台的实时预览。

---

## 7. 关系总览

```
users 1 ──< api_keys      (user_id)
users 1 ──< short_links    (user_id)
users 1 ──< user_settings  (user_id, 1:1)
users 1 ──< access_logs    (user_id)
(无外键约束，user_id 为逻辑关联，便于分库与软删除)
```

---

## 8. ClickHouse（非 MySQL，仅作分析，供参考）

访问次数/分析统计落 ClickHouse，不在 MySQL 中：

```sql
-- ClickHouse
CREATE TABLE click_events (
  code        String,
  user_id     UInt64 DEFAULT 0,
  ip          String DEFAULT '',
  ua          String DEFAULT '',
  referer     String DEFAULT '',
  created_at  DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (code, created_at);
```

建表脚本：`server/deploy/sql/clickhouse.sql`。
