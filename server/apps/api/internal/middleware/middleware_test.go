package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"server/apps/api/internal/config"
	"server/apps/api/internal/svc"
	"server/common/model"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/golang-jwt/jwt/v4"
)

func testSvcCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	var c config.Config
	conf.MustLoad("../../etc/api-api.yaml", &c)
	return &svc.ServiceContext{Config: c}
}

// testSvcCtxWithModels 构建带 MySQL Models 的 ServiceContext（不含 ClickHouse，避免无 CH 时 panic）。
// 供 ActionLog 中间件测试使用；其 Insert 为 best-effort（忽略错误），MySQL 不可用时也不会失败。
func testSvcCtxWithModels(t *testing.T) *svc.ServiceContext {
	t.Helper()
	var c config.Config
	conf.MustLoad("../../etc/api-api.yaml", &c)
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &svc.ServiceContext{
		Config: c,
		Mysql:  conn,
		Models: model.NewModels(conn),
	}
}

func issueJWT(secret string, uid int64) string {
	claims := jwt.MapClaims{"uid": uid, "exp": jwt.NewNumericDate(jwt.TimeFunc().Add(3600 * time.Second))}
	return mustSign(claims, secret)
}

func mustSign(claims jwt.MapClaims, secret string) string {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := tok.SignedString([]byte(secret))
	return s
}

func TestAuthMiddleware_Bearer(t *testing.T) {
	svcCtx := testSvcCtx(t)
	mw := NewAuthMiddleware(svcCtx)

	var gotUID interface{}
	called := false
	next := func(w http.ResponseWriter, r *http.Request) {
		called = true
		gotUID = r.Context().Value("uid")
		w.WriteHeader(http.StatusOK)
	}
	h := mw(next)

	req := httptest.NewRequest(http.MethodGet, "/api/short-links", nil)
	req.Header.Set("Authorization", "Bearer "+issueJWT(svcCtx.Config.Auth.AccessSecret, 42))
	rec := httptest.NewRecorder()
	h(rec, req)

	if !called {
		t.Fatal("next not called")
	}
	if gotUID != float64(42) {
		t.Fatalf("uid in context = %v, want 42", gotUID)
	}
}

func TestAuthMiddleware_APIKey(t *testing.T) {
	svcCtx := testSvcCtx(t)
	mw := NewAuthMiddleware(svcCtx)

	var gotKey interface{}
	called := false
	next := func(w http.ResponseWriter, r *http.Request) {
		called = true
		gotKey = r.Context().Value(APIKeyKey)
		w.WriteHeader(http.StatusOK)
	}
	h := mw(next)

	req := httptest.NewRequest(http.MethodPost, "/api/short-links", nil)
	req.Header.Set("X-API-Key", "slk_abcdef")
	rec := httptest.NewRecorder()
	h(rec, req)

	if !called {
		t.Fatal("next not called")
	}
	if gotKey != "slk_abcdef" {
		t.Fatalf("api_key in context = %v, want slk_abcdef", gotKey)
	}
}

func TestAuthMiddleware_Missing(t *testing.T) {
	svcCtx := testSvcCtx(t)
	mw := NewAuthMiddleware(svcCtx)

	called := false
	next := func(w http.ResponseWriter, r *http.Request) { called = true }
	h := mw(next)

	req := httptest.NewRequest(http.MethodGet, "/api/short-links", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if called {
		t.Fatal("next should NOT be called without auth")
	}
	// httpx.ErrorCtx 默认写 400（项目未注册自定义 error handler），与线上行为一致。
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAuthMiddleware_NonAPIPathPassthrough(t *testing.T) {
	svcCtx := testSvcCtx(t)
	mw := NewAuthMiddleware(svcCtx)

	called := false
	next := func(w http.ResponseWriter, r *http.Request) { called = true; w.WriteHeader(200) }
	h := mw(next)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if !called {
		t.Fatal("non-/api path should pass through without auth")
	}
}

func TestActionLogMiddleware(t *testing.T) {
	svcCtx := testSvcCtxWithModels(t)
	mw := NewActionLogMiddleware(svcCtx)

	called := false
	next := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, "ok")
	}
	h := mw(next)

	req := httptest.NewRequest(http.MethodPost, "/api/short-links", nil)
	req.Header.Set("Authorization", "Bearer "+issueJWT(svcCtx.Config.Auth.AccessSecret, 7))
	rec := httptest.NewRecorder()
	h(rec, req)

	if !called {
		t.Fatal("next not called")
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status not propagated: %d", rec.Code)
	}
	// skipPaths 中的端点不应写日志（此处 /api/short-links 不在 skip 列表，应写入）；
	// 仅验证中间件不阻塞、状态正确透传即可（写入为 best-effort）。
}
