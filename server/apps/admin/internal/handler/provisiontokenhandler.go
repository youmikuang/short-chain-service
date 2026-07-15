package handler

import (
	"net/http"

	"server/apps/admin/internal/logic"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ProvisionTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ProvisionTokenReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewProvisionTokenLogic(r.Context(), svcCtx)
		resp, err := l.ProvisionToken(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
