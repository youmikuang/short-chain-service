package interceptors

import (
	"context"
	"server/common/errorx"

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
