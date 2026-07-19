package logic

import (
	"context"
	"server/apps/rpc/internal/svc"
	"server/apps/rpc/pb"
	"server/common/errorx"
)

type DeleteslinkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteslinkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteslinkLogic {
	return &DeleteslinkLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteslinkLogic) Deleteslink(in *pb.DeleteslinkReq) (*pb.DeleteslinkResp, error) {
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
	return &pb.DeleteslinkResp{Ok: true}, nil
}
