package logic

import (
	"context"

	"server/apps/jump/internal/svc"
	"server/apps/jump/internal/types"
	pb "server/apps/rpc/pb"

	"google.golang.org/grpc/metadata"
)

type ResolveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResolveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveLogic {
	return &ResolveLogic{ctx: ctx, svcCtx: svcCtx}
}

// Resolve 调用 rpc 核心解析短链。访问者 IP / Referer 通过 gRPC metadata 透传，
// 由 rpc.Resolve 写入 click_events（访问明细日志）。命中黑名单时 rpc 返回 Blocked。
func (l *ResolveLogic) Resolve(req *types.ResolveReq) (resp *types.ResolveResp, err error) {
	ctx := l.ctx
	var pairs []string
	if req.Ip != "" {
		pairs = append(pairs, "x-client-ip", req.Ip)
	}
	if req.Referer != "" {
		pairs = append(pairs, "x-referer", req.Referer)
	}
	if len(pairs) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
	}
	out, err := l.svcCtx.ShortLinkRpc.Resolve(ctx, &pb.ResolveReq{Code: req.Code})
	if err != nil {
		return nil, err
	}
	return &types.ResolveResp{LongURL: out.LongUrl, Blocked: out.Blocked}, nil
}
