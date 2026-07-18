package model

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ShortLink struct {
	Id        int64  `db:"id"`
	Code      string `db:"code"`
	LongURL   string `db:"long_url"`
	UserId    int64  `db:"user_id"`
	Clicks    int64  `db:"clicks"`
	Status    int64  `db:"status"`
	Source    string `db:"source"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type ShortLinkModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewShortLinkModel(conn sqlx.SqlConn) *ShortLinkModel {
	return &ShortLinkModel{conn: conn, table: "`short_links`"}
}

const shortLinkRows = "id, code, long_url, user_id, clicks, status, source, created_at, updated_at"

// shortLinkJoinRows 联表查询时给列加表别名，避免与 users.id 冲突
const shortLinkJoinRows = "sl.id, sl.code, sl.long_url, sl.user_id, sl.clicks, sl.status, sl.source, sl.created_at, sl.updated_at"

func (m *ShortLinkModel) Insert(ctx context.Context, data *ShortLink) (sql.Result, error) {
	query := "insert into " + m.table + " (code, long_url, user_id, clicks, status, source) values (?, ?, ?, ?, ?, ?)"
	return m.conn.Exec(query, data.Code, data.LongURL, data.UserId, data.Clicks, data.Status, data.Source)
}

func (m *ShortLinkModel) FindOneByCode(ctx context.Context, code string) (*ShortLink, error) {
	query := "select " + shortLinkRows + " from " + m.table + " where code = ? limit 1"
	var resp ShortLink
	err := m.conn.QueryRow(&resp, query, code)
	return &resp, err
}

// FindOneByUserAndURL 查询同一用户是否已对相同长链接生成过短链（去重复用）
func (m *ShortLinkModel) FindOneByUserAndURL(ctx context.Context, userId int64, longURL string) (*ShortLink, error) {
	query := "select " + shortLinkRows + " from " + m.table + " where user_id = ? and long_url = ? limit 1"
	var resp ShortLink
	err := m.conn.QueryRow(&resp, query, userId, longURL)
	return &resp, err
}

// FindPage 管理后台分页列表
func (m *ShortLinkModel) FindPage(ctx context.Context, page, pageSize int64) ([]ShortLink, int64, error) {
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
	query := "select " + shortLinkRows + " from " + m.table + " order by id desc limit ? offset ?"
	var items []ShortLink
	if err := m.conn.QueryRows(&items, query, pageSize, offset); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (m *ShortLinkModel) IncrClicks(ctx context.Context, code string) error {
	query := "update " + m.table + " set clicks = clicks + 1 where code = ?"
	_, err := m.conn.Exec(query, code)
	return err
}

// UpdateSource 复用同一用户+同长链接的短链时，将 source 更新为「最后生成」的来源
// （web / api 谁后生成算谁的），并刷新 updated_at 以反映最近一次更新。
func (m *ShortLinkModel) UpdateSource(ctx context.Context, code, source string) error {
	query := "update " + m.table + " set source = ?, updated_at = now() where code = ?"
	_, err := m.conn.Exec(query, source, code)
	return err
}

func (m *ShortLinkModel) Delete(ctx context.Context, code string) error {
	query := "delete from " + m.table + " where code = ?"
	_, err := m.conn.Exec(query, code)
	return err
}

// FindPageByUser 用户维度分页列表（仅本人创建的短链）。
// search 可选，按 code / long_url 模糊匹配；sort 可选（"asc"/"desc"）按创建时间排序，缺省按 id 倒序（最新在前）。
func (m *ShortLinkModel) FindPageByUser(ctx context.Context, userId, page, pageSize int64, search, sort string) ([]ShortLink, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	where := " WHERE user_id = ?"
	args := []interface{}{userId}
	if search != "" {
		where += " AND (code LIKE ? OR long_url LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	var total int64
	if err := m.conn.QueryRow(&total, "select count(*) from "+m.table+where, args...); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	order := "id desc"
	if sort == "asc" {
		order = "created_at asc"
	} else if sort == "desc" {
		order = "created_at desc"
	}
	query := "select " + shortLinkRows + " from " + m.table + where + " order by " + order + " limit ? offset ?"
	listArgs := append(append([]interface{}{}, args...), pageSize, offset)
	var items []ShortLink
	if err := m.conn.QueryRows(&items, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// CountWhere 按状态统计短链数量
func (m *ShortLinkModel) CountWhere(ctx context.Context, status int64) (int64, error) {
	var total int64
	err := m.conn.QueryRow(&total, "select count(*) from "+m.table+" where status = ?", status)
	return total, err
}

// SumClicks 全站累计访问次数
func (m *ShortLinkModel) SumClicks(ctx context.Context) (int64, error) {
	var total int64
	err := m.conn.QueryRow(&total, "select coalesce(sum(clicks), 0) from "+m.table)
	return total, err
}

// ShortLinkWithUser 联表后的短链记录（含创建者昵称/邮箱）
type ShortLinkWithUser struct {
	ShortLink
	UserName  string `db:"user_name"`
	UserEmail string `db:"user_email"`
}

// FindPageWithUser 管理后台分页列表（联表取用户信息）
func (m *ShortLinkModel) FindPageWithUser(ctx context.Context, page, pageSize int64) ([]ShortLinkWithUser, int64, error) {
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
	query := "select " + shortLinkJoinRows + ", COALESCE(u.nickname,'') as user_name, COALESCE(u.email,'') as user_email from " +
		m.table + " sl left join `users` u on sl.user_id = u.id order by sl.id desc limit ? offset ?"
	var items []ShortLinkWithUser
	if err := m.conn.QueryRows(&items, query, pageSize, offset); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
