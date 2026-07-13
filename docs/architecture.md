# 短链服务基础架构设计

> 版本：v2（无 K8s，Nginx 负载均衡 + 开放 API + 后台管理）
> 更新时间：2026-07-13

---

## 一、架构总览

```
【 用户 / 浏览器 】
        │
        ▼
【 Nginx 负载均衡 + TLS终止 + 限流/WAF 】
   └─ 路由:
      /r/{code}  → 短链跳转服务
      /api/*     → 业务 API(注册/OAuth/Key/短链CRUD/黑名单)
      /admin/*   → Vue3 管理后台(静态托管)
      /          → Vue3 官网(注册/登录/授权/申请Key)
        │
        ▼
【 Go 服务多实例 (Nginx 轮询/最少连接, 固定实例) 】
   ├─ 跳转 /r/{code}:
   │     Redis缓存 → 命中302 + 发Kafka(点击事件)
   │                → 未命中 SingleFlight 回源 + 布隆防穿透
   │                → 跳转前校验 域名黑名单(Redis Set) → 命中则拦截
   ├─ 业务API:
   │     邮箱注册/登录 · GitHub OAuth · API Key 申请与校验(按Key限流)
   │     短链 CRUD · 域名黑名单管理(admin用)
   └─ Kafka 消费者: 点击事件 → ClickHouse(统计) + 聚合计数
        │
        ▼
【 Kafka 】  ◄── 点击流 / 审计事件(削峰解耦)
   ├─► ClickHouse(单机Docker)  ← 访问次数/分析统计
   └─► (可选)落库聚合
【 MySQL 】  ← 用户/API Key/短链/黑名单 持久化
【 Redis 】  ← 跳转缓存 + 黑名单Set + 限流计数 + 实时点击数
        │
        ▼
【 web/  Vue3 官网 】   【 admin/ Vue3 管理后台 】
        │
        ▼
【 sdk/go  Go 调用包 】  ← 第三方带 Key 鉴权调用短链 API
```

---

## 二、技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go 1.26 · Gin/Echo · go-redis · GORM/sqlc · sarama(Kafka) · clickhouse-go |
| 存储 | MySQL(业务) · Redis(缓存/黑名单/限流) · ClickHouse(统计) · Kafka(消息) |
| 前端 | Vue3 + Vite + Pinia + Vue Router · Element Plus(admin) / Naive UI |
| 鉴权 | JWT(会话) · GitHub OAuth2 · 邮箱 SMTP 验证 · API Key(Bearer) |
| 部署 | Docker Compose 一键起 MySQL/Redis/Kafka/ClickHouse/Nginx |

---

## 三、目录划分

```
short-chain-service/
├── api/              # Go 后端(跳转+业务API+Kafka消费者)
│   ├── cmd/server/   # 启动入口
│   ├── internal/
│   │   ├── router/
│   │   ├── handler/      # 跳转/注册/OAuth/Key/短链/黑名单
│   │   ├── service/      # 短链/用户/Key/统计/黑名单
│   │   ├── idgen/        # 短码生成(Snowflake, 固定实例分配workerId)
│   │   ├── mq/           # Kafka 生产/消费
│   │   ├── repo/         # MySQL/Redis/ClickHouse 访问
│   │   └── middleware/   # API Key 鉴权 + 限流
│   └── configs/
├── web/              # Vue3 官网: 邮箱注册/登录/GitHub授权/申请Key
├── admin/            # Vue3 管理后台: 链接管理/访问次数/域名黑名单
├── sdk/go/           # 可发布的 Go SDK(独立 go.mod, 供 import)
├── docs/             # 架构 & API 文档
├── deploy/           # docker-compose + nginx.conf
└── go.mod
```

---

## 四、核心数据流

### 1. 短链跳转 + 黑名单拦截
- 查 Redis 缓存 → 命中直接 302；未命中 SingleFlight 回源 MySQL 并回填（随机 TTL 防雪崩）。
- 回源/跳转前：从 `long_url` 提取域名，`SISMEMBER blacklist` 校验，**命中则返回拦截页（不跳转）**，防钓鱼/滥用。
- 每次跳转发 Kafka 点击事件。

### 2. 访问次数统计
- 实时：`Redis INCR short_link:{code}:clicks`（接口/后台即时展示）。
- 历史分析：Kafka → 消费者写 ClickHouse（`click_events` 表，按 code/时间聚合）。

### 3. 鉴权与 Key 调用
- 邮箱注册（SMTP 验证码）→ 登录发 JWT；GitHub OAuth 回调建号/绑定。
- 登录后申请 API Key（仅展示一次，库存 hash）。
- 调用 API 必须带 `X-API-Key`，中间件校验 + 按 Key 限流。

### 4. 短码生成（Nginx 固定实例下的解法）
- 不再有 K8s 动态扩缩，**每个实例固定分配 workerId**（环境变量/配置），Snowflake 可直接用；再 Base62 编码 + 混淆防遍历。
- 若以后想更短，可切 Leaf-segment 号段模式。

---

## 五、关键数据模型（初版）

- `users(id, email unique, password_hash, github_id, nickname, status)`
- `api_keys(id, user_id, key_hash, name, rate_limit, status)`
- `short_links(id, code unique, long_url, domain, user_id, api_key_id, status)`
- `domain_blacklist(id, domain, reason, creator_id)`
- ClickHouse `click_events(code, user_id, ip, ua, referer, event_time)`

---

## 六、与上一版的关键变化

1. **去掉 K8s** → Nginx 多实例负载均衡，Snowflake workerId 改为固定分配。
2. **新增账号体系**（邮箱 + GitHub OAuth + API Key），短链变为"需 Key 调用"的开放服务。
3. **新增域名黑名单**，在跳转前拦截，admin 可管理。
4. **新增访问统计**，Kafka → ClickHouse（单机 Docker）做分析，Redis 做实时计数。
5. **新增三套前端/包**：`web` 官网、`admin` 管理后台（均 Vue3）、`sdk/go` 可发布 Go 包。

---

## 七、落地建议顺序

1. `deploy/docker-compose.yml` + `nginx.conf`：先把依赖（MySQL/Redis/Kafka/ClickHouse）跑起来。
2. `api/` 骨架：路由 + 短链生成跳转 + Redis + 黑名单最小可用。
3. `web/` 与 `admin/` 的 Vue3 脚手架。
4. `sdk/go` 包：封装带 Key 鉴权的调用客户端。
