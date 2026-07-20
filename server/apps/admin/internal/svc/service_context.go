package svc

import (
	"database/sql"

	"server/apps/admin/internal/config"
	pb "server/apps/rpc/pb"
	"server/common/clickhouse"
	"server/common/model"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 管理后台上下文
type ServiceContext struct {
	Config   config.Config
	Mysql    sqlx.SqlConn
	Models   *model.Models
	Redis    *redis.Client
	ClickHouse *sql.DB
	RpcLog   *model.RpcLogModel
	slinkRpc pb.SlinkClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.BlacklistRedis.Host,
		Password: c.BlacklistRedis.Pass,
		DB:       c.BlacklistRedis.DB,
	})
	conn := sqlx.NewMysql(c.Mysql.DataSource)
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
		Config:      c,
		Mysql:       conn,
		Models:      model.NewModels(conn),
		Redis:       rdb,
		ClickHouse:  chDB,
		RpcLog:      model.NewRpcLogModel(chDB),
		slinkRpc:    pb.NewslinkClient(zrpc.MustNewClient(c.Rpc).Conn()),
	}
}
