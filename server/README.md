# server

short-chain-service 的 **Go 后端（go-zero 单体仓库）**，对应 `docs/architecture.md` v3 设计。

## 目录结构

```
server/
├── apps/
│   ├── api/                # 短链开放 API 服务（对外 HTTP + 对内 gRPC 双通道）
│   │   ├── api/            # HTTP 网关: desc/*.api + 手写 handler/logic/types
│   │   └── rpc/            # 短链核心 gRPC 服务: pb/*.proto + logic
│   └── admin/              # 管理后台服务（链接管理/统计/黑名单 API）
│       └── api/
├── common/                 # ctxdata / errorx / interceptors / tool / xfilters
├── deploy/                 # docker / k8s / prometheus
├── go.mod
└── Makefile
```

## 调用方式（双通道）

- **HTTP API（外部第三方）**：`POST /api/short-links`，Header `X-API-Key`，经 `apps/api/api` 网关鉴权后调用 `apps/api/rpc` 核心服务。
- **RPC / gRPC（内部 Go 服务）**：直连 `apps/api/rpc` 的 `ShortLink` gRPC 服务（端口仅内网暴露，走 `common/interceptors` 内部鉴权）。

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

- **手写部分**：`common/*`、`*/internal/logic/*`、`*/internal/handler/*`、`*/internal/types`、`rpc/internal/server/server.go`、`*/internal/config`、`*/internal/svc`、`main.go`、`.proto`/`.api` 描述文件。
- **生成部分**：仅 `apps/api/rpc/pb/*.pb.go`（由 `make gen-rpc` 经 `protoc` 生成）。`logic`/`server`/`handler`/`types` 均已手写，**不要用 `goctl` 的 `--zrpc_out` / `goctl api go` 重新生成，否则会覆盖手写实现**。
- 业务逻辑集中在 `*/internal/logic/*.go`：短码生成（Snowflake+Base62）、域名黑名单校验（Redis Set）、缓存/回源、点击计数、Kafka 事件等，均落在 `apps/api/rpc` 核心服务，保证 HTTP 与 RPC 调用逻辑一致。
- 存储（MySQL 落库、ClickHouse 消费、Kafka 生产）以 `TODO` 标注，按 `docs/architecture.md` 第四节接入。
