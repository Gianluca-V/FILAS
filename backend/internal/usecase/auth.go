package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/gianluca-v/filas-backend/internal/auth"
	"github.com/gianluca-v/filas-backend/internal/domain"
)

// jwtIssuer is satisfied by *auth.JWTService; kept narrow for testability.
type jwtIssuer interface {
	Generate(adminID int64) (string, error)
}

// dummyBcryptHash is a fixed, precomputed bcrypt hash with no corresponding
// real password. It exists purely to be compared against on the
// user-not-found login path (gate #36 risk finding: timing oracle).
// Without it, an unknown username short-circuits on the repository lookup
// and returns almost instantly, while a KNOWN username with a wrong
// password burns ~50-100ms doing a real bcrypt compare — a measurable,
// attacker-observable timing difference that leaks which usernames exist
// (username enumeration). Comparing against this dummy hash on the
// not-found path burns roughly the same CPU time as the real
// wrong-password path, closing the side channel. The hash itself is
// computed once at package init (bcrypt.GenerateFromPassword is the
// expensive part we do NOT want to repeat per request); only the
// comparison runs per login attempt.
var dummyBcryptHash = mustPrecomputeDummyBcryptHash()

func mustPrecomputeDummyBcryptHash() string {
	hash, err := auth.HashPassword("dummy-password-for-login-timing-defense")
	if err != nil {
		// auth.HashPassword only fails for inputs over bcrypt's 72-byte
		// limit; the literal above is well under that, so this is
		// unreachable in practice. Panicking on init is preferable to
		// silently disabling the timing defense.
		panic("usecase: failed to precompute dummy bcrypt hash for login timing defense: " + err.Error())
	}
	return hash
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
		// Timing-oracle defense (gate #36): burn a bcrypt comparison here so
		// this path costs about the same as the wrong-password path below,
		// instead of returning near-instantly and leaking which usernames
		// exist via response-time measurement. The result is intentionally
		// discarded — dummyBcryptHash never matches any real password.
		auth.VerifyBcrypt(dummyBcryptHash, password)
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
		// Best-effort migrate-on-login: the caller already correctly
		// authenticated via the legacy scheme, so a failure hashing or
		// persisting the upgraded bcrypt hash must NOT fail this login (it
		// would simply retry on the next legacy login). It must also NOT be
		// silently discarded (gate #36 risk finding) — a persistently
		// failing migration (e.g. a schema drift or a permanently
		// unreachable replica) needs to be observable server-side instead
		// of vanishing forever.
		newHash, hashErr := auth.HashPassword(password)
		if hashErr != nil {
			log.Printf("usecase: migrate-on-login: failed to hash password for admin %d: %v", admin.ID, hashErr)
		} else if updErr := s.repo.UpdatePassword(ctx, admin.ID, admin.Password, newHash); updErr != nil {
			log.Printf("usecase: migrate-on-login: failed to persist upgraded password for admin %d: %v", admin.ID, updErr)
		}
	}

	token, err := s.jwt.Generate(int64(admin.ID))
	if err != nil {
		return "", fmt.Errorf("usecase: generate login token: %w", err)
	}
	return token, nil
}
