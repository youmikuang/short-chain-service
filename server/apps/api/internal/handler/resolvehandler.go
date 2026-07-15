package handler

import (
	"net/http"

	"server/apps/api/internal/logic"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	"server/common/errorx"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ResolveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ResolveReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewResolveLogic(r.Context(), svcCtx)
		resp, err := l.Resolve(&req)
		if err != nil {
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
