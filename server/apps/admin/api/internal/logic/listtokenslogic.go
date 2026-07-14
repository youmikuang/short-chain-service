package logic

import (
	"context"
	"server/apps/admin/api/internal/svc"
	"server/apps/admin/api/internal/types"
	"server/common/errorx"
)

type ListTokensLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTokensLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTokensLogic {
	return &ListTokensLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListTokens API Key 分页列表（联表用户信息）
func (l *ListTokensLogic) ListTokens(req *types.ListTokensReq) (resp *types.ListTokensResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	rows, total, derr := l.svcCtx.Models.ApiKey.FindPageWithUser(l.ctx, page, size)
	if derr != nil {
		return nil, errorx.Internal(derr.Error())
	}
	resp = &types.ListTokensResp{
		Total: total,
		Items: make([]types.TokenItem, 0, len(rows)),
	}
	for _, r := range rows {
		remaining := r.Quota - r.Used
		if remaining < 0 {
			remaining = 0
		}
		resp.Items = append(resp.Items, types.TokenItem{
			Id:         r.Id,
			TokenId:    "tk_live_" + r.Prefix + "...",
			UserName:   r.UserName,
			UserEmail:  r.UserEmail,
			UsageLimit: r.Quota,
			Remaining:  remaining,
			CreatedAt:  r.CreatedAt,
			Status:     int32(r.Status),
		})
	}
	return resp, nil
}
