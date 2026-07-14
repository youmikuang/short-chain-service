package handler

import (
	"net/http"

	"server/apps/admin/api/internal/logic"
	"server/apps/admin/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DashboardHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewDashboardLogic(r.Context(), svcCtx)
		resp, err := l.Dashboard()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
