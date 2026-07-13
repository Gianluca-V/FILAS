package usecase

import (
	"context"
	"fmt"

	"github.com/gianluca-v/filas-backend/internal/auth"
	"github.com/gianluca-v/filas-backend/internal/domain"
)

// AdminService implements admin CRUD (login lives separately in
// AuthService). Consumed by internal/handler/rest.
type AdminService struct {
	repo domain.AdminRepository
}

// NewAdminService wires an AdminService to its repository.
func NewAdminService(repo domain.AdminRepository) *AdminService {
	return &AdminService{repo: repo}
}

func (s *AdminService) List(ctx context.Context) ([]domain.Admin, error) {
	return s.repo.List(ctx)
}

func (s *AdminService) Get(ctx context.Context, id int) (domain.Admin, error) {
	return s.repo.Get(ctx, id)
}

// Create validates username/password are present, bcrypt-hashes the
// password, and persists a new admin. Unlike legacy createUser, the caller
// does not supply an ID — the seed schema's AUTO_INCREMENT PRIMARY KEY
// (task 2.1) assigns it (see internal/repository/mysql.AdminRepository.Create).
func (s *AdminService) Create(ctx context.Context, username, rawPassword string) (domain.Admin, error) {
	if username == "" || rawPassword == "" {
		return domain.Admin{}, fmt.Errorf("username and password are required: %w", domain.ErrValidation)
	}
	hash, err := auth.HashPassword(rawPassword)
	if err != nil {
		return domain.Admin{}, fmt.Errorf("usecase: hash password for new admin: %w", err)
	}
	created, err := s.repo.Create(ctx, domain.Admin{Username: username, Password: hash})
	if err != nil {
		return domain.Admin{}, fmt.Errorf("usecase: create admin: %w", err)
	}
	return created, nil
}

// Update bcrypt-hashes the new password before persisting (fixes the
// legacy floatval($password) bug that stored a mangled number instead of a
// usable credential).
func (s *AdminService) Update(ctx context.Context, id int, username, rawPassword string) error {
	hash, err := auth.HashPassword(rawPassword)
	if err != nil {
		return fmt.Errorf("usecase: hash password for admin update: %w", err)
	}
	if err := s.repo.Update(ctx, id, username, hash); err != nil {
		return fmt.Errorf("usecase: update admin %d: %w", id, err)
	}
	return nil
}

func (s *AdminService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("usecase: delete admin %d: %w", id, err)
	}
	return nil
}
