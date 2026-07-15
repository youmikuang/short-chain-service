package logic

import (
	"context"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/common/errorx"
)

type ListBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBlacklistLogic {
	return &ListBlacklistLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListBlacklist 域名黑名单分页列表
func (l *ListBlacklistLogic) ListBlacklist(req *types.ListBlacklistReq) (resp *types.ListBlacklistResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	rows, total, derr := l.svcCtx.Models.DomainBlacklist.FindPage(l.ctx, page, size)
	if derr != nil {
		return nil, errorx.Internal(derr.Error())
	}
	resp = &types.ListBlacklistResp{
		Total: total,
		Items: make([]types.BlacklistItem, 0, len(rows)),
	}
	for _, r := range rows {
		resp.Items = append(resp.Items, types.BlacklistItem{
			Domain:    r.Domain,
			Reason:    r.Reason,
			Attempts:  r.Attempts,
			CreatedAt: r.CreatedAt,
		})
	}
	return resp, nil
}
