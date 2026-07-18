package handler

import (
	"net/http"

	"server/apps/api/internal/logic"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListMyLinksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListMyLinksReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewListMyLinksLogic(r.Context(), svcCtx)
		resp, err := l.ListMyLinks(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
