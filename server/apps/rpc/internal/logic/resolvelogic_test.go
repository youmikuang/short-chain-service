package logic

import (
	"context"
	"testing"

	"server/apps/rpc/internal/config"
	"server/apps/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
)

// TestClickHouseProbe 验证 rpc 的 click_events 写入链路可用：
// 加载 etc/shortlink.yaml 的真实配置，构建 ServiceContext（含 ClickHouse 连接），
// 调用 ResolveLogic.ProbeClickHouse 插入并读回一条探针记录。
//
// 运行：
//   cd server && go test ./apps/rpc/internal/logic/ -run TestClickHouseProbe -v
//
// 若失败，错误会直接指出是 insert 还是 read-back 环节（通常根因是 ClickHouse 的
// Host/Port 配错：clickhouse-go 走原生协议，必须用 9000，不能用 HTTP 的 8123）。
func TestClickHouseProbe(t *testing.T) {
	var c config.Config
	// 测试在包目录执行，配置位于 ../../etc/shortlink.yaml
	conf.MustLoad("../../etc/shortlink.yaml", &c)

	ctx := svc.NewServiceContext(c)
	l := NewResolveLogic(context.Background(), ctx)

	if err := l.ProbeClickHouse(context.Background()); err != nil {
		t.Fatalf("ClickHouse probe failed: %v", err)
	}
	t.Log("ClickHouse insert verified: probe row written and read back successfully")
}
