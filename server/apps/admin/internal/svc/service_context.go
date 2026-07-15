package svc

import (
	"server/apps/admin/internal/config"
	"server/common/model"
	pb "server/apps/rpc/pb"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 管理后台上下文
type ServiceContext struct {
	Config       config.Config
	Mysql        sqlx.SqlConn
	Models       *model.Models
	Redis        *redis.Client
	ShortLinkRpc pb.ShortLinkClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.BlacklistRedis.Host,
		Password: c.BlacklistRedis.Pass,
		DB:       c.BlacklistRedis.DB,
	})
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:       c,
		Mysql:        conn,
		Models:       model.NewModels(conn),
		Redis:        rdb,
		ShortLinkRpc: pb.NewShortLinkClient(zrpc.MustNewClient(c.Rpc).Conn()),
	}
}
