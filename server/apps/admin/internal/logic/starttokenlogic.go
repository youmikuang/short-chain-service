package logic

import (
	"context"

	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/common/errorx"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStartTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartTokenLogic {
	return &StartTokenLogic{ctx: ctx, svcCtx: svcCtx}
}

// StartToken 重新启用被吊销的 API Key（status = 1）
func (l *StartTokenLogic) StartToken(req *types.StartTokenReq) (resp *types.StartTokenResp, err error) {
	if req.Id <= 0 {
		return nil, errorx.BadParam("id required")
	}
	if err := l.svcCtx.Models.ApiKey.SetStatus(l.ctx, req.Id, 1); err != nil {
		logx.Errorf("StartToken SetStatus failed: %v", err)
		return nil, errorx.Internal("start token failed")
	}
	return &types.StartTokenResp{Ok: true}, nil
}
