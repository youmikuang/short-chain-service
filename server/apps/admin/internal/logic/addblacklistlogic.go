package logic

import (
	"context"
	"errors"

	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/common/errorx"
	"server/common/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	// 去重：已存在则直接返回，避免触发唯一键冲突并把原始 DB 错误外泄给前端
	if _, derr := l.svcCtx.Models.DomainBlacklist.FindOneByDomain(l.ctx, req.Domain); derr == nil {
		return nil, errorx.BadParam("domain already in blacklist")
	} else if !errors.Is(derr, sqlx.ErrNotFound) {
		logx.Errorf("AddBlacklist FindOneByDomain failed: %v", derr)
		return nil, errorx.Internal("query blacklist failed")
	}
	if _, ierr := l.svcCtx.Models.DomainBlacklist.Insert(l.ctx, &model.DomainBlacklist{
		Domain:   req.Domain,
		Reason:   req.Reason,
		Attempts: 0,
	}); ierr != nil {
		logx.Errorf("AddBlacklist Insert failed: %v", ierr)
		return nil, errorx.Internal("insert blacklist failed")
	}
	// Redis Set 为 rpc 回退校验用的缓存，同步失败不应阻断主流程（MySQL 才是源）
	if l.svcCtx.Config.BlacklistRedisKey != "" {
		if rerr := l.svcCtx.Redis.SAdd(l.ctx, l.svcCtx.Config.BlacklistRedisKey, req.Domain).Err(); rerr != nil {
			logx.Errorf("AddBlacklist Redis SAdd failed: %v", rerr)
		}
	}
	return &types.AddBlacklistResp{Ok: true}, nil
}
