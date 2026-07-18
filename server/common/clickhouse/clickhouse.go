package clickhouse

import (
	"database/sql"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// Config ClickHouse 连接配置（对应各服务 etc 中的 ClickHouse 段）
type Config struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// NewClient 打开一个 *sql.DB（惰性连接，不在此处 Ping）。
// clickhouse-go 已注册 "clickhouse" 驱动，dsn 形如：
//
//	clickhouse://user:pass@host:9000/db
func NewClient(c Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s",
		c.Username, c.Password, c.Host, c.Port, c.Database)
	return sql.Open("clickhouse", dsn)
}
