package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	Mysql struct {
		DataSource string
	}
	BlacklistRedis struct {
		Host string
		Pass string
		DB   int
	}
	BlacklistRedisKey string
	Rpc               zrpc.RpcClientConf
	Admin             struct {
		Username string
		Password string
	}
	// ShortDomain 短链对外域名
	ShortDomain string
}
