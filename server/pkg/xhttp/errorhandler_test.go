package xhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"server/pkg/errorx"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcErr 构造一个带 ErrorInfo(biz_code) 的 gRPC status 错误，模拟 rpc 核心返回的本项目业务错误。
func grpcErr(t *testing.T, code codes.Code, bizCode int, msg string) error {
	t.Helper()
	st := status.New(code, msg)
	st, err := st.WithDetails(&errdetails.ErrorInfo{
		Metadata: map[string]string{"biz_code": strconv.Itoa(bizCode)},
	})
	if err != nil {
		t.Fatalf("WithDetails: %v", err)
	}
	return st.Err()
}

func decode(t *testing.T, body []byte) ErrorBody {
	t.Helper()
	var e ErrorBody
	if err := json.Unmarshal(body, &e); err != nil {
		t.Fatalf("unmarshal ErrorBody: %v (body=%s)", err, body)
	}
	return e
}

func TestErrorHandler_LocalErrorx(t *testing.T) {
	cases := []struct {
		err  error
		http int
		code int
	}{
		{errorx.Blacklisted("evil.com"), http.StatusForbidden, errorx.CodeBlacklisted},
		{errorx.Forbidden("no"), http.StatusForbidden, errorx.CodeForbidden},
		{errorx.Unauthorized("nope"), http.StatusUnauthorized, errorx.CodeUnauthorized},
		{errorx.BadParam("bad"), http.StatusBadRequest, errorx.CodeInvalidParam},
		{errorx.NotFound("x"), http.StatusNotFound, errorx.CodeNotFound},
		{errorx.RateLimited("slow"), http.StatusTooManyRequests, errorx.CodeRateLimited},
		{errorx.Internal("boom"), http.StatusInternalServerError, errorx.CodeInternal},
	}
	for _, c := range cases {
		code, body := ErrorHandler(context.Background(), c.err)
		if code != c.http {
			t.Fatalf("%v: http = %d, want %d", c.err, code, c.http)
		}
		if e := decode(t, mustJSON(body)); e.Code != c.code {
			t.Fatalf("%v: body code = %d, want %d", c.err, e.Code, c.code)
		}
		if e := decode(t, mustJSON(body)); e.Msg == "" {
			t.Fatalf("%v: empty msg", c.err)
		}
	}
}

func TestErrorHandler_GRPCBlacklisted(t *testing.T) {
	// 模拟 rpc 核心返回黑名单拦截：PermissionDenied + 业务码 10007
	err := grpcErr(t, codes.PermissionDenied, errorx.CodeBlacklisted, "domain blacklisted: www.zhipin.com")
	code, body := ErrorHandler(context.Background(), err)
	if code != http.StatusForbidden {
		t.Fatalf("http = %d, want 403", code)
	}
	e := decode(t, mustJSON(body))
	if e.Code != errorx.CodeBlacklisted {
		t.Fatalf("code = %d, want 10007", e.Code)
	}
	if e.Msg != "domain blacklisted: www.zhipin.com" {
		t.Fatalf("msg = %q, want clean message without gRPC prefix", e.Msg)
	}
}

func TestErrorHandler_GRPCUnknown(t *testing.T) {
	// 未携带 detail 的 Unknown（兜底）：HTTP 500，msg 保持干净
	err := status.Error(codes.Unknown, "some internal failure")
	code, body := ErrorHandler(context.Background(), err)
	if code != http.StatusInternalServerError {
		t.Fatalf("http = %d, want 500", code)
	}
	e := decode(t, mustJSON(body))
	if e.Code != errorx.CodeInternal {
		t.Fatalf("code = %d, want 10001", e.Code)
	}
	if e.Msg == "" {
		t.Fatal("empty msg")
	}
}

func TestErrorHandler_PlainError(t *testing.T) {
	// 非 errorx、非 gRPC status 的普通错误 → 400 + CodeInvalidParam
	err := assertErr{}
	code, body := ErrorHandler(context.Background(), err)
	if code != http.StatusBadRequest {
		t.Fatalf("http = %d, want 400", code)
	}
	e := decode(t, mustJSON(body))
	if e.Code != errorx.CodeInvalidParam {
		t.Fatalf("code = %d, want 10002", e.Code)
	}
}

type assertErr struct{}

func (assertErr) Error() string { return "plain error" }

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
