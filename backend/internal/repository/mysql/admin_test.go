package mysql_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/repository/mysql"
)

const adminSelectCols = "SELECT ID, username, password, salt FROM admins"

func TestAdminRepository_List_MapsRowsToDomain(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "username", "password", "salt"}).
		AddRow(1543, "FilasAdmin", "9a0d462f...", "16308d78...")
	mock.ExpectQuery(regexp.QuoteMeta(adminSelectCols)).WillReturnRows(rows)

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	want := []domain.Admin{{ID: 1543, Username: "FilasAdmin", Password: "9a0d462f...", Salt: "16308d78..."}}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("List() = %+v, want %+v", got, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestAdminRepository_List_ReturnsEmptySliceWhenNoRows(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "username", "password", "salt"})
	mock.ExpectQuery(regexp.QuoteMeta(adminSelectCols)).WillReturnRows(rows)

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("List() = %+v, want empty slice", got)
	}
}

func TestAdminRepository_Get_ReturnsMatchingAdmin(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "username", "password", "salt"}).
		AddRow(1543, "FilasAdmin", "9a0d462f...", "16308d78...")
	mock.ExpectQuery(regexp.QuoteMeta(adminSelectCols + " WHERE ID = ?")).
		WithArgs(1543).
		WillReturnRows(rows)

	got, err := repo.Get(context.Background(), 1543)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	want := domain.Admin{ID: 1543, Username: "FilasAdmin", Password: "9a0d462f...", Salt: "16308d78..."}
	if got != want {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestAdminRepository_Get_ReturnsDomainNotFoundWhenNoRows(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(adminSelectCols + " WHERE ID = ?")).
		WithArgs(999999).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "username", "password", "salt"}))

	_, err := repo.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestAdminRepository_GetByUsername_ReturnsMatchingAdmin(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "username", "password", "salt"}).
		AddRow(1543, "FilasAdmin", "9a0d462f...", "16308d78...")
	mock.ExpectQuery(regexp.QuoteMeta(adminSelectCols + " WHERE username = ?")).
		WithArgs("FilasAdmin").
		WillReturnRows(rows)

	got, err := repo.GetByUsername(context.Background(), "FilasAdmin")
	if err != nil {
		t.Fatalf("GetByUsername() error = %v, want nil", err)
	}
	if got.ID != 1543 {
		t.Errorf("GetByUsername() = %+v, want ID 1543", got)
	}
}

func TestAdminRepository_GetByUsername_ReturnsDomainNotFoundWhenNoRows(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(adminSelectCols + " WHERE username = ?")).
		WithArgs("nobody").
		WillReturnRows(sqlmock.NewRows([]string{"ID", "username", "password", "salt"}))

	_, err := repo.GetByUsername(context.Background(), "nobody")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByUsername() error = %v, want %v", err, domain.ErrNotFound)
	}
}

// TestAdminRepository_List_PropagatesQueryError and its GetByUsername/Get
// siblings restore the PR2-sibling WillReturnError convention (gate #36
// reliability finding): every resource's repository tests cover the
// query-error path for List/Get, and admins additionally needs
// GetByUsername (used by the login flow) and UpdatePassword (used by
// migrate-on-login).
func TestAdminRepository_List_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(adminSelectCols)).WillReturnError(dbErr)

	_, err := repo.List(context.Background())
	if !errors.Is(err, dbErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestAdminRepository_Get_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(adminSelectCols + " WHERE ID = ?")).
		WithArgs(1543).
		WillReturnError(dbErr)

	_, err := repo.Get(context.Background(), 1543)
	if !errors.Is(err, dbErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestAdminRepository_GetByUsername_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(adminSelectCols + " WHERE username = ?")).
		WithArgs("FilasAdmin").
		WillReturnError(dbErr)

	_, err := repo.GetByUsername(context.Background(), "FilasAdmin")
	if !errors.Is(err, dbErr) {
		t.Errorf("GetByUsername() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestAdminRepository_Create_AssignsAutoIncrementID(t *testing.T) {
	// Task 2.1: the seed schema patches admins with PRIMARY KEY +
	// AUTO_INCREMENT — the INSERT does not supply an ID, and Create reads
	// LastInsertId() to populate the returned domain.Admin.
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO admins (username, password) VALUES (?, ?)")).
		WithArgs("newadmin", "bcrypt-hash").
		WillReturnResult(sqlmock.NewResult(9001, 1))

	got, err := repo.Create(context.Background(), domain.Admin{Username: "newadmin", Password: "bcrypt-hash"})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (from LastInsertId)", got.ID)
	}
	if got.Username != "newadmin" || got.Password != "bcrypt-hash" {
		t.Errorf("Create() = %+v, want Username/Password preserved", got)
	}
}

func TestAdminRepository_Create_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	dbErr := errors.New("duplicate entry")
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO admins (username, password) VALUES (?, ?)")).
		WithArgs("newadmin", "bcrypt-hash").
		WillReturnError(dbErr)

	_, err := repo.Create(context.Background(), domain.Admin{Username: "newadmin", Password: "bcrypt-hash"})
	if !errors.Is(err, dbErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestAdminRepository_Update_ExecutesWithUsernameAndPasswordHash(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE admins SET username = ?, password = ? WHERE ID = ?")).
		WithArgs("renamed", "new-bcrypt-hash", 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), 5, "renamed", "new-bcrypt-hash"); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
}

func TestAdminRepository_Update_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectExec(regexp.QuoteMeta("UPDATE admins SET username = ?, password = ? WHERE ID = ?")).
		WithArgs("renamed", "new-bcrypt-hash", 5).
		WillReturnError(dbErr)

	err := repo.Update(context.Background(), 5, "renamed", "new-bcrypt-hash")
	if !errors.Is(err, dbErr) {
		t.Errorf("Update() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestAdminRepository_Delete_ExecutesWithID(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM admins WHERE ID = ?")).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
}

func TestAdminRepository_Delete_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM admins WHERE ID = ?")).
		WithArgs(7).
		WillReturnError(dbErr)

	err := repo.Delete(context.Background(), 7)
	if !errors.Is(err, dbErr) {
		t.Errorf("Delete() error = %v, want it to wrap %v", err, dbErr)
	}
}

// TestAdminRepository_UpdatePassword_ExecutesConditionalOnOldHash is the
// TOCTOU fix (gate #36 risk finding): the migrate-on-login rehash UPDATE is
// now conditional on the password column still holding the exact legacy
// hash that was read and verified, via "AND password = ?". This prevents a
// concurrent password rotation from being silently reverted by a
// last-write-wins UPDATE — if the row changed between read and write, the
// WHERE clause simply matches 0 rows instead of overwriting the newer
// value.
func TestAdminRepository_UpdatePassword_ExecutesConditionalOnOldHash(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE admins SET password = ? WHERE ID = ? AND password = ?")).
		WithArgs("migrated-bcrypt-hash", 1543, "legacy-sha256-hex").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdatePassword(context.Background(), 1543, "legacy-sha256-hex", "migrated-bcrypt-hash"); err != nil {
		t.Fatalf("UpdatePassword() error = %v, want nil", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestAdminRepository_UpdatePassword_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewAdminRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectExec(regexp.QuoteMeta("UPDATE admins SET password = ? WHERE ID = ? AND password = ?")).
		WithArgs("migrated-bcrypt-hash", 1543, "legacy-sha256-hex").
		WillReturnError(dbErr)

	err := repo.UpdatePassword(context.Background(), 1543, "legacy-sha256-hex", "migrated-bcrypt-hash")
	if !errors.Is(err, dbErr) {
		t.Errorf("UpdatePassword() error = %v, want it to wrap %v", err, dbErr)
	}
}
