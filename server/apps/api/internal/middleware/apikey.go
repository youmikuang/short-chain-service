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

// NewAPIKeyMiddleware 校验请求头 X-API-Key（仅对 /api/* 生效，/r/* 跳转为公开端点）。
// 校验通过后把 key 对应的 user_id 写入 context，供后续 logic 使用。
func NewAPIKeyMiddleware(svcCtx *svc.ServiceContext) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/api/") {
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
