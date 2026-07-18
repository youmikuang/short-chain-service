package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"server/apps/api/internal/svc"
	"server/common/errorx"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// APIKeyUIDKey 是 API Key 鉴权后写入 context 的用户ID key
const APIKeyUIDKey = "uid"

// APIKeyKey 是 API Key 鉴权后写入 context 的原始 Key（供 logic 透传给 rpc）
const APIKeyKey = "api_key"

// NewAuthMiddleware 校验请求身份：Bearer JWT 优先（web 前端，无需 API Key），
// 缺失时回退到 X-API-Key（第三方）。注意：API Key 的合法性校验放在 rpc 核心服务，
// 此处仅把原始 key 透传（写入 context），不在此做哈希比对。
// 仅对 /api/* 生效，/r/* 跳转为公开端点（不在此中间件处理）。
func NewAuthMiddleware(svcCtx *svc.ServiceContext) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/api/") {
				next(w, r)
				return
			}
			// 1) Bearer JWT：web 前端鉴权，有效即可创建短链；key 与 web 无关，直接忽略
			if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
				uid, err := parseJWTUid(r, svcCtx.Config.Auth.AccessSecret)
				if err != nil {
					httpx.ErrorCtx(r.Context(), w, errorx.Unauthorized("invalid Authorization"))
					return
				}
				ctx := context.WithValue(r.Context(), "uid", uid)
				next(w, r.WithContext(ctx))
				return
			}
			// 2) 回退 X-API-Key：原始 key 透传，交由 rpc 校验
			if key := r.Header.Get("X-API-Key"); key != "" {
				ctx := context.WithValue(r.Context(), APIKeyKey, key)
				next(w, r.WithContext(ctx))
				return
			}
			httpx.ErrorCtx(r.Context(), w, errorx.Unauthorized("missing Authorization or X-API-Key"))
		}
	}
}

// parseJWTUid 校验 Authorization: Bearer <token> 并返回 uid（与 go-zero JWT 中间件同算法 HS256）。
func parseJWTUid(r *http.Request, secret string) (float64, error) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return 0, errors.New("missing bearer token")
	}
	tokenStr := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}
	v, ok := claims["uid"]
	if !ok {
		return 0, errors.New("uid missing in token")
	}
	switch n := v.(type) {
	case float64:
		return n, nil
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, err
		}
		return f, nil
	default:
		return 0, errors.New("invalid uid type")
	}
}
