package errorx

import "errors"

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

func New(code int, msg string) *Error { return &Error{Code: code, Msg: msg} }

func Internal(msg string) *Error        { return New(CodeInternal, msg) }
func BadParam(msg string) *Error         { return New(CodeInvalidParam, msg) }
func Unauthorized(msg string) *Error     { return New(CodeUnauthorized, msg) }
func Forbidden(msg string) *Error        { return New(CodeForbidden, msg) }
func NotFound(msg string) *Error         { return New(CodeNotFound, msg) }
func RateLimited(msg string) *Error      { return New(CodeRateLimited, msg) }
func Blacklisted(domain string) *Error   { return New(CodeBlacklisted, "domain blacklisted: " + domain) }

// Is 判断是否为指定错误码
func Is(err error, target *Error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == target.Code
	}
	return false
}
