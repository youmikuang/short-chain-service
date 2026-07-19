package logic

import (
	"context"

	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	"server/common/errorx"
)

type ListMyLinksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyLinksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyLinksLogic {
	return &ListMyLinksLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListMyLinks 返回当前登录用户创建的短链（JWT 鉴权）
func (l *ListMyLinksLogic) ListMyLinks(req *types.ListMyLinksReq) (*types.ListMyLinksResp, error) {
	uid, err := uidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	rows, total, err := l.svcCtx.Models.Slink.FindPageByUser(l.ctx, uid, req.Page, req.Size, req.Search, req.Sort)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	items := make([]types.MyLinkItem, 0, len(rows))
	domain := l.svcCtx.Config.ShortDomain
	if domain == "" {
		domain = "https://s.gaoheng.top"
	}
	for _, r := range rows {
		items = append(items, types.MyLinkItem{
			Code:      r.Code,
			SUrl:      domain + "/r/" + r.Code,
			LongURL:   r.LongURL,
			Clicks:    r.Clicks,
			Status:    int32(r.Status),
			Source:    r.Source,
			CreatedAt: r.CreatedAt,
		})
	}
	return &types.ListMyLinksResp{Total: total, Items: items}, nil
}
