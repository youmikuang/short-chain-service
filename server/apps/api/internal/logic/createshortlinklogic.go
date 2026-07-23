package logic

import (
	"context"

	"server/apps/api/internal/middleware"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	pb "server/apps/rpc/pb"
)

type CreateSlinkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateSlinkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSlinkLogic {
	return &CreateSlinkLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreateSlink 网关：已通过 API Key/JWT 鉴权，委托 rpc 核心服务
func (l *CreateSlinkLogic) CreateSlink(req *types.CreateSlinkReq) (resp *types.CreateSlinkResp, err error) {
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

	out, err := l.svcCtx.SlinkRpc.CreateSlink(l.ctx, &pb.CreateSlinkReq{
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
	return &types.CreateSlinkResp{
		Code:     out.Code,
		ShortURL: domain + "/r/" + out.Code,
		LongURL:  out.LongUrl,
	}, nil
}
