package logic

import (
	"context"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/common/errorx"
)

type RevokeTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRevokeTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeTokenLogic {
	return &RevokeTokenLogic{ctx: ctx, svcCtx: svcCtx}
}

// RevokeToken 吊销指定 API Key（status = 0）
func (l *RevokeTokenLogic) RevokeToken(req *types.RevokeTokenReq) (resp *types.RevokeTokenResp, err error) {
	if req.Id <= 0 {
		return nil, errorx.BadParam("id required")
	}
	if err := l.svcCtx.Models.ApiKey.UpdateStatus(l.ctx, req.Id, 0, 0); err != nil {
		return nil, errorx.Internal(err.Error())
	}
	return &types.RevokeTokenResp{Ok: true}, nil
}
