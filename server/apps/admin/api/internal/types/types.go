package types

// ---------------------------------------------------------------------------
// 通用
// ---------------------------------------------------------------------------
type AdminLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminLoginResp struct {
	Token string `json:"token"`
}

// ---------------------------------------------------------------------------
// Dashboard
// ---------------------------------------------------------------------------
type KpiItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
	Badge string `json:"badge"`
}

type TrafficPoint struct {
	Date  string `json:"date"`
	Value int64  `json:"value"`
}

type AdminActionItem struct {
	Title string `json:"title"`
	Meta  string `json:"meta"`
	Time  string `json:"time"`
}

type DashboardResp struct {
	Kpis    []KpiItem        `json:"kpis"`
	Traffic []TrafficPoint   `json:"traffic"`
	Actions []AdminActionItem `json:"actions"`
}

// ---------------------------------------------------------------------------
// Links
// ---------------------------------------------------------------------------
type ListLinksReq struct {
	Page int64 `form:"page"`
	Size int64 `form:"size"`
}

type LinkItem struct {
	Code      string `json:"code"`
	LongURL   string `json:"long_url"`
	ShortURL  string `json:"short_url"`
	Clicks    int64  `json:"clicks"`
	Status    int32  `json:"status"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
	CreatedAt string `json:"created_at"`
}

type ListLinksResp struct {
	Total int64      `json:"total"`
	Items []LinkItem `json:"items"`
}

// ---------------------------------------------------------------------------
// Blacklist
// ---------------------------------------------------------------------------
type ListBlacklistReq struct {
	Page int64 `form:"page"`
	Size int64 `form:"size"`
}

type BlacklistItem struct {
	Domain    string `json:"domain"`
	Reason    string `json:"reason"`
	Attempts  int64  `json:"attempts"`
	CreatedAt string `json:"created_at"`
}

type ListBlacklistResp struct {
	Total int64            `json:"total"`
	Items []BlacklistItem  `json:"items"`
}

type AddBlacklistReq struct {
	Domain string `json:"domain"`
	Reason string `json:"reason"`
}

type AddBlacklistResp struct {
	Ok bool `json:"ok"`
}

// ---------------------------------------------------------------------------
// Tokens
// ---------------------------------------------------------------------------
type ListTokensReq struct {
	Page int64 `form:"page"`
	Size int64 `form:"size"`
}

type TokenItem struct {
	Id         int64  `json:"id"`
	TokenId    string `json:"token_id"`
	UserName   string `json:"user_name"`
	UserEmail  string `json:"user_email"`
	UsageLimit int64  `json:"usage_limit"`
	Remaining  int64  `json:"remaining"`
	CreatedAt  string `json:"created_at"`
	Status     int32  `json:"status"`
}

type ListTokensResp struct {
	Total int64       `json:"total"`
	Items []TokenItem `json:"items"`
}

type ProvisionTokenReq struct {
	UserId int64  `json:"user_id"`
	Name   string `json:"name"`
	Quota  int64  `json:"quota"`
}

type ProvisionTokenResp struct {
	Ok      bool   `json:"ok"`
	TokenId string `json:"token_id"`
	Token   string `json:"token"`
}

type RevokeTokenReq struct {
	Id int64 `json:"id"`
}

type RevokeTokenResp struct {
	Ok bool `json:"ok"`
}
