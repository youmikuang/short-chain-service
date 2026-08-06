package handler

import (
	"net/http"

	"server/apps/api/internal/logic"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	"server/pkg/xhttp"
)

func CreateSlinkHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.CreateSlinkReq) (*types.CreateSlinkResp, error) {
		return logic.NewCreateSlinkLogic(r.Context(), svcCtx).CreateSlink(req)
	})
}
