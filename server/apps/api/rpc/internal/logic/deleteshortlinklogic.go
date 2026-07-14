package logic

import (
	"context"
	"server/apps/api/rpc/internal/svc"
	"server/apps/api/rpc/pb"
	"server/common/errorx"
)

type DeleteShortLinkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteShortLinkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteShortLinkLogic {
	return &DeleteShortLinkLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteShortLinkLogic) DeleteShortLink(in *pb.DeleteShortLinkReq) (*pb.DeleteShortLinkResp, error) {
	code := in.GetCode()
	if code == "" {
		return nil, errorx.BadParam("code required")
	}
	if err := l.svcCtx.Models.ShortLink.Delete(l.ctx, code); err != nil {
		return nil, errorx.Internal(err.Error())
	}
	// 删除 Redis 缓存
	l.svcCtx.Redis.Del(l.ctx, "short_link:"+code)
	l.svcCtx.Redis.Del(l.ctx, "short_link:"+code+":clicks")
	return &pb.DeleteShortLinkResp{Ok: true}, nil
}
