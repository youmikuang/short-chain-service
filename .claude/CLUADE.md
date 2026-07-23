# 项目说明

## 这个项目
这是一个 短链服务（Short Chain Service）：提供短链接生成、跳转、访问统计与开放 API 调用的后端服务，配套 Vue3 官网与管理后台。

## 技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go 1.26 · Gin/Echo · go-redis · GORM/sqlc · sarama(Kafka) · clickhouse-go |
| 存储 | MySQL(业务) · Redis(缓存/黑名单/限流) · ClickHouse(统计) · Kafka(消息) |
| 前端 | Vue3 + Vite + Pinia + Vue Router |
| 鉴权 | JWT(会话) · GitHub OAuth2 · 邮箱 SMTP · API Key(Bearer) |
| 部署 | Docker Compose（MySQL/Redis/Kafka/ClickHouse/Nginx） |

---


## 目录结构

```
short-chain-service/
├── docs/              # 架构 & API 文档
├── web/               # Vue3 官网：注册/登录/GitHub 授权/申请 Key
├── admin-web/         # Vue3 管理后台
├── server/            # Go 后端（go-zero 单体仓库）
│   ├── apps/
│   │   ├── rpc/       # 短链核心 gRPC 服务（生成/解析/删除/批量）
│   │   ├── api/       # 业务 API 网关（用户体系 + 短链 CRUD 委托 rpc）
│   │   ├── admin/     # 管理后台 API（黑名单 / Token / 链接 / Dashboard）
│   │   └── jump/      # 跳转服务（解析短码并 302，委托 rpc）
│   ├── common/        # 跨服务公共：model / errorx / tool / clickhouse / interceptors
│   └── deploy/        # docker-compose + nginx + k8s + sql
└── README.md
```

> `sdk/go/` 仍规划中，尚未落地（详见 `docs/architecture.md`）。

---

## 重要约束
- 所有的 API 必须要有单元测试
- 构建的时候先启动 RPC 服务在启动其他的
- ClickHouse rpc_logs 存储的是创建短链时候的记录
- ClickHouse click_events 存储的是短链服务点击的记录。
- mysql action_logs 表是操作web 或者 admin 时候的记录。
- 每次在生成完代码后都需要验证各个服务启动有没有报错，如果有报错就要进行修改
- 根据 TDD 进行代码层面进行约束，每次接口和方法新增后都要有对应的测试

## 测试

所有 API 必须有对应的单元测试（TDD 约束）。测试分两类：

### 1. 纯单元测试（无需任何基础设施）

- `server/common/tool/tool_test.go`：URL 归一化 / 域名提取 / Sha256 / Base62 / 随机串 / 雪花 ID。
- `server/common/errorx/errorx_test.go`：统一错误码与 `Is` 判定。
- `server/apps/api/internal/logic/security_test.go`：密码哈希、JWT 签发校验、`uidFromCtx`、boolToInt。
- `server/apps/jump/internal/logic/resolvelogic_test.go`：用 `mockSlinkClient` 验证 Resolve 的 metadata 透传与黑名单/错误分支，**无需启动 rpc 服务**。

### 2. 集成测试（依赖 MySQL / Redis / ClickHouse）

各 app 的 `logic` 包加载 `etc/*.yaml` 真实配置构建 `ServiceContext`，直连本地
`127.0.0.1:3306`(MySQL) / `127.0.0.1:6379`(Redis) 与配置中的 ClickHouse，覆盖：

- `rpc`：CreateSlink（web/rpc 路径、黑名单、去重）、GetByCode、Resolve（含黑名单）、BatchCreate、DeleteSlink。
- `api`：Register / Login / ListMyLinks / CreateAPIKey / ListAPIKeys / RevokeAPIKey / Profile / Settings / ChangePassword / GitHubAuthURL / GitHubCallback / UsageTrends / Logs；以及网关委托 rpc 的 **CreateSlink / GetByCode**（用 `mockSlinkClient` 校验 uid/api_key 透传与 ShortURL 拼接，无需启动 rpc）。
- `admin`：AdminLogin / 黑名单增删查 / ListLinks / Dashboard / Token 签发与吊销启用重置。

> 注意：远程 ClickHouse 偶发握手失败。依赖 CH 且**未做降级**的测试（如 `TestClickHouseProbe`、
> `TestDashboard`）在 CH 不可达时会 `t.Skip` 而非报错；而 `UsageTrends` / `Logs` 在 CH 不可达时
> 由业务逻辑优雅降级为空结果，仍断言通过。

### 3. 适配层测试（middleware / handler / interceptor，覆盖 100% 接口）

接口不仅包含 logic 业务方法，也包含其外层适配器（鉴权中间件、HTTP handler、gRPC 拦截器）。
这些同样按 TDD 约束补齐：

- **gRPC 拦截器** `server/common/interceptors/interceptors_test.go`（纯单元，无需基础设施）：
  - `UnaryServerInterceptor`：Resolve 公开端点放行、缺失 metadata / 错误 token 拒绝、正确 token 放行。
  - `UnaryClientInterceptor`：网关→rpc 每次调用附带 `x-internal-token`。
  - `UnaryServerLogInterceptor`：不阻塞请求、成功/错误均透传（rpc_logs 异步写入）。
- **api 网关中间件** `server/apps/api/internal/middleware/middleware_test.go`（纯单元）：
  - `AuthMiddleware`：Bearer JWT、X-API-Key、缺鉴权 401、非 `/api` 路径透传。
  - `ActionLogMiddleware`：请求写入 `action_logs`（best-effort，MySQL 不可用时也不阻塞）。
- **api 网关 handler** `server/apps/api/internal/handler/handler_test.go`：
  - `CreateSlinkHandler`（mock rpc，校验 ShortURL 拼接与缺参 400）、`GetByCodeHandler`（mock rpc，路径参数经 `pathvar` 注入）、`ListMyLinksHandler`（真实 MySQL，注入 uid）。
- **admin handler** `server/apps/admin/internal/handler/handler_test.go`（纯单元，仅依赖配置）：
  - `AdminLoginHandler`：正确凭据签发 JWT、错误凭据 / 非法 JSON 返回 400。
- **jump handler** `server/apps/jump/internal/handler/resolvehandler_test.go`（mock rpc，无需基础设施）：
  - 纯函数 `cleanIP` / `clientIP` / `parseForwardedFor` / `buildShortURL`（含 IPv6、代理头、Forwarded 头）。
  - `ResolveHandler`：302 跳转、短码不存在映射 404、命中黑名单、rpc 其它错误（均经 `pathvar` 注入 `:code`）。

> 说明：handler 成功路径委托 logic（logic 已 100% 覆盖），handler 测试聚焦请求解析、错误映射
> （`httpx.ErrorCtx` 默认写 400）与响应写出；受保护的管理后台路由（Dashboard / Links / Token 等）
> 的成功路径由 `adminlogic_test.go` 覆盖，handler 层仅测登录这一公开入口。

### 运行方式

```bash
cd server
# 全部后端测试
go test ./apps/... ./common/...
# 单包
go test ./apps/rpc/internal/logic/ -v
go test ./apps/api/internal/logic/ -run TestRegister -v
# 仅纯单元测试（无需基础设施）
go test ./common/tool/ ./common/errorx/ ./apps/jump/internal/logic/ \
       ./common/interceptors/ ./apps/jump/internal/handler/ ./apps/api/internal/middleware/
```

> 集成测试前置：本地需有 MySQL + Redis（`docker compose -f server/deploy/docker/docker-compose.yml up -d`），
> 且 `etc/*.yaml` 中的 ClickHouse 可达；rpc 服务无需预先启动（api/jump 中委托 rpc 的部分通过
> 真实 gRPC 或 mock 覆盖，见各测试文件注释）。

---

## 构建和测试

```bash
# RPC 服务
go build server/apps/rpc
# api 接口
go build server/apps/api
# admin 接口
go build server/apps/admin
# jump 跳转服务
go build server/apps/jump
# web 前端启动
cd web && pnpm dev
# 后台前端启动
cd admin-web && pnpm dev
```