package auth_test

import (
	"testing"
	"time"

	"github.com/marvinf95/bitesense/backend/internal/auth"
)

func TestHashAndVerify(t *testing.T) {
	hash, err := auth.HashPassword("super-secret-pw")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	ok, err := auth.VerifyPassword("super-secret-pw", hash)
	if err != nil || !ok {
		t.Fatalf("verify good: %v %v", ok, err)
	}
	ok, _ = auth.VerifyPassword("wrong", hash)
	if ok {
		t.Fatalf("verify bad must be false")
	}
}

func TestHashRejectsShort(t *testing.T) {
	if _, err := auth.HashPassword("short"); err == nil {
		t.Fatalf("expected error for short password")
	}
}

func TestIssueAndParseAccess(t *testing.T) {
	svc := auth.NewService(nil, []byte("0123456789abcdef0123456789abcdef"), time.Minute, time.Hour)
	tok, err := svc.IssueAccess("user-123", "de")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := svc.ParseAccess(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != "user-123" || claims.Locale != "de" {
		t.Fatalf("bad claims: %+v", claims)
	}
}
