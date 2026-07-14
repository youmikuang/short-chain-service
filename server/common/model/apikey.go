package model

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ApiKey struct {
	Id        int64  `db:"id"`
	UserId    int64  `db:"user_id"`
	Name      string `db:"name"`
	KeyHash   string `db:"key_hash"`
	Prefix    string `db:"prefix"`
	Quota     int64  `db:"quota"`
	Used      int64  `db:"used"`
	Status    int64  `db:"status"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type ApiKeyModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewApiKeyModel(conn sqlx.SqlConn) *ApiKeyModel {
	return &ApiKeyModel{conn: conn, table: "`api_keys`"}
}

const apiKeyRows = "id, user_id, name, key_hash, prefix, quota, used, status, created_at, updated_at"

// apiKeyJoinRows 联表查询时给列加表别名，避免与 users.id 冲突
const apiKeyJoinRows = "ak.id, ak.user_id, ak.name, ak.key_hash, ak.prefix, ak.quota, ak.used, ak.status, ak.created_at, ak.updated_at"

func (m *ApiKeyModel) Insert(ctx context.Context, data *ApiKey) (sql.Result, error) {
	query := "insert into " + m.table + " (user_id, name, key_hash, prefix, quota, used, status) values (?, ?, ?, ?, ?, ?, ?)"
	return m.conn.Exec(query, data.UserId, data.Name, data.KeyHash, data.Prefix, data.Quota, data.Used, data.Status)
}

func (m *ApiKeyModel) FindOneById(ctx context.Context, id int64) (*ApiKey, error) {
	query := "select " + apiKeyRows + " from " + m.table + " where id = ? limit 1"
	var resp ApiKey
	err := m.conn.QueryRow(&resp, query, id)
	return &resp, err
}

func (m *ApiKeyModel) FindOneByHash(ctx context.Context, keyHash string) (*ApiKey, error) {
	query := "select " + apiKeyRows + " from " + m.table + " where key_hash = ? limit 1"
	var resp ApiKey
	err := m.conn.QueryRow(&resp, query, keyHash)
	return &resp, err
}

func (m *ApiKeyModel) FindByUser(ctx context.Context, userId int64) ([]ApiKey, error) {
	query := "select " + apiKeyRows + " from " + m.table + " where user_id = ? and status = 1 order by id desc"
	var items []ApiKey
	err := m.conn.QueryRows(&items, query, userId)
	return items, err
}

func (m *ApiKeyModel) UpdateStatus(ctx context.Context, id, userId, status int64) error {
	query := "update " + m.table + " set status = ? where id = ? and user_id = ?"
	_, err := m.conn.Exec(query, status, id, userId)
	return err
}

// CountWhere 按状态统计 key 数量
func (m *ApiKeyModel) CountWhere(ctx context.Context, status int64) (int64, error) {
	var total int64
	err := m.conn.QueryRow(&total, "select count(*) from "+m.table+" where status = ?", status)
	return total, err
}

// FindPageWithUser 管理后台分页列表（联表取用户昵称/邮箱）
func (m *ApiKeyModel) FindPageWithUser(ctx context.Context, page, pageSize int64) ([]ApiKeyWithUser, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var total int64
	if err := m.conn.QueryRow(&total, "select count(*) from "+m.table); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	query := "select " + apiKeyJoinRows + ", u.nickname as user_name, u.email as user_email from " +
		m.table + " ak left join `users` u on ak.user_id = u.id order by ak.id desc limit ? offset ?"
	var items []ApiKeyWithUser
	if err := m.conn.QueryRows(&items, query, pageSize, offset); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

type ApiKeyWithUser struct {
	ApiKey
	UserName  string `db:"user_name"`
	UserEmail string `db:"user_email"`
}
