package handler

import (
	"net/http"

	"server/apps/api/internal/logic"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	"server/pkg/xhttp"
)

func GetByCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.GetByCodeReq) (*types.GetByCodeResp, error) {
		return logic.NewGetByCodeLogic(r.Context(), svcCtx).GetByCode(req)
	})
}
