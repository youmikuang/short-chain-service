# 短链服务架构文档

---

## 一、架构总览

```
【 用户 / 浏览器 】
        │
        ▼
【 Nginx（生产）/ Vite 代理（开发） 】
   · 开发：web(5173) / admin-web(5174) 经 Vite proxy 转发到后端
   · 生产：Nginx 做 TLS 终止 + 路由 + 限流
   ├─ /api/*     → apps/api        (HTTP 网关，:8888)
   ├─ /admin/*   → apps/admin      (管理后台 API，:8889)
   ├─ /r/{code}  → apps/api        (短链跳转，:8888，再调 rpc)
   └─ /          → web / admin-web 静态资源
        │
        ▼
【 Go 服务（go-zero 单体仓库 server/，共享一个 go.mod）】
   ├─ apps/api    (HTTP 网关 :8888)
   │     · 用户体系：注册/登录/GitHub OAuth/API Key/资料/设置/改密（本地逻辑，无 gRPC）
   │     · 短链 CRUD / 跳转：经鉴权后内部 gRPC 调用 apps/rpc
   │     · 日志/用量：查 ClickHouse `click_events`（按用户隔离）
   ├─ apps/admin  (管理后台 API :8889)
   │     · 登录 / 仪表盘 / 链接管理 / 域名黑名单 / Token 管理
   └─ apps/rpc    (短链核心 gRPC :8081，仅内网，不对外暴露)
         · slink 服务：生成/查询/批量/删除/解析跳转
         · Snowflake + Base62 短码；Redis 缓存；域名黑名单校验
        │
        ▼
【 存储 】
   · MySQL    ：用户 / API Key / 用户设置 / 短链 / 黑名单 / handler 操作日志(`action_logs`)
   · Redis    ：跳转缓存(short_link:{code}) + 黑名单 Set + 实时点击计数
   · Kafka    ：[规划中] 点击流异步通道（配置已就位，业务未接入）
   · ClickHouse：短链访问明细 `click_events`（已接入，rpc.Resolve 异步写入，/api/logs 与 /api/usage-trends 直接查询）+ RPC 调用日志 `rpc_logs`（rpc 拦截器异步写入）
```

> 设计取舍
> - 用户体系（注册/登录/OAuth/Key/资料/设置）与"会话/凭证"强相关，放在 `apps/api` 网关本地处理，延迟最低、无需跨进程。
> - 短链的生成/解析/删除等核心逻辑统一在 `apps/rpc`，HTTP 与（未来的）内部 gRPC 调用都落到同一实现，避免逻辑分叉。

---

## 二、服务与职责

### apps/api（HTTP 网关，:8888）
| 能力 | 说明 |
|---|---|
| 用户体系 | 注册、登录、GitHub OAuth、API Key 管理、资料、偏好设置、改密 —— 纯 HTTP，本地落库 |
| 短链 CRUD | `POST /api/short-links`、`GET /api/short-links/:code`、`GET /api/short-links`（我的列表）等，鉴权后调 rpc |
| 跳转 | `GET /r/:code` → 调 rpc `Resolve`，命中黑名单由 handler 拦截 |
| 日志/用量 | `GET /api/logs`、`GET /api/usage-trends`，查 ClickHouse `click_events` |

### apps/admin（管理后台 API，:8889）
登录、仪表盘、链接管理（`listlinks`）、域名黑名单（`listblacklist`）、后台 Token 管理（`listtokens`/`provisiontoken`/`revoketoken`）。供 `admin-web` 前端调用。

### apps/rpc（短链核心 gRPC，:8081）
`slink` 服务，方法见 proto：
- `Createslink`：黑名单校验 → Snowflake+Base62 生成 → 落 MySQL + 写 Redis 缓存
- `GetByCode`：按 code 查询
- `BatchCreate`：批量生成
- `Deleteslink`：删除
- `Resolve`：跳转解析（Redis→MySQL 回源、域名黑名单校验、实时计数、写访问明细）

---

## 三、技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go · go-zero（rest + zrpc）· go-redis · sqlx · 内置限流/熔断/中间件 |
| 存储 | MySQL（业务）· Redis（缓存/黑名单/计数）· ClickHouse（访问明细 `click_events`）· Kafka（[规划中] 异步管道） |
| 前端 | Vue3 + Vite + Pinia + Vue Router；`web/`（官网）、`admin-web/`（管理后台） |
| 鉴权 | JWT（会话）· GitHub OAuth2 · API Key（X-API-Key） |
| 部署 | Docker Compose 起 MySQL/Redis/Kafka/ClickHouse/Nginx；开发用 Vite 代理 |

---

## 四、目录结构（实际）

```
short-chain-service/
├── server/                  # Go 后端 · go-zero 单体仓库（一个 go.mod）
│   ├── apps/
│   │   ├── api/             # HTTP 网关 (:8888)
│   │   │   ├── api/         # .api 描述 + 生成代码
│   │   │   ├── internal/    # handler / logic / middleware / svc / types
│   │   │   └── etc/api-api.yaml
│   │   ├── admin/           # 管理后台 API (:8889)
│   │   │   ├── api/  internal/  etc/
│   │   └── rpc/             # 短链核心 gRPC (:8081)
│   │       ├── pb/          # slink.proto + 生成
│   │       ├── internal/    # logic(rpc 核心) / server / svc
│   │       └── etc/slink.yaml
│   ├── common/              # 跨模块公共代码（避免循环依赖）
│   │   ├── ctxdata/         # JWT / Context 上下文解析
│   │   ├── errorx/          # 统一错误码
│   │   ├── interceptors/    # gRPC 拦截器 / API 中间件
│   │   ├── model/           # 数据模型：MySQL(sqlx) user/apikey/usersettings/
│   │   │                   #   slink/domainblacklist/actionlog；
│   │   │                   #   ClickHouse slink_visit(click_events)/rpclog(rpc_logs)
│   │   ├── tool/            # 工具类（加密/Base62/Snowflake/域名提取）
│   │   └── xfilters/        # 通用过滤器
│   └── deploy/              # 部署配置
│       ├── docker/          # Dockerfile
│       ├── k8s/             # K8s yaml（可选）
│       ├── sql/             # schema.sql（MySQL）/ clickhouse.sql
│       └── prometheus/
├── web/                     # Vue3 官网（邮箱注册/登录/GitHub 授权/申请 Key）
├── admin-web/               # Vue3 管理后台前端（构建产物由 apps/admin 静态托管）
├── docs/                    # 架构 & API 文档
└── README.md
```

> 说明：`sdk/go`（可发布 Go SDK）当前**尚未创建**；Nginx 生产配置为规划项，开发期由 Vite proxy 转发。

---

## 五、鉴权模型

两套机制并存，由 `apps/api/internal/handler/register.go` 路由分组控制：

- **API Key 中间件**（`middleware/apikey.go`）：对所有 `/api/*` 强制要求 `X-API-Key`，校验 `sha256(key)` 命中 `api_keys` 表且 `status=1`，通过后将 `user_id` 写入 context 供 logic 使用。
- **JWT 中间件**（`rest.WithJwt`）：保护"用户自己的资源"路由（资料、Key 管理、设置、日志等）。

跳过表 `apiKeySkipPaths`：仅用 JWT 鉴权、不需要 API Key 的路由，API Key 中间件直接放行，交由 JWT 中间件处理。当前包含：
`GET /api/short-links`、`GET|POST /api/keys`、`DELETE /api/keys/:id`、`GET|POST /api/profile`、`POST /api/profile/password`、`GET|PUT /api/settings`、`GET /api/usage-trends`、`GET /api/logs`。

> 注意：短链**创建**（`POST /api/short-links`）需 API Key；**列出我的短链**（`GET /api/short-links`）走 JWT。

---

## 六、核心数据流

### 1. 短链创建
```
POST /api/short-links (X-API-Key 或 JWT)
  → api.CreateslinkLogic
  → rpc.Createslink
       · 提取域名 → 命中 domain_blacklist（MySQL 优先，回退 Redis Set）则拒绝
       · 短码 = Base62(Snowflake.NextID())
       · 落 MySQL(short_links) + 写 Redis(short_link:{code} → long_url，随机 TTL 防雪崩)
  → 返回 { code, long_url }
```

### 2. 短链跳转 + 黑名单拦截
```
GET /r/:code
  → api.ResolveLogic → rpc.Resolve
       · Redis 命中直接取 long_url；未命中回源 MySQL 并回填（含 owner user_id 缓存）
       · 提取 long_url 域名 → SISMEMBER 黑名单 → 命中返回 Blocked（handler 拦截，不跳转）
       · 实时计数：Redis INCR short_link:{code}:clicks + MySQL IncrClicks
       · 异步写 ClickHouse `click_events`（code/long_url/user_id/ip/referer/status），驱动 /api/logs 与 /api/usage-trends
  → 302 跳转 long_url
```

### 3. 访问统计（当前实现）
- **实时计数**：Redis `short_link:{code}:clicks`（列表/详情即时展示）。
- **访问明细（已接入 ClickHouse）**：每次跳转由 `rpc.Resolve` **异步**写入 ClickHouse 表 `click_events`（code / long_url / user_id / ip / referer / status / created_at，按 `user_id` 隔离）；`/api/logs` 分页查询、`/api/usage-trends` 按天聚合均直接查 ClickHouse。表结构见 `deploy/sql/clickhouse.sql`。
- **[规划中] 异步管道**：`slink.yaml` 仍保留 `KafkaBrokers` + `ClickEventsTopic` 配置，未来可改为「resolve → Kafka → 消费者批量写 ClickHouse」进一步解耦；当前为 rpc 直接异步写 ClickHouse。

### 4. 用户体系（纯 HTTP，无 RPC）
- 邮箱注册（SMTP 验证码）→ 登录发 JWT；GitHub OAuth 回调建号/绑定。
- 登录后可在 `/api/keys` 管理 API Key（明文仅返回一次，入库存 `sha256`）。
- 偏好设置 `user_settings`（邮件通知/安全告警/营销通讯），`/api/settings` 读写。

### 5. 接口清单（apps/api）

| 方法 & 路径 | 鉴权 | 说明 |
|---|---|---|
| POST /api/auth/register | 公开 | 邮箱注册 → JWT |
| POST /api/auth/login | 公开 | 邮箱登录 → JWT |
| GET /api/auth/github | 公开 | GitHub OAuth 授权 URL |
| GET /api/auth/github/callback | 公开 | GitHub 回调 → JWT |
| POST /api/short-links | API Key | 创建短链（核心逻辑在 rpc） |
| GET /api/short-links | JWT | 列出我的短链 |
| GET /api/short-links/:code | API Key | 按 code 查询 |
| GET /r/:code | 公开 | 短链跳转（调 rpc.Resolve） |
| GET /api/keys | JWT | 列出我的 API Key |
| POST /api/keys | JWT | 创建 API Key |
| DELETE /api/keys/:id | JWT | 吊销 API Key |
| GET /api/profile | JWT | 获取资料 |
| POST /api/profile | JWT | 更新资料 |
| POST /api/profile/password | JWT | 修改密码 |
| GET /api/settings | JWT | 读取偏好设置 |
| PUT /api/settings | JWT | 更新偏好设置 |
| GET /api/usage-trends | JWT | 用量趋势（按天聚合） |
| GET /api/logs | JWT | 短链访问明细日志 |

> 字段命名约定：请求/响应 JSON 统一使用 **camelCase**（前端 `web/`、`admin-web/` 均按 camelCase 收发）。

---

## 七、关键数据模型（MySQL: short_chain）

| 表 | 用途 |
|---|---|
| `users` | 账号体系（email unique / github_id / password_hash / nickname / status） |
| `api_keys` | API Key（key_hash=sha256，prefix 用于展示，quota/used 限流） |
| `user_settings` | 用户偏好（email_notif / security_alerts / marketing_comm，按 user_id） |
| `short_links` | 短链核心（code unique / long_url / user_id / clicks / status） |
| `domain_blacklist` | 域名黑名单（创建/跳转时校验，admin 可管理） |
| `action_logs` | handler 操作日志（web / admin-web 经网关与 admin 服务产生的操作记录） |
| `click_events` | 短链访问明细（ClickHouse，每次 `/r/:code` 跳转异步写入） |
| `rpc_logs` | 短链核心 gRPC 调用日志（ClickHouse，rpc 拦截器异步写入） |

> **ClickHouse `click_events` 表**（非 MySQL，已启用）：短链被访问的明细（驱动 `/api/logs`、`/api/usage-trends`，按 `user_id` 隔离），由 `rpc.Resolve` 异步写入，`/api/logs` 分页与 `/api/usage-trends` 按天聚合直接查询。MySQL 中原有的 `short_link_visits` 表已废弃移除；表结构见 `deploy/sql/clickhouse.sql`。

---

## 八、短码生成

- 算法：**Snowflake（IdGen.NextID）+ Base62 编码**（见 `common/tool` + `rpc/internal/logic/createslinklogic.go`）。
- 缓存：`short_link:{code}` 存 long_url，TTL = 30min + 随机 0~10min，防集中失效/雪崩；另缓存 `short_link:{code}:uid` 供访问明细按用户隔离。
- 去重：`short_links.code` 唯一索引。

---

## 九、落地建议顺序

1. `server/deploy/docker/`：先起依赖（MySQL/Redis；Kafka/ClickHouse 按需）。
2. `apps/rpc`：短链核心（生成/解析/删除）+ Redis + 黑名单最小可用。
3. `apps/api`：网关骨架（用户体系 + 短链 CRUD/跳转 + 日志/用量）。
4. `apps/admin`：管理后台 API（链接/黑名单/Token）。
5. `web/` 与 `admin-web/`：Vue3 前端；`admin-web` 构建产物交由 `apps/admin` 静态托管。
6. [规划] `sdk/go`：封装 HTTP 客户端（带 Key）+ gRPC 客户端；[规划] Kafka→ClickHouse 异步分析接入。

---

> 敏感配置（数据库 / ClickHouse / JWT 密钥 / GitHub OAuth 等）集中在各 `etc/*.yaml` 与部署配置中管理，不在此文档列出。
