# server

short-chain-service 的 **Go 后端（go-zero 单体仓库）**，对应 `docs/architecture.md` v3 设计。

## 目录结构

布局参考 [go-zero-box](https://github.com/prf16/go-zero-box) 的工程模板约定：
API 描述文件集中在顶层 `api/`，跨服务复用代码放在 `pkg/`，各独立服务置于 `apps/`。

```
server/
├── api/                   # API 描述文件（按模块平铺，对应 go-zero-box 的 api/）
│   ├── shortlink.api       # 短链开放 API 定义（api 服务）
│   ├── user.api            # 用户体系 API 定义（api 服务）
│   └── admin.api           # 管理后台 API 定义（admin 服务）
├── apps/                  # 各独立服务（每个是一个独立二进制，对应 go-zero-box 的 app/）
│   ├── rpc/               # 短链核心 gRPC 服务（生成/解析/删除/批量）
│   │   ├── pb/            # shortlink.proto + 生成代码
│   │   ├── etc/           # slink.yaml
│   │   ├── internal/      # config / logic / server / svc
│   │   └── main.go
│   ├── api/               # 业务 API 网关（用户体系 + 短链 CRUD 委托 rpc）
│   │   ├── etc/           # api-api.yaml
│   │   ├── internal/      # config / handler / logic / middleware / svc / types
│   │   └── main.go
│   ├── admin/             # 管理后台 API（黑名单 / Token / 链接 / Dashboard）
│   │   ├── etc/           # admin-api.yaml
│   │   ├── internal/      # config / handler / logic / middleware / svc / types
│   │   └── main.go
│   └── jump/              # 跳转服务（解析短码并 302，委托 rpc）
│       ├── etc/           # jump-api.yaml
│       ├── internal/      # config / handler / logic / svc / types
│       └── main.go
├── pkg/                    # 跨服务公共代码（对应 go-zero-box 的 pkg/）
│   ├── clickhouse/        # ClickHouse 连接封装
│   ├── ctxdata/           # 上下文数据（uid/api_key 透传）
│   ├── errorx/            # 统一错误码
│   ├── interceptors/      # gRPC 拦截器（鉴权 / 日志）
│   ├── model/             # 数据模型层（gorm models）
│   ├── tool/              # 工具函数（URL 归一化 / 雪花 ID / Base62 等）
│   ├── xfilters/          # 过滤器
│   └── xhttp/             # HTTP 响应模板（统一错误响应）
├── deploy/                 # docker / k8s / prometheus / sql
├── runtime/                # 运行时目录（日志等）
├── go.mod
└── Makefile
```

## 调用方式（双通道）

- **HTTP API（外部第三方）**：`POST /api/short-links`，Header `X-API-Key`，经 `apps/api` 网关鉴权后调用 `apps/rpc` 核心服务。
- **RPC / gRPC（内部 Go 服务）**：直连 `apps/rpc` 的 `slink` gRPC 服务（端口仅内网暴露，走 `pkg/interceptors` 内部鉴权）。

## 本地开发

```sh
# 1. 起依赖（MySQL/Redis/Kafka/ClickHouse/etcd/Nginx）
docker compose -f deploy/docker/docker-compose.yml up -d

# 2. 安装 protoc + 插件（或统一用 goctl）
# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 3. 仅生成 .proto 的 pb 代码（pb.go + grpc.pb.go）
make gen
make tidy

# 4. 运行（需先 etcd 与依赖就绪）
make run-rpc      # 短链核心 gRPC 服务 :8081
make run-api      # HTTP 网关           :8888
make run-admin    # 管理后台 API        :8889
```

## 说明

- **手写部分**：`pkg/*`、`*/internal/logic/*`、`*/internal/handler/*`、`*/internal/types`、`rpc/internal/server/server.go`、`*/internal/config`、`*/internal/svc`、`main.go`、`.proto`/`.api` 描述文件。
- **生成部分**：仅 `apps/rpc/pb/*.pb.go`（由 `make gen-rpc` 经 `protoc` 生成）。`logic`/`server`/`handler`/`types` 均已手写，**不要用 `goctl` 的 `--zrpc_out` / `goctl api go` 重新生成，否则会覆盖手写实现**。
- 业务逻辑集中在 `*/internal/logic/*.go`：短码生成（Snowflake+Base62）、域名黑名单校验（Redis Set）、缓存/回源、点击计数、Kafka 事件等，均落在 `apps/rpc` 核心服务，保证 HTTP 与 RPC 调用逻辑一致。
- 存储（MySQL 落库、ClickHouse 消费、Kafka 生产）以 `TODO` 标注，按 `docs/architecture.md` 第四节接入。
