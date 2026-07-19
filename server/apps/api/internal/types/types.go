package types

type CreateslinkReq struct {
	LongURL string `json:"long_url"`
}

type CreateslinkResp struct {
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

// --- 用户自己的短链列表 ---
type ListMyLinksReq struct {
	Page   int64  `form:"page,optional"`
	Size   int64  `form:"size,optional"`
	Search string `form:"search,optional"`
	Sort   string `form:"sort,optional"`
}

type MyLinkItem struct {
	Code      string `json:"code"`
	SUrl      string `json:"s_url"`
	LongURL   string `json:"long_url"`
	Clicks    int64  `json:"clicks"`
	Status    int32  `json:"status"`
	Source    string `json:"source"`
	CreatedAt string `json:"created_at"`
}

type ListMyLinksResp struct {
	Total int64        `json:"total"`
	Items []MyLinkItem `json:"items"`
}
