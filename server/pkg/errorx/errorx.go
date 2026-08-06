package errorx

import (
	"errors"
	"strconv"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 统一业务错误码
const (
	CodeOK           = 0
	CodeInternal     = 10001
	CodeInvalidParam = 10002
	CodeUnauthorized = 10003
	CodeForbidden    = 10004
	CodeNotFound     = 10005
	CodeRateLimited  = 10006
	CodeBlacklisted  = 10007
)

// Error 统一错误结构
type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string { return e.Msg }

// GRPCStatus 将业务错误转换为 gRPC status：
//   - 业务码 → 标准 gRPC code（Blacklisted/Forbidden → PermissionDenied 等），
//     网关侧可据此映射出正确的 HTTP 状态码（如黑名单拦截 → 403，而非 500）；
//   - 通过 errdetails.ErrorInfo 携带原始业务码（biz_code），
//     网关侧据此还原出精确的 ErrorBody.Code（如 10007），而 msg 保持干净、
//     不出现 "rpc error: code = Unknown desc = ..." 这类 gRPC 前缀。
func (e *Error) GRPCStatus() *status.Status {
	st := status.New(grpcCodeOfBiz(e.Code), e.Msg)
	if d, err := st.WithDetails(&errdetails.ErrorInfo{
		Metadata: map[string]string{"biz_code": strconv.Itoa(e.Code)},
	}); err == nil {
		return d
	}
	return st
}

// grpcCodeOfBiz 业务错误码 → 标准 gRPC code
func grpcCodeOfBiz(code int) codes.Code {
	switch code {
	case CodeInvalidParam:
		return codes.InvalidArgument
	case CodeUnauthorized:
		return codes.Unauthenticated
	case CodeForbidden, CodeBlacklisted:
		return codes.PermissionDenied
	case CodeNotFound:
		return codes.NotFound
	case CodeRateLimited:
		return codes.ResourceExhausted
	case CodeInternal:
		return codes.Internal
	default:
		return codes.Internal
	}
}

func New(code int, msg string) *Error { return &Error{Code: code, Msg: msg} }

func Internal(msg string) *Error       { return New(CodeInternal, msg) }
func BadParam(msg string) *Error       { return New(CodeInvalidParam, msg) }
func Unauthorized(msg string) *Error   { return New(CodeUnauthorized, msg) }
func Forbidden(msg string) *Error      { return New(CodeForbidden, msg) }
func NotFound(msg string) *Error       { return New(CodeNotFound, msg) }
func RateLimited(msg string) *Error    { return New(CodeRateLimited, msg) }
func Blacklisted(domain string) *Error { return New(CodeBlacklisted, "domain blacklisted: "+domain) }

// Is 判断是否为指定错误码
func Is(err error, target *Error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == target.Code
	}
	return false
}
