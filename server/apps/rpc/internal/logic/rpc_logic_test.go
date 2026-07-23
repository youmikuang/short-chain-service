package logic

import (
	"context"
	"testing"

	"server/apps/rpc/internal/config"
	"server/apps/rpc/internal/svc"
	pb "server/apps/rpc/pb"
	"server/common/model"
	"server/common/tool"

	"github.com/zeromicro/go-zero/core/conf"
)

// newTestSvc 加载 etc/slink.yaml 真实配置构建 ServiceContext（含 MySQL/Redis/ClickHouse 连接）。
// 测试在包目录执行，配置位于 ../../etc/slink.yaml。
func newTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()
	var c config.Config
	conf.MustLoad("../../etc/slink.yaml", &c)
	return svc.NewServiceContext(c)
}

// cleanupLink 删除短链及其 Redis 缓存，避免污染测试库。
func cleanupLink(t *testing.T, ctx *svc.ServiceContext, code string) {
	t.Helper()
	_ = ctx.Models.Slink.Delete(context.Background(), code)
	ctx.Redis.Del(context.Background(),
		"short_link:"+code,
		"short_link:"+code+":uid",
		"short_link:"+code+":source",
		"short_link:"+code+":clicks",
	)
}

func TestCreateSlink_WebPath(t *testing.T) {
	ctx := newTestSvc(t)
	l := NewCreateSlinkLogic(context.Background(), ctx)

	longURL := "https://example.com/web-" + tool.RandString(8)
	resp, err := l.CreateSlink(&pb.CreateSlinkReq{LongUrl: longURL, UserId: 1001})
	if err != nil {
		t.Fatalf("CreateSlink(web) failed: %v", err)
	}
	if resp.Code == "" {
		t.Fatal("empty code")
	}
	defer cleanupLink(t, ctx, resp.Code)

	row, derr := ctx.Models.Slink.FindOneByCode(context.Background(), resp.Code)
	if derr != nil {
		t.Fatalf("FindOneByCode failed: %v", derr)
	}
	if row.LongURL != longURL {
		t.Fatalf("LongURL mismatch: %q vs %q", row.LongURL, longURL)
	}
	if row.Source != "web" {
		t.Fatalf("source = %q, want web", row.Source)
	}
	if got, rerr := ctx.Redis.Get(context.Background(), "short_link:"+resp.Code).Result(); rerr != nil || got != longURL {
		t.Fatalf("redis cache missing: got=%q err=%v", got, rerr)
	}
}

func TestCreateSlink_Unauthorized(t *testing.T) {
	ctx := newTestSvc(t)
	l := NewCreateSlinkLogic(context.Background(), ctx)
	// user_id=0 且无 api key → 未授权
	if _, err := l.CreateSlink(&pb.CreateSlinkReq{LongUrl: "https://example.com/x"}); err == nil {
		t.Fatal("expected Unauthorized")
	}
}

func TestCreateSlink_EmptyAndInvalidURL(t *testing.T) {
	ctx := newTestSvc(t)
	l := NewCreateSlinkLogic(context.Background(), ctx)
	if _, err := l.CreateSlink(&pb.CreateSlinkReq{UserId: 1}); err == nil {
		t.Fatal("empty long_url should be BadParam")
	}
	if _, err := l.CreateSlink(&pb.CreateSlinkReq{LongUrl: "not-a-url", UserId: 1}); err == nil {
		t.Fatal("invalid long_url should be BadParam")
	}
}

func TestCreateSlink_Blacklisted(t *testing.T) {
	ctx := newTestSvc(t)
	l := NewCreateSlinkLogic(context.Background(), ctx)

	domain := "blacklisted-" + tool.RandString(6) + ".example"
	if _, err := ctx.Models.DomainBlacklist.Insert(context.Background(), &model.DomainBlacklist{
		Domain: domain, Reason: "test",
	}); err != nil {
		t.Fatalf("insert blacklist failed: %v", err)
	}
	defer func() { _ = ctx.Models.DomainBlacklist.Delete(context.Background(), domain) }()

	if _, err := l.CreateSlink(&pb.CreateSlinkReq{LongUrl: "https://" + domain + "/p", UserId: 1}); err == nil {
		t.Fatal("expected Blacklisted error")
	}
}

func TestCreateSlink_Dedup(t *testing.T) {
	ctx := newTestSvc(t)
	l := NewCreateSlinkLogic(context.Background(), ctx)

	longURL := "https://example.com/dedup-" + tool.RandString(8)
	r1, err := l.CreateSlink(&pb.CreateSlinkReq{LongUrl: longURL, UserId: 2001})
	if err != nil {
		t.Fatalf("first CreateSlink failed: %v", err)
	}
	defer cleanupLink(t, ctx, r1.Code)

	r2, err := l.CreateSlink(&pb.CreateSlinkReq{LongUrl: longURL, UserId: 2001})
	if err != nil {
		t.Fatalf("second CreateSlink failed: %v", err)
	}
	if r1.Code != r2.Code {
		t.Fatalf("dedup failed: %q vs %q", r1.Code, r2.Code)
	}
}

func TestGetByCode(t *testing.T) {
	ctx := newTestSvc(t)
	cl := NewCreateSlinkLogic(context.Background(), ctx)
	gl := NewGetByCodeLogic(context.Background(), ctx)

	longURL := "https://example.com/getbycode-" + tool.RandString(8)
	cr, err := cl.CreateSlink(&pb.CreateSlinkReq{LongUrl: longURL, UserId: 3001})
	if err != nil {
		t.Fatalf("CreateSlink failed: %v", err)
	}
	defer cleanupLink(t, ctx, cr.Code)

	resp, err := gl.GetByCode(&pb.GetByCodeReq{Code: cr.Code})
	if err != nil {
		t.Fatalf("GetByCode failed: %v", err)
	}
	if resp.LongUrl != longURL {
		t.Fatalf("LongUrl = %q, want %q", resp.LongUrl, longURL)
	}
	if resp2, err := gl.GetByCode(&pb.GetByCodeReq{Code: cr.Code}); err != nil || resp2.LongUrl != longURL {
		t.Fatalf("GetByCode cache hit failed: %v", err)
	}
	if _, err := gl.GetByCode(&pb.GetByCodeReq{Code: "nope" + tool.RandString(6)}); err == nil {
		t.Fatal("expected NotFound")
	}
	if _, err := gl.GetByCode(&pb.GetByCodeReq{}); err == nil {
		t.Fatal("empty code should be BadParam")
	}
}

func TestResolve(t *testing.T) {
	ctx := newTestSvc(t)
	cl := NewCreateSlinkLogic(context.Background(), ctx)
	rl := NewResolveLogic(context.Background(), ctx)

	longURL := "https://example.com/resolve-" + tool.RandString(8)
	cr, err := cl.CreateSlink(&pb.CreateSlinkReq{LongUrl: longURL, UserId: 4001})
	if err != nil {
		t.Fatalf("CreateSlink failed: %v", err)
	}
	defer cleanupLink(t, ctx, cr.Code)

	resp, err := rl.Resolve(&pb.ResolveReq{Code: cr.Code})
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if resp.Blocked {
		t.Fatal("unexpected blocked")
	}
	if resp.LongUrl != longURL {
		t.Fatalf("LongUrl = %q, want %q", resp.LongUrl, longURL)
	}
	if _, err := rl.Resolve(&pb.ResolveReq{Code: "missing" + tool.RandString(6)}); err == nil {
		t.Fatal("expected NotFound for missing code")
	}
}

func TestResolve_Blacklisted(t *testing.T) {
	ctx := newTestSvc(t)
	rl := NewResolveLogic(context.Background(), ctx)

	domain := "blk-" + tool.RandString(6) + ".example"
	if _, err := ctx.Models.DomainBlacklist.Insert(context.Background(), &model.DomainBlacklist{
		Domain: domain, Reason: "test",
	}); err != nil {
		t.Fatalf("insert blacklist failed: %v", err)
	}
	defer func() { _ = ctx.Models.DomainBlacklist.Delete(context.Background(), domain) }()

	code := "blk" + tool.RandString(5)
	if _, err := ctx.Models.Slink.Insert(context.Background(), &model.Slink{
		Code: code, LongURL: "https://" + domain + "/x", UserId: 1, Status: 1, Source: "web",
	}); err != nil {
		t.Fatalf("insert slink failed: %v", err)
	}
	defer cleanupLink(t, ctx, code)

	resp, err := rl.Resolve(&pb.ResolveReq{Code: code})
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if !resp.Blocked {
		t.Fatal("expected Blocked=true for blacklisted domain")
	}
}

func TestBatchCreate(t *testing.T) {
	ctx := newTestSvc(t)
	l := NewBatchCreateLogic(context.Background(), ctx)

	urls := []string{
		"https://example.com/batch1-" + tool.RandString(6),
		"https://example.com/batch2-" + tool.RandString(6),
	}
	resp, err := l.BatchCreate(&pb.BatchCreateReq{LongUrls: urls, UserId: 5001})
	if err != nil {
		t.Fatalf("BatchCreate failed: %v", err)
	}
	if len(resp.Items) != len(urls) {
		t.Fatalf("got %d items, want %d", len(resp.Items), len(urls))
	}
	for _, it := range resp.Items {
		cleanupLink(t, ctx, it.Code)
	}
}

func TestDeleteSlink(t *testing.T) {
	ctx := newTestSvc(t)
	cl := NewCreateSlinkLogic(context.Background(), ctx)
	dl := NewDeleteslinkLogic(context.Background(), ctx)

	longURL := "https://example.com/del-" + tool.RandString(8)
	cr, err := cl.CreateSlink(&pb.CreateSlinkReq{LongUrl: longURL, UserId: 6001})
	if err != nil {
		t.Fatalf("CreateSlink failed: %v", err)
	}
	if _, err := dl.Deleteslink(&pb.DeleteslinkReq{Code: cr.Code}); err != nil {
		t.Fatalf("Deleteslink failed: %v", err)
	}
	if _, derr := ctx.Models.Slink.FindOneByCode(context.Background(), cr.Code); !isNotFound(derr) {
		t.Fatalf("expected NotFound after delete, got %v", derr)
	}
	if _, err := dl.Deleteslink(&pb.DeleteslinkReq{}); err == nil {
		t.Fatal("empty code should be BadParam")
	}
}
