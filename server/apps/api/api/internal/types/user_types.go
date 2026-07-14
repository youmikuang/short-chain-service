package types

type RegisterReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type RegisterResp struct {
	UserID int64  `json:"user_id"`
	Token  string `json:"token"`
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
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
type ChangePasswordResp struct {
	Ok bool `json:"ok"`
}

// --- 用户偏好设置 ---
type GetSettingsResp struct {
	EmailNotif     bool `json:"email_notif"`
	SecurityAlerts bool `json:"security_alerts"`
	MarketingComm  bool `json:"marketing_comm"`
}
type UpdateSettingsReq struct {
	EmailNotif     bool `json:"email_notif"`
	SecurityAlerts bool `json:"security_alerts"`
	MarketingComm  bool `json:"marketing_comm"`
}
type UpdateSettingsResp struct {
	EmailNotif     bool `json:"email_notif"`
	SecurityAlerts bool `json:"security_alerts"`
	MarketingComm  bool `json:"marketing_comm"`
}

// --- 用量趋势 ---
type UsageTrendsReq struct {
	Days int64 `form:"days"`
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
	Search   string `form:"search"`
	Page     int64  `form:"page"`
	PageSize int64  `form:"page_size"`
}
type LogItem struct {
	Timestamp string `json:"timestamp"`
	Endpoint  string `json:"endpoint"`
	Status    int64  `json:"status"`
	Latency   string `json:"latency"`
}
type LogsResp struct {
	Total int64      `json:"total"`
	Items []LogItem `json:"items"`
}
