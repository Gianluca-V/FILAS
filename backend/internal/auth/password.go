package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// IsBcryptHash reports whether hash looks like a bcrypt digest ($2a$/$2b$/
// $2y$ prefix), as opposed to a legacy sha256(salt+password) hex digest.
// Admin rows migrate lazily (see usecase.AuthService.Login): a legacy row's
// Password column holds a 64-char hex string with none of these prefixes.
func IsBcryptHash(hash string) bool {
	return strings.HasPrefix(hash, "$2a$") ||
		strings.HasPrefix(hash, "$2b$") ||
		strings.HasPrefix(hash, "$2y$")
}

// HashPassword bcrypt-hashes raw at bcrypt.DefaultCost, for both new admin
// creation and the migrate-on-login upgrade path.
func HashPassword(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyBcrypt reports whether raw matches the given bcrypt hash.
func VerifyBcrypt(hash, raw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw)) == nil
}

// VerifyLegacySHA256 reproduces the legacy PHP scheme
// (hash('sha256', $salt . $rawPassword), characterized live against
// FilasServer/admins.php — see backend/docs/legacy-quirks.md §3) so a
// not-yet-migrated admin row can still log in. The comparison is
// constant-time to avoid leaking hash-match timing.
func VerifyLegacySHA256(salt, raw, storedHex string) bool {
	sum := sha256.Sum256([]byte(salt + raw))
	computed := hex.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(computed), []byte(storedHex)) == 1
}
