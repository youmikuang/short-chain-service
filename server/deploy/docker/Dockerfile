# 多阶段构建 go-zero 服务镜像（以 api 网关为例，rpc/admin 同理）
FROM golang:1.26 AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/slink-api ./apps/api/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/slink-rpc ./apps/api/rpc
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/admin-api ./apps/admin/api

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/bin/slink-api /app/bin/slink-rpc /app/bin/admin-api /app/bin/
COPY --from=builder /app/apps/api/api/etc /app/etc/api-api.yaml
COPY --from=builder /app/apps/api/rpc/etc /app/etc/slink.yaml
COPY --from=builder /app/apps/admin/api/etc /app/etc/admin-api.yaml
EXPOSE 8888 8081 8889
ENTRYPOINT ["/app/bin/slink-api", "-f", "/app/etc/api-api.yaml"]
