package svc

import (
	"database/sql"

	"server/apps/api/internal/config"
	"server/common/clickhouse"
	"server/common/interceptors"
	"server/common/model"
	pb "server/apps/rpc/pb"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext api 网关上下文
// 用户体系为纯 HTTP 接口（本地逻辑，直连 MySQL）；
// 短链核心仍通过 gRPC 调用 rpc 核心服务。
type ServiceContext struct {
	Config         config.Config
	Mysql          sqlx.SqlConn
	Models         *model.Models
	ShortLinkRpc   pb.ShortLinkClient
	ClickHouse     *sql.DB
	ClickHouseVisit *model.ShortLinkVisitModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	cli := zrpc.MustNewClient(c.Rpc, zrpc.WithUnaryClientInterceptor(interceptors.UnaryClientInterceptor()))
	chDB, err := clickhouse.NewClient(clickhouse.Config{
		Host:     c.ClickHouse.Host,
		Port:     c.ClickHouse.Port,
		Database: c.ClickHouse.Database,
		Username: c.ClickHouse.Username,
		Password: c.ClickHouse.Password,
	})
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:         c,
		Mysql:          conn,
		Models:         model.NewModels(conn),
		ShortLinkRpc:   pb.NewShortLinkClient(cli.Conn()),
		ClickHouse:     chDB,
		ClickHouseVisit: model.NewShortLinkVisitModel(chDB),
	}
}
