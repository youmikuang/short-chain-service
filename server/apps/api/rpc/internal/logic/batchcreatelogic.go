package logic

import (
	"context"
	"server/apps/api/rpc/internal/svc"
	"server/apps/api/rpc/pb"
)

type BatchCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchCreateLogic {
	return &BatchCreateLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *BatchCreateLogic) BatchCreate(in *pb.BatchCreateReq) (*pb.BatchCreateResp, error) {
	resp := &pb.BatchCreateResp{}
	for _, u := range in.GetLongUrls() {
		r, err := NewCreateShortLinkLogic(l.ctx, l.svcCtx).CreateShortLink(&pb.CreateShortLinkReq{
			LongUrl: u,
			UserId:  in.GetUserId(),
		})
		if err != nil {
			return nil, err
		}
		resp.Items = append(resp.Items, r)
	}
	return resp, nil
}
