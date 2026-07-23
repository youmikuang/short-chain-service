package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"server/apps/api/internal/config"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	"server/common/model"
	pb "server/apps/rpc/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest/pathvar"
	"google.golang.org/grpc"
)

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
