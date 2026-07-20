package handler

import (
	"net/http"

	"server/apps/admin/internal/logic"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func StartTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.StartTokenReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewStartTokenLogic(r.Context(), svcCtx)
		resp, err := l.StartToken(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
