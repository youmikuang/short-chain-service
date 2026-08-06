// Package xhttp 提供 HTTP handler 的通用模板封装，
// 消除各服务 handler 中重复的「解析请求 → 调用 logic → 写响应」样板代码。
package xhttp

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// Handle 统一 go-zero handler 模板：httpx.Parse 解析请求（body/query/path），
// 成功后执行业务函数 call，按 err 写出 ErrorCtx / OkJsonCtx。
// Req/Resp 由 call 的签名自动推导，调用方无需显式指定类型参数。
func Handle[Req, Resp any](call func(r *http.Request, req *Req) (Resp, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Req
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		resp, err := call(r, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// HandleNoReq 是 Handle 的无请求参数版本（如 GetProfile / ListAPIKeys 等只依赖 JWT 上下文的接口）。
func HandleNoReq[Resp any](call func(r *http.Request) (Resp, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := call(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
