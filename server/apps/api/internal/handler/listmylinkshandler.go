package handler

import (
	"net/http"

	"server/apps/api/internal/logic"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	"server/pkg/xhttp"
)

func ListMyLinksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.ListMyLinksReq) (*types.ListMyLinksResp, error) {
		return logic.NewListMyLinksLogic(r.Context(), svcCtx).ListMyLinks(req)
	})
}
