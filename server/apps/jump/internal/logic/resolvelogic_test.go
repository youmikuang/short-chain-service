package logic

import (
	"context"
	"testing"

	"server/apps/jump/internal/svc"
	"server/apps/jump/internal/types"
	pb "server/apps/rpc/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// mockSlinkClient 实现 pb.SlinkClient 接口，用于 jump 服务的单元测试（无需启动 rpc 核心）。
// 仅覆盖 Resolve；其余方法由内嵌的 nil 接口提供（测试中不会调用）。
type mockSlinkClient struct {
	pb.SlinkClient
	lastReq *pb.ResolveReq
	lastCtx context.Context
	resp    *pb.ResolveResp
	err     error
}

func (m *mockSlinkClient) Resolve(ctx context.Context, in *pb.ResolveReq, opts ...grpc.CallOption) (*pb.ResolveResp, error) {
	m.lastCtx = ctx
	m.lastReq = in
	return m.resp, m.err
}

func newJumpTestSvc(client pb.SlinkClient) *svc.ServiceContext {
	return &svc.ServiceContext{SlinkRpc: client}
}

func TestJumpResolve(t *testing.T) {
	mock := &mockSlinkClient{
		resp: &pb.ResolveResp{LongUrl: "https://example.com/x", Blocked: false},
	}
	l := NewResolveLogic(context.Background(), newJumpTestSvc(mock))

	resp, err := l.Resolve(&types.ResolveReq{Code: "abc123"})
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if resp.LongURL != "https://example.com/x" {
		t.Fatalf("LongURL = %q, want https://example.com/x", resp.LongURL)
	}
	if resp.Blocked {
		t.Fatal("unexpected blocked")
	}
	if mock.lastReq.GetCode() != "abc123" {
		t.Fatalf("rpc received code %q, want abc123", mock.lastReq.GetCode())
	}
}

func TestJumpResolveBlocked(t *testing.T) {
	mock := &mockSlinkClient{resp: &pb.ResolveResp{Blocked: true}}
	l := NewResolveLogic(context.Background(), newJumpTestSvc(mock))
	resp, err := l.Resolve(&types.ResolveReq{Code: "bad"})
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if !resp.Blocked {
		t.Fatal("expected Blocked=true")
	}
}

func TestJumpResolveMetadata(t *testing.T) {
	mock := &mockSlinkClient{resp: &pb.ResolveResp{LongUrl: "https://e.com"}}
	l := NewResolveLogic(context.Background(), newJumpTestSvc(mock))

	_, err := l.Resolve(&types.ResolveReq{Code: "abc", Ip: "1.2.3.4", Referer: "https://refer.com"})
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	// 校验 metadata 透传（jump 把 IP / Referer 写入 gRPC metadata）
	md, ok := metadata.FromOutgoingContext(mock.lastCtx)
	if !ok {
		t.Fatal("metadata not attached to outgoing context")
	}
	if got := md.Get("x-client-ip"); len(got) == 0 || got[0] != "1.2.3.4" {
		t.Fatalf("x-client-ip metadata = %v, want [1.2.3.4]", got)
	}
	if got := md.Get("x-referer"); len(got) == 0 || got[0] != "https://refer.com" {
		t.Fatalf("x-referer metadata = %v, want [https://refer.com]", got)
	}
}

func TestJumpResolveRPCError(t *testing.T) {
	mock := &mockSlinkClient{err: context.Canceled}
	l := NewResolveLogic(context.Background(), newJumpTestSvc(mock))
	if _, err := l.Resolve(&types.ResolveReq{Code: "abc"}); err == nil {
		t.Fatal("expected rpc error to propagate")
	}
}
