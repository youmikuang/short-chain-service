package middleware

import (
	"net/http"
	"server/apps/api/api/internal/svc"
	"server/common/model"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// skipPaths 这些自身用于查看日志/统计的端点不写访问日志，避免自递归。
var skipPaths = map[string]bool{
	"/api/logs":         true,
	"/api/usage-trends": true,
}

// responseRecorder 包装 http.ResponseWriter 以捕获状态码
type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// uidFromToken 仅解码（不校验）JWT 载荷取 uid，用于访问日志；
// 真正的鉴权由 jwt/apikey 中间件完成，此处只做记录。
func uidFromToken(r *http.Request) int64 {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return 0
	}
	tokenStr := auth[len("Bearer "):]
	token, _, err := jwt.NewParser().ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return 0
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if v, ok := claims["uid"].(float64); ok {
			return int64(v)
		}
	}
	return 0
}

// NewAccessLogMiddleware 记录每次 HTTP 请求到 access_logs（驱动 /api/logs 与 /api/usage-trends）。
func NewAccessLogMiddleware(svcCtx *svc.ServiceContext) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if skipPaths[r.URL.Path] {
				next(w, r)
				return
			}
			start := time.Now()
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			next(rec, r)
			latency := time.Since(start).Milliseconds()

			uid := uidFromToken(r)
			if uid == 0 {
				if v, ok := r.Context().Value(APIKeyUIDKey).(float64); ok {
					uid = int64(v)
				}
			}
			_, _ = svcCtx.Models.AccessLog.Insert(r.Context(), &model.AccessLog{
				UserId:    uid,
				Method:    r.Method,
				Endpoint:  r.URL.Path,
				Status:    int64(rec.status),
				LatencyMs: latency,
			})
		}
	}
}
