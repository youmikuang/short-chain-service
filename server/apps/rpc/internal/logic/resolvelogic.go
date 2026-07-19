package logic

import (
	"context"
	"server/apps/rpc/internal/svc"
	"server/apps/rpc/pb"
	"server/common/errorx"
	"server/common/model"
	"server/common/tool"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
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
	start := time.Now()
	code := in.GetCode()

	// 从网关透传的 gRPC metadata 取访问者 IP / Referer（网关从 HTTP 请求提取）
	ip, referer := "", ""
	if md, ok := metadata.FromIncomingContext(l.ctx); ok {
		if v := md.Get("x-client-ip"); len(v) > 0 {
			ip = v[0]
		}
		if v := md.Get("x-referer"); len(v) > 0 {
			referer = v[0]
		}
	}
	longURL, err := l.svcCtx.Redis.Get(l.ctx, "short_link:"+code).Result()
	// 短链所属用户（用于写入访问明细日志，按用户隔离）
	var ownerId int64
	var source string
	if err == redis.Nil {
		// 回源 MySQL
		row, derr := l.svcCtx.Models.Slink.FindOneByCode(l.ctx, code)
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
		} else if row, derr := l.svcCtx.Models.Slink.FindOneByCode(l.ctx, code); derr == nil {
			ownerId = row.UserId
			l.svcCtx.Redis.Set(l.ctx, "short_link:"+code+":uid", ownerId, redisCacheTTL())
		}
		// source 优先从缓存读取，缺失再回源
		if s, e := l.svcCtx.Redis.Get(l.ctx, "short_link:"+code+":source").Result(); e == nil {
			source = s
		} else if row, derr := l.svcCtx.Models.Slink.FindOneByCode(l.ctx, code); derr == nil {
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
	_ = l.svcCtx.Models.Slink.IncrClicks(l.ctx, code)

	// 写入短链访问明细到 ClickHouse（异步，不阻塞跳转；访问日志允许少量丢失）
	latencyMs := time.Since(start).Milliseconds()
	go func() {
		if err := l.svcCtx.ClickHouseVisit.Insert(context.Background(), &model.SlinkVisit{
			Code:      code,
			LongURL:   longURL,
			UserId:    ownerId,
			IP:        ip,
			Referer:   referer,
			Status:    200,
			Source:    source,
			LatencyMs: latencyMs,
		}); err != nil {
			logx.Errorf("ClickHouse visit insert failed for code %s: %v", code, err)
		}
	}()

	return &pb.ResolveResp{LongUrl: longURL, Blocked: false}, nil
}

// ProbeClickHouse 自检方法：写入一条探针记录到 click_events 并读回，
// 用于验证 ClickHouse 写入链路（配置/端口/驱动）是否正常。不在正常请求路径中调用，
// 仅供测试 / 运维排查使用（如 go test ./apps/rpc/internal/logic/ -run TestClickHouseProbe）。
func (l *ResolveLogic) ProbeClickHouse(ctx context.Context) error {
	probe := &model.SlinkVisit{
		Code:    "probe-" + time.Now().Format("20060102150405"),
		LongURL: "https://example.com/probe",
		UserId:  0,
		Status:  200,
		Source:  "rpc",
	}
	if err := l.svcCtx.ClickHouseVisit.Insert(ctx, probe); err != nil {
		return errorx.Internal("clickhouse insert failed: " + err.Error())
	}
	rows, _, err := l.svcCtx.ClickHouseVisit.FindPageByUser(ctx, 0, 1, 50, probe.Code, "")
	if err != nil {
		return errorx.Internal("clickhouse read-back failed: " + err.Error())
	}
	for _, r := range rows {
		if r.Code == probe.Code {
			return nil
		}
	}
	return errorx.Internal("clickhouse probe row not found after insert")
}

// isBlacklisted 校验域名是否命中黑名单（优先 MySQL，回退 Redis）
func (l *ResolveLogic) isBlacklisted(domain string) (bool, error) {
	if _, err := l.svcCtx.Models.DomainBlacklist.FindOneByDomain(l.ctx, domain); err == nil {
		return true, nil
	} else if !isNotFound(err) {
		return false, err
	}
	// 回退 Redis Set（admin 写入时同步 SADD）
	return l.svcCtx.Redis.SIsMember(l.ctx, l.svcCtx.Config.RedisKey, domain).Result()
}
