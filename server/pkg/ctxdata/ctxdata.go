package ctxdata

import "context"

type ctxKey string

const userIDKey ctxKey = "uid"

// GetUserID 从 context 取用户ID（由 JWT 中间件注入）
func GetUserID(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(userIDKey).(int64)
	return v, ok
}

// WithUserID 注入用户ID到 context
func WithUserID(ctx context.Context, uid int64) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}
