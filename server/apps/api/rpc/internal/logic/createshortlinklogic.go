package logic

import (
	"context"
	"math/rand"
	"time"

	"server/apps/api/rpc/internal/svc"
	"server/apps/api/rpc/pb"
	"server/common/errorx"
	"server/common/model"
	"server/common/tool"
)

type CreateShortLinkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateShortLinkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateShortLinkLogic {
	return &CreateShortLinkLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *CreateShortLinkLogic) CreateShortLink(in *pb.CreateShortLinkReq) (*pb.CreateShortLinkResp, error) {
	if in.GetLongUrl() == "" {
		return nil, errorx.BadParam("long_url required")
	}
	domain, err := tool.ExtractDomain(in.GetLongUrl())
	if err != nil {
		return nil, errorx.BadParam("invalid long_url")
	}
	blocked, err := l.isBlacklisted(domain)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, errorx.Blacklisted(domain)
	}

	// 生成短码：Snowflake + Base62
	code := tool.Base62Encode(l.svcCtx.IdGen.NextID())

	// 落库 MySQL(short_links)
	if _, err := l.svcCtx.Models.ShortLink.Insert(l.ctx, &model.ShortLink{
		Code:    code,
		LongURL: in.GetLongUrl(),
		UserId:  in.GetUserId(),
		Status:  1,
	}); err != nil {
		return nil, errorx.Internal(err.Error())
	}

	// 写 Redis 缓存(short_link:{code} -> long_url)，随机 TTL 防集中失效
	l.svcCtx.Redis.Set(l.ctx, "short_link:"+code, in.GetLongUrl(), redisCacheTTL())

	return &pb.CreateShortLinkResp{Code: code, LongUrl: in.GetLongUrl()}, nil
}

// isBlacklisted 校验域名是否命中黑名单（优先 MySQL，回退 Redis）
func (l *CreateShortLinkLogic) isBlacklisted(domain string) (bool, error) {
	if _, err := l.svcCtx.Models.DomainBlacklist.FindOneByDomain(l.ctx, domain); err == nil {
		return true, nil
	} else if !isNotFound(err) {
		return false, err
	}
	// 回退 Redis Set（admin 写入时同步 SADD）
	return l.svcCtx.Redis.SIsMember(l.ctx, l.svcCtx.Config.BlacklistRedisKey, domain).Result()
}

func redisCacheTTL() time.Duration {
	return 30*time.Minute + time.Duration(rand.Intn(600))*time.Second
}
