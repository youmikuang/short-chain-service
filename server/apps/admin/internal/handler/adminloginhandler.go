package handler

import (
	"net/http"

	"server/apps/admin/internal/logic"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/pkg/xhttp"
)

func AdminLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.AdminLoginReq) (*types.AdminLoginResp, error) {
		return logic.NewAdminLoginLogic(r.Context(), svcCtx).AdminLogin(req)
	})
}
