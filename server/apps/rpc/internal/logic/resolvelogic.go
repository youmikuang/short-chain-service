package logic

import (
	"context"
	"server/apps/rpc/internal/svc"
	"server/apps/rpc/pb"
	"server/common/errorx"
	"server/common/model"
	"server/common/tool"

	"github.com/redis/go-redis/v9"
)

type ResolveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResolveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveLogic {
	return &ResolveLogic{ctx: ctx, svcCtx: svcCtx}
}

// Resolve 跳转解析：缓存命中直接返回；未命中回源 MySQL；跳转前校验域名黑名单
func (l *ResolveLogic) Resolve(in *pb.ResolveReq) (*pb.ResolveResp, error) {
	code := in.GetCode()
	longURL, err := l.svcCtx.Redis.Get(l.ctx, "short_link:"+code).Result()
	// 短链所属用户（用于写入访问明细日志，按用户隔离）
	var ownerId int64
	var source string
	if err == redis.Nil {
		// 回源 MySQL
		row, derr := l.svcCtx.Models.ShortLink.FindOneByCode(l.ctx, code)
		if isNotFound(derr) {
			return nil, errorx.NotFound("code not found")
		} else if derr != nil {
			return nil, errorx.Internal(derr.Error())
		}
		longURL = row.LongURL
		ownerId = row.UserId
		source = row.Source
		l.svcCtx.Redis.Set(l.ctx, "short_link:"+code, longURL, redisCacheTTL())
		l.svcCtx.Redis.Set(l.ctx, "short_link:"+code+":uid", ownerId, redisCacheTTL())
		l.svcCtx.Redis.Set(l.ctx, "short_link:"+code+":source", source, redisCacheTTL())
	} else if err != nil {
		return nil, err
	} else {
		// 缓存命中：尝试从 uid 缓存取所属用户，缺失则回源补齐
		if v, e := l.svcCtx.Redis.Get(l.ctx, "short_link:"+code+":uid").Int64(); e == nil {
			ownerId = v
		} else if row, derr := l.svcCtx.Models.ShortLink.FindOneByCode(l.ctx, code); derr == nil {
			ownerId = row.UserId
			l.svcCtx.Redis.Set(l.ctx, "short_link:"+code+":uid", ownerId, redisCacheTTL())
		}
		// source 优先从缓存读取，缺失再回源
		if s, e := l.svcCtx.Redis.Get(l.ctx, "short_link:"+code+":source").Result(); e == nil {
			source = s
		} else if row, derr := l.svcCtx.Models.ShortLink.FindOneByCode(l.ctx, code); derr == nil {
			source = row.Source
			l.svcCtx.Redis.Set(l.ctx, "short_link:"+code+":source", source, redisCacheTTL())
		}
	}

	// 域名黑名单校验
	domain, derr := tool.ExtractDomain(longURL)
	if derr == nil {
		blocked, berr := l.isBlacklisted(domain)
		if berr == nil && blocked {
			return &pb.ResolveResp{Blocked: true}, nil // 命中黑名单，不跳转
		}
	}

	// 实时点击计数：Redis incr + MySQL 落库（保证 admin 列表准确）
	l.svcCtx.Redis.Incr(l.ctx, "short_link:"+code+":clicks")
	_ = l.svcCtx.Models.ShortLink.IncrClicks(l.ctx, code)

	// 写入短链访问明细到 ClickHouse（异步，不阻塞跳转；访问日志允许少量丢失）
	go func() {
		_ = l.svcCtx.ClickHouseVisit.Insert(context.Background(), &model.ShortLinkVisit{
			Code:    code,
			LongURL: longURL,
			UserId:  ownerId,
			Status:  200,
			Source:  source,
		})
	}()

	return &pb.ResolveResp{LongUrl: longURL, Blocked: false}, nil
}

// isBlacklisted 校验域名是否命中黑名单（优先 MySQL，回退 Redis）
func (l *ResolveLogic) isBlacklisted(domain string) (bool, error) {
	if _, err := l.svcCtx.Models.DomainBlacklist.FindOneByDomain(l.ctx, domain); err == nil {
		return true, nil
	} else if !isNotFound(err) {
		return false, err
	}
	// 回退 Redis Set（admin 写入时同步 SADD）
	return l.svcCtx.Redis.SIsMember(l.ctx, l.svcCtx.Config.BlacklistRedisKey, domain).Result()
}
