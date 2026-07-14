package model

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type User struct {
	Id           int64  `db:"id"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
	Nickname     string `db:"nickname"`
	GithubId     string `db:"github_id"`
	Avatar       string `db:"avatar"`
	Status       int64  `db:"status"`
	CreatedAt    string `db:"created_at"`
	UpdatedAt    string `db:"updated_at"`
}

type UserModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewUserModel(conn sqlx.SqlConn) *UserModel {
	return &UserModel{conn: conn, table: "`users`"}
}

const userRows = "id, email, password_hash, nickname, github_id, avatar, status, created_at, updated_at"

func (m *UserModel) Insert(ctx context.Context, data *User) (sql.Result, error) {
	query := "insert into " + m.table + " (email, password_hash, nickname, github_id, avatar, status) values (?, ?, ?, ?, ?, ?)"
	return m.conn.Exec(query, data.Email, data.PasswordHash, data.Nickname, data.GithubId, data.Avatar, data.Status)
}

func (m *UserModel) FindOneById(ctx context.Context, id int64) (*User, error) {
	query := "select " + userRows + " from " + m.table + " where id = ? limit 1"
	var resp User
	err := m.conn.QueryRow(&resp, query, id)
	return &resp, err
}

func (m *UserModel) FindOneByEmail(ctx context.Context, email string) (*User, error) {
	query := "select " + userRows + " from " + m.table + " where email = ? limit 1"
	var resp User
	err := m.conn.QueryRow(&resp, query, email)
	return &resp, err
}

func (m *UserModel) FindOneByGithubId(ctx context.Context, githubId string) (*User, error) {
	query := "select " + userRows + " from " + m.table + " where github_id = ? limit 1"
	var resp User
	err := m.conn.QueryRow(&resp, query, githubId)
	return &resp, err
}

func (m *UserModel) Update(ctx context.Context, data *User) error {
	query := "update " + m.table + " set email = ?, password_hash = ?, nickname = ?, github_id = ?, avatar = ?, status = ? where id = ?"
	_, err := m.conn.Exec(query, data.Email, data.PasswordHash, data.Nickname, data.GithubId, data.Avatar, data.Status, data.Id)
	return err
}
