package domain

import "context"

// Admin is an administrator account. Password holds whatever the DB
// currently has: a bcrypt hash ($2a$/$2b$/$2y$ prefix) once migrated, or a
// legacy sha256(salt+password) hex digest before its first successful
// login post-migration (see usecase.AuthService.Login and
// internal/auth/password.go). Salt is only meaningful for the legacy
// scheme; a bcrypt hash carries its own salt internally.
type Admin struct {
	ID       int
	Username string
	Password string
	Salt     string
}

// AdminRepository is implemented by internal/repository/mysql.
type AdminRepository interface {
	List(ctx context.Context) ([]Admin, error)
	Get(ctx context.Context, id int) (Admin, error)
	GetByUsername(ctx context.Context, username string) (Admin, error)
	// Create inserts a new admin and returns it with ID populated from the
	// database's AUTO_INCREMENT (see PR3 task 2.1: the seed schema patches
	// admins with PRIMARY KEY + AUTO_INCREMENT — the DB assigns IDs, the
	// caller never supplies one).
	Create(ctx context.Context, a Admin) (Admin, error)
	// Update overwrites username and password (already a bcrypt hash) for
	// the given ID. Mirrors legacy updateUser: it does not check whether a
	// row with that ID exists (always reports success — see
	// backend/docs/legacy-quirks.md for the live-characterized quirk).
	Update(ctx context.Context, id int, username, passwordHash string) error
	// Delete removes the admin with the given ID. Same no-existence-check
	// quirk as Update, preserved for contract fidelity.
	Delete(ctx context.Context, id int) error
	// UpdatePassword persists only the password column, used by the
	// migrate-on-login flow to transparently upgrade a legacy sha256 row to
	// bcrypt without touching username. The write is CONDITIONAL on oldHash
	// (the exact value that was read and verified before calling this
	// method): if the row's password has changed since then — e.g. a
	// concurrent password rotation — the update matches 0 rows instead of
	// clobbering the newer value with a stale last-write-wins overwrite
	// (TOCTOU fix, gate #36 risk finding).
	UpdatePassword(ctx context.Context, id int, oldHash, newHash string) error
}
