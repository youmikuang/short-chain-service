package middleware

import (
	"context"
	"net/http"
	"server/apps/api/internal/svc"
	"server/common/errorx"
	"server/common/tool"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// APIKeyUIDKey 是 API Key 鉴权后写入 context 的用户ID key
const APIKeyUIDKey = "uid"

// APIKeyKey 是 API Key 鉴权后写入 context 的原始 Key（供 logic 透传给 rpc）
const APIKeyKey = "api_key"

// apiKeySkipPaths 是仅用 JWT 鉴权、不需要 X-API-Key 的路由（method+path）。
// 这些路由由 go-zero 的 JWT 中间件保护，API Key 中间件应直接放行。
var apiKeySkipPaths = map[string]bool{
	"GET /api/short-links":    true, // 用户自己的短链列表
	"GET /api/keys":           true, // 列出当前用户的 API Key
	"POST /api/keys":          true, // 创建 API Key
	"DELETE /api/keys/:id":    true, // 吊销 API Key
	"GET /api/profile":        true, // 资料读取
	"POST /api/profile":       true, // 资料更新
	"POST /api/profile/password": true, // 改密码
	"GET /api/settings":       true, // 设置读取
	"PUT /api/settings":       true, // 设置更新
	"GET /api/usage-trends":   true, // 用量趋势
	"GET /api/logs":           true, // 访问日志
}

// NewAPIKeyMiddleware 校验请求头 X-API-Key（仅对 /api/* 生效，/r/* 跳转为公开端点）。
// 校验通过后把 key 对应的 user_id 写入 context，供后续 logic 使用。
// 注意：仅用 JWT 鉴权的路由（见 apiKeySkipPaths）会被直接放行，交由 JWT 中间件处理。
func NewAPIKeyMiddleware(svcCtx *svc.ServiceContext) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/api/") {
				next(w, r)
				return
			}
			if apiKeySkipPaths[r.Method+" "+r.URL.Path] {
				next(w, r)
				return
			}
			key := r.Header.Get("X-API-Key")
			if key == "" {
				httpx.ErrorCtx(r.Context(), w, errorx.Unauthorized("missing X-API-Key"))
				return
			}
			row, err := svcCtx.Models.ApiKey.FindOneByHash(r.Context(), tool.Sha256Hex(key))
			if err != nil || row.Status != 1 {
				httpx.ErrorCtx(r.Context(), w, errorx.Unauthorized("invalid X-API-Key"))
				return
			}
			// 写入 user_id 与原始 key 到 context，供 CreateShortLink 透传给 rpc
			ctx := r.Context()
			ctx = context.WithValue(ctx, APIKeyUIDKey, float64(row.UserId))
			ctx = context.WithValue(ctx, APIKeyKey, key)
			next(w, r.WithContext(ctx))
		}
	}
}
