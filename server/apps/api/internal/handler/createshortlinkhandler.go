package handler

import (
	"net/http"

	"server/apps/api/internal/logic"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateSlinkHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateSlinkReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewCreateSlinkLogic(r.Context(), svcCtx)
		resp, err := l.CreateSlink(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
