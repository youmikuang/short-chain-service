package handler

import (
	"net/http"

	"server/apps/admin/internal/logic"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/pkg/xhttp"
)

func ListTokensHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.ListTokensReq) (*types.ListTokensResp, error) {
		return logic.NewListTokensLogic(r.Context(), svcCtx).ListTokens(req)
	})
}
