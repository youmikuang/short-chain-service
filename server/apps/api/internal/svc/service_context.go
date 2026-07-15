package svc

import (
	"server/apps/api/internal/config"
	"server/common/model"
	pb "server/apps/rpc/pb"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext api 网关上下文
// 用户体系为纯 HTTP 接口（本地逻辑，直连 MySQL）；
// 短链核心仍通过 gRPC 调用 rpc 核心服务。
type ServiceContext struct {
	Config       config.Config
	Mysql        sqlx.SqlConn
	Models       *model.Models
	ShortLinkRpc pb.ShortLinkClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	cli := zrpc.MustNewClient(c.Rpc)
	return &ServiceContext{
		Config:       c,
		Mysql:        conn,
		Models:       model.NewModels(conn),
		ShortLinkRpc: pb.NewShortLinkClient(cli.Conn()),
	}
}
