package handler

import (
	"net/http"

	"server/apps/admin/internal/logic"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/pkg/xhttp"
)

func DeleteBlacklistHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.DeleteBlacklistReq) (*types.DeleteBlacklistResp, error) {
		return logic.NewDeleteBlacklistLogic(r.Context(), svcCtx).DeleteBlacklist(req)
	})
}
