package model

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DomainBlacklist struct {
	Id        int64  `db:"id"`
	Domain    string `db:"domain"`
	Reason    string `db:"reason"`
	Attempts  int64  `db:"attempts"`
	CreatedAt string `db:"created_at"`
}

type DomainBlacklistModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewDomainBlacklistModel(conn sqlx.SqlConn) *DomainBlacklistModel {
	return &DomainBlacklistModel{conn: conn, table: "`domain_blacklist`"}
}

const domainBlacklistRows = "id, domain, reason, attempts, created_at"

func (m *DomainBlacklistModel) Insert(ctx context.Context, data *DomainBlacklist) (sql.Result, error) {
	query := "insert into " + m.table + " (domain, reason, attempts) values (?, ?, ?)"
	return m.conn.Exec(query, data.Domain, data.Reason, data.Attempts)
}

func (m *DomainBlacklistModel) FindOneByDomain(ctx context.Context, domain string) (*DomainBlacklist, error) {
	query := "select " + domainBlacklistRows + " from " + m.table + " where domain = ? limit 1"
	var resp DomainBlacklist
	err := m.conn.QueryRow(&resp, query, domain)
	return &resp, err
}

// Count 黑名单总数
func (m *DomainBlacklistModel) Count(ctx context.Context) (int64, error) {
	var total int64
	err := m.conn.QueryRow(&total, "select count(*) from "+m.table)
	return total, err
}

// FindPage 管理后台分页列表
func (m *DomainBlacklistModel) FindPage(ctx context.Context, page, pageSize int64) ([]DomainBlacklist, int64, error) {
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
	query := "select " + domainBlacklistRows + " from " + m.table + " order by id desc limit ? offset ?"
	var items []DomainBlacklist
	if err := m.conn.QueryRows(&items, query, pageSize, offset); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// IncrAttempts 命中拦截时累加尝试次数
func (m *DomainBlacklistModel) IncrAttempts(ctx context.Context, domain string) error {
	query := "update " + m.table + " set attempts = attempts + 1 where domain = ?"
	_, err := m.conn.Exec(query, domain)
	return err
}
