package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Mysql struct {
		DataSource string
	}
	BlacklistRedis struct {
		Host string
		Pass string
		DB   int
	}
	BlacklistRedisKey string
	ClickEventsTopic  string
	KafkaBrokers      []string
}
