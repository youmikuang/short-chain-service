package handler

import (
	"net"
	"net/http"
	"strings"

	"server/apps/jump/internal/logic"
	"server/apps/jump/internal/svc"
	"server/apps/jump/internal/types"
	"server/common/errorx"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// clientIP 从 HTTP 请求提取真实客户端 IP（兼容反向代理 X-Forwarded-For / X-Real-IP）
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.IndexByte(xff, ','); idx >= 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// buildShortURL 根据当前请求还原本次访问的短链完整地址，
// 例如 code=fViTMk 时返回 https://s.gaoheng.top/r/fViTMk。
// 优先使用 X-Forwarded-Proto / 反向代理透传的 scheme 与 host，回退到请求自身信息。
func buildShortURL(r *http.Request, code string) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	return scheme + "://" + host + "/r/" + code
}

func ResolveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ResolveReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		// 提取访问者 IP；Referer 记录为本次访问的短链地址本身（如 https://s.gaoheng.top/r/fViTMk）
		req.Ip = clientIP(r)
		req.Referer = buildShortURL(r, req.Code)
		l := logic.NewResolveLogic(r.Context(), svcCtx)
		resp, err := l.Resolve(&req)
		if err != nil {
			logx.Errorf("jump resolve failed for code %s: %v", req.Code, err)
			// 短码不存在时 rpc 返回 "code not found"，这里映射为 404（而非默认 500），
			// 便于前端 / CDN 做缓存与重试判断；其余错误（如 rpc 不可达）维持 500。
			if strings.Contains(err.Error(), "not found") {
				httpx.WriteJson(w, http.StatusNotFound, map[string]any{"message": "short link not found"})
				return
			}
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if resp.Blocked {
			httpx.ErrorCtx(r.Context(), w, errorx.New(errorx.CodeBlacklisted, "link blocked by blacklist"))
			return
		}
		// 302 跳转
		http.Redirect(w, r, resp.LongURL, http.StatusFound)
	}
}
