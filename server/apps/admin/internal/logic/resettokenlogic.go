package logic

import (
	"context"

	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/common/errorx"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetTokenLogic {
	return &ResetTokenLogic{ctx: ctx, svcCtx: svcCtx}
}

// ResetToken 将指定 API Key 的已用量清零（重置配额使用）
func (l *ResetTokenLogic) ResetToken(req *types.ResetTokenReq) (resp *types.ResetTokenResp, err error) {
	if req.Id <= 0 {
		return nil, errorx.BadParam("id required")
	}
	if err := l.svcCtx.Models.ApiKey.ResetUsage(l.ctx, req.Id); err != nil {
		logx.Errorf("ResetToken ResetUsage failed: %v", err)
		return nil, errorx.Internal("reset token failed")
	}
	return &types.ResetTokenResp{Ok: true}, nil
}
