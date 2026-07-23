package errorx

import (
	"errors"
	"testing"
)

func TestErrorCodes(t *testing.T) {
	cases := []struct {
		err  error
		code int
	}{
		{Internal("x"), CodeInternal},
		{BadParam("x"), CodeInvalidParam},
		{Unauthorized("x"), CodeUnauthorized},
		{Forbidden("x"), CodeForbidden},
		{NotFound("x"), CodeNotFound},
		{RateLimited("x"), CodeRateLimited},
		{Blacklisted("evil.com"), CodeBlacklisted},
	}
	for _, c := range cases {
		var e *Error
		if !errors.As(c.err, &e) {
			t.Fatalf("%v is not *Error", c.err)
		}
		if e.Code != c.code {
			t.Fatalf("got code %d, want %d", e.Code, c.code)
		}
		if e.Error() == "" {
			t.Fatalf("error message empty for code %d", c.code)
		}
	}
}

func TestIs(t *testing.T) {
	if !Is(Internal("a"), Internal("b")) {
		t.Fatal("Is should match same code")
	}
	if Is(NotFound("a"), Internal("b")) {
		t.Fatal("Is should not match different code")
	}
	if Is(errors.New("plain"), Internal("b")) {
		t.Fatal("Is should not match non-*Error")
	}
	// 包装后仍可识别
	wrapped := errors.Join(Internal("inner"), errors.New("other"))
	if !Is(wrapped, Internal("x")) {
		t.Fatal("Is should find *Error inside joined error")
	}
}

func TestBlacklistedMessage(t *testing.T) {
	err := Blacklisted("evil.com")
	if err.Error() != "domain blacklisted: evil.com" {
		t.Fatalf("unexpected message: %q", err.Error())
	}
}
