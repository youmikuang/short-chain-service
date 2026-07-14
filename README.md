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
├── go.mod             # Go 模块定义
└── README.md
```

> 以下模块按架构设计规划中，尚未落地（详见 `docs/architecture.md`）：
> - `api/`：Go 后端（跳转 + 业务 API + Kafka 消费者）
> - `admin/`：Vue3 管理后台
> - `sdk/go/`：可发布的 Go SDK
> - `deploy/`：docker-compose + nginx.conf

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

## 五、后端与部署（规划中）

后端采用 **Nginx 负载均衡 + Go 多实例（固定实例）** 的部署形态，依赖通过 Docker Compose 一键启动：

```sh
# 规划中的启动方式（待 deploy/ 落地）
docker compose -f deploy/docker-compose.yml up -d   # MySQL/Redis/Kafka/ClickHouse/Nginx
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

## 六、短链调用示例（规划中 API）

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
