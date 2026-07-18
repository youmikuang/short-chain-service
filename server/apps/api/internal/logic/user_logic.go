package logic

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	"server/common/errorx"
	"server/common/model"
	"server/common/tool"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// uidFromCtx 从 JWT 上下文取出用户ID
// go-zero 的 jwt 中间件会把每个 claim 写入 request context（key 为 claim 名），
// 例如 token 载荷 {"uid": 1} 可通过 ctx.Value("uid") 取到。
// 注意：go-zero v1.7.3 的 token 解析器使用了 jwt.WithJSONNumber()，
// 因此数值型 claim 在 context 中是 json.Number（而非 float64），需兼容处理。
func uidFromCtx(ctx context.Context) (int64, error) {
	v := ctx.Value("uid")
	if v == nil {
		return 0, errorx.Unauthorized("uid missing in token")
	}
	switch t := v.(type) {
	case float64:
		return int64(t), nil
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			return 0, errorx.Unauthorized("invalid uid")
		}
		return n, nil
	case int64:
		return t, nil
	case int:
		return int64(t), nil
	case string:
		n, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return 0, errorx.Unauthorized("invalid uid")
		}
		return n, nil
	default:
		return 0, errorx.Unauthorized("invalid uid")
	}
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *RegisterLogic) Register(req *types.RegisterReq) (*types.RegisterResp, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errorx.BadParam("email and password required")
	}
	if !strings.Contains(req.Email, "@") {
		return nil, errorx.BadParam("invalid email format")
	}
	if _, err := l.svcCtx.Models.User.FindOneByEmail(l.ctx, req.Email); err == nil {
		return nil, errorx.BadParam("email already registered")
	} else if !errors.Is(err, sqlx.ErrNotFound) {
		return nil, errorx.Internal(err.Error())
	}

	hashed, err := hashPassword(req.Password)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	nickname := req.Email[:strings.Index(req.Email, "@")]
	res, err := l.svcCtx.Models.User.Insert(l.ctx, &model.User{
		Email:        req.Email,
		PasswordHash: hashed,
		Nickname:     nickname,
		Status:       1,
	})
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	userID, err := res.LastInsertId()
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	token, err := issueToken(l.svcCtx.Config.Auth.AccessSecret, l.svcCtx.Config.Auth.AccessExpire, userID)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	return &types.RegisterResp{UserID: userID, Token: token}, nil
}

// ---------------------------------------------------------------------------
// Login
// ---------------------------------------------------------------------------

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginResp, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errorx.BadParam("email and password required")
	}
	user, err := l.svcCtx.Models.User.FindOneByEmail(l.ctx, req.Email)
	if errors.Is(err, sqlx.ErrNotFound) {
		return nil, errorx.Unauthorized("invalid credentials")
	} else if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	if err := checkPassword(user.PasswordHash, req.Password); err != nil {
		return nil, errorx.Unauthorized("invalid credentials")
	}
	token, err := issueToken(l.svcCtx.Config.Auth.AccessSecret, l.svcCtx.Config.Auth.AccessExpire, user.Id)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	return &types.LoginResp{Token: token, UserID: user.Id, Nickname: user.Nickname}, nil
}

// ---------------------------------------------------------------------------
// GitHub OAuth
// ---------------------------------------------------------------------------

type GitHubAuthURLLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGitHubAuthURLLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GitHubAuthURLLogic {
	return &GitHubAuthURLLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *GitHubAuthURLLogic) GitHubAuthURL(req *types.GitHubAuthURLReq) (*types.GitHubAuthURLResp, error) {
	cfg := l.svcCtx.Config.Github
	if cfg.ClientID == "" {
		return nil, errorx.Internal("github oauth not configured")
	}
	redirect := req.Redirect
	if redirect == "" {
		redirect = cfg.RedirectURL
	}
	u := "https://github.com/login/oauth/authorize?client_id=" + cfg.ClientID +
		"&redirect_uri=" + url.QueryEscape(redirect) + "&scope=read:user"
	return &types.GitHubAuthURLResp{Url: u}, nil
}

type GitHubCallbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGitHubCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GitHubCallbackLogic {
	return &GitHubCallbackLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *GitHubCallbackLogic) GitHubCallback(req *types.GitHubCallbackReq) (*types.LoginResp, error) {
	if req.Code == "" {
		return nil, errorx.BadParam("code required")
	}
	cfg := l.svcCtx.Config.Github
	if cfg.ClientID == "" {
		return nil, errorx.Internal("github oauth not configured")
	}

	// 1) 用 code 换 access_token
	tokenURL := "https://github.com/login/oauth/access_token"
	form := url.Values{
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"code":          {req.Code},
	}
	httpReq, _ := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, errorx.Internal("github token exchange failed: " + err.Error())
	}
	defer resp.Body.Close()
	var tok struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil || tok.AccessToken == "" {
		return nil, errorx.Internal("github token exchange failed: " + tok.Error)
	}

	// 2) 拉取 GitHub 用户
	userReq, _ := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	userReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	userReq.Header.Set("Accept", "application/json")
	uresp, err := client.Do(userReq)
	if err != nil {
		return nil, errorx.Internal("github user fetch failed: " + err.Error())
	}
	defer uresp.Body.Close()
	var gh struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(uresp.Body).Decode(&gh); err != nil {
		return nil, errorx.Internal("github user decode failed")
	}

	ghID := strconv.FormatInt(gh.ID, 10)
	user, err := l.svcCtx.Models.User.FindOneByGithubId(l.ctx, ghID)
	if errors.Is(err, sqlx.ErrNotFound) {
		// 首次登录：创建账号（email 可能为空，用 noreply 兜底）
		email := gh.Email
		if email == "" {
			email = gh.Login + "@users.noreply.github.com"
		}
		res, insErr := l.svcCtx.Models.User.Insert(l.ctx, &model.User{
			Email:    email,
			Nickname: gh.Login,
			GithubId: ghID,
			Avatar:   gh.AvatarURL,
			Status:   1,
		})
		if insErr != nil {
			return nil, errorx.Internal(insErr.Error())
		}
		id, _ := res.LastInsertId()
		user = &model.User{Id: id, Email: email, Nickname: gh.Login, Avatar: gh.AvatarURL}
	} else if err != nil {
		return nil, errorx.Internal(err.Error())
	}

	token, err := issueToken(l.svcCtx.Config.Auth.AccessSecret, l.svcCtx.Config.Auth.AccessExpire, user.Id)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	return &types.LoginResp{Token: token, UserID: user.Id, Nickname: user.Nickname}, nil
}

// ---------------------------------------------------------------------------
// API Key
// ---------------------------------------------------------------------------

type CreateAPIKeyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAPIKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAPIKeyLogic {
	return &CreateAPIKeyLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *CreateAPIKeyLogic) CreateAPIKey(req *types.CreateAPIKeyReq) (*types.CreateAPIKeyResp, error) {
	uid, err := uidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	if req.Name == "" {
		return nil, errorx.BadParam("name required")
	}
	// 明文 key 仅展示一次；入库存哈希
	key := "slk_" + tool.RandString(32)
	keyHash := tool.Sha256Hex(key)
	res, err := l.svcCtx.Models.ApiKey.Insert(l.ctx, &model.ApiKey{
		UserId:  uid,
		Name:    req.Name,
		KeyHash: keyHash,
		Prefix:  key[:8],
		Status:  1,
	})
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	id, _ := res.LastInsertId()
	return &types.CreateAPIKeyResp{Key: key, Name: req.Name, Id: id}, nil
}

type ListAPIKeysLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAPIKeysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAPIKeysLogic {
	return &ListAPIKeysLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ListAPIKeysLogic) ListAPIKeys(req *types.ListAPIKeysReq) (*types.ListAPIKeysResp, error) {
	uid, err := uidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	rows, err := l.svcCtx.Models.ApiKey.FindByUser(l.ctx, uid)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	items := make([]types.APIKeyItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, types.APIKeyItem{
			Id:        r.Id,
			Name:      r.Name,
			Status:    int32(r.Status),
			CreatedAt: r.CreatedAt,
		})
	}
	return &types.ListAPIKeysResp{Items: items}, nil
}

type RevokeAPIKeyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRevokeAPIKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeAPIKeyLogic {
	return &RevokeAPIKeyLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *RevokeAPIKeyLogic) RevokeAPIKey(req *types.RevokeAPIKeyReq) (*types.RevokeAPIKeyResp, error) {
	uid, err := uidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.Models.ApiKey.UpdateStatus(l.ctx, req.Id, uid, 0); err != nil {
		return nil, errorx.Internal(err.Error())
	}
	return &types.RevokeAPIKeyResp{Ok: true}, nil
}

// ---------------------------------------------------------------------------
// Profile
// ---------------------------------------------------------------------------

type GetProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProfileLogic {
	return &GetProfileLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *GetProfileLogic) GetProfile() (*types.GetProfileResp, error) {
	uid, err := uidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	user, err := l.svcCtx.Models.User.FindOneById(l.ctx, uid)
	if errors.Is(err, sqlx.ErrNotFound) {
		return nil, errorx.NotFound("user not found")
	} else if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	return &types.GetProfileResp{UserID: user.Id, Email: user.Email, Nickname: user.Nickname}, nil
}

type UpdateProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *UpdateProfileLogic) UpdateProfile(req *types.UpdateProfileReq) (*types.UpdateProfileResp, error) {
	uid, err := uidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	if req.Email != "" && !strings.Contains(req.Email, "@") {
		return nil, errorx.BadParam("invalid email format")
	}
	user, err := l.svcCtx.Models.User.FindOneById(l.ctx, uid)
	if errors.Is(err, sqlx.ErrNotFound) {
		return nil, errorx.NotFound("user not found")
	} else if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if err := l.svcCtx.Models.User.Update(l.ctx, user); err != nil {
		return nil, errorx.Internal(err.Error())
	}
	return &types.UpdateProfileResp{UserID: user.Id, Email: user.Email, Nickname: user.Nickname}, nil
}

type ChangePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordReq) (*types.ChangePasswordResp, error) {
	uid, err := uidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	if req.NewPassword == "" {
		return nil, errorx.BadParam("new_password required")
	}
	user, err := l.svcCtx.Models.User.FindOneById(l.ctx, uid)
	if errors.Is(err, sqlx.ErrNotFound) {
		return nil, errorx.NotFound("user not found")
	} else if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	if err := checkPassword(user.PasswordHash, req.CurrentPassword); err != nil {
		return nil, errorx.Unauthorized("current password incorrect")
	}
	hashed, err := hashPassword(req.NewPassword)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	user.PasswordHash = hashed
	if err := l.svcCtx.Models.User.Update(l.ctx, user); err != nil {
		return nil, errorx.Internal(err.Error())
	}
	return &types.ChangePasswordResp{Ok: true}, nil
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

type GetSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSettingsLogic {
	return &GetSettingsLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *GetSettingsLogic) GetSettings() (*types.GetSettingsResp, error) {
	uid, err := uidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	row, err := l.svcCtx.Models.UserSettings.FindOneByUserId(l.ctx, uid)
	if errors.Is(err, sqlx.ErrNotFound) {
		// 不存在则返回默认偏好
		return &types.GetSettingsResp{EmailNotif: true, SecurityAlerts: true, MarketingComm: false}, nil
	} else if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	return &types.GetSettingsResp{
		EmailNotif:     row.EmailNotif == 1,
		SecurityAlerts: row.SecurityAlerts == 1,
		MarketingComm:  row.MarketingComm == 1,
	}, nil
}

type UpdateSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSettingsLogic {
	return &UpdateSettingsLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *UpdateSettingsLogic) UpdateSettings(req *types.UpdateSettingsReq) (*types.UpdateSettingsResp, error) {
	uid, err := uidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	row := &model.UserSettings{
		UserId:         uid,
		EmailNotif:     boolToInt(req.EmailNotif),
		SecurityAlerts: boolToInt(req.SecurityAlerts),
		MarketingComm:  boolToInt(req.MarketingComm),
	}
	if err := l.svcCtx.Models.UserSettings.Upsert(l.ctx, row); err != nil {
		return nil, errorx.Internal(err.Error())
	}
	return &types.UpdateSettingsResp{
		EmailNotif:     req.EmailNotif,
		SecurityAlerts: req.SecurityAlerts,
		MarketingComm:  req.MarketingComm,
	}, nil
}

// ---------------------------------------------------------------------------
// Usage trends & logs
// ---------------------------------------------------------------------------

type UsageTrendsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUsageTrendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UsageTrendsLogic {
	return &UsageTrendsLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *UsageTrendsLogic) UsageTrends(req *types.UsageTrendsReq) (*types.UsageTrendsResp, error) {
	days := req.Days
	if days <= 0 {
		days = 30
	}
	counts, err := l.svcCtx.ClickHouseVisit.CountByDay(l.ctx, int(days))
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	// 补齐连续日期（缺失的天补 0）
	now := time.Now()
	items := make([]types.UsagePoint, 0, days)
	for i := int(days) - 1; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		items = append(items, types.UsagePoint{Day: d, Value: counts[d]})
	}
	return &types.UsageTrendsResp{Items: items}, nil
}

type LogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogsLogic {
	return &LogsLogic{ctx: ctx, svcCtx: svcCtx}
}
func (l *LogsLogic) Logs(req *types.LogsReq) (*types.LogsResp, error) {
	uid, err := uidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	rows, total, err := l.svcCtx.ClickHouseVisit.FindPageByUser(l.ctx, uid, req.Page, req.PageSize, req.Search)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	items := make([]types.LogItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, types.LogItem{
			Timestamp: r.CreatedAt.Format(time.DateTime),
			Code:      r.Code,
			LongURL:   r.LongURL,
			Status:    r.Status,
			IP:        r.IP,
		})
	}
	return &types.LogsResp{Total: total, Items: items}, nil
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
