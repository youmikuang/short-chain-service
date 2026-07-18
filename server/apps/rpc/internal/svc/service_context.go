package svc

import (
	"database/sql"

	"server/apps/rpc/internal/config"
	"server/common/clickhouse"
	"server/common/model"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ServiceContext 短链核心服务上下文
type ServiceContext struct {
	Config           config.Config
	Mysql            sqlx.SqlConn
	Models           *model.Models
	Redis            *redis.Client
	ClickHouse       *sql.DB
	ClickHouseVisit  *model.ShortLinkVisitModel
	ClickHouseRpcLog *model.RpcLogModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.BlacklistRedis.Host,
		Password: c.BlacklistRedis.Pass,
		DB:       c.BlacklistRedis.DB,
	})
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	// workerId 由部署时固定分配（环境变量/配置），此处取实例号
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
		Config:           c,
		Mysql:            conn,
		Models:           model.NewModels(conn),
		Redis:            rdb,
		ClickHouse:       chDB,
		ClickHouseVisit:  model.NewShortLinkVisitModel(chDB),
		ClickHouseRpcLog: model.NewRpcLogModel(chDB),
	}
}
