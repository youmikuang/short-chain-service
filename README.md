# short-chain-service

短链服务（Short Chain Service）：提供短链接生成、跳转、访问统计与开放 API 调用的后端服务，配套 Vue3 官网与管理后台。

> 架构设计详见 [`docs/architecture.md`](docs/architecture.md)。

---

## 一、功能特性

- **短链生成与跳转**：将长链接压缩为短码，访问 `/r/{code}` 时 302 跳转至原链接。
- **域名黑名单**：跳转前校验目标域名，命中黑名单则拦截，防钓鱼/滥用。
- **账号体系**：邮箱注册/登录（SMTP 验证码）+ GitHub OAuth。
- **开放 API（API Key）**：登录后申请 API Key，调用短链接口需携带 `X-API-Key` 并按 Key 限流。
- **访问统计**：Redis 实时计数 + Kafka → ClickHouse 做历史分析。
- **管理后台**：链接管理、访问次数查看、域名黑名单管理。

---

## 二、技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go 1.26 · Gin/Echo · go-redis · GORM/sqlc · sarama(Kafka) · clickhouse-go |
| 存储 | MySQL(业务) · Redis(缓存/黑名单/限流) · ClickHouse(统计) · Kafka(消息) |
| 前端 | Vue3 + Vite + Pinia + Vue Router |
| 鉴权 | JWT(会话) · GitHub OAuth2 · 邮箱 SMTP · API Key(Bearer) |
| 部署 | Docker Compose（MySQL/Redis/Kafka/ClickHouse/Nginx） |

---

## 三、目录结构

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

## 四、快速开始（前端 web）

`web/` 是一个标准的 Vue3 + Vite 项目，当前已可独立开发与构建。

### 环境要求

- Node.js `^22.18.0 || >=24.12.0`
- 包管理器：pnpm（推荐）

### 安装依赖

```sh
cd web
pnpm install
```

### 本地开发（热更新）

```sh
pnpm dev
```

### 类型检查 + 生产构建

```sh
pnpm build
```

### 预览构建产物

```sh
pnpm preview
```

### 运行单元测试（Vitest）

```sh
pnpm test:unit
```

### 代码格式化

```sh
pnpm format
```

---

## 五、后端与部署

后端采用 **Nginx 负载均衡 + Go 多实例（固定实例）** 的部署形态，依赖通过 Docker Compose 一键启动：

```sh
# 启动依赖（MySQL/Redis/Kafka/ClickHouse/Nginx）
docker compose -f server/deploy/docker/docker-compose.yml up -d
```

各服务构建（构建顺序：先 RPC，再其它）：

```sh
cd server
go build ./apps/rpc && go build ./apps/api && go build ./apps/admin && go build ./apps/jump
```

### 后端单元测试

所有 API 均有对应单元测试（详见 `.claude/CLUADE.md` 的「测试」一节）。纯单元测试无需基础设施；
集成测试需本地 MySQL + Redis（`127.0.0.1:3306` / `127.0.0.1:6379`）与配置中的 ClickHouse：

```sh
cd server
go test ./apps/... ./common/...
# 仅纯单元测试（无需基础设施）
go test ./common/tool/ ./common/errorx/ ./apps/jump/internal/logic/ \
       ./common/interceptors/ ./apps/jump/internal/handler/ ./apps/api/internal/middleware/
```

Nginx 路由规划：

| 路径 | 说明 |
|---|---|
| `/r/{code}` | 短链跳转服务 |
| `/api/*` | 业务 API（注册/OAuth/Key/短链 CRUD/黑名单） |
| `/admin/*` | Vue3 管理后台（静态托管） |
| `/` | Vue3 官网（注册/登录/授权/申请 Key） |

> 详细数据流、数据模型与落地顺序见 [`docs/architecture.md`](docs/architecture.md) 第四、五、七节。

---

## 六、短链调用示例

调用短链创建接口需携带 API Key：

```sh
curl -X POST https://your-domain/api/short-links \
  -H "X-API-Key: <YOUR_API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"long_url": "https://example.com/some/very/long/path"}'
```

跳转访问：

```
GET https://your-domain/r/{code}   → 302 跳转到原链接
```

---

## 七、相关文档

- [架构设计](docs/architecture.md)
- [web 前端说明](web/README.md)

---

## 八、License

待定。
