package interceptors

import (
	"context"
	"errors"
	"server/common/errorx"
	"server/common/model"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// InternalServiceToken 内部服务间调用约定的 token（生产建议 mTLS + 服务标识）
const InternalServiceToken = "internal-rpc-secret"

// UnaryServerInterceptor 校验内部调用身份（gRPC 服务端）
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errorx.Unauthorized("missing metadata")
		}
		tokens := md.Get("x-internal-token")
		if len(tokens) == 0 || tokens[0] != InternalServiceToken {
			return nil, errorx.Unauthorized("invalid internal token")
		}
		return handler(ctx, req)
	}
}

// UnaryClientInterceptor 在网关（api）→ 核心 rpc 的每次调用中附带内部服务 token，
// 与服务端的 UnaryServerInterceptor 校验配对。缺少该 metadata 时 rpc 会拒绝所有调用。
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-internal-token", InternalServiceToken)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// UnaryServerLogInterceptor 记录每次 RPC 调用到 ClickHouse rpc_logs（异步，不阻塞处理）。
// 作为最外层拦截器注册，可覆盖鉴权失败等所有入口调用。
func UnaryServerLogInterceptor(logModel *model.RpcLogModel) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		latency := time.Since(start).Milliseconds()

		status := int64(0)
		errMsg := ""
		if err != nil {
			var e *errorx.Error
			if errors.As(err, &e) {
				status = int64(e.Code)
			} else {
				status = -1
			}
			errMsg = err.Error()
		}

		go func() {
			_ = logModel.Insert(context.Background(), &model.RpcLog{
				Method:    info.FullMethod,
				UserId:    0,
				Status:    status,
				LatencyMs: latency,
				Error:     errMsg,
			})
		}()
		return resp, err
	}
}
