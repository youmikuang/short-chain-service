-- Short Chain Service — 演示/种子数据
-- 作用于 admin 后台与 web 前端共用的 short_chain 库。
-- 在 schema.sql 建表之后执行：
--   mysql -h 127.0.0.1 -u root -p short_chain < deploy/sql/seed.sql
--
-- 说明：
--   * 带 UNIQUE 约束的表使用 INSERT IGNORE，可重复执行而不报错。
--   * access_logs 无唯一键，仅首次灌入即可（重复执行会产生重复日志）。
--   * 用户密码为占位哈希，admin 后台使用 admin-api.yaml 中的 Admin 凭据登录，不依赖本表。

-- ---------------------------------------------------------------------------
-- 迁移：补齐新增列（兼容已在运行的旧库）
-- schema.sql 使用 CREATE TABLE IF NOT EXISTS，对已存在的表不会追加列，
-- 这里用 information_schema 判断后按需 ALTER，可重复执行且不报错。
-- ---------------------------------------------------------------------------
SET @s = (SELECT IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'api_keys' AND COLUMN_NAME = 'quota') = 0,
  'ALTER TABLE api_keys ADD COLUMN quota BIGINT NOT NULL DEFAULT 0',
  'SELECT 1'));
PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @s = (SELECT IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'api_keys' AND COLUMN_NAME = 'used') = 0,
  'ALTER TABLE api_keys ADD COLUMN used BIGINT NOT NULL DEFAULT 0',
  'SELECT 1'));
PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @s = (SELECT IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'domain_blacklist' AND COLUMN_NAME = 'attempts') = 0,
  'ALTER TABLE domain_blacklist ADD COLUMN attempts BIGINT NOT NULL DEFAULT 0',
  'SELECT 1'));
PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ---------------------------------------------------------------------------
-- users（账号体系）
-- ---------------------------------------------------------------------------
INSERT IGNORE INTO users (id, email, password_hash, nickname, github_id, avatar, status, created_at) VALUES
-- 以下账号密码统一为 password123（bcrypt 哈希由 golang.org/x/crypto/bcrypt 生成）
(1, 'john@enterprise.co', '$2a$10$Ql5OF34j4tJbwzi.lrmJj.j5oYEQWeKmGlmfsMBHqErWwuYSKTnLe', 'John Doe',       NULL, '', 1, NOW() - INTERVAL 120 DAY),
(2, 'sarah@infra-tech.io', '$2a$10$Ql5OF34j4tJbwzi.lrmJj.j5oYEQWeKmGlmfsMBHqErWwuYSKTnLe', 'Sarah Al-Farsi', NULL, '', 1, NOW() - INTERVAL 110 DAY),
(3, 'marcus@shadow.dev',   '$2a$10$Ql5OF34j4tJbwzi.lrmJj.j5oYEQWeKmGlmfsMBHqErWwuYSKTnLe', 'Marcus K.',      NULL, '', 1, NOW() - INTERVAL 90 DAY),
(4, 'api@datalabs.com',    '$2a$10$Ql5OF34j4tJbwzi.lrmJj.j5oYEQWeKmGlmfsMBHqErWwuYSKTnLe', 'Data Labs Inc',  NULL, '', 1, NOW() - INTERVAL 80 DAY),
(5, 'emma@growth.io',      '$2a$10$Ql5OF34j4tJbwzi.lrmJj.j5oYEQWeKmGlmfsMBHqErWwuYSKTnLe', 'Emma Watson',    NULL, '', 1, NOW() - INTERVAL 60 DAY),
(6, 'liwei@corp.cn',       '$2a$10$Ql5OF34j4tJbwzi.lrmJj.j5oYEQWeKmGlmfsMBHqErWwuYSKTnLe', 'Li Wei',        NULL, '', 1, NOW() - INTERVAL 40 DAY);

-- ---------------------------------------------------------------------------
-- short_links（短链核心表；status: 1=Active, 2=Flagged, 0=Expired）
-- ---------------------------------------------------------------------------
INSERT IGNORE INTO short_links (code, long_url, user_id, clicks, status, created_at) VALUES
('gh28a', 'https://github.com/enterprise/project-alpha-deployment-v2',        1, 12402, 1, NOW() - INTERVAL 30 DAY),
('ns88z', 'https://notion.so/workspace/design-specs-2025-v1',                  2,   842, 2, NOW() - INTERVAL 29 DAY),
('md92x', 'https://medium.com/@ryankim/how-to-scale-infrastructure-quickly',  3,  2190, 1, NOW() - INTERVAL 28 DAY),
('aws42', 'https://aws.amazon.com/console/s3/buckets/backup-001',              4, 15002, 1, NOW() - INTERVAL 27 DAY),
('tw10k', 'https://twitter.com/devnews/status/1234567890',                    5,   530, 1, NOW() - INTERVAL 20 DAY),
('yt55p', 'https://youtube.com/watch?v=shortchain-demo-2025',                  6,  9800, 1, NOW() - INTERVAL 18 DAY),
('fb33m', 'https://facebook.com/share/launch-post',                           1,     0, 0, NOW() - INTERVAL 15 DAY),
('ig77c', 'https://instagram.com/p/summer-campaign',                          2,   312, 1, NOW() - INTERVAL 12 DAY),
('lk21d', 'https://linkedin.com/posts/infra-automation',                      3,  1450, 1, NOW() - INTERVAL 10 DAY),
('rd09q', 'https://reddit.com/r/selfhosted/comments/abc',                     4,   220, 2, NOW() - INTERVAL 8 DAY),
('st44n', 'https://stackoverflow.com/questions/12345/vue3-tips',              5,   670, 1, NOW() - INTERVAL 5 DAY),
('doc8h', 'https://docs.shortchain.io/guides/getting-started',                6,  4300, 1, NOW() - INTERVAL 3 DAY);

-- ---------------------------------------------------------------------------
-- api_keys（status: 1=Active, 0=Revoked；quota/used 为月度额度）
-- key_hash 为占位 sha256（64 位十六进制），prefix 为前 5 位
-- ---------------------------------------------------------------------------
INSERT IGNORE INTO api_keys (user_id, name, key_hash, prefix, quota, used, status, created_at) VALUES
(1, 'prod-token', 'a1b2c3d4e5f60718293a4b5c6d7e8f9011223344556677889900aabbccddeeff00', 'a1b2c',   50000,  14000, 1, NOW() - INTERVAL 100 DAY),
(2, 'enterprise', 'b2c3d4e5f60718293a4b5c6d7e8f9011223344556677889900aabbccddeeff0011', 'b2c3d', 1000000,1000000, 0, NOW() - INTERVAL 95 DAY),
(3, 'shadow',     'c3d4e5f60718293a4b5c6d7e8f9011223344556677889900aabbccddeeff001122', 'c3d4e',   10000,   4500, 0, NOW() - INTERVAL 80 DAY),
(4, 'datalabs',   'd4e5f60718293a4b5c6d7e8f9011223344556677889900aabbccddeeff00112233', 'd4e5f',  250000,  29850, 1, NOW() - INTERVAL 70 DAY),
(5, 'growth',     'e5f60718293a4b5c6d7e8f9011223344556677889900aabbccddeeff0011223344', 'e5f60',   50000,  12000, 1, NOW() - INTERVAL 50 DAY),
(6, 'corp',       'f60718293a4b5c6d7e8f9011223344556677889900aabbccddeeff001122334455', 'f6071',  120000,   3000, 1, NOW() - INTERVAL 35 DAY),
(1, 'ci-bot',     '0718293a4b5c6d7e8f9011223344556677889900aabbccddeeff00112233445566', '07182',  200000,  88000, 1, NOW() - INTERVAL 25 DAY),
(2, 'analytics',  '18293a4b5c6d7e8f9011223344556677889900aabbccddeeff0011223344556677', '18293',   80000,      0, 1, NOW() - INTERVAL 15 DAY),
(3, 'legacy',     '293a4b5c6d7e8f9011223344556677889900aabbccddeeff001122334455667788', '293a4',   30000,  29500, 1, NOW() - INTERVAL 10 DAY),
(4, 'mobile',     '3a4b5c6d7e8f9011223344556677889900aabbccddeeff00112233445566778899', '3a4b5',   60000,  41000, 1, NOW() - INTERVAL 6 DAY),
(5, 'webhook',    '4b5c6d7e8f9011223344556677889900aabbccddeeff0011223344556677889900', '4b5c6',   40000,   1200, 1, NOW() - INTERVAL 3 DAY),
(6, 'test',       '5c6d7e8f9011223344556677889900aabbccddeeff0011223344556677889900aa', '5c6d7',   10000,      0, 0, NOW() - INTERVAL 1 DAY);

-- ---------------------------------------------------------------------------
-- domain_blacklist（域名黑名单；attempts 为命中拦截次数）
-- ---------------------------------------------------------------------------
INSERT IGNORE INTO domain_blacklist (domain, reason, attempts, created_at) VALUES
('secure-login-update.cc', 'Phishing', 42901, NOW() - INTERVAL 260 DAY),
('spam-generator.net',     'Spam',    11482, NOW() - INTERVAL 240 DAY),
('malware-drop-site.org',  'Malware',  2910, NOW() - INTERVAL 190 DAY),
('free-tokens-now.biz',     'Phishing',  8554, NOW() - INTERVAL 150 DAY),
('crypto-airdrop-scam.io',  'Phishing', 12330, NOW() - INTERVAL 120 DAY),
('fake-invoice-pro.com',    'Phishing',  6720, NOW() - INTERVAL 90 DAY),
('clickbait-news.xyz',      'Spam',      3400, NOW() - INTERVAL 60 DAY),
('trojan-download.win',     'Malware',    980, NOW() - INTERVAL 40 DAY),
('survey-pay-me.ru',        'Spam',      2100, NOW() - INTERVAL 20 DAY),
('login-verify-now.cc',     'Phishing', 15600, NOW() - INTERVAL 5 DAY);

-- ---------------------------------------------------------------------------
-- access_logs（近 7 天访问日志，驱动 dashboard 流量趋势图）
-- ---------------------------------------------------------------------------
INSERT INTO access_logs (user_id, method, endpoint, status, latency_ms, created_at) VALUES
-- 6 天前
(1, 'GET',  '/api/resolve',   200, 12, NOW() - INTERVAL 6 DAY),
(2, 'POST', '/api/shorten',   200, 31, NOW() - INTERVAL 6 DAY),
(3, 'GET',  '/api/resolve',   200, 18, NOW() - INTERVAL 6 DAY),
-- 5 天前
(4, 'GET',  '/api/stats',     200, 22, NOW() - INTERVAL 5 DAY),
(5, 'POST', '/api/shorten',   200, 40, NOW() - INTERVAL 5 DAY),
(1, 'GET',  '/api/resolve',   200, 9,  NOW() - INTERVAL 5 DAY),
(6, 'GET',  '/api/resolve',   429, 5,  NOW() - INTERVAL 5 DAY),
(2, 'GET',  '/api/stats',     200, 27, NOW() - INTERVAL 5 DAY),
-- 4 天前
(3, 'GET',  '/api/resolve',   200, 14, NOW() - INTERVAL 4 DAY),
(4, 'POST', '/api/shorten',   200, 33, NOW() - INTERVAL 4 DAY),
-- 3 天前
(5, 'GET',  '/api/resolve',   200, 11, NOW() - INTERVAL 3 DAY),
(6, 'GET',  '/api/stats',     200, 19, NOW() - INTERVAL 3 DAY),
(1, 'POST', '/api/shorten',   200, 38, NOW() - INTERVAL 3 DAY),
(2, 'GET',  '/api/resolve',   200, 8,  NOW() - INTERVAL 3 DAY),
(3, 'GET',  '/api/resolve',   404, 6,  NOW() - INTERVAL 3 DAY),
(4, 'GET',  '/api/stats',     200, 24, NOW() - INTERVAL 3 DAY),
(5, 'GET',  '/api/resolve',   200, 13, NOW() - INTERVAL 3 DAY),
-- 2 天前
(6, 'POST', '/api/shorten',   200, 29, NOW() - INTERVAL 2 DAY),
(1, 'GET',  '/api/resolve',   200, 10, NOW() - INTERVAL 2 DAY),
(2, 'GET',  '/api/stats',     200, 21, NOW() - INTERVAL 2 DAY),
(3, 'GET',  '/api/resolve',   200, 15, NOW() - INTERVAL 2 DAY),
(4, 'GET',  '/api/resolve',   429, 4,  NOW() - INTERVAL 2 DAY),
-- 1 天前
(5, 'GET',  '/api/resolve',   200, 12, NOW() - INTERVAL 1 DAY),
(6, 'POST', '/api/shorten',   200, 35, NOW() - INTERVAL 1 DAY),
(1, 'GET',  '/api/stats',     200, 23, NOW() - INTERVAL 1 DAY),
(2, 'GET',  '/api/resolve',   200, 9,  NOW() - INTERVAL 1 DAY),
(3, 'GET',  '/api/resolve',   200, 17, NOW() - INTERVAL 1 DAY),
(4, 'GET',  '/api/resolve',   200, 11, NOW() - INTERVAL 1 DAY),
-- 今天
(5, 'POST', '/api/shorten',   200, 31, NOW()),
(6, 'GET',  '/api/resolve',   200, 8,  NOW()),
(1, 'GET',  '/api/stats',     200, 20, NOW()),
(2, 'GET',  '/api/resolve',   200, 14, NOW()),
(3, 'GET',  '/api/resolve',   200, 10, NOW()),
(4, 'GET',  '/api/resolve',   200, 12, NOW()),
(5, 'GET',  '/api/resolve',   429, 5,  NOW()),
(6, 'GET',  '/api/stats',     200, 18, NOW());
