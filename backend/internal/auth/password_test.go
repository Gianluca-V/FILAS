package auth_test

import (
	"testing"

	"github.com/gianluca-v/filas-backend/internal/auth"
)

func TestIsBcryptHash(t *testing.T) {
	tests := []struct {
		name string
		hash string
		want bool
	}{
		{"bcrypt 2a prefix", "$2a$10$abcdefghijklmnopqrstuv", true},
		{"bcrypt 2b prefix", "$2b$10$abcdefghijklmnopqrstuv", true},
		{"bcrypt 2y prefix", "$2y$10$abcdefghijklmnopqrstuv", true},
		{"legacy sha256 hex digest", "9a0d462f4985ce28e354f318e2346a3d61881336330e55cca8d9697cad9f1686", false},
		{"empty string", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := auth.IsBcryptHash(tt.hash); got != tt.want {
				t.Errorf("IsBcryptHash(%q) = %v, want %v", tt.hash, got, tt.want)
			}
		})
	}
}

func TestHashPassword_ProducesVerifiableBcryptHash(t *testing.T) {
	hash, err := auth.HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v, want nil", err)
	}
	if !auth.IsBcryptHash(hash) {
		t.Errorf("HashPassword() = %q, want a bcrypt-prefixed hash", hash)
	}
	if !auth.VerifyBcrypt(hash, "correct-horse-battery-staple") {
		t.Errorf("VerifyBcrypt() = false for the password that was just hashed, want true")
	}
}

func TestVerifyBcrypt_RejectsWrongPassword(t *testing.T) {
	hash, err := auth.HashPassword("the-real-password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v, want nil", err)
	}
	if auth.VerifyBcrypt(hash, "not-the-real-password") {
		t.Errorf("VerifyBcrypt() = true for a wrong password, want false")
	}
}

func TestVerifyLegacySHA256_AcceptsMatchingHash(t *testing.T) {
	// Reproduces PHP: hash('sha256', $salt . $rawPassword) — characterized
	// live against the seeded synthetic admin (see
	// sdd/migrate-go-vue/apply-progress obs #31 methodology, admins.php).
	salt := "16308d7827b79fa10077d9a137027f97"
	raw := "filas-local-dev-2026"
	stored := "9a0d462f4985ce28e354f318e2346a3d61881336330e55cca8d9697cad9f1686"

	if !auth.VerifyLegacySHA256(salt, raw, stored) {
		t.Errorf("VerifyLegacySHA256() = false for a known-matching salt/password/hash triple, want true")
	}
}

func TestVerifyLegacySHA256_RejectsWrongPassword(t *testing.T) {
	salt := "16308d7827b79fa10077d9a137027f97"
	stored := "9a0d462f4985ce28e354f318e2346a3d61881336330e55cca8d9697cad9f1686"

	if auth.VerifyLegacySHA256(salt, "wrong-password", stored) {
		t.Errorf("VerifyLegacySHA256() = true for a wrong password, want false")
	}
}

func TestVerifyLegacySHA256_RejectsWrongSalt(t *testing.T) {
	stored := "9a0d462f4985ce28e354f318e2346a3d61881336330e55cca8d9697cad9f1686"

	if auth.VerifyLegacySHA256("not-the-real-salt", "filas-local-dev-2026", stored) {
		t.Errorf("VerifyLegacySHA256() = true for a wrong salt, want false")
	}
}
