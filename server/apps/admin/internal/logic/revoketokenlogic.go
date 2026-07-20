package logic

import (
	"context"

	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/common/errorx"

	"github.com/zeromicro/go-zero/core/logx"
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
	// 按 id 直接置状态；原 UpdateStatus 带 user_id 约束，对真实 token 永不命中。
	if err := l.svcCtx.Models.ApiKey.SetStatus(l.ctx, req.Id, 0); err != nil {
		logx.Errorf("RevokeToken SetStatus failed: %v", err)
		return nil, errorx.Internal("revoke token failed")
	}
	return &types.RevokeTokenResp{Ok: true}, nil
}
