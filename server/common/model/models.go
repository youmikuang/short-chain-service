package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

// Models 聚合所有数据访问模型，由各个服务（api 网关 / rpc 核心 / admin）
// 使用各自的 sqlx 连接初始化一次后注入 ServiceContext。
type Models struct {
	User           *UserModel
	ApiKey         *ApiKeyModel
	UserSettings   *UserSettingsModel
	AccessLog       *AccessLogModel
	ShortLink       *ShortLinkModel
	DomainBlacklist *DomainBlacklistModel
}

func NewModels(conn sqlx.SqlConn) *Models {
	return &Models{
		User:            NewUserModel(conn),
		ApiKey:          NewApiKeyModel(conn),
		UserSettings:    NewUserSettingsModel(conn),
		AccessLog:       NewAccessLogModel(conn),
		ShortLink:       NewShortLinkModel(conn),
		DomainBlacklist: NewDomainBlacklistModel(conn),
	}
}
