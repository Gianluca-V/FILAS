package usecase_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/auth"
	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

type fakeAdminRepoForAuth struct {
	byUsername          map[string]domain.Admin
	err                 error
	updatePasswordID    int
	updatePasswordCalls int
	updatedOldHash      string
	updatedHash         string
	updatePasswordErr   error
}

func (f *fakeAdminRepoForAuth) List(ctx context.Context) ([]domain.Admin, error) { return nil, nil }
func (f *fakeAdminRepoForAuth) Get(ctx context.Context, id int) (domain.Admin, error) {
	return domain.Admin{}, nil
}
func (f *fakeAdminRepoForAuth) GetByUsername(ctx context.Context, username string) (domain.Admin, error) {
	if f.err != nil {
		return domain.Admin{}, f.err
	}
	a, ok := f.byUsername[username]
	if !ok {
		return domain.Admin{}, domain.ErrNotFound
	}
	return a, nil
}
func (f *fakeAdminRepoForAuth) Create(ctx context.Context, a domain.Admin) (domain.Admin, error) {
	return domain.Admin{}, nil
}
func (f *fakeAdminRepoForAuth) Update(ctx context.Context, id int, username, passwordHash string) error {
	return nil
}
func (f *fakeAdminRepoForAuth) Delete(ctx context.Context, id int) error { return nil }
func (f *fakeAdminRepoForAuth) UpdatePassword(ctx context.Context, id int, oldHash, newHash string) error {
	f.updatePasswordCalls++
	f.updatePasswordID = id
	f.updatedOldHash = oldHash
	f.updatedHash = newHash
	if f.updatePasswordErr != nil {
		return f.updatePasswordErr
	}
	// Simulate the real conditional UPDATE (mysql.AdminRepository.UpdatePassword,
	// "... WHERE ID = ? AND password = ?"): only mutate the stored row if its
	// Password still matches oldHash, and mirror that mutation back into
	// byUsername so a second Login() call in the same test sees the
	// migrated bcrypt row and takes the bcrypt path.
	for username, a := range f.byUsername {
		if a.ID == id && a.Password == oldHash {
			a.Password = newHash
			f.byUsername[username] = a
		}
	}
	return nil
}

type fakeJWTIssuer struct {
	token string
	err   error
}

func (f fakeJWTIssuer) Generate(adminID int64) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.token, nil
}

func TestAuthService_Login_BcryptRow_AcceptsCorrectPassword(t *testing.T) {
	hash, err := auth.HashPassword("s3cret-pw")
	if err != nil {
		t.Fatalf("auth.HashPassword() error = %v", err)
	}
	repo := &fakeAdminRepoForAuth{byUsername: map[string]domain.Admin{
		"admin": {ID: 42, Username: "admin", Password: hash},
	}}
	svc := usecase.NewAuthService(repo, fakeJWTIssuer{token: "signed-token"})

	token, err := svc.Login(context.Background(), "admin", "s3cret-pw")
	if err != nil {
		t.Fatalf("Login() error = %v, want nil", err)
	}
	if token != "signed-token" {
		t.Errorf("Login() = %q, want %q", token, "signed-token")
	}
	if repo.updatedHash != "" {
		t.Errorf("UpdatePassword should NOT be called for an already-bcrypt row, got hash %q", repo.updatedHash)
	}
}

func TestAuthService_Login_BcryptRow_RejectsWrongPassword(t *testing.T) {
	hash, _ := auth.HashPassword("s3cret-pw")
	repo := &fakeAdminRepoForAuth{byUsername: map[string]domain.Admin{
		"admin": {ID: 42, Username: "admin", Password: hash},
	}}
	svc := usecase.NewAuthService(repo, fakeJWTIssuer{token: "signed-token"})

	_, err := svc.Login(context.Background(), "admin", "wrong-pw")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("Login() error = %v, want %v", err, domain.ErrUnauthorized)
	}
}

func TestAuthService_Login_LegacyRow_AcceptsCorrectPasswordAndMigratesToBcrypt(t *testing.T) {
	// Characterized live against FilasServer/admins.php (see
	// backend/docs/legacy-quirks.md §3): salt/hash/password triple for the
	// seeded synthetic admin.
	repo := &fakeAdminRepoForAuth{byUsername: map[string]domain.Admin{
		"FilasAdmin": {
			ID:       1543,
			Username: "FilasAdmin",
			Password: "9a0d462f4985ce28e354f318e2346a3d61881336330e55cca8d9697cad9f1686",
			Salt:     "16308d7827b79fa10077d9a137027f97",
		},
	}}
	svc := usecase.NewAuthService(repo, fakeJWTIssuer{token: "signed-token"})

	token, err := svc.Login(context.Background(), "FilasAdmin", "filas-local-dev-2026")
	if err != nil {
		t.Fatalf("Login() error = %v, want nil", err)
	}
	if token != "signed-token" {
		t.Errorf("Login() = %q, want %q", token, "signed-token")
	}

	if repo.updatePasswordID != 1543 {
		t.Errorf("UpdatePassword called with id = %d, want 1543 (migrate-on-login)", repo.updatePasswordID)
	}
	if !auth.IsBcryptHash(repo.updatedHash) {
		t.Errorf("UpdatePassword called with hash %q, want a bcrypt hash", repo.updatedHash)
	}
	if !auth.VerifyBcrypt(repo.updatedHash, "filas-local-dev-2026") {
		t.Errorf("the migrated bcrypt hash does not verify against the original password")
	}
	// TOCTOU fix (gate #36): the persistence call must be conditioned on the
	// exact legacy hash that was read and verified, so a concurrent
	// password rotation cannot be silently reverted by a last-write-wins
	// UPDATE (see mysql.AdminRepository.UpdatePassword's "AND password = ?").
	if repo.updatedOldHash != "9a0d462f4985ce28e354f318e2346a3d61881336330e55cca8d9697cad9f1686" {
		t.Errorf("UpdatePassword called with oldHash = %q, want the legacy hash that was actually verified", repo.updatedOldHash)
	}
}

// TestAuthService_Login_SequentialLogins_MigratesOnceThenIsIdempotent is the
// coverage regression fix (gate #36): two Login calls against the SAME repo
// state — the first (legacy hash) migrates and returns a token, the second
// uses the now-bcrypt row and does NOT call UpdatePassword again.
func TestAuthService_Login_SequentialLogins_MigratesOnceThenIsIdempotent(t *testing.T) {
	repo := &fakeAdminRepoForAuth{byUsername: map[string]domain.Admin{
		"FilasAdmin": {
			ID:       1543,
			Username: "FilasAdmin",
			Password: "9a0d462f4985ce28e354f318e2346a3d61881336330e55cca8d9697cad9f1686",
			Salt:     "16308d7827b79fa10077d9a137027f97",
		},
	}}
	svc := usecase.NewAuthService(repo, fakeJWTIssuer{token: "signed-token"})

	firstToken, err := svc.Login(context.Background(), "FilasAdmin", "filas-local-dev-2026")
	if err != nil {
		t.Fatalf("first Login() error = %v, want nil", err)
	}
	if firstToken != "signed-token" {
		t.Errorf("first Login() = %q, want %q", firstToken, "signed-token")
	}
	if repo.updatePasswordCalls != 1 {
		t.Fatalf("after first login, UpdatePassword calls = %d, want 1", repo.updatePasswordCalls)
	}
	migratedHash := repo.byUsername["FilasAdmin"].Password
	if !auth.IsBcryptHash(migratedHash) {
		t.Fatalf("after first login, stored password = %q, want a bcrypt hash", migratedHash)
	}

	secondToken, err := svc.Login(context.Background(), "FilasAdmin", "filas-local-dev-2026")
	if err != nil {
		t.Fatalf("second Login() error = %v, want nil", err)
	}
	if secondToken != "signed-token" {
		t.Errorf("second Login() = %q, want %q", secondToken, "signed-token")
	}
	if repo.updatePasswordCalls != 1 {
		t.Errorf("after second login, UpdatePassword calls = %d, want still 1 (no re-rehash on the bcrypt path)", repo.updatePasswordCalls)
	}
	if repo.byUsername["FilasAdmin"].Password != migratedHash {
		t.Errorf("second login changed the stored hash: got %q, want unchanged %q", repo.byUsername["FilasAdmin"].Password, migratedHash)
	}
}

// TestAuthService_Login_UnknownUsername_AndWrongPassword_BothReturnUnauthorized
// locks the timing-oracle fix (gate #36 risk finding): the not-found path
// now performs a dummy bcrypt comparison (see the unexported dummy-compare
// call inside Login) so it costs roughly the same as the wrong-password
// path below, instead of returning near-instantly. This test asserts the
// OBSERVABLE behavior (both paths reach ErrUnauthorized uniformly); the
// dummy-compare call itself is exercised by this test and is visible in
// coverage — timing is deliberately NOT asserted here (unit tests cannot
// reliably measure microsecond-scale bcrypt cost differences).
func TestAuthService_Login_UnknownUsername_AndWrongPassword_BothReturnUnauthorized(t *testing.T) {
	hash, _ := auth.HashPassword("s3cret-pw")
	repo := &fakeAdminRepoForAuth{byUsername: map[string]domain.Admin{
		"admin": {ID: 42, Username: "admin", Password: hash},
	}}
	svc := usecase.NewAuthService(repo, fakeJWTIssuer{token: "signed-token"})

	_, unknownErr := svc.Login(context.Background(), "nobody-such-user", "whatever")
	if !errors.Is(unknownErr, domain.ErrUnauthorized) {
		t.Errorf("Login(unknown username) error = %v, want %v", unknownErr, domain.ErrUnauthorized)
	}

	_, wrongPwErr := svc.Login(context.Background(), "admin", "wrong-pw")
	if !errors.Is(wrongPwErr, domain.ErrUnauthorized) {
		t.Errorf("Login(wrong password) error = %v, want %v", wrongPwErr, domain.ErrUnauthorized)
	}
}

func TestAuthService_Login_LegacyRow_RejectsWrongPasswordWithoutMigrating(t *testing.T) {
	repo := &fakeAdminRepoForAuth{byUsername: map[string]domain.Admin{
		"FilasAdmin": {
			ID:       1543,
			Username: "FilasAdmin",
			Password: "9a0d462f4985ce28e354f318e2346a3d61881336330e55cca8d9697cad9f1686",
			Salt:     "16308d7827b79fa10077d9a137027f97",
		},
	}}
	svc := usecase.NewAuthService(repo, fakeJWTIssuer{token: "signed-token"})

	_, err := svc.Login(context.Background(), "FilasAdmin", "wrong-password")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("Login() error = %v, want %v", err, domain.ErrUnauthorized)
	}
	if repo.updatePasswordID != 0 {
		t.Errorf("UpdatePassword should NOT be called on failed login, but was called with id = %d", repo.updatePasswordID)
	}
}

func TestAuthService_Login_UnknownUsername_ReturnsUnauthorizedUniformly(t *testing.T) {
	// No user-enumeration signal: unknown username and wrong password both
	// map to the same ErrUnauthorized (design §2.6 point 5).
	repo := &fakeAdminRepoForAuth{byUsername: map[string]domain.Admin{}}
	svc := usecase.NewAuthService(repo, fakeJWTIssuer{token: "signed-token"})

	_, err := svc.Login(context.Background(), "nobody", "whatever")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("Login() error = %v, want %v", err, domain.ErrUnauthorized)
	}
}

func TestAuthService_Login_MigratePersistFailureStillReturnsTokenAndLogsFailure(t *testing.T) {
	// Best-effort migration: a transient persist failure must not lock the
	// admin out of a login they otherwise correctly authenticated for. The
	// migration is naturally retried on the next successful legacy login.
	// Gate #36 risk finding: the failure must not be silently discarded —
	// it is logged so a PERSISTENTLY-failing migration is observable
	// server-side instead of vanishing.
	repo := &fakeAdminRepoForAuth{
		byUsername: map[string]domain.Admin{
			"FilasAdmin": {
				ID:       1543,
				Username: "FilasAdmin",
				Password: "9a0d462f4985ce28e354f318e2346a3d61881336330e55cca8d9697cad9f1686",
				Salt:     "16308d7827b79fa10077d9a137027f97",
			},
		},
		updatePasswordErr: errors.New("db unavailable"),
	}
	svc := usecase.NewAuthService(repo, fakeJWTIssuer{token: "signed-token"})

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	token, err := svc.Login(context.Background(), "FilasAdmin", "filas-local-dev-2026")
	if err != nil {
		t.Fatalf("Login() error = %v, want nil (persist failure must not fail the login)", err)
	}
	if token != "signed-token" {
		t.Errorf("Login() = %q, want %q", token, "signed-token")
	}

	logged := buf.String()
	if !strings.Contains(logged, "1543") {
		t.Errorf("log output = %q, want it to reference admin ID 1543", logged)
	}
	if !strings.Contains(logged, "db unavailable") {
		t.Errorf("log output = %q, want it to contain the underlying error", logged)
	}
}

// TestAuthService_Login_MigrateHashFailureStillReturnsTokenAndLogsFailure
// exercises the OTHER half of the best-effort migration: bcrypt.
// GenerateFromPassword itself failing (bcrypt rejects passwords over 72
// bytes) must not fail the login (the caller already correctly
// authenticated via the legacy scheme) but must be logged, not discarded.
func TestAuthService_Login_MigrateHashFailureStillReturnsTokenAndLogsFailure(t *testing.T) {
	longPassword := strings.Repeat("x", 73) // bcrypt's hard limit is 72 bytes
	salt := "16308d7827b79fa10077d9a137027f97"
	sum := sha256.Sum256([]byte(salt + longPassword))
	legacyHash := hex.EncodeToString(sum[:])

	repo := &fakeAdminRepoForAuth{byUsername: map[string]domain.Admin{
		"FilasAdmin": {ID: 1543, Username: "FilasAdmin", Password: legacyHash, Salt: salt},
	}}
	svc := usecase.NewAuthService(repo, fakeJWTIssuer{token: "signed-token"})

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	token, err := svc.Login(context.Background(), "FilasAdmin", longPassword)
	if err != nil {
		t.Fatalf("Login() error = %v, want nil (a migration hash failure must not fail an otherwise-valid login)", err)
	}
	if token != "signed-token" {
		t.Errorf("Login() = %q, want %q", token, "signed-token")
	}
	if repo.updatePasswordCalls != 0 {
		t.Errorf("UpdatePassword should not be called when hashing failed, but was called %d time(s)", repo.updatePasswordCalls)
	}

	logged := buf.String()
	if !strings.Contains(logged, "1543") {
		t.Errorf("log output = %q, want it to reference admin ID 1543", logged)
	}
}

func TestAuthService_Login_PropagatesJWTGenerationError(t *testing.T) {
	hash, _ := auth.HashPassword("s3cret-pw")
	repo := &fakeAdminRepoForAuth{byUsername: map[string]domain.Admin{
		"admin": {ID: 42, Username: "admin", Password: hash},
	}}
	genErr := errors.New("signing key unavailable")
	svc := usecase.NewAuthService(repo, fakeJWTIssuer{err: genErr})

	_, err := svc.Login(context.Background(), "admin", "s3cret-pw")
	if !errors.Is(err, genErr) {
		t.Errorf("Login() error = %v, want it to wrap %v", err, genErr)
	}
}
