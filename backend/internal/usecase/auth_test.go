package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/auth"
	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

type fakeAdminRepoForAuth struct {
	byUsername        map[string]domain.Admin
	err               error
	updatePasswordID  int
	updatedHash       string
	updatePasswordErr error
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
func (f *fakeAdminRepoForAuth) UpdatePassword(ctx context.Context, id int, passwordHash string) error {
	f.updatePasswordID = id
	f.updatedHash = passwordHash
	return f.updatePasswordErr
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
	// sdd/migrate-go-vue/apply-progress): salt/hash/password triple for the
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

func TestAuthService_Login_MigratePersistFailureStillReturnsToken(t *testing.T) {
	// Best-effort migration: a transient persist failure must not lock the
	// admin out of a login they otherwise correctly authenticated for. The
	// migration is naturally retried on the next successful legacy login.
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

	token, err := svc.Login(context.Background(), "FilasAdmin", "filas-local-dev-2026")
	if err != nil {
		t.Fatalf("Login() error = %v, want nil (persist failure must not fail the login)", err)
	}
	if token != "signed-token" {
		t.Errorf("Login() = %q, want %q", token, "signed-token")
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
