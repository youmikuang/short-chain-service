package middleware

import (
	"net/http"
	"server/common/model"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// responseRecorder 包装 http.ResponseWriter 以捕获状态码
type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// uidFromToken 仅解码（不校验）JWT 载荷取 uid，用于操作日志；
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

// NewActionLogMiddleware 记录 admin 服务每次 HTTP 请求（admin-web 操作日志）到 MySQL action_logs。
func NewActionLogMiddleware(models *model.Models) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			next(rec, r)
			latency := time.Since(start).Milliseconds()

			_, _ = models.ActionLog.Insert(r.Context(), &model.ActionLog{
				UserId:    uidFromToken(r),
				Method:    r.Method,
				Endpoint:  r.URL.Path,
				Status:    int64(rec.status),
				LatencyMs: latency,
			})
		}
	}
}
