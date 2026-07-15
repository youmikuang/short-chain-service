package logic

import (
	"context"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	pb "server/apps/rpc/pb"
)

type GetByCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetByCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetByCodeLogic {
	return &GetByCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *GetByCodeLogic) GetByCode(req *types.GetByCodeReq) (resp *types.GetByCodeResp, err error) {
	out, err := l.svcCtx.ShortLinkRpc.GetByCode(l.ctx, &pb.GetByCodeReq{Code: req.Code})
	if err != nil {
		return nil, err
	}
	return &types.GetByCodeResp{
		Code:    out.Code,
		LongURL: out.LongUrl,
		Clicks:  out.Clicks,
		Status:  out.Status,
	}, nil
}
