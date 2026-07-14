# 短链服务基础架构设计

> 版本：v3（go-zero 重构：Nginx 负载均衡 + 开放 API + 后台管理）
> 更新时间：2026-07-14

---

## 一、架构总览

```
【 用户 / 浏览器 】
        │
        ▼
【 Nginx 负载均衡 + TLS终止 + 限流/WAF 】
   └─ 路由:
      /r/{code}  → 短链跳转服务 (server/apps/api)
      /api/*     → 业务 API (server/apps/api: 注册/OAuth/Key/短链CRUD/黑名单)
      /admin/*   → 管理后台 API (server/apps/admin) + Vue3 静态托管
      /          → Vue3 官网 (web/)
        │
        ▼
【 Go 服务 (go-zero 单体仓库, 多实例, Nginx 固定实例) 】
   ├─ apps/api (短链开放 API 服务):
   │     ├─ api/   HTTP 网关: 纯 HTTP 接口
   │     │     · 用户体系: 注册/登录/GitHub OAuth/API Key 管理/资料 (本地逻辑，不依赖 gRPC)
   │     │     · 跳转 /r/{code}: Redis缓存 → 命中302 + 发Kafka(点击事件)
   │     │     │                → 未命中 SingleFlight 回源 + 布隆防穿透
   │     │     │                → 跳转前校验 域名黑名单(Redis Set) → 命中则拦截
   │     │     · 短链 CRUD: 鉴权(API Key/JWT) 后内部 gRPC 调用 rpc/ 核心服务
   │     ├─ rpc/   短链核心服务(gRPC/zrpc): 仅承载短链生成/查询/删除/批量 (ShortLink)
   │     └─ Kafka 消费者: 点击事件 → ClickHouse(统计) + 聚合计数
   └─ apps/admin (管理后台服务):
         链接管理 · 访问次数 · 域名黑名单 的 API (供 admin 前端调用)
        │
        ▼
【 Kafka 】  ◄── 点击流 / 审计事件(削峰解耦)
   ├─► ClickHouse(单机Docker)  ← 访问次数/分析统计
   └─► (可选)落库聚合
【 MySQL 】  ← 用户/API Key/短链/黑名单 持久化
【 Redis 】  ← 跳转缓存 + 黑名单Set + 限流计数 + 实时点击数
        │
        ▼
【 web/  Vue3 官网 】   【 admin 前端  Vue3 管理后台 】
        │
        ▼
【 sdk/go  Go 调用包 】  ← 第三方: HTTP 客户端(带 Key) + gRPC 客户端(内部短链服务)
        │
        ▼
   注:
   · 用户体系(注册/登录/OAuth/Key/资料) = 纯 HTTP 接口，由 apps/api/api 本地处理，无 gRPC 调用
   · 短链生成提供两种调用方式 ——
       · 外部第三方: HTTP API (POST /api/short-links, 带 X-API-Key) → api/ 网关 → rpc/ (gRPC)
       · 内部 Go 服务: 直连 apps/api/rpc 的 gRPC 服务 (ShortLink) → 走内部网络
```

---

## 二、技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go 1.26 · go-zero（api + zrpc）· go-redis · sqlx · sarama(Kafka) · clickhouse-go |
| 存储 | MySQL(业务) · Redis(缓存/黑名单/限流) · ClickHouse(统计) · Kafka(消息) |
| 前端 | Vue3 + Vite + Pinia + Vue Router · Element Plus(admin) / Naive UI |
| 鉴权 | JWT(会话) · GitHub OAuth2 · 邮箱 SMTP 验证 · API Key(Bearer) |
| 部署 | Docker Compose 一键起 MySQL/Redis/Kafka/ClickHouse/Nginx |

---

## 三、目录划分

```
short-chain-service/
├── server/            # Go 后端 · go-zero 单体仓库 (共享一个 go.mod)
│   ├── apps/          # 业务服务总目录
│   │   ├── api/       # 短链开放 API 服务 (HTTP 网关 + 短链核心 gRPC)
│   │   │   ├── api/   # HTTP 网关层: 纯 HTTP 接口
│   │   │   │           #  · 用户体系(注册/登录/GitHub OAuth/API Key/资料) = 本地逻辑，无 gRPC
│   │   │   │           #  · 短链 CRUD 经 API Key/JWT 鉴权后内部 gRPC 调用 rpc/
│   │   │   └── rpc/   # 短链核心 gRPC 服务: pb/shortlink.proto + 生成; 生成/查询/删除短链
│   │   └── admin/     # 管理后台服务 (链接管理/统计/黑名单 API)
│   │       ├── api/
│   │       └── rpc/
│   ├── common/        # 全局公共代码 (避免模块间循环依赖)
│   │   ├── ctxdata/       # JWT / Context 上下文解析
│   │   ├── errorx/        # 全局统一错误码
│   │   ├── interceptors/  # gRPC 拦截器 / API 中间件
│   │   ├── tool/          # 工具类 (加密/随机/时间)
│   │   └── xfilters/      # 敏感词过滤 / 通用过滤器
│   ├── deploy/        # 部署配置
│   │   ├── docker/        # Dockerfile
│   │   ├── k8s/           # K8s yaml (可选)
│   │   └── prometheus/    # 监控配置
│   ├── go.mod
│   └── go.sum
├── web/              # Vue3 官网: 邮箱注册/登录/GitHub授权/申请Key
├── admin/            # Vue3 管理后台前端 (构建产物由 server/apps/admin 静态托管)
├── sdk/go/           # 可发布的 Go SDK (独立 go.mod): HTTP 客户端(带 Key) + gRPC 客户端
├── docs/             # 架构 & API 文档
└── go.mod            # 根模块 (可用 Go Workspace 聚合 server/ 与 sdk/go/)
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

### 5. 接口调用方式

#### 5.1 用户体系（纯 HTTP，无 RPC）
用户体系**只提供 HTTP 接口，不暴露 gRPC**，所有逻辑在 `apps/api/api` 网关本地处理，不依赖 `apps/api/rpc` 核心服务：

- `POST /api/auth/register` 邮箱注册 → 返回 JWT
- `POST /api/auth/login` 邮箱登录 → 返回 JWT
- `GET  /api/auth/github` 获取 GitHub OAuth 授权 URL
- `GET  /api/auth/github/callback` GitHub 回调 → 返回 JWT
- `POST /api/keys` 创建 API Key（JWT 鉴权）
- `GET  /api/keys` 列出我的 API Key（JWT 鉴权）
- `DELETE /api/keys/:id` 吊销 API Key（JWT 鉴权）
- `GET  /api/profile` 获取当前用户资料（JWT 鉴权）

> 设计取舍：用户体系与"会话/凭证"强相关（JWT 签发、OAuth、Key 哈希），放在网关本地处理最直观、延迟最低，无需跨进程 gRPC；其数据模型（`users` / `api_keys`）由网关直接落库。

#### 5.2 短链调用方式（HTTP API + gRPC 双通道）
短链的生成/管理能力同时以两种方式对外提供，**二者最终都落到 `apps/api/rpc` 核心服务**，保证业务逻辑唯一、一致：

- **HTTP API（外部第三方调用）**
  - 入口：`POST /api/short-links`（及短链 CRUD 接口），请求头携带 `X-API-Key`，由 `apps/api/api` 网关鉴权 + 按 Key 限流后，内部调用 `apps/api/rpc`（gRPC）。
  - 适用：浏览器 / 脚本 / 异构语言服务，经 Nginx 公网暴露。
  - 客户端：`sdk/go` 提供封装好的 HTTP 客户端（自动带 Key、重试）。
- **RPC / gRPC（内部 Go 服务调用）**
  - 入口：`apps/api/rpc` 暴露的 `ShortLink` gRPC 服务（如 `CreateShortLink` / `GetByCode` / `BatchCreate` / `DeleteShortLink`）。
  - 适用：单体仓库内其他微服务、或同内网的其他 Go 服务，直连 gRPC，免去每次 HTTP 编解码与 Key 校验开销。
  - 鉴权：内部调用走 `common/interceptors` 做服务标识 / mTLS 校验，也可按需携带 Key；**gRPC 端口不对外网暴露（仅内网）**。
  - 客户端：`sdk/go` 同时提供 gRPC 客户端 stub。
- **一致性**：无论走 HTTP 还是 RPC，短码生成（Snowflake + Base62）、黑名单校验、Redis 缓存、Kafka 事件均在同一 `rpc` 核心实现，避免逻辑分叉。

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

### v3 本次调整（相对 v2）

1. **后端框架由 Gin/Echo 切换为 go-zero**：用 `.api` 描述文件 + `goctl` 生成 HTTP 代码，内部服务间用 `zrpc`（gRPC）通信；内置 `sqlx`/`redis`/`kafka` 集成与限流、熔断、中间件能力。
2. **`api` 与 `admin` 合并进 `server/` 单体仓库**：二者作为 `server/apps/` 下的两个业务模块（`apps/api` 短链开放 API 服务、`apps/admin` 管理后台服务），共享同一个 `go.mod` 与 `common/` 公共库，避免重复依赖与循环引用。
3. **公共能力下沉到 `common/`**：`ctxdata`（JWT 解析）、`errorx`（统一错误码）、`interceptors`（拦截器/中间件）、`tool`、`xfilters` 等跨模块复用。
4. **部署配置归入 `server/deploy/`**：`docker/`、`k8s/`、`prometheus/` 统一管理；根 `deploy/` 不再单独存在。
5. **用户体系为纯 HTTP 接口**：注册/登录/GitHub OAuth/API Key 管理/资料全部由 `apps/api/api` 网关本地处理，**不提供 gRPC**，不依赖 `apps/api/rpc`。
6. **短链生成提供双调用方式**：HTTP API（外部带 Key）+ gRPC RPC（内部 Go 服务直连），核心逻辑统一在 `apps/api/rpc`，`sdk/go` 同时提供两种客户端。

---

## 七、落地建议顺序

1. `server/deploy/docker/`：先把依赖（MySQL/Redis/Kafka/ClickHouse）用 Docker 跑起来。
2. `server/apps/api` 骨架：`goctl` 生成 `.api` + 跳转 `/r/{code}` + Redis + 黑名单最小可用。
3. `server/apps/admin` 骨架：管理后台 API（链接管理/统计/黑名单）。
4. `web/` 与 `admin/` 的 Vue3 脚手架；`admin` 构建产物交由 `server/apps/admin` 静态托管。
5. `sdk/go` 包：封装 HTTP 客户端（带 Key）+ gRPC 客户端两种调用方式。





## ClickHouse 密码

账号 default ， 密码：L4Pt8w6sxk6ueNZ