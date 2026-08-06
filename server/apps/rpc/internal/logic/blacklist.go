package logic

import (
	"context"

	"server/apps/rpc/internal/svc"
	"server/pkg/tool"
)

// isDomainBlacklisted 校验 host 是否命中域名黑名单（优先 MySQL，回退 Redis Set）。
// 黑名单按「域名及其父域」匹配：黑名单存 zhipin.com 时，www.zhipin.com 等子域同样命中。
// Redis 查询失败不阻断（黑名单以 MySQL 为准，Redis 仅为 admin 同步的加速缓存）。
func isDomainBlacklisted(ctx context.Context, svcCtx *svc.ServiceContext, host string) (bool, error) {
	for _, d := range tool.DomainCandidates(host) {
		if _, err := svcCtx.Models.DomainBlacklist.FindOneByDomain(ctx, d); err == nil {
			return true, nil
		} else if !isNotFound(err) {
			return false, err
		}
		if ok, err := svcCtx.Redis.SIsMember(ctx, svcCtx.Config.RedisKey, d).Result(); err == nil && ok {
			return true, nil
		}
	}
	return false, nil
}
