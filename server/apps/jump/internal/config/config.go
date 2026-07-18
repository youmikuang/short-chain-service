package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config jump 服务配置：仅需要 REST 服务配置与 rpc 核心客户端。
// 跳转服务不直连 MySQL / ClickHouse / 鉴权，访问明细由 rpc.Resolve 异步写入 ClickHouse。
type Config struct {
	rest.RestConf
	Rpc zrpc.RpcClientConf
}
