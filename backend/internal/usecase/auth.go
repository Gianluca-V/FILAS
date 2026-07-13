package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gianluca-v/filas-backend/internal/auth"
	"github.com/gianluca-v/filas-backend/internal/domain"
)

// jwtIssuer is satisfied by *auth.JWTService; kept narrow for testability.
type jwtIssuer interface {
	Generate(adminID int64) (string, error)
}

// AuthService implements admin login, including the bcrypt +
// migrate-on-login flow from design §2.6. Consumed by internal/handler/rest.
type AuthService struct {
	repo domain.AdminRepository
	jwt  jwtIssuer
}

// NewAuthService wires an AuthService to its admin repository and JWT
// issuer.
func NewAuthService(repo domain.AdminRepository, jwt jwtIssuer) *AuthService {
	return &AuthService{repo: repo, jwt: jwt}
}

// Login verifies username/password and returns a signed JWT on success.
//
// Password verification tries bcrypt first; if the stored hash is not
// bcrypt-shaped, it falls back to the legacy sha256(salt+password) scheme.
// A successful legacy verification transparently re-hashes the password to
// bcrypt and persists it (migrate-on-login) — the admin never notices, and
// no forced reset is required. The persist step is best-effort: a failure
// there does not fail the login (it would simply retry on the next legacy
// login), since the password was already correctly verified.
//
// Unknown username and wrong password both return the same
// domain.ErrUnauthorized with no distinguishing detail, avoiding a
// user-enumeration signal.
func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	admin, err := s.repo.GetByUsername(ctx, username)
	if errors.Is(err, domain.ErrNotFound) {
		return "", domain.ErrUnauthorized
	}
	if err != nil {
		return "", fmt.Errorf("usecase: login lookup: %w", err)
	}

	if auth.IsBcryptHash(admin.Password) {
		if !auth.VerifyBcrypt(admin.Password, password) {
			return "", domain.ErrUnauthorized
		}
	} else {
		if !auth.VerifyLegacySHA256(admin.Salt, password, admin.Password) {
			return "", domain.ErrUnauthorized
		}
		if newHash, hashErr := auth.HashPassword(password); hashErr == nil {
			_ = s.repo.UpdatePassword(ctx, admin.ID, newHash)
		}
	}

	token, err := s.jwt.Generate(int64(admin.ID))
	if err != nil {
		return "", fmt.Errorf("usecase: generate login token: %w", err)
	}
	return token, nil
}
