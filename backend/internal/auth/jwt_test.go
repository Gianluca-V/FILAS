package auth_test

import (
	"testing"
	"time"

	"github.com/gianluca-v/filas-backend/internal/auth"
)

func TestGenerateAndParse_RoundTrip(t *testing.T) {
	svc := auth.NewJWTService("test-secret", time.Hour)

	token, err := svc.Generate(1543)
	if err != nil {
		t.Fatalf("Generate() error = %v, want nil", err)
	}
	if token == "" {
		t.Fatal("Generate() returned empty token")
	}

	adminID, err := svc.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if adminID != 1543 {
		t.Errorf("Parse() adminID = %d, want %d", adminID, 1543)
	}
}

func TestParse_DifferentAdminID(t *testing.T) {
	svc := auth.NewJWTService("test-secret", time.Hour)

	token, err := svc.Generate(7)
	if err != nil {
		t.Fatalf("Generate() error = %v, want nil", err)
	}

	adminID, err := svc.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if adminID != 7 {
		t.Errorf("Parse() adminID = %d, want %d", adminID, 7)
	}
}

func TestParse_RejectsInvalidToken(t *testing.T) {
	svc := auth.NewJWTService("test-secret", time.Hour)

	_, err := svc.Parse("not-a-jwt")
	if err == nil {
		t.Fatal("Parse() error = nil, want error for malformed token")
	}
}

func TestParse_RejectsWrongSecret(t *testing.T) {
	issuer := auth.NewJWTService("secret-a", time.Hour)
	verifier := auth.NewJWTService("secret-b", time.Hour)

	token, err := issuer.Generate(1)
	if err != nil {
		t.Fatalf("Generate() error = %v, want nil", err)
	}

	_, err = verifier.Parse(token)
	if err == nil {
		t.Fatal("Parse() error = nil, want error for token signed with different secret")
	}
}

func TestParse_RejectsExpiredToken(t *testing.T) {
	svc := auth.NewJWTService("test-secret", -time.Hour) // already expired

	token, err := svc.Generate(1)
	if err != nil {
		t.Fatalf("Generate() error = %v, want nil", err)
	}

	_, err = svc.Parse(token)
	if err == nil {
		t.Fatal("Parse() error = nil, want error for expired token")
	}
}
