package handler

import (
	"net/http"
	"net/url"
	"strconv"

	"server/apps/api/internal/logic"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	"server/common/tool"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegisterReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewRegisterLogic(r.Context(), svcCtx)
		resp, err := l.Register(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

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
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateAPIKeyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewCreateAPIKeyLogic(r.Context(), svcCtx)
		resp, err := l.CreateAPIKey(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func ListAPIKeysHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewListAPIKeysLogic(r.Context(), svcCtx)
		resp, err := l.ListAPIKeys(&types.ListAPIKeysReq{})
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func RevokeAPIKeyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RevokeAPIKeyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewRevokeAPIKeyLogic(r.Context(), svcCtx)
		resp, err := l.RevokeAPIKey(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func GetProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetProfileLogic(r.Context(), svcCtx)
		resp, err := l.GetProfile()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func UpdateProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateProfileReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewUpdateProfileLogic(r.Context(), svcCtx)
		resp, err := l.UpdateProfile(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func ChangePasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChangePasswordReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewChangePasswordLogic(r.Context(), svcCtx)
		resp, err := l.ChangePassword(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func GetSettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetSettingsLogic(r.Context(), svcCtx)
		resp, err := l.GetSettings()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func UpdateSettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateSettingsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewUpdateSettingsLogic(r.Context(), svcCtx)
		resp, err := l.UpdateSettings(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func UsageTrendsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UsageTrendsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewUsageTrendsLogic(r.Context(), svcCtx)
		resp, err := l.UsageTrends(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func LogsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LogsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewLogsLogic(r.Context(), svcCtx)
		resp, err := l.Logs(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
