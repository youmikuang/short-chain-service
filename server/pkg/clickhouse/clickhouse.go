package clickhouse

import (
	"database/sql"
	"fmt"

	_ "github.com/ClickHouse/clickhouse-go/v2"
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
//
// 这里额外配置连接池与超时，避免每次请求都重建 TCP/TLS 连接（远程 ClickHouse
// 跨公网时握手代价很高），并让空闲连接被复用，显著降低查询延迟。
func NewClient(c Config) (*sql.DB, error) {
	// 仅保留驱动层可识别的 dial_timeout；read_timeout/write_timeout 会被驱动当作服务端
	// SETTINGS 转发，而当前 ClickHouse 版本不认，会报 "Unknown setting"。慢查询的超时
	// 由调用方用 context.WithTimeout 兜底（见 user_logic.go 的 3s）。
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s?dial_timeout=5s",
		c.Username, c.Password, c.Host, c.Port, c.Database)
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, err
	}
	// 连接池：保持一定数量的空闲长连接，跨请求复用，省去重复握手。
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	// 0 表示连接永不过期；clickhouse-go 会自动重连被服务端关闭的连接。
	db.SetConnMaxLifetime(0)
	return db, nil
}
