package logic

import (
	"context"
	"server/apps/api/rpc/internal/svc"
	"server/apps/api/rpc/pb"
	"server/common/errorx"

	"github.com/redis/go-redis/v9"
)

type GetByCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetByCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetByCodeLogic {
	return &GetByCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *GetByCodeLogic) GetByCode(in *pb.GetByCodeReq) (*pb.GetByCodeResp, error) {
	code := in.GetCode()
	if code == "" {
		return nil, errorx.BadParam("code required")
	}
	// 查缓存
	longURL, err := l.svcCtx.Redis.Get(l.ctx, "short_link:"+code).Result()
	if err == redis.Nil {
		// 回源 MySQL
		row, derr := l.svcCtx.Models.ShortLink.FindOneByCode(l.ctx, code)
		if isNotFound(derr) {
			return nil, errorx.NotFound("code not found")
		} else if derr != nil {
			return nil, errorx.Internal(derr.Error())
		}
		// 回填缓存
		l.svcCtx.Redis.Set(l.ctx, "short_link:"+code, row.LongURL, redisCacheTTL())
		return &pb.GetByCodeResp{
			Code:    row.Code,
			LongUrl: row.LongURL,
			Clicks:  row.Clicks,
			Status:  int32(row.Status),
		}, nil
	} else if err != nil {
		return nil, err
	}
	clicks, _ := l.svcCtx.Redis.Get(l.ctx, "short_link:"+code+":clicks").Int64()
	return &pb.GetByCodeResp{Code: code, LongUrl: longURL, Clicks: clicks, Status: 1}, nil
}
