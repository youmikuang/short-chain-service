# 短链服务架构文档

> 依据当前代码梳理（server/ 单体仓库 + web/ 官网 + admin-web/ 后台）。

## 一、架构总览

```
【 用户 / 浏览器 】
        │
        ▼
【 Nginx（生产）/ Vite 代理（开发） 】
   ├─ /api/*     → apps/api     (HTTP 网关 :8888)
   ├─ /admin/*   → apps/admin   (管理后台 API :8889)
   ├─ /r/{code}  → apps/jump     (短链跳转 :8890，再调 rpc)
   └─ /          → web / admin-web 静态资源
        │
        ▼
【 Go 服务（go-zero 单体仓库 server/）】
   ├─ apps/api    (:8888) 用户体系 + 短链 CRUD（调 rpc）+ 查 ClickHouse click_events
   ├─ apps/jump   (:8890) GET /r/{code} → rpc.Resolve → 302/404/黑名单拦截
   ├─ apps/admin  (:8889) 登录/仪表盘/链接/黑名单/Token；仪表盘查 ClickHouse rpc_logs
   └─ apps/rpc    (:8081, 仅内网) slink 核心：生成/查询/批量/删除/解析；拦截器写 rpc_logs
        │
        ▼
【 存储 】
   · MySQL     ：用户 / API Key / 用户设置 / 短链(short_links) / 黑名单 / action_logs
   · Redis     ：跳转缓存(short_link:{code}) + 黑名单 Set + 实时点击计数
   · Kafka     ：[规划中] 点击流异步通道（配置就位，业务未接入）
   · ClickHouse：click_events（访问明细，rpc.Resolve 异步写，api 查询）
                 rpc_logs（RPC 日志，rpc 拦截器异步写，admin 仪表盘查询）
```

> 跳转（高并发读）单独拆为 `apps/jump`，与网关解耦，便于独立扩容。

---

## 二、服务与职责

### apps/api（:8888）
| 能力 | 说明 |
|---|---|
| 用户体系 | 注册/登录/GitHub OAuth/API Key/资料/设置/改密，纯 HTTP 本地落库 |
| 短链 CRUD | `POST /api/short-links`、`GET /api/short-links/:code`、`GET /api/short-links`，鉴权后调 rpc |
| 日志/用量 | `GET /api/logs`、`GET /api/usage-trends`，查 ClickHouse `click_events` |

> `/r/:code` 跳转已拆分到 `apps/jump`，不在 api。

### apps/jump（:8890）
- `ResolveHandler`：解析 code，提取客户端 IP（`X-Forwarded-For`/`X-Real-IP`），`Referer` 置为本次短链完整地址。
- `ResolveLogic` → rpc `Resolve`：返回 `LongURL` 做 302；`Blocked` 命中黑名单报错；短码不存在映射 **404**（便于 CDN 缓存）。

### apps/admin（:8889）
登录、仪表盘、链接管理、域名黑名单、Token 管理（供 `admin-web`）。
- 仪表盘流量趋势：`action_logs`（MySQL 每日操作量）+ `rpc_logs`（ClickHouse 每日生成量）双序列，聚合成近 7 天 `TrafficPoint[]`。

### apps/rpc（:8081）
`slink` 服务（`pb/shortlink.proto`）：`Createslink` / `GetByCode` / `BatchCreate` / `Deleteslink` / `Resolve`（含黑名单校验）。拦截器异步写 ClickHouse `rpc_logs`。

---

## 三、技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go · go-zero(rest+zrpc) · go-redis · sqlx(MySQL) · database/sql+clickhouse-go(ClickHouse) |
| 存储 | MySQL · Redis · ClickHouse(click_events + rpc_logs) · Kafka[规划] |
| 前端 | Vue3 + Vite + Pinia + Vue Router；`web/`、`admin-web/` |
| 鉴权 | 用户侧 JWT + GitHub OAuth2 + API Key；管理后台 独立 Admin JWT（Pinia+localStorage） |
| 部署 | Docker Compose；开发用 Vite 代理 |

---

## 四、目录结构（实际）

```
short-chain-service/
├── server/
│   ├── apps/
│   │   ├── api/     # HTTP 网关 :8888（api/ internal/ etc/）
│   │   ├── admin/   # 管理后台 API :8889
│   │   ├── jump/    # 短链跳转 :8890（internal/{handler,logic,svc,types,config}）
│   │   └── rpc/     # 短链核心 gRPC :8081（pb/ internal/ etc/）
│   ├── common/
│   │   ├── clickhouse/   # ClickHouse 连接工厂（连接池/超时）
│   │   ├── ctxdata/ errorx/ interceptors/ xfilters/
│   │   ├── model/        # MySQL: user/apikey/usersettings/slink/domainblacklist/actionlog
│   │   │                # ClickHouse: shortlink_visit(click_events)/rpclog(rpc_logs)
│   │   └── tool/         # 加密/Base62/Snowflake/域名提取
│   └── deploy/           # docker / k8s / sql / prometheus
├── web/            # Vue3 官网
├── admin-web/      # Vue3 管理后台
│   └── src/{api,stores/auth.ts,router(守卫),components(TrafficChart/TopNavBar/...),views(login/dashboard/links/blacklist/tokens)}
├── docs/  README.md
```

---

## 五、鉴权模型

### 用户侧（apps/api）
`register.go` 路由分组：
- `slinkRoutes`（API Key **或** JWT）：`POST /api/short-links`、`GET /api/short-links/:code`
- `jwtRoutes`（JWT）：`/api/keys`、`/api/profile*`、`/api/settings`、`/api/usage-trends`、`/api/logs`、`GET /api/short-links`
- `publicRoutes`：`/api/auth/register`、`/api/auth/login`、`/api/auth/github*`（`/r/:code` 在 jump，公开）

### 管理后台侧（admin-web）
- `stores/auth.ts`（Pinia）：`token` + `username`，持久化到 `localStorage`（`admin_token`/`admin_username`）。
- `router/index.ts`：`beforeEach` 守卫——未登录访问非登录页跳 `/login`，已登录访问 `/login` 跳 `/dashboard`。
- `TopNavBar` 接入 auth store 处理登出并跳转 `/login`；`main.ts` 不再启动自动登录。
- 后端 `apps/admin` 登录接口返回 JWT，`client.ts` 自动带 `Authorization: Bearer <token>`。

---

## 六、核心数据流

### 1. 短链创建
```
POST /api/short-links (API Key 或 JWT)
  → api.CreateslinkLogic（取 user_id/api_key）→ rpc.Createslink
       · 域名黑名单校验（MySQL 优先，回退 Redis Set）
       · 短码 = Base62(Snowflake.NextID())
       · 落 MySQL(short_links) + 写 Redis(short_link:{code}，随机 TTL 防雪崩)
  → 返回 { code, shortURL=ShortDomain/r/{code}, longURL }
```

### 2. 短链跳转（apps/jump）
```
GET /r/:code → jump.ResolveLogic → rpc.Resolve
       · Redis 命中取 long_url；未命中回源 MySQL 并回填
       · 域名黑名单校验 → 命中返回 Blocked（handler 拦截，不跳转）
       · 实时计数：Redis INCR + MySQL IncrClicks
       · 异步写 ClickHouse click_events（code/long_url/user_id/ip/referer/status）
  → 302 跳转 / 404（短码不存在）/ 黑名单错误
```

### 3. 访问统计
- **实时计数**：Redis `short_link:{code}:clicks`（列表/详情即时展示）。
- **访问明细**：每次跳转 `rpc.Resolve` 异步写 `click_events`（按 `user_id` 隔离）；`/api/logs` 分页、`/api/usage-trends` 按天聚合直接查 ClickHouse。
- **RPC 日志**：rpc 拦截器异步写 `rpc_logs`；admin 仪表盘直接查其每日生成量。
- **[规划]** Kafka→ClickHouse 异步管道（`slink.yaml` 仍保留 `KafkaBrokers`+`ClickEventsTopic`）。

### 4. 用户体系（纯 HTTP，无 RPC）
邮箱注册（SMTP 验证码）→ 登录发 JWT；GitHub OAuth 回调建号/绑定（**state 防 CSRF**，见 §6）；`/api/keys` 管理 API Key（明文仅返回一次，入库存 `sha256`）；`user_settings` 偏好设置。

### 5. 接口清单（apps/api）
| 方法 & 路径 | 鉴权 | 说明 |
|---|---|---|
| POST /api/auth/register | 公开 | 注册 → JWT |
| POST /api/auth/login | 公开 | 登录 → JWT |
| GET /api/auth/github | 公开 | GitHub 授权 URL（下发 `gh_oauth_state` Cookie + 带 `state`） |
| GET /api/auth/github/callback | 公开 | GitHub 回调：校验 `state` → 换 token → 302 `/login?token=...` |
| POST /api/short-links | API Key/JWT | 创建短链（核心在 rpc） |
| GET /api/short-links | JWT | 我的短链列表 |
| GET /api/short-links/:code | API Key/JWT | 按 code 查询 |
| GET /r/:code | 公开(jump) | 短链跳转（调 rpc.Resolve） |
| GET /api/keys · POST /api/keys · DELETE /api/keys/:id | JWT | API Key 管理 |
| GET/POST /api/profile · POST /api/profile/password | JWT | 资料/改密 |
| GET/PUT /api/settings | JWT | 偏好设置 |
| GET /api/usage-trends · GET /api/logs | JWT | 用量趋势 / 访问明细 |

> JSON 统一 **camelCase**（前端 `web/`、`admin-web/` 均按 camelCase 收发）。

### 6. GitHub OAuth 登录流程（state 防 CSRF）
后端无 session，采用「随机 `state` + HttpOnly Cookie」在换 token **之前**完成校验，可真正防 CSRF：

```
① 前端 stores/auth.ts 调 GET /api/auth/github?redirect=<前端回调地址>
      → GitHubAuthURLHandler 生成 state=tool.RandString(32)
      → 写入 gh_oauth_state Cookie（HttpOnly + SameSite=Lax + 10min 过期）
      → 把 state 拼入 GitHub 授权 URL 返回
② 浏览器跳 GitHub 授权
      → GitHub 302 回 redirect?code=...&state=...
        （顶级跳转中浏览器自动带 gh_oauth_state Cookie，SameSite=Lax 允许）
③ GitHubCallbackHandler 先校验 state 查询参数 === Cookie 值
      · 不一致 / 缺失 → 直接拒绝（302 /login?error=invalid+oauth+state），绝不换 token
      · 一致        → 用 code 换 access_token、建号/绑定、签发 JWT
                     → 302 <WebBaseURL>/login?token=...&user_id=...&nickname=...
                     → 清除一次性 gh_oauth_state Cookie
④ 前端 /login 页 onMounted 读取 ?token=...，落库并跳首页
```

关键配置（`apps/api/etc/*.yaml`）：
- `Github.RedirectURL`：GitHub App 登记的 **Authorization callback URL**，须与前端传入的 `redirect` 完全一致（如 `http://localhost:5173/api/auth/github/callback`）。
- `WebBaseURL`：回调 302 跳转的前端 SPA 基地址（如 `http://localhost:5173`）。

> 前端无需手动管理 state：Cookie 由浏览器在回调跳转时自动携带，校验全部在后端完成。

---

## 七、关键数据模型

**MySQL（short_chain）**
| 表 | 用途 |
|---|---|
| `users` | 账号体系（email unique / github_id / password_hash / nickname / status） |
| `api_keys` | API Key（key_hash=sha256，prefix 展示，quota/used 限流） |
| `user_settings` | 用户偏好（email_notif / security_alerts / marketing_comm） |
| `short_links` | 短链核心（code unique / long_url / user_id / clicks / status） |
| `domain_blacklist` | 域名黑名单（admin 可管理） |
| `action_logs` | handler 操作日志（web / admin-web 操作记录；驱动仪表盘每日操作量） |

**ClickHouse（short_chain）**
| 表 | 用途 |
|---|---|
| `click_events` | 短链访问明细（每次 `/r/:code` 异步写入；驱动 `/api/logs`、`/api/usage-trends`，按 user_id 隔离） |
| `rpc_logs` | 短链核心 gRPC 调用日志（rpc 拦截器异步写入；admin 仪表盘查每日生成量） |

> ClickHouse 表结构见 `deploy/sql/clickhouse.sql`。MySQL 原有 `short_link_visits` 表已废弃移除。

---

## 八、短码生成
- 算法：**Snowflake（IdGen.NextID）+ Base62 编码**（见 `common/tool` + `rpc/internal/logic/createslinklogic.go`）。
- 缓存：`short_link:{code}` 存 long_url，TTL = 30min + 随机 0~10min 防雪崩；另缓存 `short_link:{code}:uid` 供访问明细按用户隔离。
- 去重：`short_links.code` 唯一索引。

---

## 九、前端要点
- **web/**：`token.vue`、`logs.vue` 的日志表头已由 `Endpoint` 改为 **Shortened URL**（当前展示 `/r/{code}` 路径）。
- **admin-web/**：
  - `dashboard.vue` 的 "System Traffic (7 days)" 由柱状图改为 **带曲率双折线图**（`components/TrafficChart.vue`）：Actions（action_logs 每日量）+ RPC（rpc_logs 每日量）。
  - 登录态由 `stores/auth.ts`（Pinia）管理，含 `/login` 页与路由守卫。

---

## 十、落地建议顺序
1. `deploy/docker/`：先起依赖（MySQL/Redis；Kafka/ClickHouse 按需）。
2. `apps/rpc`：短链核心 + Redis + 黑名单最小可用。
3. `apps/api`：网关（用户体系 + 短链 CRUD/跳转 + 日志/用量）。
4. `apps/jump`：独立跳转服务（调 rpc.Resolve）。
5. `apps/admin`：管理后台 API（链接/黑名单/Token/仪表盘）。
6. `web/` 与 `admin-web/`：Vue3 前端；`admin-web` 构建产物交由 `apps/admin` 静态托管。
7. [规划] `sdk/go`、Kafka→ClickHouse 异步分析接入。

> 敏感配置（数据库 / ClickHouse / JWT 密钥 / GitHub OAuth 等）集中在各 `etc/*.yaml` 与部署配置中管理，不在此文档列出。
