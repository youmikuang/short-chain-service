package handler

import (
	"net/http"

	"server/apps/admin/internal/logic"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RevokeTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RevokeTokenReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewRevokeTokenLogic(r.Context(), svcCtx)
		resp, err := l.RevokeToken(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
