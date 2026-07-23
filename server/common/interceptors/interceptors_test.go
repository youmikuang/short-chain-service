package interceptors

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"server/common/model"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// fakeHandler 返回一个总是成功并返回固定响应的 UnaryHandler。
func fakeHandler(resp interface{}, err error) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return resp, err
	}
}

func TestUnaryServerInterceptor_ResolvePassthrough(t *testing.T) {
	ic := UnaryServerInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/shortlink.Slink/Resolve"}
	called := false
	h := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}
	resp, err := ic(context.Background(), nil, info, h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("Resolve 应放行，handler 未被调用")
	}
	if resp != "ok" {
		t.Fatalf("resp = %v, want ok", resp)
	}
}

func TestUnaryServerInterceptor_MissingMetadata(t *testing.T) {
	ic := UnaryServerInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/shortlink.Slink/CreateSlink"}
	_, err := ic(context.Background(), nil, info, fakeHandler("x", nil))
	if err == nil {
		t.Fatal("缺少 metadata 应返回错误")
	}
}

func TestUnaryServerInterceptor_InvalidToken(t *testing.T) {
	ic := UnaryServerInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/shortlink.Slink/CreateSlink"}
	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.New(map[string]string{"x-internal-token": "wrong"}))
	_, err := ic(ctx, nil, info, fakeHandler("x", nil))
	if err == nil {
		t.Fatal("错误的内部 token 应返回错误")
	}
}

func TestUnaryServerInterceptor_ValidToken(t *testing.T) {
	ic := UnaryServerInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/shortlink.Slink/CreateSlink"}
	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.New(map[string]string{"x-internal-token": InternalServiceToken}))
	called := false
	h := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}
	if _, err := ic(ctx, nil, info, h); err != nil {
		t.Fatalf("正确 token 不应报错: %v", err)
	}
	if !called {
		t.Fatal("正确 token 时 handler 未被调用")
	}
}

func TestUnaryClientInterceptor_AppendsToken(t *testing.T) {
	ic := UnaryClientInterceptor()
	var captured context.Context
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		captured = ctx
		return nil
	}
	if err := ic(context.Background(), "/shortlink.Slink/CreateSlink", nil, nil, nil, invoker); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	md, ok := metadata.FromOutgoingContext(captured)
	if !ok {
		t.Fatal("outgoing context 缺少 metadata")
	}
	toks := md.Get("x-internal-token")
	if len(toks) == 0 || toks[0] != InternalServiceToken {
		t.Fatalf("x-internal-token = %v, want [%s]", toks, InternalServiceToken)
	}
}

func TestUnaryServerLogInterceptor(t *testing.T) {
	// 使用惰性 *sql.DB（不会真正连接），Insert 在 goroutine 中失败被忽略，不阻塞请求。
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/shortlink?parseTime=true")
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	logModel := model.NewRpcLogModel(db)

	ic := UnaryServerLogInterceptor(logModel)
	info := &grpc.UnaryServerInfo{FullMethod: "/shortlink.Slink/CreateSlink"}

	// 成功路径：应透传响应，不阻塞。
	resp, err := ic(context.Background(), nil, info, fakeHandler("ok", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("resp = %v, want ok", resp)
	}

	// 错误路径：应透传错误，且错误码被记录（status 映射）。
	target := errors.New("boom")
	_, err = ic(context.Background(), nil, info, fakeHandler(nil, target))
	if err == nil {
		t.Fatal("错误应被透传")
	}
}
