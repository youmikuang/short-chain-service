package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"

	"server/apps/api/rpc/internal/config"
	"server/apps/api/rpc/internal/server"
	"server/apps/api/rpc/internal/svc"
	"server/common/interceptors"

	pb "server/apps/api/rpc/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/shortlink.yaml", "the config file")

// resolveConfigFile 在给定路径找不到配置时，回退到相对于源码所在目录的路径，
// 这样无论以何种工作目录（go run / GoLand 临时构建 / 直接执行二进制）运行都能定位到 etc/ 下的配置。
func resolveConfigFile(f string) string {
	if _, err := os.Stat(f); err == nil {
		return f
	}
	if !filepath.IsAbs(f) {
		if _, thisFile, _, ok := runtime.Caller(0); ok {
			if candidate := filepath.Join(filepath.Dir(thisFile), f); candidate != f {
				if _, err := os.Stat(candidate); err == nil {
					return candidate
				}
			}
		}
	}
	return f
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(resolveConfigFile(*configFile), &c)

	ctx := svc.NewServiceContext(c)
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterShortLinkServer(grpcServer, server.NewShortLinkServer(ctx))
	})
	// 内部服务间调用鉴权
	s.AddUnaryInterceptors(interceptors.UnaryServerInterceptor())

	defer s.Stop()
	s.Start()
}
