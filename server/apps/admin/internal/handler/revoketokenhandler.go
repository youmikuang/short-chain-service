package handler

import (
	"net/http"

	"server/apps/admin/internal/logic"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/pkg/xhttp"
)

func RevokeTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.RevokeTokenReq) (*types.RevokeTokenResp, error) {
		return logic.NewRevokeTokenLogic(r.Context(), svcCtx).RevokeToken(req)
	})
}
