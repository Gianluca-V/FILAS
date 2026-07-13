package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// AdminRepository implements domain.AdminRepository on top of sqlx with
// parameterized queries.
type AdminRepository struct {
	db *sqlx.DB
}

// NewAdminRepository wires an AdminRepository to an open connection pool.
func NewAdminRepository(db *sqlx.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// adminRow uses sql.NullString for username/password/salt even though a
// real admin always has all three: the schema declares them `text DEFAULT
// NULL` (see backend/db/init/01-schema.sql), and scanning a NULL directly
// into a plain string errors — the exact bug class fixed for Product.Image
// in PR2 (gate #34). A NULL here degrades to "" rather than 500ing.
type adminRow struct {
	ID       int            `db:"ID"`
	Username sql.NullString `db:"username"`
	Password sql.NullString `db:"password"`
	Salt     sql.NullString `db:"salt"`
}

func (r adminRow) toDomain() domain.Admin {
	return domain.Admin{ID: r.ID, Username: r.Username.String, Password: r.Password.String, Salt: r.Salt.String}
}

const adminSelect = "SELECT ID, username, password, salt FROM admins"

func (r *AdminRepository) List(ctx context.Context) ([]domain.Admin, error) {
	var rows []adminRow
	if err := r.db.SelectContext(ctx, &rows, adminSelect); err != nil {
		return nil, fmt.Errorf("mysql: list admins: %w", err)
	}
	admins := make([]domain.Admin, 0, len(rows))
	for _, row := range rows {
		admins = append(admins, row.toDomain())
	}
	return admins, nil
}

func (r *AdminRepository) Get(ctx context.Context, id int) (domain.Admin, error) {
	var row adminRow
	err := r.db.GetContext(ctx, &row, adminSelect+" WHERE ID = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Admin{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Admin{}, fmt.Errorf("mysql: get admin %d: %w", id, err)
	}
	return row.toDomain(), nil
}

func (r *AdminRepository) GetByUsername(ctx context.Context, username string) (domain.Admin, error) {
	var row adminRow
	err := r.db.GetContext(ctx, &row, adminSelect+" WHERE username = ?", username)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Admin{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Admin{}, fmt.Errorf("mysql: get admin by username: %w", err)
	}
	return row.toDomain(), nil
}

// Create inserts a new admin WITHOUT supplying an ID — the seed schema's
// AUTO_INCREMENT PRIMARY KEY (task 2.1) assigns it, unlike legacy
// createUser which required the caller to supply one. The assigned ID is
// read back via LastInsertId() and populated on the returned domain.Admin.
func (r *AdminRepository) Create(ctx context.Context, a domain.Admin) (domain.Admin, error) {
	res, err := r.db.ExecContext(ctx, "INSERT INTO admins (username, password) VALUES (?, ?)", a.Username, a.Password)
	if err != nil {
		return domain.Admin{}, fmt.Errorf("mysql: create admin: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Admin{}, fmt.Errorf("mysql: read new admin id: %w", err)
	}
	a.ID = int(id)
	return a, nil
}

// Update overwrites username and password (already a bcrypt hash) for the
// given ID. Like legacy updateUser, it does not check row existence first
// — a nonexistent ID is a no-op UPDATE that still reports success. This
// quirk was live-characterized (see sdd/migrate-go-vue/apply-progress) and
// preserved deliberately rather than invented as new behavior.
func (r *AdminRepository) Update(ctx context.Context, id int, username, passwordHash string) error {
	if _, err := r.db.ExecContext(ctx, "UPDATE admins SET username = ?, password = ? WHERE ID = ?", username, passwordHash, id); err != nil {
		return fmt.Errorf("mysql: update admin %d: %w", id, err)
	}
	return nil
}

// Delete removes the admin with the given ID, same no-existence-check
// quirk as Update.
func (r *AdminRepository) Delete(ctx context.Context, id int) error {
	if _, err := r.db.ExecContext(ctx, "DELETE FROM admins WHERE ID = ?", id); err != nil {
		return fmt.Errorf("mysql: delete admin %d: %w", id, err)
	}
	return nil
}

// UpdatePassword persists only the password column, used by the
// migrate-on-login flow (usecase.AuthService.Login) to upgrade a legacy
// sha256 row to bcrypt without touching username.
func (r *AdminRepository) UpdatePassword(ctx context.Context, id int, passwordHash string) error {
	if _, err := r.db.ExecContext(ctx, "UPDATE admins SET password = ? WHERE ID = ?", passwordHash, id); err != nil {
		return fmt.Errorf("mysql: update admin %d password: %w", id, err)
	}
	return nil
}
