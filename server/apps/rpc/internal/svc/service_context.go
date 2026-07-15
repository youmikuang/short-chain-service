package svc

import (
	"server/apps/rpc/internal/config"
	"server/common/model"
	"server/common/tool"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ServiceContext 短链核心服务上下文
type ServiceContext struct {
	Config config.Config
	Mysql  sqlx.SqlConn
	Models *model.Models
	IdGen  *tool.Snowflake
	Redis  *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.BlacklistRedis.Host,
		Password: c.BlacklistRedis.Pass,
		DB:       c.BlacklistRedis.DB,
	})
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	// workerId 由部署时固定分配（环境变量/配置），此处取实例号
	return &ServiceContext{
		Config: c,
		Mysql:  conn,
		Models: model.NewModels(conn),
		IdGen:  tool.NewSnowflake(1),
		Redis:  rdb,
	}
}
