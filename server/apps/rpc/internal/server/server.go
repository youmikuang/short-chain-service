package server

import (
	"context"

	"server/apps/rpc/internal/logic"
	"server/apps/rpc/internal/svc"
	"server/apps/rpc/pb"
)

// ShortLinkServer 短链核心 gRPC 服务实现（手写，连接 logic 与 pb）
type ShortLinkServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedShortLinkServer
}

func NewShortLinkServer(svcCtx *svc.ServiceContext) *ShortLinkServer {
	return &ShortLinkServer{svcCtx: svcCtx}
}

func (s *ShortLinkServer) CreateShortLink(ctx context.Context, in *pb.CreateShortLinkReq) (*pb.CreateShortLinkResp, error) {
	return logic.NewCreateShortLinkLogic(ctx, s.svcCtx).CreateShortLink(in)
}

func (s *ShortLinkServer) GetByCode(ctx context.Context, in *pb.GetByCodeReq) (*pb.GetByCodeResp, error) {
	return logic.NewGetByCodeLogic(ctx, s.svcCtx).GetByCode(in)
}

func (s *ShortLinkServer) BatchCreate(ctx context.Context, in *pb.BatchCreateReq) (*pb.BatchCreateResp, error) {
	return logic.NewBatchCreateLogic(ctx, s.svcCtx).BatchCreate(in)
}

func (s *ShortLinkServer) DeleteShortLink(ctx context.Context, in *pb.DeleteShortLinkReq) (*pb.DeleteShortLinkResp, error) {
	return logic.NewDeleteShortLinkLogic(ctx, s.svcCtx).DeleteShortLink(in)
}

func (s *ShortLinkServer) Resolve(ctx context.Context, in *pb.ResolveReq) (*pb.ResolveResp, error) {
	return logic.NewResolveLogic(ctx, s.svcCtx).Resolve(in)
}
