package handler

import (
	"net/http"

	"server/apps/admin/internal/logic"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/pkg/xhttp"
)

func StartTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.StartTokenReq) (*types.StartTokenResp, error) {
		return logic.NewStartTokenLogic(r.Context(), svcCtx).StartToken(req)
	})
}
