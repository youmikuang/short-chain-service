package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"server/apps/api/internal/config"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	pb "server/apps/rpc/pb"
	"server/pkg/errorx"
	"server/pkg/model"
	"server/pkg/xhttp"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestMain 注册统一错误处理器，使 handler 测试覆盖真实的错误映射（如黑名单 → 403）。
func TestMain(m *testing.M) {
	httpx.SetErrorHandlerCtx(xhttp.ErrorHandler)
	os.Exit(m.Run())
}

// mockSlinkClient 实现 pb.SlinkClient，仅覆盖 CreateSlink / GetByCode，供网关 handler 测试（无需启动 rpc 核心）。
type mockSlinkClient struct {
	pb.SlinkClient
	createResp *pb.CreateSlinkResp
	getResp    *pb.GetByCodeResp
	createErr  error
	getErr     error
}

func (m *mockSlinkClient) CreateSlink(ctx context.Context, in *pb.CreateSlinkReq, opts ...grpc.CallOption) (*pb.CreateSlinkResp, error) {
	return m.createResp, m.createErr
}

func (m *mockSlinkClient) GetByCode(ctx context.Context, in *pb.GetByCodeReq, opts ...grpc.CallOption) (*pb.GetByCodeResp, error) {
	return m.getResp, m.getErr
}

func loadAPIConfig(t *testing.T) config.Config {
	t.Helper()
	var c config.Config
	conf.MustLoad("../../etc/api-api.yaml", &c)
	return c
}

// newAPITestSvcMySQL 构建带 MySQL Models 的 ServiceContext（不含 ClickHouse，避免无 CH 时 panic）。
func newAPITestSvcMySQL(t *testing.T) *svc.ServiceContext {
	t.Helper()
	c := loadAPIConfig(t)
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &svc.ServiceContext{Config: c, Mysql: conn, Models: model.NewModels(conn)}
}

func TestCreateSlinkHandler_OK(t *testing.T) {
	c := loadAPIConfig(t)
	mock := &mockSlinkClient{createResp: &pb.CreateSlinkResp{Code: "abc123"}}
	svcCtx := &svc.ServiceContext{Config: c, SlinkRpc: mock}

	body, _ := json.Marshal(types.CreateSlinkReq{LongURL: "https://example.com/long"})
	req := httptest.NewRequest(http.MethodPost, "/api/short-links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	CreateSlinkHandler(svcCtx)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp types.CreateSlinkResp
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Code != "abc123" {
		t.Fatalf("Code = %q, want abc123", resp.Code)
	}
	if !bytes.HasSuffix([]byte(resp.ShortURL), []byte("/r/abc123")) {
		t.Fatalf("ShortURL = %q, want suffix /r/abc123", resp.ShortURL)
	}
}

func TestCreateSlinkHandler_BadJSON(t *testing.T) {
	c := loadAPIConfig(t)
	svcCtx := &svc.ServiceContext{Config: c, SlinkRpc: &mockSlinkClient{}}
	req := httptest.NewRequest(http.MethodPost, "/api/short-links", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	CreateSlinkHandler(svcCtx)(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGetByCodeHandler_OK(t *testing.T) {
	c := loadAPIConfig(t)
	mock := &mockSlinkClient{getResp: &pb.GetByCodeResp{Code: "abc", LongUrl: "https://example.com", Clicks: 5, Status: 1}}
	svcCtx := &svc.ServiceContext{Config: c, SlinkRpc: mock}

	req := httptest.NewRequest(http.MethodGet, "/api/short-links/abc", nil)
	req = pathvar.WithVars(req, map[string]string{"code": "abc"})
	rec := httptest.NewRecorder()
	GetByCodeHandler(svcCtx)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp types.GetByCodeResp
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.LongURL != "https://example.com" {
		t.Fatalf("LongURL = %q, want https://example.com", resp.LongURL)
	}
}

func TestListMyLinksHandler_OK(t *testing.T) {
	ctx := newAPITestSvcMySQL(t)
	r := httptest.NewRequest(http.MethodGet, "/api/short-links?page=1&size=10", nil)
	r = r.WithContext(context.WithValue(r.Context(), "uid", float64(123)))
	rec := httptest.NewRecorder()
	ListMyLinksHandler(ctx)(rec, r)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp types.ListMyLinksResp
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Total < 0 {
		t.Fatalf("Total = %d, want >= 0", resp.Total)
	}
}

// TestCreateSlinkHandler_Blacklisted 验证：rpc 核心返回黑名单拦截（gRPC PermissionDenied + 业务码 10007）
// 经统一错误处理器映射为 HTTP 403，且响应体为干净 JSON（无 "rpc error:" 前缀）。
func TestCreateSlinkHandler_Blacklisted(t *testing.T) {
	c := loadAPIConfig(t)
	// 模拟 rpc 核心透传的黑名单错误（与 errorx.Blacklisted 经 GRPCStatus 一致）
	mock := &mockSlinkClient{createErr: errorx.Blacklisted("www.zhipin.com").GRPCStatus().Err()}
	svcCtx := &svc.ServiceContext{Config: c, SlinkRpc: mock}

	body, _ := json.Marshal(types.CreateSlinkReq{LongURL: "https://www.zhipin.com/job/123"})
	req := httptest.NewRequest(http.MethodPost, "/api/short-links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	CreateSlinkHandler(svcCtx)(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v (body=%s)", err, rec.Body.String())
	}
	if resp.Code != errorx.CodeBlacklisted {
		t.Fatalf("code = %d, want %d", resp.Code, errorx.CodeBlacklisted)
	}
	if resp.Msg != "domain blacklisted: www.zhipin.com" {
		t.Fatalf("msg = %q, want clean message without gRPC prefix", resp.Msg)
	}
}

// TestCreateSlinkHandler_RPCInternal 验证：rpc 不可达（gRPC Internal）映射为 HTTP 500 而非裸 500 文本。
func TestCreateSlinkHandler_RPCInternal(t *testing.T) {
	c := loadAPIConfig(t)
	mock := &mockSlinkClient{createErr: status.Error(codes.Internal, "rpc unavailable")}
	svcCtx := &svc.ServiceContext{Config: c, SlinkRpc: mock}

	body, _ := json.Marshal(types.CreateSlinkReq{LongURL: "https://example.com/x"})
	req := httptest.NewRequest(http.MethodPost, "/api/short-links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	CreateSlinkHandler(svcCtx)(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}
