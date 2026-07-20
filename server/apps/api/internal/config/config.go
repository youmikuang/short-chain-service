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
	// WebBaseURL 是前端 SPA 基地址；GitHub 回调在后端完成 OAuth 交换后，
	// 会把浏览器 302 重定向到这里（/login?token=...），由前端读取并落库。
	WebBaseURL string
	Rpc         zrpc.RpcClientConf // 指向 apps/api/rpc 核心服务（仅短链核心）
	ShortDomain string             // ShortDomain 短链对外域名
	ClickHouse  struct {
		Host     string
		Port     int
		Database string
		Username string
		Password string
	}
}
