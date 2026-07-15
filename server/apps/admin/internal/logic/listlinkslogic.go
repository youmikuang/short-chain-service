package logic

import (
	"context"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/common/errorx"
)

type ListLinksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLinksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLinksLogic {
	return &ListLinksLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListLinks 链接管理列表（MySQL 分页 + 联表用户信息）
func (l *ListLinksLogic) ListLinks(req *types.ListLinksReq) (resp *types.ListLinksResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	rows, total, derr := l.svcCtx.Models.ShortLink.FindPageWithUser(l.ctx, page, size)
	if derr != nil {
		return nil, errorx.Internal(derr.Error())
	}
	resp = &types.ListLinksResp{
		Total: total,
		Items: make([]types.LinkItem, 0, len(rows)),
	}
	for _, r := range rows {
		resp.Items = append(resp.Items, types.LinkItem{
			Code:      r.Code,
			LongURL:   r.LongURL,
			ShortURL:  "slnk.it/" + r.Code,
			Clicks:    r.Clicks,
			Status:    int32(r.Status),
			UserName:  r.UserName,
			UserEmail: r.UserEmail,
			CreatedAt: r.CreatedAt,
		})
	}
	return resp, nil
}
