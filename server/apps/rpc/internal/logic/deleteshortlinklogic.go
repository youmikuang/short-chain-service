package logic

import (
	"context"
	"server/apps/rpc/internal/svc"
	"server/apps/rpc/pb"
	"server/pkg/errorx"
)

type DeleteSlinkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteSlinkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSlinkLogic {
	return &DeleteSlinkLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteSlinkLogic) DeleteSlink(in *pb.DeleteSlinkReq) (*pb.DeleteSlinkResp, error) {
	code := in.GetCode()
	if code == "" {
		return nil, errorx.BadParam("code required")
	}
	if err := l.svcCtx.Models.Slink.Delete(l.ctx, code); err != nil {
		return nil, errorx.Internal(err.Error())
	}
	// 删除 Redis 缓存
	l.svcCtx.Redis.Del(l.ctx, "short_link:"+code)
	l.svcCtx.Redis.Del(l.ctx, "short_link:"+code+":clicks")
	return &pb.DeleteSlinkResp{Ok: true}, nil
}
