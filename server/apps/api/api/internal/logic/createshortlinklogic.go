package logic

import (
	"context"

	"server/apps/api/api/internal/middleware"
	"server/apps/api/api/internal/svc"
	"server/apps/api/api/internal/types"
	pb "server/apps/api/rpc/pb"
)

type CreateShortLinkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateShortLinkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateShortLinkLogic {
	return &CreateShortLinkLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreateShortLink 网关：已通过 API Key/JWT 鉴权，委托 rpc 核心服务
func (l *CreateShortLinkLogic) CreateShortLink(req *types.CreateShortLinkReq) (resp *types.CreateShortLinkResp, err error) {
	uid := int64(0)
	if v, ok := l.ctx.Value(middleware.APIKeyUIDKey).(float64); ok {
		uid = int64(v)
	} else if v, ok := l.ctx.Value("uid").(float64); ok {
		uid = int64(v)
	}
	apiKey := ""
	if v, ok := l.ctx.Value(middleware.APIKeyKey).(string); ok {
		apiKey = v
	}

	out, err := l.svcCtx.ShortLinkRpc.CreateShortLink(l.ctx, &pb.CreateShortLinkReq{
		LongUrl: req.LongURL,
		UserId:  uid,
		ApiKey:  apiKey,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateShortLinkResp{Code: out.Code, LongURL: out.LongUrl}, nil
}
