package handler

import (
	"net/http"

	"server/apps/jump/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers 注册跳转路由。跳转是公开端点，无需鉴权中间件。
func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	publicRoutes := []rest.Route{
		{Method: http.MethodGet, Path: "/r/:code", Handler: ResolveHandler(svcCtx)},
	}
	server.AddRoutes(publicRoutes)
}
