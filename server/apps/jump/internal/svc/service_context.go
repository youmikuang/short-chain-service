package svc

import (
	"server/apps/jump/internal/config"
	pb "server/apps/rpc/pb"

	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext jump 服务上下文：仅持有 rpc 核心客户端，用于解析短链并跳转。
type ServiceContext struct {
	Config       config.Config
	ShortLinkRpc pb.ShortLinkClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	cli := zrpc.MustNewClient(c.Rpc)
	return &ServiceContext{
		Config:       c,
		ShortLinkRpc: pb.NewShortLinkClient(cli.Conn()),
	}
}
