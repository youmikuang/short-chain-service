package handler

import (
	"net/http"

	"server/apps/admin/internal/logic"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/pkg/xhttp"
)

func ListBlacklistHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.ListBlacklistReq) (*types.ListBlacklistResp, error) {
		return logic.NewListBlacklistLogic(r.Context(), svcCtx).ListBlacklist(req)
	})
}
