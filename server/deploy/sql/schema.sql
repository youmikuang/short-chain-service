-- Short Chain Service — MySQL schema
-- Database: short_chain  (utf8mb4)
--
-- Apply with:
--   mysql -h 127.0.0.1 -u root -p short_chain < deploy/sql/schema.sql
-- or let docker-compose create the DB and run this file once.

CREATE DATABASE IF NOT EXISTS short_chain DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE short_chain;

-- ---------------------------------------------------------------------------
-- users: 账号体系（邮箱注册 / GitHub 绑定）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id            BIGINT       NOT NULL AUTO_INCREMENT,
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL DEFAULT '',
    nickname      VARCHAR(128) NOT NULL DEFAULT '',
    github_id     VARCHAR(64)  NOT NULL DEFAULT '',
    avatar        VARCHAR(512) NOT NULL DEFAULT '',
    status        TINYINT      NOT NULL DEFAULT 1,
    created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_email (email),
    UNIQUE KEY uk_github_id (github_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ---------------------------------------------------------------------------
-- api_keys: 第三方调用短链所需的 API Key（明文仅返回一次，入库存 sha256）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS api_keys (
    id         BIGINT       NOT NULL AUTO_INCREMENT,
    user_id    BIGINT       NOT NULL,
    name       VARCHAR(128) NOT NULL DEFAULT '',
    key_hash   VARCHAR(255) NOT NULL,
    prefix     VARCHAR(16)  NOT NULL DEFAULT '',
    quota      BIGINT       NOT NULL DEFAULT 0,
    used       BIGINT       NOT NULL DEFAULT 0,
    status     TINYINT      NOT NULL DEFAULT 1,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_user_id (user_id),
    KEY idx_key_hash (key_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ---------------------------------------------------------------------------
-- user_settings: 用户偏好（邮件通知 / 安全告警 / 营销通讯）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_settings (
    user_id         BIGINT NOT NULL,
    email_notif     TINYINT NOT NULL DEFAULT 1,
    security_alerts TINYINT NOT NULL DEFAULT 1,
    marketing_comm  TINYINT NOT NULL DEFAULT 0,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ---------------------------------------------------------------------------
-- short_links: 短链核心表（code 为短码，long_url 为长链）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS short_links (
    id         BIGINT       NOT NULL AUTO_INCREMENT,
    code       VARCHAR(32)  NOT NULL,
    long_url   TEXT         NOT NULL,
    user_id    BIGINT       NOT NULL DEFAULT 0,
    clicks     BIGINT       NOT NULL DEFAULT 0,
    status     TINYINT      NOT NULL DEFAULT 1,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_code (code),
    KEY idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ---------------------------------------------------------------------------
-- domain_blacklist: 域名黑名单（创建短链时命中则拒绝）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS domain_blacklist (
    id         BIGINT       NOT NULL AUTO_INCREMENT,
    domain     VARCHAR(255) NOT NULL,
    reason     VARCHAR(255) NOT NULL DEFAULT '',
    attempts   BIGINT       NOT NULL DEFAULT 0,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_domain (domain)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ---------------------------------------------------------------------------
-- access_logs: HTTP 访问日志（网关自身访问记录，仅供运维排查）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS access_logs (
    id         BIGINT       NOT NULL AUTO_INCREMENT,
    user_id    BIGINT       NOT NULL DEFAULT 0,
    method     VARCHAR(16)  NOT NULL DEFAULT '',
    endpoint   VARCHAR(255) NOT NULL DEFAULT '',
    status     INT          NOT NULL DEFAULT 0,
    latency_ms INT          NOT NULL DEFAULT 0,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_created_at (created_at),
    KEY idx_endpoint (endpoint)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
