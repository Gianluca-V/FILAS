package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

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

// --- alg-confusion guards (fix #4a from the PR1 gate review) ---

func TestParse_RejectsNoneAlgorithmToken(t *testing.T) {
	svc := auth.NewJWTService("test-secret", time.Hour)

	claims := jwt.MapClaims{
		"exp":  time.Now().Add(time.Hour).Unix(),
		"data": 1543,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("SignedString() error = %v, want nil", err)
	}

	if _, err := svc.Parse(signed); err == nil {
		t.Fatal("Parse() error = nil, want rejection of alg=none token")
	}
}

func TestParse_RejectsRS256Token(t *testing.T) {
	svc := auth.NewJWTService("test-secret", time.Hour)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() error = %v, want nil", err)
	}

	claims := jwt.MapClaims{
		"exp":  time.Now().Add(time.Hour).Unix(),
		"data": 1543,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v, want nil", err)
	}

	if _, err := svc.Parse(signed); err == nil {
		t.Fatal("Parse() error = nil, want rejection of RS256-signed token (alg confusion)")
	}
}

// --- legacy PHP contract-fidelity (fix #4b) ---

func TestParse_AcceptsLegacyPHPClaimShape(t *testing.T) {
	// Mirrors FilasServer/JWT.php's CreateToken payload EXACTLY:
	// {"exp": <ts>, "data": <adminID>} — no "iat", no other claims.
	secret := "test-secret"
	claims := jwt.MapClaims{
		"exp":  time.Now().Add(3 * time.Hour).Unix(),
		"data": 1543,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v, want nil", err)
	}

	svc := auth.NewJWTService(secret, time.Hour)
	adminID, err := svc.Parse(signed)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil for legacy-shaped token", err)
	}
	if adminID != 1543 {
		t.Errorf("Parse() adminID = %d, want %d", adminID, 1543)
	}
}

// --- malformed claims (fix #4c) ---

func TestParse_RejectsTokenMissingDataClaim(t *testing.T) {
	secret := "test-secret"
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		// "data" intentionally absent
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v, want nil", err)
	}

	svc := auth.NewJWTService(secret, time.Hour)
	if _, err := svc.Parse(signed); err == nil {
		t.Fatal("Parse() error = nil, want rejection when data claim is absent")
	}
}

func TestParse_RejectsStringDataClaim(t *testing.T) {
	secret := "test-secret"
	claims := jwt.MapClaims{
		"exp":  time.Now().Add(time.Hour).Unix(),
		"data": "1543", // wrong type: string instead of number
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v, want nil", err)
	}

	svc := auth.NewJWTService(secret, time.Hour)
	if _, err := svc.Parse(signed); err == nil {
		t.Fatal("Parse() error = nil, want rejection for string-typed data claim")
	}
}

func TestParse_RejectsFloatDataClaim(t *testing.T) {
	secret := "test-secret"
	claims := jwt.MapClaims{
		"exp":  time.Now().Add(time.Hour).Unix(),
		"data": 1543.7, // wrong type: non-integer float instead of integer
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v, want nil", err)
	}

	svc := auth.NewJWTService(secret, time.Hour)
	if _, err := svc.Parse(signed); err == nil {
		t.Fatal("Parse() error = nil, want rejection for float-typed data claim")
	}
}

func TestParse_RejectsTokenMissingExpClaim(t *testing.T) {
	secret := "test-secret"
	claims := jwt.MapClaims{
		"data": 1543,
		// "exp" intentionally absent
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v, want nil", err)
	}

	svc := auth.NewJWTService(secret, time.Hour)
	if _, err := svc.Parse(signed); err == nil {
		t.Fatal("Parse() error = nil, want rejection when exp claim is absent")
	}
}
