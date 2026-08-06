package handler

import (
	"net/http"

	"server/apps/admin/internal/logic"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/pkg/xhttp"
)

func DashboardHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.HandleNoReq(func(r *http.Request) (*types.DashboardResp, error) {
		return logic.NewDashboardLogic(r.Context(), svcCtx).Dashboard()
	})
}
