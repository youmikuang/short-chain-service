package server

import (
	"context"

	"server/apps/rpc/internal/logic"
	"server/apps/rpc/internal/svc"
	"server/apps/rpc/pb"
)

// slinkServer 短链核心 gRPC 服务实现（手写，连接 logic 与 pb）
type slinkServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedslinkServer
}

func NewslinkServer(svcCtx *svc.ServiceContext) *slinkServer {
	return &slinkServer{svcCtx: svcCtx}
}

func (s *slinkServer) CreateSlink(ctx context.Context, in *pb.CreateSlinkReq) (*pb.CreateSlinkResp, error) {
	return logic.NewCreateSlinkLogic(ctx, s.svcCtx).CreateSlink(in)
}

func (s *slinkServer) GetByCode(ctx context.Context, in *pb.GetByCodeReq) (*pb.GetByCodeResp, error) {
	return logic.NewGetByCodeLogic(ctx, s.svcCtx).GetByCode(in)
}

func (s *slinkServer) BatchCreate(ctx context.Context, in *pb.BatchCreateReq) (*pb.BatchCreateResp, error) {
	return logic.NewBatchCreateLogic(ctx, s.svcCtx).BatchCreate(in)
}

func (s *slinkServer) Deleteslink(ctx context.Context, in *pb.DeleteslinkReq) (*pb.DeleteslinkResp, error) {
	return logic.NewDeleteslinkLogic(ctx, s.svcCtx).Deleteslink(in)
}

func (s *slinkServer) Resolve(ctx context.Context, in *pb.ResolveReq) (*pb.ResolveResp, error) {
	return logic.NewResolveLogic(ctx, s.svcCtx).Resolve(in)
}
