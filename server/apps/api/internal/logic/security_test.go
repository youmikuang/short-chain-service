package logic

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"server/common/errorx"

	"github.com/golang-jwt/jwt/v4"
)

func TestHashAndCheckPassword(t *testing.T) {
	hashed, err := hashPassword("secret123")
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	if hashed == "secret123" {
		t.Fatal("password not hashed")
	}
	if err := checkPassword(hashed, "secret123"); err != nil {
		t.Fatalf("checkPassword(valid) failed: %v", err)
	}
	if err := checkPassword(hashed, "wrong"); err == nil {
		t.Fatal("checkPassword(invalid) should fail")
	}
}

func TestIssueToken(t *testing.T) {
	const secret = "test-secret"
	token, err := issueToken(secret, 3600, 42)
	if err != nil {
		t.Fatalf("issueToken: %v", err)
	}
	parsed, err := jwt.Parse(token, func(tk *jwt.Token) (interface{}, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("parse token failed: %v", err)
	}
	claims := parsed.Claims.(jwt.MapClaims)
	if int64(claims["uid"].(float64)) != 42 {
		t.Fatalf("uid claim = %v, want 42", claims["uid"])
	}
	if _, ok := claims["exp"]; !ok {
		t.Fatal("missing exp claim")
	}
	// 错误密钥应校验失败
	if _, err := jwt.Parse(token, func(tk *jwt.Token) (interface{}, error) {
		return []byte("wrong"), nil
	}); err == nil {
		t.Fatal("token verified with wrong secret")
	}
}

func TestUidFromCtx(t *testing.T) {
	cases := []struct {
		name    string
		ctx     context.Context
		want    int64
		wantErr bool
	}{
		{"float64", context.WithValue(context.Background(), "uid", float64(7)), 7, false},
		{"json.Number", context.WithValue(context.Background(), "uid", json.Number("9")), 9, false},
		{"int64", context.WithValue(context.Background(), "uid", int64(11)), 11, false},
		{"int", context.WithValue(context.Background(), "uid", int(13)), 13, false},
		{"string", context.WithValue(context.Background(), "uid", "15"), 15, false},
		{"missing", context.Background(), 0, true},
		{"invalid type", context.WithValue(context.Background(), "uid", true), 0, true},
		{"bad string", context.WithValue(context.Background(), "uid", "abc"), 0, true},
	}
	for _, c := range cases {
		got, err := uidFromCtx(c.ctx)
		if c.wantErr {
			if err == nil {
				t.Fatalf("%s: expected error", c.name)
			}
			var e *errorx.Error
			if !errors.As(err, &e) {
				t.Fatalf("%s: error not *errorx.Error: %v", c.name, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", c.name, err)
		}
		if got != c.want {
			t.Fatalf("%s: got %d, want %d", c.name, got, c.want)
		}
	}
}

func TestBoolToInt(t *testing.T) {
	if boolToInt(true) != 1 || boolToInt(false) != 0 {
		t.Fatal("boolToInt wrong")
	}
}
