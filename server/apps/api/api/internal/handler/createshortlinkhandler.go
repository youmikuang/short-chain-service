package handler

import (
	"net/http"

	"server/apps/api/api/internal/logic"
	"server/apps/api/api/internal/svc"
	"server/apps/api/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateShortLinkHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateShortLinkReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewCreateShortLinkLogic(r.Context(), svcCtx)
		resp, err := l.CreateShortLink(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
