package jwt

import (
	"testing"
	"time"
)

func TestGenerateAndParse(t *testing.T) {
	m := New("test-secret", time.Hour)
	token, err := m.Generate(42, "SUPER_ADMIN")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	claims, err := m.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != 42 || claims.Role != "SUPER_ADMIN" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestParseWrongSecret(t *testing.T) {
	m := New("test-secret", time.Hour)
	other := New("other-secret", time.Hour)
	token, _ := m.Generate(1, "USER")
	if _, err := other.Parse(token); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}
