package logic

import (
	"context"
	"errors"
	"server/apps/rpc/internal/svc"
	"server/apps/rpc/pb"
	"server/pkg/errorx"

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
	if errors.Is(err, redis.Nil) {
		// 回源 MySQL
		row, err := l.svcCtx.Models.Slink.FindOneByCode(l.ctx, code)
		if isNotFound(err) {
			return nil, errorx.NotFound("code not found")
		} else if err != nil {
			return nil, errorx.Internal(err.Error())
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
