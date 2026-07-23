package logic

import (
	"context"
	"math/rand"
	"time"

	"server/apps/rpc/internal/svc"
	"server/apps/rpc/pb"
	"server/common/errorx"
	"server/common/model"
	"server/common/tool"
)

type CreateSlinkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateSlinkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSlinkLogic {
	return &CreateSlinkLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *CreateSlinkLogic) CreateSlink(in *pb.CreateSlinkReq) (*pb.CreateSlinkResp, error) {
	if in.GetLongUrl() == "" {
		return nil, errorx.BadParam("long_url required")
	}
	// 身份判定（web 与第三方两条互不干扰的路径）：
	//   - web：网关已用 JWT 鉴权并传入 user_id，key 不参与，直接使用 user_id；
	//   - 第三方：无 JWT（user_id=0），必须带 X-API-Key，由 rpc 校验合法性并以 key 归属用户为准。
	ownerId := in.GetUserId()
	source := "rpc" // 第三方 API Key 调用（经 rpc 核心服务）
	if ownerId != 0 {
		source = "web" // 网页（JWT）调用
	} else if in.GetApiKey() == "" {
		return nil, errorx.Unauthorized("unauthorized")
	} else {
		row, err := l.svcCtx.Models.ApiKey.FindOneByHash(l.ctx, tool.Sha256Hex(in.GetApiKey()))
		if err != nil || row.Status != 1 {
			return nil, errorx.Unauthorized("invalid X-API-Key")
		}
		ownerId = row.UserId
	}
	// 归一化长链接用于去重与存储（https://baidu.com/ 与 https://baidu.com 视为同一链接）
	longURL := tool.NormalizeURL(in.GetLongUrl())
	domain, err := tool.ExtractDomain(longURL)
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

	// 去重复用：同一用户 + 同一长链接只存一条；若已存在，则把 source 更新为
	// 「最后生成」的来源（web / rpc 谁后生成算谁的），并刷新缓存。
	if exist, err := l.svcCtx.Models.Slink.FindOneByUserAndURL(l.ctx, ownerId, longURL); err == nil {
		if exist.Source != source {
			if uerr := l.svcCtx.Models.Slink.UpdateSource(l.ctx, exist.Code, source); uerr != nil {
				return nil, errorx.Internal(uerr.Error())
			}
		}
		// 刷新 Redis 缓存（long_url / uid / source），保证跳转与访问日志来源准确
		l.svcCtx.Redis.Set(l.ctx, "short_link:"+exist.Code, exist.LongURL, redisCacheTTL())
		l.svcCtx.Redis.Set(l.ctx, "short_link:"+exist.Code+":uid", ownerId, redisCacheTTL())
		l.svcCtx.Redis.Set(l.ctx, "short_link:"+exist.Code+":source", source, redisCacheTTL())
		return &pb.CreateSlinkResp{Code: exist.Code, LongUrl: exist.LongURL}, nil
	} else if !isNotFound(err) {
		return nil, errorx.Internal(err.Error())
	}

	// 生成短码：随机 Base62（默认 6 位，冲突则加长重试），比雪花 ID 更短
	code, err := l.genCode()
	if err != nil {
		return nil, err
	}

	// 落库 MySQL(short_links)
	if _, err := l.svcCtx.Models.Slink.Insert(l.ctx, &model.Slink{
		Code:    code,
		LongURL: longURL,
		UserId:  ownerId,
		Status:  1,
		Source:  source,
	}); err != nil {
		return nil, errorx.Internal(err.Error())
	}

	// 写 Redis 缓存(short_link:{code} -> long_url)，随机 TTL 防集中失效
	l.svcCtx.Redis.Set(l.ctx, "short_link:"+code, longURL, redisCacheTTL())

	return &pb.CreateSlinkResp{Code: code, LongUrl: in.GetLongUrl()}, nil
}

// genCode 生成短码：随机 Base62 串，默认 6 位；若已存在则加长 1 位重试，最多到 8 位。
func (l *CreateSlinkLogic) genCode() (string, error) {
	for length := 6; length <= 8; length++ {
		for i := 0; i < 5; i++ {
			code := tool.RandString(length)
			if _, err := l.svcCtx.Models.Slink.FindOneByCode(l.ctx, code); err != nil {
				if isNotFound(err) {
					return code, nil
				}
				return "", errorx.Internal(err.Error())
			}
			// code 已存在，换一个重试
		}
	}
	return "", errorx.Internal("failed to generate unique short code")
}

// isBlacklisted 校验域名是否命中黑名单（优先 MySQL，回退 Redis）
func (l *CreateSlinkLogic) isBlacklisted(domain string) (bool, error) {
	if _, err := l.svcCtx.Models.DomainBlacklist.FindOneByDomain(l.ctx, domain); err == nil {
		return true, nil
	} else if !isNotFound(err) {
		return false, err
	}
	// 回退 Redis Set（admin 写入时同步 SADD）
	return l.svcCtx.Redis.SIsMember(l.ctx, l.svcCtx.Config.RedisKey, domain).Result()
}

func redisCacheTTL() time.Duration {
	return 30*time.Minute + time.Duration(rand.Intn(600))*time.Second
}
