package types

type RegisterReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type RegisterResp struct {
	UserID int64  `json:"user_id"`
	Token  string `json:"token"`
	ApiKey string `json:"api_key"`
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginResp struct {
	Token    string `json:"token"`
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
}

type GitHubAuthURLReq struct {
	Redirect string `form:"redirect"`
}
type GitHubAuthURLResp struct {
	Url string `json:"url"`
}

type GitHubCallbackReq struct {
	Code  string `form:"code"`
	State string `form:"state"`
}

type CreateAPIKeyReq struct {
	Name string `json:"name"`
}
type CreateAPIKeyResp struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Id   int64  `json:"id"`
}

type ListAPIKeysReq struct{}
type APIKeyItem struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	Status    int32  `json:"status"`
	CreatedAt string `json:"created_at"`
}
type ListAPIKeysResp struct {
	Items []APIKeyItem `json:"items"`
}

type RevokeAPIKeyReq struct {
	Id int64 `path:"id"`
}
type RevokeAPIKeyResp struct {
	Ok bool `json:"ok"`
}

type GetProfileResp struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

// --- 资料更新 ---
type UpdateProfileReq struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}
type UpdateProfileResp struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

// --- 修改密码 ---
type ChangePasswordReq struct {
	CurrentPassword string `json:"currentPassword,optional"`
	NewPassword     string `json:"newPassword,optional"`
}
type ChangePasswordResp struct {
	Ok bool `json:"ok"`
}

// --- 用户偏好设置 ---
type GetSettingsResp struct {
	EmailNotif     bool `json:"emailNotif"`
	SecurityAlerts bool `json:"securityAlerts"`
	MarketingComm  bool `json:"marketingComm"`
}
type UpdateSettingsReq struct {
	EmailNotif     bool `json:"emailNotif,optional"`
	SecurityAlerts bool `json:"securityAlerts,optional"`
	MarketingComm  bool `json:"marketingComm,optional"`
}
type UpdateSettingsResp struct {
	EmailNotif     bool `json:"emailNotif"`
	SecurityAlerts bool `json:"securityAlerts"`
	MarketingComm  bool `json:"marketingComm"`
}

// --- 用量趋势 ---
type UsageTrendsReq struct {
	Days int64 `form:"days,optional"`
}
type UsagePoint struct {
	Day   string `json:"day"`
	Value int64  `json:"value"`
}
type UsageTrendsResp struct {
	Items []UsagePoint `json:"items"`
}

// --- 访问日志 ---
type LogsReq struct {
	Search   string `form:"search,optional"`
	Source   string `form:"source,optional"`
	Page     int64  `form:"page,optional"`
	PageSize int64  `form:"page_size,optional"`
}
type LogItem struct {
	Timestamp string `json:"timestamp"`
	Code      string `json:"code"`
	LongURL   string `json:"long_url"`
	Status    int64  `json:"status"`
	IP        string `json:"ip"`
	Source    string `json:"source"`
}
type LogsResp struct {
	Total int64      `json:"total"`
	Items []LogItem `json:"items"`
}
