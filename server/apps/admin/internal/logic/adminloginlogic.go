package logic

import (
	"context"
	"time"

	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/common/errorx"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLoginLogic {
	return &AdminLoginLogic{ctx: ctx, svcCtx: svcCtx}
}

// AdminLogin 校验管理后台凭据并签发 JWT
func (l *AdminLoginLogic) AdminLogin(req *types.AdminLoginReq) (resp *types.AdminLoginResp, err error) {
	if req.Username != l.svcCtx.Config.Admin.Username || req.Password != l.svcCtx.Config.Admin.Password {
		return nil, errorx.Unauthorized("invalid username or password")
	}

	claims := jwt.MapClaims{
		"uid": 0,
		"exp": time.Now().Add(time.Duration(l.svcCtx.Config.Auth.AccessExpire) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
	if err != nil {
		logx.Errorf("AdminLogin sign token failed: %v", err)
		return nil, errorx.Internal("sign token failed")
	}
	return &types.AdminLoginResp{Token: signed}, nil
}
