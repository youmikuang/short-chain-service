package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"server/apps/admin/internal/config"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"

	"github.com/zeromicro/go-zero/core/conf"
)

func loadAdminConfig(t *testing.T) config.Config {
	t.Helper()
	var c config.Config
	conf.MustLoad("../../etc/admin-api.yaml", &c)
	return c
}

func TestAdminLoginHandler_OK(t *testing.T) {
	c := loadAdminConfig(t)
	svcCtx := &svc.ServiceContext{Config: c}

	body, _ := json.Marshal(types.AdminLoginReq{Username: c.Admin.Username, Password: c.Admin.Password})
	req := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	AdminLoginHandler(svcCtx)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
	}
	var resp types.AdminLoginResp
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("empty token")
	}
}

func TestAdminLoginHandler_BadCreds(t *testing.T) {
	c := loadAdminConfig(t)
	svcCtx := &svc.ServiceContext{Config: c}

	body, _ := json.Marshal(types.AdminLoginReq{Username: "nope", Password: "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	AdminLoginHandler(svcCtx)(rec, req)

	// errorx.Unauthorized 经 httpx.ErrorCtx 默认写 400
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAdminLoginHandler_BadJSON(t *testing.T) {
	c := loadAdminConfig(t)
	svcCtx := &svc.ServiceContext{Config: c}

	req := httptest.NewRequest(http.MethodPost, "/admin/api/login", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	AdminLoginHandler(svcCtx)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
