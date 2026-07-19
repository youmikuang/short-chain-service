# server Makefile —— go-zero 单体仓库构建脚本

.PHONY: gen gen-rpc build tidy run-api run-rpc run-admin run-jump docker

# 安装工具
# protoc + 插件: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#                go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# 或统一用 goctl: go install github.com/zeromicro/go-zero/tools/goctl@latest

# 仅生成 .proto 的 pb 代码（pb.go + grpc.pb.go）。
# 注意：logic / server / handler / types 均已手写，勿用 --zrpc_out / goctl api go 覆盖。
gen-rpc:
	protoc --go_out=paths=source_relative:. \
	       --go-grpc_out=paths=source_relative:. \
	       apps/api/rpc/pb/slink.proto
	# 若使用 goctl：
	# goctl rpc protoc apps/api/rpc/pb/slink.proto \
	#   --go_out=apps/api/rpc/pb --go-grpc_out=apps/api/rpc/pb

# api / admin 的 types/handler/logic/register 已手写完整，无需生成。
# 如需重新生成桩代码，请先备份 internal/ 再执行：
# goctl api go -api apps/api/api/desc/slink.api -dir apps/api/api --style gozero
# goctl api go -api apps/admin/api/desc/admin.api -dir apps/admin/api --style gozero

gen: gen-rpc

tidy:
	go mod tidy

build:
	go build ./apps/api && go build ./apps/rpc && go build ./apps/admin && go build ./apps/jump

run-api:
	go run apps/api -f apps/api/etc/api-api.yaml

run-rpc:
	go run apps/rpc -f apps/rpc/etc/slink.yaml

run-admin:
	go run apps/admin -f apps/admin/etc/admin-api.yaml

run-jump:
	go run apps/jump -f apps/jump/etc/jump-api.yaml

docker:
	docker build -f deploy/docker/Dockerfile -t slink-server .
