package xhttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"server/pkg/errorx"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorBody 统一错误响应体
type ErrorBody struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// ErrorHandler 统一 HTTP 错误响应：
//   - *errorx.Error（本地业务错误）→ 按业务码映射 HTTP 状态；
//   - gRPC status 错误（rpc 透传）→ 按 gRPC code 映射 HTTP 状态，msg 取干净的 desc
//     （避免把 "rpc error: code = Unknown desc = ..." 原样吐给前端）；
//   - 其他错误 → 400。
//
// 通过 httpx.SetErrorHandlerCtx(xhttp.ErrorHandler) 在各服务 main 中注册。
func ErrorHandler(_ context.Context, err error) (int, any) {
	var e *errorx.Error
	if errors.As(err, &e) {
		return httpStatusOfBiz(e.Code), ErrorBody{Code: e.Code, Msg: e.Msg}
	}
	if st, ok := status.FromError(err); ok && st.Code() != codes.OK {
		code := bizCodeFromDetails(st)
		if code == 0 {
			code = bizCodeOfGRPC(st.Code())
		}
		return httpStatusOfGRPC(st.Code()), ErrorBody{Code: code, Msg: st.Message()}
	}
	return http.StatusBadRequest, ErrorBody{Code: errorx.CodeInvalidParam, Msg: err.Error()}
}

// httpStatusOfBiz 业务错误码 → HTTP 状态码
func httpStatusOfBiz(code int) int {
	switch code {
	case errorx.CodeInvalidParam:
		return http.StatusBadRequest
	case errorx.CodeUnauthorized:
		return http.StatusUnauthorized
	case errorx.CodeForbidden, errorx.CodeBlacklisted:
		return http.StatusForbidden
	case errorx.CodeNotFound:
		return http.StatusNotFound
	case errorx.CodeRateLimited:
		return http.StatusTooManyRequests
	case errorx.CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}

// httpStatusOfGRPC gRPC code → HTTP 状态码（与 go-zero rest/internal/errcode 保持一致）
func httpStatusOfGRPC(code codes.Code) int {
	switch code {
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.AlreadyExists, codes.Aborted:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// bizCodeOfGRPC gRPC code → 业务错误码（用于响应体 code 字段）
func bizCodeOfGRPC(code codes.Code) int {
	switch code {
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return errorx.CodeInvalidParam
	case codes.Unauthenticated:
		return errorx.CodeUnauthorized
	case codes.PermissionDenied:
		return errorx.CodeForbidden
	case codes.NotFound:
		return errorx.CodeNotFound
	case codes.ResourceExhausted:
		return errorx.CodeRateLimited
	default:
		return errorx.CodeInternal
	}
}

// bizCodeFromDetails 从 gRPC status 的 errdetails.ErrorInfo 中提取原始业务码（biz_code）。
// 仅当 rpc 返回的是本项目的 *errorx.Error（带 detail）时才有值，用于还原精确的 ErrorBody.Code。
func bizCodeFromDetails(st *status.Status) int {
	for _, d := range st.Details() {
		if ei, ok := d.(*errdetails.ErrorInfo); ok {
			if v, ok := ei.Metadata["biz_code"]; ok {
				if n, err := strconv.Atoi(v); err == nil {
					return n
				}
			}
		}
	}
	return 0
}
