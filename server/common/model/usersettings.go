package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserSettings struct {
	UserId         int64  `db:"user_id"`
	EmailNotif     int64  `db:"email_notif"`
	SecurityAlerts int64  `db:"security_alerts"`
	MarketingComm  int64  `db:"marketing_comm"`
	CreatedAt      string `db:"created_at"`
	UpdatedAt      string `db:"updated_at"`
}

type UserSettingsModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewUserSettingsModel(conn sqlx.SqlConn) *UserSettingsModel {
	return &UserSettingsModel{conn: conn, table: "`user_settings`"}
}

const userSettingsRows = "user_id, email_notif, security_alerts, marketing_comm, created_at, updated_at"

// Upsert 写入或更新用户偏好（以 user_id 为主键）
func (m *UserSettingsModel) Upsert(ctx context.Context, data *UserSettings) error {
	query := "insert into " + m.table +
		" (user_id, email_notif, security_alerts, marketing_comm) values (?, ?, ?, ?) " +
		"on duplicate key update email_notif = values(email_notif), security_alerts = values(security_alerts), marketing_comm = values(marketing_comm)"
	_, err := m.conn.Exec(query, data.UserId, data.EmailNotif, data.SecurityAlerts, data.MarketingComm)
	return err
}

func (m *UserSettingsModel) FindOneByUserId(ctx context.Context, userId int64) (*UserSettings, error) {
	query := "select " + userSettingsRows + " from " + m.table + " where user_id = ? limit 1"
	var resp UserSettings
	err := m.conn.QueryRow(&resp, query, userId)
	return &resp, err
}
