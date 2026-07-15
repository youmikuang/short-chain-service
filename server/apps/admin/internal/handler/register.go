package handler

import (
	"net/http"
	"server/apps/admin/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers 注册管理后台路由
func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	// 公开接口（无需 JWT）
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/admin/api/login", Handler: AdminLoginHandler(svcCtx)},
	})

	// 受保护接口（需 JWT）
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/admin/api/dashboard", Handler: DashboardHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/admin/api/links", Handler: ListLinksHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/admin/api/blacklist", Handler: ListBlacklistHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/admin/api/blacklist", Handler: AddBlacklistHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/admin/api/tokens", Handler: ListTokensHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/admin/api/tokens", Handler: ProvisionTokenHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/admin/api/tokens/revoke", Handler: RevokeTokenHandler(svcCtx)},
	}, rest.WithJwt(svcCtx.Config.Auth.AccessSecret))
}
