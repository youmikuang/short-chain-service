package logic

import (
	"context"
	"errors"
	"testing"

	"server/apps/admin/internal/config"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/common/model"
	"server/common/tool"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func isNotFoundAdmin(err error) bool {
	return errors.Is(err, sqlx.ErrNotFound)
}

// requireClickHouse 在 ClickHouse 不可用时跳过测试（远程 CH 偶发握手失败）。
// Dashboard 同步依赖 ClickHouse，CH 不可达时整体报错，故以 skip 保护集成测试。
func requireClickHouse(t *testing.T, ctx *svc.ServiceContext) {
	t.Helper()
	if err := ctx.ClickHouse.Ping(); err != nil {
		t.Skipf("ClickHouse unavailable, skip: %v", err)
	}
}

func newAdminTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()
	var c config.Config
	conf.MustLoad("../../etc/admin-api.yaml", &c)
	return svc.NewServiceContext(c)
}

func cleanupBlacklist(t *testing.T, ctx *svc.ServiceContext, domain string) {
	t.Helper()
	_ = ctx.Models.DomainBlacklist.Delete(context.Background(), domain)
	if ctx.Config.BlacklistRedisKey != "" {
		_ = ctx.Redis.SRem(context.Background(), ctx.Config.BlacklistRedisKey, domain).Err()
	}
}

func cleanupLinkByCode(t *testing.T, ctx *svc.ServiceContext, code string) {
	t.Helper()
	_ = ctx.Models.Slink.Delete(context.Background(), code)
}

func cleanupApiKey(t *testing.T, ctx *svc.ServiceContext, id int64) {
	t.Helper()
	_, _ = ctx.Mysql.Exec("delete from `api_keys` where id = ?", id)
}

func TestAdminLogin(t *testing.T) {
	ctx := newAdminTestSvc(t)
	l := NewAdminLoginLogic(context.Background(), ctx)

	resp, err := l.AdminLogin(&types.AdminLoginReq{Username: ctx.Config.Admin.Username, Password: ctx.Config.Admin.Password})
	if err != nil {
		t.Fatalf("AdminLogin failed: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("empty token")
	}
	// 错误凭据
	if _, err := l.AdminLogin(&types.AdminLoginReq{Username: "admin", Password: "wrong"}); err == nil {
		t.Fatal("wrong password should be Unauthorized")
	}
	// 空凭据
	if _, err := l.AdminLogin(&types.AdminLoginReq{}); err == nil {
		t.Fatal("empty credentials should be Unauthorized")
	}
}

func TestAddAndDeleteBlacklist(t *testing.T) {
	ctx := newAdminTestSvc(t)
	al := NewAddBlacklistLogic(context.Background(), ctx)
	dl := NewDeleteBlacklistLogic(context.Background(), ctx)

	domain := "adm-bl-" + tool.RandString(8) + ".example"
	resp, err := al.AddBlacklist(&types.AddBlacklistReq{Domain: domain, Reason: "spam"})
	if err != nil {
		t.Fatalf("AddBlacklist failed: %v", err)
	}
	if !resp.Ok {
		t.Fatal("AddBlacklist Ok=false")
	}
	defer cleanupBlacklist(t, ctx, domain)

	// 落库校验
	if _, derr := ctx.Models.DomainBlacklist.FindOneByDomain(context.Background(), domain); derr != nil {
		t.Fatalf("FindOneByDomain failed: %v", derr)
	}
	// Redis 缓存应已写入
	if ok, rerr := ctx.Redis.SIsMember(context.Background(), ctx.Config.BlacklistRedisKey, domain).Result(); rerr != nil || !ok {
		t.Fatalf("redis blacklist member missing: ok=%v err=%v", ok, rerr)
	}
	// 重复添加 → BadParam
	if _, err := al.AddBlacklist(&types.AddBlacklistReq{Domain: domain}); err == nil {
		t.Fatal("duplicate domain should be BadParam")
	}
	// 删除
	if _, err := dl.DeleteBlacklist(&types.DeleteBlacklistReq{Domain: domain}); err != nil {
		t.Fatalf("DeleteBlacklist failed: %v", err)
	}
	if _, derr := ctx.Models.DomainBlacklist.FindOneByDomain(context.Background(), domain); !isNotFoundAdmin(derr) {
		t.Fatalf("expected NotFound after delete, got %v", derr)
	}
	// 空 domain
	if _, err := al.AddBlacklist(&types.AddBlacklistReq{}); err == nil {
		t.Fatal("empty domain should be BadParam")
	}
	if _, err := dl.DeleteBlacklist(&types.DeleteBlacklistReq{}); err == nil {
		t.Fatal("empty domain should be BadParam")
	}
}

func TestListBlacklist(t *testing.T) {
	ctx := newAdminTestSvc(t)
	al := NewAddBlacklistLogic(context.Background(), ctx)
	ll := NewListBlacklistLogic(context.Background(), ctx)

	domain := "adm-list-" + tool.RandString(8) + ".example"
	if _, err := al.AddBlacklist(&types.AddBlacklistReq{Domain: domain}); err != nil {
		t.Fatalf("AddBlacklist failed: %v", err)
	}
	defer cleanupBlacklist(t, ctx, domain)

	resp, err := ll.ListBlacklist(&types.ListBlacklistReq{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("ListBlacklist failed: %v", err)
	}
	found := false
	for _, it := range resp.Items {
		if it.Domain == domain {
			found = true
		}
	}
	if !found {
		t.Fatal("added domain not in list")
	}
}

func TestListLinks(t *testing.T) {
	ctx := newAdminTestSvc(t)
	ll := NewListLinksLogic(context.Background(), ctx)

	code := "adml" + tool.RandString(5)
	if _, err := ctx.Models.Slink.Insert(context.Background(), &model.Slink{
		Code: code, LongURL: "https://example.com/adminlist", UserId: 1, Status: 1, Source: "web",
	}); err != nil {
		t.Fatalf("insert slink failed: %v", err)
	}
	defer cleanupLinkByCode(t, ctx, code)

	resp, err := ll.ListLinks(&types.ListLinksReq{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("ListLinks failed: %v", err)
	}
	found := false
	for _, it := range resp.Items {
		if it.Code == code {
			found = true
		}
	}
	if !found {
		t.Fatal("inserted link not in admin list")
	}
}

func TestDashboard(t *testing.T) {
	ctx := newAdminTestSvc(t)
	requireClickHouse(t, ctx)
	l := NewDashboardLogic(context.Background(), ctx)
	resp, err := l.Dashboard()
	if err != nil {
		t.Fatalf("Dashboard failed: %v", err)
	}
	if len(resp.Kpis) != 4 {
		t.Fatalf("Kpis len = %d, want 4", len(resp.Kpis))
	}
	if len(resp.Traffic) != 7 {
		t.Fatalf("Traffic len = %d, want 7", len(resp.Traffic))
	}
}

func TestProvisionAndManageTokens(t *testing.T) {
	ctx := newAdminTestSvc(t)
	pl := NewProvisionTokenLogic(context.Background(), ctx)
	rl := NewListTokensLogic(context.Background(), ctx)
	rev := NewRevokeTokenLogic(context.Background(), ctx)
	st := NewStartTokenLogic(context.Background(), ctx)
	rst := NewResetTokenLogic(context.Background(), ctx)

	prov, err := pl.ProvisionToken(&types.ProvisionTokenReq{Name: "adm-tok-" + tool.RandString(4), Quota: 100})
	if err != nil {
		t.Fatalf("ProvisionToken failed: %v", err)
	}
	if prov.Token == "" {
		t.Fatal("empty token")
	}
	// ProvisionTokenResp 不含 id，用返回 token 的哈希反查主键
	keyRow, herr := ctx.Models.ApiKey.FindOneByHash(context.Background(), tool.Sha256Hex(prov.Token))
	if herr != nil {
		t.Fatalf("FindOneByHash failed: %v", herr)
	}
	provId := keyRow.Id
	defer cleanupApiKey(t, ctx, provId)

	// 列表应包含
	list, err := rl.ListTokens(&types.ListTokensReq{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("ListTokens failed: %v", err)
	}
	has := false
	for _, it := range list.Items {
		if it.Id == provId {
			has = true
		}
	}
	if !has {
		t.Fatal("provisioned token not in list")
	}
	// 吊销
	if _, err := rev.RevokeToken(&types.RevokeTokenReq{Id: provId}); err != nil {
		t.Fatalf("RevokeToken failed: %v", err)
	}
	row, derr := ctx.Models.ApiKey.FindOneById(context.Background(), provId)
	if derr != nil {
		t.Fatalf("FindOneById failed: %v", derr)
	}
	if row.Status != 0 {
		t.Fatalf("status = %d, want 0", row.Status)
	}
	// 重新启用
	if _, err := st.StartToken(&types.StartTokenReq{Id: provId}); err != nil {
		t.Fatalf("StartToken failed: %v", err)
	}
	row2, derr := ctx.Models.ApiKey.FindOneById(context.Background(), provId)
	if derr != nil {
		t.Fatalf("FindOneById failed: %v", derr)
	}
	if row2.Status != 1 {
		t.Fatalf("status = %d, want 1 after start", row2.Status)
	}
	// 重置用量
	if _, err := rst.ResetToken(&types.ResetTokenReq{Id: provId}); err != nil {
		t.Fatalf("ResetToken failed: %v", err)
	}
	// 非法 id
	if _, err := rev.RevokeToken(&types.RevokeTokenReq{}); err == nil {
		t.Fatal("id<=0 should be BadParam")
	}
	if _, err := st.StartToken(&types.StartTokenReq{Id: -1}); err == nil {
		t.Fatal("id<=0 should be BadParam")
	}
	if _, err := rst.ResetToken(&types.ResetTokenReq{Id: 0}); err == nil {
		t.Fatal("id<=0 should be BadParam")
	}
}
