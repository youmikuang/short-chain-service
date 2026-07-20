package logic

import (
	"context"

	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/common/errorx"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBlacklistLogic {
	return &DeleteBlacklistLogic{ctx: ctx, svcCtx: svcCtx}
}

// DeleteBlacklist 删除域名黑名单：从 MySQL 删除，并同步 SREM 出 Redis Set（与 Add 的 SADD 对应）
func (l *DeleteBlacklistLogic) DeleteBlacklist(req *types.DeleteBlacklistReq) (resp *types.DeleteBlacklistResp, err error) {
	if req.Domain == "" {
		return nil, errorx.BadParam("domain required")
	}
	if derr := l.svcCtx.Models.DomainBlacklist.Delete(l.ctx, req.Domain); derr != nil {
		logx.Errorf("DeleteBlacklist Delete failed: %v", derr)
		return nil, errorx.Internal("delete blacklist failed")
	}
	// 同步移除 Redis Set，避免 rpc 回退校验时仍命中已删除的域名；
	// Redis 为缓存，同步失败不应阻断主流程
	if l.svcCtx.Config.BlacklistRedisKey != "" {
		if rerr := l.svcCtx.Redis.SRem(l.ctx, l.svcCtx.Config.BlacklistRedisKey, req.Domain).Err(); rerr != nil {
			logx.Errorf("DeleteBlacklist Redis SRem failed: %v", rerr)
		}
	}
	return &types.DeleteBlacklistResp{Ok: true}, nil
}
