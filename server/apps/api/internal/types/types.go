package types

type CreateShortLinkReq struct {
	LongURL string `json:"long_url"`
}

type CreateShortLinkResp struct {
	Code    string `json:"code"`
	LongURL string `json:"long_url"`
}

type GetByCodeReq struct {
	Code string `path:"code"`
}

type GetByCodeResp struct {
	Code    string `json:"code"`
	LongURL string `json:"long_url"`
	Clicks  int64  `json:"clicks"`
	Status  int32  `json:"status"`
}

type ResolveReq struct {
	Code string `path:"code"`
}

type ResolveResp struct {
	LongURL string `json:"long_url"`
	Blocked bool   `json:"blocked"`
}
