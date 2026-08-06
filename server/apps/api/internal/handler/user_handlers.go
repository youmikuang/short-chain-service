package handler

import (
	"net/http"
	"net/url"
	"strconv"

	"server/apps/api/internal/logic"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	"server/pkg/tool"
	"server/pkg/xhttp"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.RegisterReq) (*types.RegisterResp, error) {
		return logic.NewRegisterLogic(r.Context(), svcCtx).Register(req)
	})
}

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.LoginReq) (*types.LoginResp, error) {
		return logic.NewLoginLogic(r.Context(), svcCtx).Login(req)
	})
}

// GitHubAuthURLHandler 需要在响应中下发 state Cookie（防 CSRF），不走通用模板。
func GitHubAuthURLHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GitHubAuthURLReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		// 生成随机 state 写入 HttpOnly Cookie（SameSite=Lax，10 分钟过期），
		// 用于回调时校验，防止 CSRF。GitHub 回调是顶级跳转，Lax 模式会带上该 Cookie。
		if req.State == "" {
			req.State = tool.RandString(32)
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "gh_oauth_state",
			Value:    req.State,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   600,
		})
		l := logic.NewGitHubAuthURLLogic(r.Context(), svcCtx)
		resp, err := l.GitHubAuthURL(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// githubCallbackError 把错误以 302 方式带回前端 /login?error=...，并清除一次性 state cookie。
func githubCallbackError(w http.ResponseWriter, r *http.Request, svcCtx *svc.ServiceContext, msg string) {
	http.SetCookie(w, &http.Cookie{Name: "gh_oauth_state", Path: "/", MaxAge: -1})
	q := url.Values{}
	q.Set("error", msg)
	base := svcCtx.Config.WebBaseURL
	if base == "" {
		base = "http://localhost:5173"
	}
	http.Redirect(w, r, base+"/login?"+q.Encode(), http.StatusFound)
}

// GitHubCallbackHandler 以 302 重定向（而非 JSON）响应，不走通用模板。
func GitHubCallbackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GitHubCallbackReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// CSRF 防护：校验回调的 state 与发起授权时下发的 Cookie 一致，
		// 不一致（或缺失）说明不是用户主动发起的登录，直接拒绝，绝不换 token。
		stateCookie, cerr := r.Cookie("gh_oauth_state")
		if req.State == "" || cerr != nil || stateCookie.Value != req.State {
			githubCallbackError(w, r, svcCtx, "invalid oauth state")
			return
		}
		// 校验通过，清除一次性 state cookie
		http.SetCookie(w, &http.Cookie{Name: "gh_oauth_state", Path: "/", MaxAge: -1})

		l := logic.NewGitHubCallbackLogic(r.Context(), svcCtx)
		resp, err := l.GitHubCallback(&req)

		// GitHub 以浏览器顶级跳转方式回到本回调，SPA 无法读取 JSON 响应，
		// 因此由后端完成 OAuth 交换后 302 重定向到前端 /login 并带上 token。
		q := url.Values{}
		if err != nil {
			q.Set("error", err.Error())
		} else {
			q.Set("token", resp.Token)
			q.Set("user_id", strconv.FormatInt(resp.UserID, 10))
			q.Set("nickname", resp.Nickname)
		}
		base := svcCtx.Config.WebBaseURL
		if base == "" {
			base = "http://localhost:5173"
		}
		http.Redirect(w, r, base+"/login?"+q.Encode(), http.StatusFound)
	}
}

func CreateAPIKeyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.CreateAPIKeyReq) (*types.CreateAPIKeyResp, error) {
		return logic.NewCreateAPIKeyLogic(r.Context(), svcCtx).CreateAPIKey(req)
	})
}

func ListAPIKeysHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.HandleNoReq(func(r *http.Request) (*types.ListAPIKeysResp, error) {
		return logic.NewListAPIKeysLogic(r.Context(), svcCtx).ListAPIKeys(&types.ListAPIKeysReq{})
	})
}

func RevokeAPIKeyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.RevokeAPIKeyReq) (*types.RevokeAPIKeyResp, error) {
		return logic.NewRevokeAPIKeyLogic(r.Context(), svcCtx).RevokeAPIKey(req)
	})
}

func GetProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.HandleNoReq(func(r *http.Request) (*types.GetProfileResp, error) {
		return logic.NewGetProfileLogic(r.Context(), svcCtx).GetProfile()
	})
}

func UpdateProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.UpdateProfileReq) (*types.UpdateProfileResp, error) {
		return logic.NewUpdateProfileLogic(r.Context(), svcCtx).UpdateProfile(req)
	})
}

func ChangePasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.ChangePasswordReq) (*types.ChangePasswordResp, error) {
		return logic.NewChangePasswordLogic(r.Context(), svcCtx).ChangePassword(req)
	})
}

func GetSettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.HandleNoReq(func(r *http.Request) (*types.GetSettingsResp, error) {
		return logic.NewGetSettingsLogic(r.Context(), svcCtx).GetSettings()
	})
}

func UpdateSettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.UpdateSettingsReq) (*types.UpdateSettingsResp, error) {
		return logic.NewUpdateSettingsLogic(r.Context(), svcCtx).UpdateSettings(req)
	})
}

func UsageTrendsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.UsageTrendsReq) (*types.UsageTrendsResp, error) {
		return logic.NewUsageTrendsLogic(r.Context(), svcCtx).UsageTrends(req)
	})
}

func LogsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.LogsReq) (*types.LogsResp, error) {
		return logic.NewLogsLogic(r.Context(), svcCtx).Logs(req)
	})
}
