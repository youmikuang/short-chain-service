package handler

import (
	"net/http"

	"server/apps/admin/internal/logic"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/pkg/xhttp"
)

func ProvisionTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.ProvisionTokenReq) (*types.ProvisionTokenResp, error) {
		return logic.NewProvisionTokenLogic(r.Context(), svcCtx).ProvisionToken(req)
	})
}
