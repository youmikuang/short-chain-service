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
	RateLimit struct {
		Redis struct {
			Host string
			Pass string
			DB   int
		}
	}
	Github struct {
		ClientID     string
		ClientSecret string
		RedirectURL  string
	}
	Rpc zrpc.RpcClientConf // 指向 apps/api/rpc 核心服务（仅短链核心）
	ClickHouse struct {
		Host     string
		Port     int
		Database string
		Username string
		Password string
	}
}
