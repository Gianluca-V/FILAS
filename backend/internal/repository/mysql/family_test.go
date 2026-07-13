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

const familySelectCols = "SELECT ID, Image, Body, Category FROM family"

func TestFamilyRepository_List_MapsRowsToDomain(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewFamilyRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Image", "Body", "Category"}).
		AddRow(3, "https://example.com/3.jpg", "aqui va la descripcion 3", "Taller protegido").
		AddRow(6, nil, "Taller de musica", "Centro de dia")
	mock.ExpectQuery(regexp.QuoteMeta(familySelectCols)).WillReturnRows(rows)

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("List() returned %d items, want 2", len(got))
	}
	if got[0].Image == nil || *got[0].Image != "https://example.com/3.jpg" {
		t.Errorf("got[0].Image = %v, want pointer to URL", got[0].Image)
	}
	if got[1].Image != nil {
		t.Errorf("got[1].Image = %v, want nil (NULL in DB)", *got[1].Image)
	}
	if got[1].Category != "Centro de dia" {
		t.Errorf("got[1].Category = %q, want %q", got[1].Category, "Centro de dia")
	}
}

func TestFamilyRepository_List_ReturnsEmptySliceWhenNoRows(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewFamilyRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(familySelectCols)).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Image", "Body", "Category"}))

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("List() = %+v, want empty slice", got)
	}
}

func TestFamilyRepository_Get_ReturnsMatchingItem(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewFamilyRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Image", "Body", "Category"}).
		AddRow(3, "https://example.com/3.jpg", "aqui va la descripcion 3", "Taller protegido")
	mock.ExpectQuery(regexp.QuoteMeta(familySelectCols + " WHERE ID = ?")).
		WithArgs(3).
		WillReturnRows(rows)

	got, err := repo.Get(context.Background(), 3)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got.ID != 3 || got.Category != "Taller protegido" {
		t.Errorf("Get() = %+v, want ID=3 Category=Taller protegido", got)
	}
}

func TestFamilyRepository_Get_ReturnsDomainNotFoundWhenNoRows(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewFamilyRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(familySelectCols + " WHERE ID = ?")).
		WithArgs(999999).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Image", "Body", "Category"}))

	_, err := repo.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestFamilyRepository_List_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewFamilyRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(familySelectCols)).WillReturnError(dbErr)

	_, err := repo.List(context.Background())
	if !errors.Is(err, dbErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestFamilyRepository_Get_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewFamilyRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(familySelectCols + " WHERE ID = ?")).
		WithArgs(3).
		WillReturnError(dbErr)

	_, err := repo.Get(context.Background(), 3)
	if !errors.Is(err, dbErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, dbErr)
	}
}

const familyInsertSQL = "INSERT INTO family (Body, Category, Image) VALUES (?, ?, ?)"
const familyUpdateSQL = "UPDATE family SET Body = ?, Category = ?, Image = ? WHERE ID = ?"
const familyDeleteSQL = "DELETE FROM family WHERE ID = ?"

func TestFamilyRepository_Create_AssignsAutoIncrementID(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewFamilyRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(familyInsertSQL)).
		WithArgs("cuerpo", "Centro de dia", (*string)(nil)).
		WillReturnResult(sqlmock.NewResult(9001, 1))

	got, err := repo.Create(context.Background(), domain.FamilyItem{Body: "cuerpo", Category: "Centro de dia"})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (from LastInsertId)", got.ID)
	}
}

func TestFamilyRepository_Create_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewFamilyRepository(db)

	dbErr := errors.New("duplicate entry")
	mock.ExpectExec(regexp.QuoteMeta(familyInsertSQL)).
		WithArgs("cuerpo", "Centro de dia", (*string)(nil)).
		WillReturnError(dbErr)

	_, err := repo.Create(context.Background(), domain.FamilyItem{Body: "cuerpo", Category: "Centro de dia"})
	if !errors.Is(err, dbErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestFamilyRepository_Update_ExecutesWithBodyCategoryImage(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewFamilyRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(familyUpdateSQL)).
		WithArgs("renombrado", "Taller protegido", (*string)(nil), 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), 5, domain.FamilyItem{Body: "renombrado", Category: "Taller protegido"}); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
}

func TestFamilyRepository_Update_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewFamilyRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectExec(regexp.QuoteMeta(familyUpdateSQL)).
		WithArgs("renombrado", "Taller protegido", (*string)(nil), 5).
		WillReturnError(dbErr)

	err := repo.Update(context.Background(), 5, domain.FamilyItem{Body: "renombrado", Category: "Taller protegido"})
	if !errors.Is(err, dbErr) {
		t.Errorf("Update() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestFamilyRepository_Delete_ExecutesWithID(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewFamilyRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(familyDeleteSQL)).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
}

func TestFamilyRepository_Delete_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewFamilyRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectExec(regexp.QuoteMeta(familyDeleteSQL)).
		WithArgs(7).
		WillReturnError(dbErr)

	err := repo.Delete(context.Background(), 7)
	if !errors.Is(err, dbErr) {
		t.Errorf("Delete() error = %v, want it to wrap %v", err, dbErr)
	}
}
