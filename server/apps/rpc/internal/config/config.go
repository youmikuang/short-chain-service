package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Mysql struct {
		DataSource string
	}
	RedisConfig struct {
		Host string
		Pass string
		DB   int
	}
	RedisKey         string
	ClickEventsTopic string
	KafkaBrokers     []string
	ClickHouse       struct {
		Host     string
		Port     int
		Database string
		Username string
		Password string
	}
}
