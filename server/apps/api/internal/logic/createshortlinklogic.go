package logic

import (
	"context"

	"server/apps/api/internal/middleware"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	pb "server/apps/rpc/pb"
)

type CreateslinkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateslinkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateslinkLogic {
	return &CreateslinkLogic{ctx: ctx, svcCtx: svcCtx}
}

// Createslink 网关：已通过 API Key/JWT 鉴权，委托 rpc 核心服务
func (l *CreateslinkLogic) Createslink(req *types.CreateslinkReq) (resp *types.CreateslinkResp, err error) {
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

	out, err := l.svcCtx.SlinkRpc.Createslink(l.ctx, &pb.CreateslinkReq{
		LongUrl: req.LongURL,
		UserId:  uid,
		ApiKey:  apiKey,
	})
	if err != nil {
		return nil, err
	}
	domain := l.svcCtx.Config.ShortDomain
	if domain == "" {
		domain = "https://s.gaoheng.top"
	}
	return &types.CreateslinkResp{
		Code:     out.Code,
		ShortURL: domain + "/r/" + out.Code,
		LongURL:  out.LongUrl,
	}, nil
}
