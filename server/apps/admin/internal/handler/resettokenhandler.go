package handler

import (
	"net/http"

	"server/apps/admin/internal/logic"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/pkg/xhttp"
)

func ResetTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.ResetTokenReq) (*types.ResetTokenResp, error) {
		return logic.NewResetTokenLogic(r.Context(), svcCtx).ResetToken(req)
	})
}
