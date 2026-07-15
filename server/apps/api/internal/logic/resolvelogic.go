package logic

import (
	"context"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	pb "server/apps/rpc/pb"
)

type ResolveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResolveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveLogic {
	return &ResolveLogic{ctx: ctx, svcCtx: svcCtx}
}

// Resolve 网关：调用 rpc 解析，命中黑名单时由 handler 拦截
func (l *ResolveLogic) Resolve(req *types.ResolveReq) (resp *types.ResolveResp, err error) {
	out, err := l.svcCtx.ShortLinkRpc.Resolve(l.ctx, &pb.ResolveReq{Code: req.Code})
	if err != nil {
		return nil, err
	}
	return &types.ResolveResp{LongURL: out.LongUrl, Blocked: out.Blocked}, nil
}
