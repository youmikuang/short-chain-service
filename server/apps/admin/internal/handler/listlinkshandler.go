package handler

import (
	"net/http"

	"server/apps/admin/internal/logic"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/pkg/xhttp"
)

func ListLinksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.ListLinksReq) (*types.ListLinksResp, error) {
		return logic.NewListLinksLogic(r.Context(), svcCtx).ListLinks(req)
	})
}

func AddBlacklistHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return xhttp.Handle(func(r *http.Request, req *types.AddBlacklistReq) (*types.AddBlacklistResp, error) {
		return logic.NewAddBlacklistLogic(r.Context(), svcCtx).AddBlacklist(req)
	})
}
