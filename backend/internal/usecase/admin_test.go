package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/auth"
	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

type fakeAdminRepo struct {
	admins       []domain.Admin
	byID         map[int]domain.Admin
	err          error
	createdAdmin domain.Admin
	createErr    error
	updateCalled bool
	updateID     int
	updateUser   string
	updateHash   string
	updateErr    error
	deleteID     int
	deleteErr    error
}

func (f *fakeAdminRepo) List(ctx context.Context) ([]domain.Admin, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.admins, nil
}

func (f *fakeAdminRepo) Get(ctx context.Context, id int) (domain.Admin, error) {
	if f.err != nil {
		return domain.Admin{}, f.err
	}
	a, ok := f.byID[id]
	if !ok {
		return domain.Admin{}, domain.ErrNotFound
	}
	return a, nil
}

func (f *fakeAdminRepo) GetByUsername(ctx context.Context, username string) (domain.Admin, error) {
	return domain.Admin{}, nil
}

func (f *fakeAdminRepo) Create(ctx context.Context, a domain.Admin) (domain.Admin, error) {
	if f.createErr != nil {
		return domain.Admin{}, f.createErr
	}
	f.createdAdmin = a
	a.ID = 9001
	return a, nil
}

func (f *fakeAdminRepo) Update(ctx context.Context, id int, username, passwordHash string) error {
	f.updateCalled = true
	f.updateID = id
	f.updateUser = username
	f.updateHash = passwordHash
	return f.updateErr
}

func (f *fakeAdminRepo) Delete(ctx context.Context, id int) error {
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeAdminRepo) UpdatePassword(ctx context.Context, id int, oldHash, newHash string) error {
	return nil
}

func TestAdminService_List_ReturnsRepositoryAdmins(t *testing.T) {
	want := []domain.Admin{{ID: 1, Username: "a"}, {ID: 2, Username: "b"}}
	svc := usecase.NewAdminService(&fakeAdminRepo{admins: want})

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 2 || got[0].Username != "a" || got[1].Username != "b" {
		t.Errorf("List() = %+v, want %+v", got, want)
	}
}

func TestAdminService_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	svc := usecase.NewAdminService(&fakeAdminRepo{byID: map[int]domain.Admin{}})

	_, err := svc.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestAdminService_Create_HashesPasswordWithBcryptBeforePersisting(t *testing.T) {
	repo := &fakeAdminRepo{}
	svc := usecase.NewAdminService(repo)

	got, err := svc.Create(context.Background(), "newadmin", "raw-password")
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (repo-assigned)", got.ID)
	}
	if repo.createdAdmin.Password == "raw-password" {
		t.Errorf("Create() persisted the raw password verbatim, want a bcrypt hash")
	}
	if !auth.IsBcryptHash(repo.createdAdmin.Password) {
		t.Errorf("Create() persisted password %q, want a bcrypt hash", repo.createdAdmin.Password)
	}
	if !auth.VerifyBcrypt(repo.createdAdmin.Password, "raw-password") {
		t.Errorf("the persisted hash does not verify against the raw password")
	}
}

func TestAdminService_Create_RejectsMissingUsernameOrPassword(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{"missing username", "", "raw-password"},
		{"missing password", "newadmin", ""},
		{"missing both", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeAdminRepo{}
			svc := usecase.NewAdminService(repo)

			_, err := svc.Create(context.Background(), tt.username, tt.password)
			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Create() error = %v, want %v", err, domain.ErrValidation)
			}
			if repo.createdAdmin.Username != "" {
				t.Errorf("Create() should not have called the repository, but persisted %+v", repo.createdAdmin)
			}
		})
	}
}

func TestAdminService_Update_HashesPasswordWithBcryptBeforePersisting(t *testing.T) {
	repo := &fakeAdminRepo{}
	svc := usecase.NewAdminService(repo)

	if err := svc.Update(context.Background(), 5, "renamed", "new-raw-password"); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if repo.updateID != 5 {
		t.Errorf("Update() called repo with id = %d, want 5", repo.updateID)
	}
	if repo.updateUser != "renamed" {
		t.Errorf("Update() called repo with username = %q, want %q", repo.updateUser, "renamed")
	}
	if repo.updateHash == "new-raw-password" {
		t.Errorf("Update() persisted the raw password verbatim (the legacy floatval bug), want a bcrypt hash")
	}
	if !auth.VerifyBcrypt(repo.updateHash, "new-raw-password") {
		t.Errorf("the persisted hash does not verify against the raw password")
	}
}

// TestAdminService_Update_RejectsMissingUsernameOrPassword is the blocker
// fix from the PR3 gate review (obs #36): PUT /api/admins/:id had NO
// validation, so an omitted/empty password field produced
// auth.HashPassword("") -> a VALID bcrypt hash of the empty string,
// persisted unconditionally. Any admin holding a JWT could blank another
// admin's password and log in as them with password "". Deliberate design
// decision (documented on AdminService.Update): legacy's updateUser has NO
// isset() guard at all — it always overwrites BOTH columns unconditionally
// (mangling password via floatval when non-numeric) — so legacy never
// supported a real "username-only" partial update. Update therefore
// REQUIRES both username and password, mirroring Create's guard exactly,
// rather than making password optional.
func TestAdminService_Update_RejectsMissingUsernameOrPassword(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{"missing password", "renamed", ""},
		{"missing username", "", "new-raw-password"},
		{"missing both", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeAdminRepo{}
			svc := usecase.NewAdminService(repo)

			err := svc.Update(context.Background(), 5, tt.username, tt.password)
			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Update() error = %v, want %v", err, domain.ErrValidation)
			}
			if repo.updateCalled {
				t.Errorf("Update() called the repository with (%q, %q) — no persistence should happen on validation failure", tt.username, tt.password)
			}
		})
	}
}

// TestAdminService_Update_NeverPersistsABcryptHashOfEmptyString locks the
// exact attack this blocker closes: even if a caller sneaks an empty
// password through some other path, the persisted hash must never verify
// against "" (which is what auth.HashPassword("") would otherwise produce).
func TestAdminService_Update_NeverPersistsABcryptHashOfEmptyString(t *testing.T) {
	repo := &fakeAdminRepo{}
	svc := usecase.NewAdminService(repo)

	_ = svc.Update(context.Background(), 5, "renamed", "")

	if repo.updateCalled {
		t.Fatalf("Update() persisted a hash for an empty password: hash=%q", repo.updateHash)
	}
}

func TestAdminService_Delete_DelegatesToRepository(t *testing.T) {
	repo := &fakeAdminRepo{}
	svc := usecase.NewAdminService(repo)

	if err := svc.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
	if repo.deleteID != 7 {
		t.Errorf("Delete() called repo with id = %d, want 7", repo.deleteID)
	}
}

func TestAdminService_Delete_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeAdminRepo{deleteErr: repoErr}
	svc := usecase.NewAdminService(repo)

	err := svc.Delete(context.Background(), 7)
	if !errors.Is(err, repoErr) {
		t.Errorf("Delete() error = %v, want it to wrap %v", err, repoErr)
	}
}
