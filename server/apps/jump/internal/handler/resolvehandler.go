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

// cleanIP 清洗单个 IP 字符串：去掉 IPv6 的方括号 []、端口号，仅保留纯 IP。
// 支持 [ipv6]:port、[ipv6]、host:port 以及裸 IP 形式。
func cleanIP(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	// IPv6 形式 [ipv6]:port 或 [ipv6]
	if strings.HasPrefix(s, "[") {
		if idx := strings.LastIndexByte(s, ']'); idx >= 0 {
			return s[1:idx]
		}
	}
	// 其余情况用 SplitHostPort 去掉端口（host:port）
	if host, _, err := net.SplitHostPort(s); err == nil {
		return host
	}
	return s
}

// clientIP 提取真实客户端源 IP。
// 优先从反向代理 / CDN 透传的头部取（兼容 Nginx、Cloudflare、常规代理等多种部署），
// 都没有时回退到 TCP 连接的 RemoteAddr。
func clientIP(r *http.Request) string {
	// 候选头部按优先级排列：越靠前越接近真实客户端。
	// 不同部署链（Nginx / Cloudflare / 其他 CDN）透传的真实 IP 所在头部不同，全部兜底扫描。
	candidates := []string{
		r.Header.Get("X-Forwarded-For"),  // 多代理时为 "client, proxy1, proxy2"
		r.Header.Get("X-Real-IP"),        // Nginx 透传
		r.Header.Get("X-Client-IP"),      // 常规代理
		r.Header.Get("CF-Connecting-IP"), // Cloudflare
		r.Header.Get("True-Client-IP"),   // Cloudflare / Akamai
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		// X-Forwarded-For 可能含多个，取第一个（最原始客户端）
		if idx := strings.IndexByte(c, ','); idx >= 0 {
			c = c[:idx]
		}
		if ip := cleanIP(c); ip != "" {
			return ip
		}
	}
	// 标准 Forwarded 头：for=<client>;proto=https
	if fwd := r.Header.Get("Forwarded"); fwd != "" {
		if ip := parseForwardedFor(fwd); ip != "" {
			return ip
		}
	}
	// 最后回退到 TCP 连接地址（直连、无代理时就是真实客户端 IP）
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr 可能是裸 IP（无端口），去掉 IPv6 方括号
		return cleanIP(r.RemoteAddr)
	}
	return host
}

// parseForwardedFor 从标准 Forwarded 头解析第一个 for= 的客户端 IP
func parseForwardedFor(fwd string) string {
	for part := range strings.SplitSeq(fwd, ";") {
		part = strings.TrimSpace(part)
		if !strings.HasPrefix(strings.ToLower(part), "for=") {
			continue
		}
		v := strings.TrimSpace(part[len("for="):])
		v = strings.Trim(v, `"`)
		if idx := strings.IndexByte(v, ','); idx >= 0 {
			v = v[:idx]
		}
		return cleanIP(v)
	}
	return ""
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
