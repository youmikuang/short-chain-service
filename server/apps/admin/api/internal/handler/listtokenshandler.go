package handler

import (
	"net/http"

	"server/apps/admin/api/internal/logic"
	"server/apps/admin/api/internal/svc"
	"server/apps/admin/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListTokensHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListTokensReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewListTokensLogic(r.Context(), svcCtx)
		resp, err := l.ListTokens(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
