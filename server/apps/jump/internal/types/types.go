package types

// ResolveReq 短链跳转请求：code 取路径参数，Ip / Referer 由 handler 从 HTTP 请求提取，
// 经 gRPC metadata 透传给 rpc（不进 JSON 编解码）。
type ResolveReq struct {
	Code    string `path:"code"`
	Ip      string `json:"-"`
	Referer string `json:"-"`
}

// ResolveResp 跳转响应：LongURL 为目标长链，Blocked 表示命中域名黑名单不跳转。
type ResolveResp struct {
	LongURL string `json:"long_url"`
	Blocked bool   `json:"blocked"`
}
