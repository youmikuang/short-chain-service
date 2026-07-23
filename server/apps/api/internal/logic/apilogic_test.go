package logic

import (
	"context"
	"strings"
	"testing"

	"server/apps/api/internal/config"
	"server/apps/api/internal/svc"
	"server/apps/api/internal/types"
	"server/common/model"
	"server/common/tool"

	"github.com/zeromicro/go-zero/core/conf"
)

func newAPITestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()
	var c config.Config
	conf.MustLoad("../../etc/api-api.yaml", &c)
	return svc.NewServiceContext(c)
}

// withUID 把 uid 注入 context（模拟 JWT 中间件写入的 claim），用于构造 logic。
func withUID(uid int64) context.Context {
	return context.WithValue(context.Background(), "uid", float64(uid))
}

func cleanupUser(t *testing.T, ctx *svc.ServiceContext, uid int64) {
	t.Helper()
	_, _ = ctx.Mysql.Exec("delete from `api_keys` where user_id = ?", uid)
	_, _ = ctx.Mysql.Exec("delete from `users` where id = ?", uid)
}

func cleanupLinkByCode(t *testing.T, ctx *svc.ServiceContext, code string) {
	t.Helper()
	_ = ctx.Models.Slink.Delete(context.Background(), code)
}

func TestRegister(t *testing.T) {
	ctx := newAPITestSvc(t)
	l := NewRegisterLogic(context.Background(), ctx)

	email := "reg-" + tool.RandString(8) + "@example.com"
	resp, err := l.Register(&types.RegisterReq{Email: email, Password: "pass1234"})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if resp.UserID <= 0 {
		t.Fatalf("UserID = %d, want > 0", resp.UserID)
	}
	if resp.Token == "" {
		t.Fatal("empty token")
	}
	if !strings.HasPrefix(resp.ApiKey, "slk_") {
		t.Fatalf("ApiKey = %q, want prefix slk_", resp.ApiKey)
	}
	defer cleanupUser(t, ctx, resp.UserID)

	u, derr := ctx.Models.User.FindOneByEmail(context.Background(), email)
	if derr != nil {
		t.Fatalf("FindOneByEmail failed: %v", derr)
	}
	if u.Id != resp.UserID {
		t.Fatalf("user id mismatch")
	}
	// 重复注册
	if _, err := l.Register(&types.RegisterReq{Email: email, Password: "x"}); err == nil {
		t.Fatal("duplicate email should be BadParam")
	}
	// 空参数
	if _, err := l.Register(&types.RegisterReq{}); err == nil {
		t.Fatal("empty email/password should be BadParam")
	}
	// 非法邮箱
	if _, err := l.Register(&types.RegisterReq{Email: "notanemail", Password: "x"}); err == nil {
		t.Fatal("invalid email should be BadParam")
	}
}

func TestLogin(t *testing.T) {
	ctx := newAPITestSvc(t)
	rl := NewRegisterLogic(context.Background(), ctx)
	ll := NewLoginLogic(context.Background(), ctx)

	email := "login-" + tool.RandString(8) + "@example.com"
	reg, err := rl.Register(&types.RegisterReq{Email: email, Password: "pass1234"})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	defer cleanupUser(t, ctx, reg.UserID)

	resp, err := ll.Login(&types.LoginReq{Email: email, Password: "pass1234"})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.Token == "" || resp.UserID != reg.UserID {
		t.Fatalf("Login resp unexpected: %+v", resp)
	}
	if _, err := ll.Login(&types.LoginReq{Email: email, Password: "wrong"}); err == nil {
		t.Fatal("wrong password should be Unauthorized")
	}
	if _, err := ll.Login(&types.LoginReq{}); err == nil {
		t.Fatal("empty credentials should be BadParam")
	}
}

func TestListMyLinks(t *testing.T) {
	ctx := newAPITestSvc(t)
	rl := NewRegisterLogic(context.Background(), ctx)

	reg, err := rl.Register(&types.RegisterReq{Email: "links-" + tool.RandString(8) + "@example.com", Password: "p"})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	defer cleanupUser(t, ctx, reg.UserID)

	code := "myl" + tool.RandString(5)
	if _, err := ctx.Models.Slink.Insert(context.Background(), &model.Slink{
		Code: code, LongURL: "https://example.com/mylinks", UserId: reg.UserID, Status: 1, Source: "web",
	}); err != nil {
		t.Fatalf("insert slink failed: %v", err)
	}
	defer cleanupLinkByCode(t, ctx, code)

	ll := NewListMyLinksLogic(withUID(reg.UserID), ctx)
	resp, err := ll.ListMyLinks(&types.ListMyLinksReq{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListMyLinks failed: %v", err)
	}
	if resp.Total < 1 {
		t.Fatalf("Total = %d, want >= 1", resp.Total)
	}
	found := false
	for _, it := range resp.Items {
		if it.Code == code {
			found = true
		}
	}
	if !found {
		t.Fatal("created link not found in list")
	}
	// 缺 uid → Unauthorized
	if _, err := NewListMyLinksLogic(context.Background(), ctx).ListMyLinks(&types.ListMyLinksReq{}); err == nil {
		t.Fatal("missing uid should be Unauthorized")
	}
}

func TestCreateAndRevokeAPIKey(t *testing.T) {
	ctx := newAPITestSvc(t)
	rl := NewRegisterLogic(context.Background(), ctx)

	reg, err := rl.Register(&types.RegisterReq{Email: "key-" + tool.RandString(8) + "@example.com", Password: "p"})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	defer cleanupUser(t, ctx, reg.UserID)

	ck := NewCreateAPIKeyLogic(withUID(reg.UserID), ctx)
	created, err := ck.CreateAPIKey(&types.CreateAPIKeyReq{Name: "test-key"})
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}
	if !strings.HasPrefix(created.Key, "slk_") {
		t.Fatalf("key = %q, want prefix slk_", created.Key)
	}
	defer func() { _, _ = ctx.Mysql.Exec("delete from `api_keys` where id = ?", created.Id) }()

	lk := NewListAPIKeysLogic(withUID(reg.UserID), ctx)
	list, err := lk.ListAPIKeys(&types.ListAPIKeysReq{})
	if err != nil {
		t.Fatalf("ListAPIKeys failed: %v", err)
	}
	has := false
	for _, it := range list.Items {
		if it.Id == created.Id {
			has = true
		}
	}
	if !has {
		t.Fatal("created api key not in list")
	}

	rk := NewRevokeAPIKeyLogic(withUID(reg.UserID), ctx)
	if _, err := rk.RevokeAPIKey(&types.RevokeAPIKeyReq{Id: created.Id}); err != nil {
		t.Fatalf("RevokeAPIKey failed: %v", err)
	}
	row, derr := ctx.Models.ApiKey.FindOneById(context.Background(), created.Id)
	if derr != nil {
		t.Fatalf("FindOneById failed: %v", derr)
	}
	if row.Status != 0 {
		t.Fatalf("status = %d, want 0 after revoke", row.Status)
	}
	// 缺 name → BadParam
	if _, err := ck.CreateAPIKey(&types.CreateAPIKeyReq{}); err == nil {
		t.Fatal("empty name should be BadParam")
	}
}

func TestProfileAndSettings(t *testing.T) {
	ctx := newAPITestSvc(t)
	rl := NewRegisterLogic(context.Background(), ctx)

	email := "prof-" + tool.RandString(8) + "@example.com"
	reg, err := rl.Register(&types.RegisterReq{Email: email, Password: "p"})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	defer cleanupUser(t, ctx, reg.UserID)

	uidCtx := withUID(reg.UserID)
	gp := NewGetProfileLogic(uidCtx, ctx)
	up := NewUpdateProfileLogic(uidCtx, ctx)
	gs := NewGetSettingsLogic(uidCtx, ctx)
	us := NewUpdateSettingsLogic(uidCtx, ctx)

	p, err := gp.GetProfile()
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}
	if p.Email != email {
		t.Fatalf("email = %q, want %q", p.Email, email)
	}
	upd, err := up.UpdateProfile(&types.UpdateProfileReq{Nickname: "NewName"})
	if err != nil {
		t.Fatalf("UpdateProfile failed: %v", err)
	}
	if upd.Nickname != "NewName" {
		t.Fatalf("nickname = %q, want NewName", upd.Nickname)
	}
	s, err := gs.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}
	if !s.EmailNotif || !s.SecurityAlerts {
		t.Fatal("default settings mismatch")
	}
	s2, err := us.UpdateSettings(&types.UpdateSettingsReq{EmailNotif: false, SecurityAlerts: false, MarketingComm: true})
	if err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}
	if s2.EmailNotif || s2.SecurityAlerts || !s2.MarketingComm {
		t.Fatalf("updated settings mismatch: %+v", s2)
	}
	if _, err := NewGetProfileLogic(withUID(999999999), ctx).GetProfile(); err == nil {
		t.Fatal("non-existent user should be NotFound")
	}
}

func TestChangePassword(t *testing.T) {
	ctx := newAPITestSvc(t)
	rl := NewRegisterLogic(context.Background(), ctx)
	ll := NewLoginLogic(context.Background(), ctx)

	email := "chg-" + tool.RandString(8) + "@example.com"
	reg, err := rl.Register(&types.RegisterReq{Email: email, Password: "oldpass12"})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	defer cleanupUser(t, ctx, reg.UserID)
	uidCtx := withUID(reg.UserID)
	cp := NewChangePasswordLogic(uidCtx, ctx)

	if _, err := cp.ChangePassword(&types.ChangePasswordReq{CurrentPassword: "oldpass12", NewPassword: "newpass34"}); err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}
	if _, err := ll.Login(&types.LoginReq{Email: email, Password: "oldpass12"}); err == nil {
		t.Fatal("old password should fail after change")
	}
	if _, err := ll.Login(&types.LoginReq{Email: email, Password: "newpass34"}); err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}
	if _, err := cp.ChangePassword(&types.ChangePasswordReq{CurrentPassword: "wrong", NewPassword: "x"}); err == nil {
		t.Fatal("wrong current password should be Unauthorized")
	}
	if _, err := cp.ChangePassword(&types.ChangePasswordReq{CurrentPassword: "newpass34"}); err == nil {
		t.Fatal("empty new password should be BadParam")
	}
}

func TestGitHubAuthURL(t *testing.T) {
	ctx := newAPITestSvc(t)
	l := NewGitHubAuthURLLogic(context.Background(), ctx)
	resp, err := l.GitHubAuthURL(&types.GitHubAuthURLReq{State: "abc"})
	if err != nil {
		t.Fatalf("GitHubAuthURL failed: %v", err)
	}
	if !strings.Contains(resp.Url, "github.com/login/oauth/authorize") {
		t.Fatalf("unexpected url: %q", resp.Url)
	}
	if !strings.Contains(resp.Url, "client_id="+ctx.Config.Github.ClientID) {
		t.Fatalf("url missing client_id: %q", resp.Url)
	}
	if !strings.Contains(resp.Url, "state=abc") {
		t.Fatalf("url missing state: %q", resp.Url)
	}
}

func TestUsageTrendsAndLogs(t *testing.T) {
	ctx := newAPITestSvc(t)
	rl := NewRegisterLogic(context.Background(), ctx)

	reg, err := rl.Register(&types.RegisterReq{Email: "usage-" + tool.RandString(8) + "@example.com", Password: "p"})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	defer cleanupUser(t, ctx, reg.UserID)
	uidCtx := withUID(reg.UserID)
	ut := NewUsageTrendsLogic(uidCtx, ctx)
	lg := NewLogsLogic(uidCtx, ctx)

	utResp, err := ut.UsageTrends(&types.UsageTrendsReq{})
	if err != nil {
		t.Fatalf("UsageTrends failed: %v", err)
	}
	if len(utResp.Items) != 30 {
		t.Fatalf("UsageTrends items = %d, want 30", len(utResp.Items))
	}
	utResp7, err := ut.UsageTrends(&types.UsageTrendsReq{Days: 7})
	if err != nil {
		t.Fatalf("UsageTrends(7) failed: %v", err)
	}
	if len(utResp7.Items) != 7 {
		t.Fatalf("UsageTrends(7) items = %d, want 7", len(utResp7.Items))
	}
	logs, err := lg.Logs(&types.LogsReq{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Logs failed: %v", err)
	}
	if logs.Items == nil {
		t.Fatal("Logs items nil")
	}
}
