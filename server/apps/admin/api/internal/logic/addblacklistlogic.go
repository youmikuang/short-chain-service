package logic

import (
	"context"
	"server/apps/admin/api/internal/svc"
	"server/apps/admin/api/internal/types"
	"server/common/errorx"
	"server/common/model"
)

type AddBlacklistLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddBlacklistLogic {
	return &AddBlacklistLogic{ctx: ctx, svcCtx: svcCtx}
}

// AddBlacklist 新增域名黑名单：写入 MySQL，并同步 SADD 到 Redis Set（供 rpc 回退校验）
func (l *AddBlacklistLogic) AddBlacklist(req *types.AddBlacklistReq) (resp *types.AddBlacklistResp, err error) {
	if req.Domain == "" {
		return nil, errorx.BadParam("domain required")
	}
	if _, ierr := l.svcCtx.Models.DomainBlacklist.Insert(l.ctx, &model.DomainBlacklist{
		Domain:   req.Domain,
		Reason:   req.Reason,
		Attempts: 0,
	}); ierr != nil {
		return nil, errorx.Internal(ierr.Error())
	}
	// 同步到 Redis Set，rpc 在 MySQL 未命中时回退到此校验
	if l.svcCtx.Config.BlacklistRedisKey != "" {
		if rerr := l.svcCtx.Redis.SAdd(l.ctx, l.svcCtx.Config.BlacklistRedisKey, req.Domain).Err(); rerr != nil {
			return nil, errorx.Internal(rerr.Error())
		}
	}
	return &types.AddBlacklistResp{Ok: true}, nil
}
