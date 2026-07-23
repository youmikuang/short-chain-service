package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"server/apps/jump/internal/svc"
	pb "server/apps/rpc/pb"

	"github.com/zeromicro/go-zero/rest/pathvar"
	"google.golang.org/grpc"
)

// mockSlinkClient 实现 pb.SlinkClient，仅覆盖 Resolve，供 handler 测试（无需启动 rpc 核心）。
type mockSlinkClient struct {
	pb.SlinkClient
	resp *pb.ResolveResp
	err  error
}

func (m *mockSlinkClient) Resolve(ctx context.Context, in *pb.ResolveReq, opts ...grpc.CallOption) (*pb.ResolveResp, error) {
	return m.resp, m.err
}

func newJumpSvc(client pb.SlinkClient) *svc.ServiceContext {
	return &svc.ServiceContext{SlinkRpc: client}
}

func TestCleanIP(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"[::1]:8080", "::1"},
		{"1.2.3.4:8080", "1.2.3.4"},
		{"[2001:db8::1]", "2001:db8::1"},
		{"1.2.3.4", "1.2.3.4"},
		{"  ", ""},
	}
	for _, c := range cases {
		if got := cleanIP(c.in); got != c.want {
			t.Fatalf("cleanIP(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestClientIP(t *testing.T) {
	cases := []struct {
		name string
		setup func(r *http.Request)
		want string
	}{
		{"x-forwarded-for", func(r *http.Request) { r.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8") }, "1.2.3.4"},
		{"x-real-ip", func(r *http.Request) { r.Header.Set("X-Real-IP", "9.9.9.9") }, "9.9.9.9"},
		{"cf-connecting-ip", func(r *http.Request) { r.Header.Set("CF-Connecting-IP", "8.8.8.8") }, "8.8.8.8"},
		{"remote-addr", func(r *http.Request) { r.RemoteAddr = "1.2.3.4:5678" }, "1.2.3.4"},
		{"remote-addr-ipv6", func(r *http.Request) { r.RemoteAddr = "[::1]:1234" }, "::1"},
		{"forwarded-header", func(r *http.Request) { r.Header.Set("Forwarded", "for=7.7.7.7;proto=https") }, "7.7.7.7"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/r/abc", nil)
			c.setup(r)
			if got := clientIP(r); got != c.want {
				t.Fatalf("clientIP = %q, want %q", got, c.want)
			}
		})
	}
}

func TestParseForwardedFor(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{`for=1.2.3.4;proto=https`, "1.2.3.4"},
		{`for="[2001:db8::1]";proto=https`, "2001:db8::1"},
		{`host=example.com;for=4.4.4.4`, "4.4.4.4"},
		{`proto=https`, ""},
	}
	for _, c := range cases {
		if got := parseForwardedFor(c.in); got != c.want {
			t.Fatalf("parseForwardedFor(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBuildShortURL(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/r/abc", nil)
	r.Host = "s.example.com"
	if got := buildShortURL(r, "abc"); got != "http://s.example.com/r/abc" {
		t.Fatalf("buildShortURL = %q, want http://s.example.com/r/abc", got)
	}

	r2 := httptest.NewRequest(http.MethodGet, "/r/abc", nil)
	r2.Header.Set("X-Forwarded-Proto", "https")
	r2.Header.Set("X-Forwarded-Host", "short.test")
	if got := buildShortURL(r2, "xyz"); got != "https://short.test/r/xyz" {
		t.Fatalf("buildShortURL = %q, want https://short.test/r/xyz", got)
	}
}

func TestResolveHandler_OK(t *testing.T) {
	mock := &mockSlinkClient{resp: &pb.ResolveResp{LongUrl: "https://example.com/x", Blocked: false}}
	h := ResolveHandler(newJumpSvc(mock))

	req := httptest.NewRequest(http.MethodGet, "/r/abc?code=abc", nil)
	req = pathvar.WithVars(req, map[string]string{"code": "abc"})
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "https://example.com/x" {
		t.Fatalf("Location = %q, want https://example.com/x", loc)
	}
}

func TestResolveHandler_NotFound(t *testing.T) {
	mock := &mockSlinkClient{err: errors.New("code not found")}
	h := ResolveHandler(newJumpSvc(mock))

	req := httptest.NewRequest(http.MethodGet, "/r/abc?code=abc", nil)
	req = pathvar.WithVars(req, map[string]string{"code": "abc"})
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestResolveHandler_Blocked(t *testing.T) {
	mock := &mockSlinkClient{resp: &pb.ResolveResp{Blocked: true}}
	h := ResolveHandler(newJumpSvc(mock))

	req := httptest.NewRequest(http.MethodGet, "/r/abc?code=abc", nil)
	req = pathvar.WithVars(req, map[string]string{"code": "abc"})
	rec := httptest.NewRecorder()
	h(rec, req)

	// 命中黑名单：errorx.CodeBlacklisted 经 httpx.ErrorCtx 默认写 400
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestResolveHandler_RPCError(t *testing.T) {
	mock := &mockSlinkClient{err: errors.New("rpc unavailable")}
	h := ResolveHandler(newJumpSvc(mock))

	req := httptest.NewRequest(http.MethodGet, "/r/abc?code=abc", nil)
	req = pathvar.WithVars(req, map[string]string{"code": "abc"})
	rec := httptest.NewRecorder()
	h(rec, req)

	// 非 not found 的 rpc 错误：httpx.ErrorCtx 默认写 400
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
