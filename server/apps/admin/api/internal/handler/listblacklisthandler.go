package handler

import (
	"net/http"

	"server/apps/admin/api/internal/logic"
	"server/apps/admin/api/internal/svc"
	"server/apps/admin/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListBlacklistHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListBlacklistReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewListBlacklistLogic(r.Context(), svcCtx)
		resp, err := l.ListBlacklist(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
