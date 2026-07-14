package logic

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"server/apps/admin/api/internal/svc"
	"server/apps/admin/api/internal/types"
	"server/common/errorx"
	"server/common/model"
)

type ProvisionTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProvisionTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProvisionTokenLogic {
	return &ProvisionTokenLogic{ctx: ctx, svcCtx: svcCtx}
}

// ProvisionToken 签发新的 API Key（明文仅返回一次，入库存哈希）
func (l *ProvisionTokenLogic) ProvisionToken(req *types.ProvisionTokenReq) (resp *types.ProvisionTokenResp, err error) {
	userId := req.UserId
	if userId <= 0 {
		userId = 1
	}
	if req.Name == "" {
		req.Name = "admin-provisioned"
	}
	if req.Quota <= 0 {
		req.Quota = 100000
	}

	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return nil, errorx.Internal(err.Error())
	}
	token := hex.EncodeToString(buf)
	prefix := token[:5]
	sum := sha256.Sum256([]byte(token))
	keyHash := hex.EncodeToString(sum[:])

	if _, err := l.svcCtx.Models.ApiKey.Insert(l.ctx, &model.ApiKey{
		UserId:   userId,
		Name:     req.Name,
		KeyHash:  keyHash,
		Prefix:   prefix,
		Quota:    req.Quota,
		Used:     0,
		Status:   1,
	}); err != nil {
		return nil, errorx.Internal(err.Error())
	}

	return &types.ProvisionTokenResp{
		Ok:      true,
		TokenId: "tk_live_" + prefix + "...",
		Token:   token,
	}, nil
}
