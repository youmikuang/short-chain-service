package handler

import (
	"net/http"
	"server/apps/api/internal/middleware"
	"server/apps/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers 注册路由，按鉴权方式分组：
//   - apiKeyRoutes: 需 X-API-Key（第三方短链调用）
//   - jwtRoutes:    需 JWT（用户自己的资源：API Key / 资料）
//   - publicRoutes: 公开（注册/登录/GitHub/短链跳转）
func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	apiKeyRoutes := []rest.Route{
		{Method: http.MethodPost, Path: "/api/short-links", Handler: CreateShortLinkHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/short-links/:code", Handler: GetByCodeHandler(svcCtx)},
	}
	jwtRoutes := []rest.Route{
		{Method: http.MethodPost, Path: "/api/keys", Handler: CreateAPIKeyHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/keys", Handler: ListAPIKeysHandler(svcCtx)},
		{Method: http.MethodDelete, Path: "/api/keys/:id", Handler: RevokeAPIKeyHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/profile", Handler: GetProfileHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/profile", Handler: UpdateProfileHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/profile/password", Handler: ChangePasswordHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/settings", Handler: GetSettingsHandler(svcCtx)},
		{Method: http.MethodPut, Path: "/api/settings", Handler: UpdateSettingsHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/usage-trends", Handler: UsageTrendsHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/logs", Handler: LogsHandler(svcCtx)},
	}
	publicRoutes := []rest.Route{
		{Method: http.MethodPost, Path: "/api/auth/register", Handler: RegisterHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/auth/login", Handler: LoginHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/auth/github", Handler: GitHubAuthURLHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/auth/github/callback", Handler: GitHubCallbackHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/r/:code", Handler: ResolveHandler(svcCtx)},
	}

	server.AddRoutes(rest.WithMiddlewares([]rest.Middleware{middleware.NewAPIKeyMiddleware(svcCtx)}, apiKeyRoutes...))
	server.AddRoutes(jwtRoutes, rest.WithJwt(svcCtx.Config.Auth.AccessSecret))
	server.AddRoutes(publicRoutes)
}
