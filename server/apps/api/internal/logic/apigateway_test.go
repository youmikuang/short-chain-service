package logic

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"server/apps/api/internal/config"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	pb "server/apps/rpc/pb"

	"google.golang.org/grpc"
)

// --- mock SlinkClient：用于 api 网关委托 rpc 的单元测试（无需启动 rpc 服务） ---
type mockSlinkClient struct {
	pb.SlinkClient
	lastCreateSlink *pb.CreateSlinkReq
	lastGetByCode   *pb.GetByCodeReq
	createResp      *pb.CreateSlinkResp
	getResp         *pb.GetByCodeResp
	err             error
}

func (m *mockSlinkClient) CreateSlink(ctx context.Context, in *pb.CreateSlinkReq, opts ...grpc.CallOption) (*pb.CreateSlinkResp, error) {
	m.lastCreateSlink = in
	return m.createResp, m.err
}

func (m *mockSlinkClient) GetByCode(ctx context.Context, in *pb.GetByCodeReq, opts ...grpc.CallOption) (*pb.GetByCodeResp, error) {
	m.lastGetByCode = in
	return m.getResp, m.err
}

// newGatewaySvc 构造仅含 rpc 客户端与短链域名的 api ServiceContext（CreateSlink/GetByCode 不依赖 MySQL/CH）。
func newGatewaySvc(client pb.SlinkClient) *svc.ServiceContext {
	return &svc.ServiceContext{
		Config:   config.Config{ShortDomain: "https://s.test"},
		SlinkRpc: client,
	}
}

func TestApiCreateSlink(t *testing.T) {
	mock := &mockSlinkClient{createResp: &pb.CreateSlinkResp{Code: "abc123", LongUrl: "https://example.com/x"}}
	svcCtx := newGatewaySvc(mock)

	// 模拟 API Key 鉴权中间件写入的 context
	ctx := context.WithValue(context.Background(), "uid", float64(42))
	ctx = context.WithValue(ctx, "api_key", "slk_xxx")

	l := NewCreateSlinkLogic(ctx, svcCtx)
	resp, err := l.CreateSlink(&types.CreateSlinkReq{LongURL: "https://example.com/x"})
	if err != nil {
		t.Fatalf("CreateSlink failed: %v", err)
	}
	if resp.Code != "abc123" {
		t.Fatalf("Code = %q, want abc123", resp.Code)
	}
	if resp.LongURL != "https://example.com/x" {
		t.Fatalf("LongURL = %q", resp.LongURL)
	}
	if resp.ShortURL != "https://s.test/r/abc123" {
		t.Fatalf("ShortURL = %q, want https://s.test/r/abc123", resp.ShortURL)
	}
	// 校验 uid / api_key 正确透传给 rpc
	if mock.lastCreateSlink.GetUserId() != 42 {
		t.Fatalf("rpc received UserId %d, want 42", mock.lastCreateSlink.GetUserId())
	}
	if mock.lastCreateSlink.GetApiKey() != "slk_xxx" {
		t.Fatalf("rpc received ApiKey %q, want slk_xxx", mock.lastCreateSlink.GetApiKey())
	}
	if mock.lastCreateSlink.GetLongUrl() != "https://example.com/x" {
		t.Fatalf("rpc received LongUrl %q", mock.lastCreateSlink.GetLongUrl())
	}
}

func TestApiGetByCode(t *testing.T) {
	mock := &mockSlinkClient{getResp: &pb.GetByCodeResp{Code: "abc123", LongUrl: "https://example.com/x", Clicks: 7, Status: 1}}
	svcCtx := newGatewaySvc(mock)

	l := NewGetByCodeLogic(context.Background(), svcCtx)
	resp, err := l.GetByCode(&types.GetByCodeReq{Code: "abc123"})
	if err != nil {
		t.Fatalf("GetByCode failed: %v", err)
	}
	if resp.Code != "abc123" || resp.LongURL != "https://example.com/x" || resp.Clicks != 7 || resp.Status != 1 {
		t.Fatalf("unexpected mapping: %+v", resp)
	}
	if mock.lastGetByCode.GetCode() != "abc123" {
		t.Fatalf("rpc received Code %q", mock.lastGetByCode.GetCode())
	}
}

// --- GitHubCallback：用 httptest 风格的假 transport 模拟 GitHub，验证 OAuth 回调建号/绑定的完整链路 ---
type fakeGitHubTransport struct{}

func (fakeGitHubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body string
	switch {
	case strings.Contains(req.URL.Path, "/login/oauth/access_token"):
		body = `{"access_token":"fake-token","token_type":"bearer"}`
	case strings.Contains(req.URL.Path, "/user"):
		body = `{"id":12345,"login":"octocat","email":"octo@github.com","avatar_url":"https://avatars.github.com/octo"}`
	default:
		body = `{}`
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}, nil
}

func TestGitHubCallback(t *testing.T) {
	ctx := newAPITestSvc(t)

	// 注入假 GitHub HTTP 客户端，测试后恢复
	old := githubHTTPClient
	githubHTTPClient = &http.Client{Transport: fakeGitHubTransport{}}
	defer func() { githubHTTPClient = old }()

	l := NewGitHubCallbackLogic(context.Background(), ctx)
	resp, err := l.GitHubCallback(&types.GitHubCallbackReq{Code: "auth-code"})
	if err != nil {
		t.Fatalf("GitHubCallback failed: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("empty token")
	}
	defer cleanupUser(t, ctx, resp.UserID)

	// 二次回调（同一 GitHub id）应复用账号而非新建
	resp2, err := l.GitHubCallback(&types.GitHubCallbackReq{Code: "auth-code"})
	if err != nil {
		t.Fatalf("GitHubCallback (2nd) failed: %v", err)
	}
	if resp2.UserID != resp.UserID {
		t.Fatalf("expected same user, got %d then %d", resp.UserID, resp2.UserID)
	}
	defer cleanupUser(t, ctx, resp2.UserID)

	// 空 code → BadParam
	if _, err := l.GitHubCallback(&types.GitHubCallbackReq{}); err == nil {
		t.Fatal("empty code should be BadParam")
	}
}
