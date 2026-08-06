package errorx

import (
	"errors"
	"strconv"
	"testing"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorCodes(t *testing.T) {
	cases := []struct {
		err  error
		code int
	}{
		{Internal("x"), CodeInternal},
		{BadParam("x"), CodeInvalidParam},
		{Unauthorized("x"), CodeUnauthorized},
		{Forbidden("x"), CodeForbidden},
		{NotFound("x"), CodeNotFound},
		{RateLimited("x"), CodeRateLimited},
		{Blacklisted("evil.com"), CodeBlacklisted},
	}
	for _, c := range cases {
		var e *Error
		if !errors.As(c.err, &e) {
			t.Fatalf("%v is not *Error", c.err)
		}
		if e.Code != c.code {
			t.Fatalf("got code %d, want %d", e.Code, c.code)
		}
		if e.Error() == "" {
			t.Fatalf("error message empty for code %d", c.code)
		}
	}
}

func TestIs(t *testing.T) {
	if !Is(Internal("a"), Internal("b")) {
		t.Fatal("Is should match same code")
	}
	if Is(NotFound("a"), Internal("b")) {
		t.Fatal("Is should not match different code")
	}
	if Is(errors.New("plain"), Internal("b")) {
		t.Fatal("Is should not match non-*Error")
	}
	// 包装后仍可识别
	wrapped := errors.Join(Internal("inner"), errors.New("other"))
	if !Is(wrapped, Internal("x")) {
		t.Fatal("Is should find *Error inside joined error")
	}
}

func TestBlacklistedMessage(t *testing.T) {
	err := Blacklisted("evil.com")
	if err.Error() != "domain blacklisted: evil.com" {
		t.Fatalf("unexpected message: %q", err.Error())
	}
}

// TestGRPCStatus 验证业务错误经 gRPC 边界的映射：
//   - 业务码 → 标准 gRPC code（Blacklisted → PermissionDenied）；
//   - 原始业务码通过 errdetails.ErrorInfo(biz_code) 透传，供网关还原。
func TestGRPCStatus(t *testing.T) {
	cases := []struct {
		err  error
		code codes.Code
		biz  int
	}{
		{Blacklisted("evil.com"), codes.PermissionDenied, CodeBlacklisted},
		{Forbidden("no"), codes.PermissionDenied, CodeForbidden},
		{Unauthorized("nope"), codes.Unauthenticated, CodeUnauthorized},
		{BadParam("bad"), codes.InvalidArgument, CodeInvalidParam},
		{NotFound("x"), codes.NotFound, CodeNotFound},
		{RateLimited("slow"), codes.ResourceExhausted, CodeRateLimited},
		{Internal("boom"), codes.Internal, CodeInternal},
	}
	for _, c := range cases {
		st := status.Convert(c.err)
		if st.Code() != c.code {
			t.Fatalf("%v: grpc code = %v, want %v", c.err, st.Code(), c.code)
		}
		got := 0
		for _, d := range st.Details() {
			if ei, ok := d.(*errdetails.ErrorInfo); ok {
				if v, ok := ei.Metadata["biz_code"]; ok {
					n, _ := strconv.Atoi(v)
					got = n
				}
			}
		}
		if got != c.biz {
			t.Fatalf("%v: biz_code = %d, want %d", c.err, got, c.biz)
		}
	}
}
