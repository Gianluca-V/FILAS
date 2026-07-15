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

const organizationSelectCols = "SELECT ID, Title, Description, Image FROM organizations"

func TestOrganizationRepository_List_MapsRowsToDomain(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrganizationRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Title", "Description", "Image"}).
		AddRow(1, "organizacions de prueba 1", "Descripcion de prueba 1", "https://example.com/1.jpg").
		AddRow(2, "sin descripcion", nil, "https://example.com/2.jpg")
	mock.ExpectQuery(regexp.QuoteMeta(organizationSelectCols)).WillReturnRows(rows)

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("List() returned %d items, want 2", len(got))
	}
	if got[0].Description == nil || *got[0].Description != "Descripcion de prueba 1" {
		t.Errorf("got[0].Description = %v, want pointer to description", got[0].Description)
	}
	if got[1].Description != nil {
		t.Errorf("got[1].Description = %v, want nil (NULL in DB)", *got[1].Description)
	}
}

func TestOrganizationRepository_List_ReturnsEmptySliceWhenNoRows(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrganizationRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(organizationSelectCols)).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Title", "Description", "Image"}))

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("List() = %+v, want empty slice", got)
	}
}

func TestOrganizationRepository_Get_ReturnsMatchingItem(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrganizationRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Title", "Description", "Image"}).
		AddRow(1, "organizacions de prueba 1", "Descripcion de prueba 1", "https://example.com/1.jpg")
	mock.ExpectQuery(regexp.QuoteMeta(organizationSelectCols + " WHERE ID = ?")).
		WithArgs(1).
		WillReturnRows(rows)

	got, err := repo.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got.ID != 1 || got.Title != "organizacions de prueba 1" {
		t.Errorf("Get() = %+v, want ID=1 Title=organizacions de prueba 1", got)
	}
}

func TestOrganizationRepository_Get_ReturnsDomainNotFoundWhenNoRows(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrganizationRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(organizationSelectCols + " WHERE ID = ?")).
		WithArgs(999999).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Title", "Description", "Image"}))

	_, err := repo.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestOrganizationRepository_List_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrganizationRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(organizationSelectCols)).WillReturnError(dbErr)

	_, err := repo.List(context.Background())
	if !errors.Is(err, dbErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestOrganizationRepository_Get_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrganizationRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(organizationSelectCols + " WHERE ID = ?")).
		WithArgs(1).
		WillReturnError(dbErr)

	_, err := repo.Get(context.Background(), 1)
	if !errors.Is(err, dbErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, dbErr)
	}
}

const organizationInsertSQL = "INSERT INTO organizations (Title, Description, Image) VALUES (?, ?, ?)"
const organizationUpdateSQL = "UPDATE organizations SET Title = ?, Description = ?, Image = ? WHERE ID = ?"
const organizationDeleteSQL = "DELETE FROM organizations WHERE ID = ?"

func TestOrganizationRepository_Create_AssignsAutoIncrementID(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrganizationRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(organizationInsertSQL)).
		WithArgs("Nueva organizacion", (*string)(nil), "https://example.com/n.jpg").
		WillReturnResult(sqlmock.NewResult(9001, 1))

	got, err := repo.Create(context.Background(), domain.Organization{Title: "Nueva organizacion", Image: "https://example.com/n.jpg"})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (from LastInsertId)", got.ID)
	}
}

func TestOrganizationRepository_Create_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrganizationRepository(db)

	dbErr := errors.New("duplicate entry")
	mock.ExpectExec(regexp.QuoteMeta(organizationInsertSQL)).
		WithArgs("Nueva organizacion", (*string)(nil), "https://example.com/n.jpg").
		WillReturnError(dbErr)

	_, err := repo.Create(context.Background(), domain.Organization{Title: "Nueva organizacion", Image: "https://example.com/n.jpg"})
	if !errors.Is(err, dbErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestOrganizationRepository_Update_ExecutesWithTitleDescriptionImage(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrganizationRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(organizationUpdateSQL)).
		WithArgs("renombrada", (*string)(nil), "https://example.com/r.jpg", 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), 5, domain.Organization{Title: "renombrada", Image: "https://example.com/r.jpg"}); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
}

func TestOrganizationRepository_Update_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrganizationRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectExec(regexp.QuoteMeta(organizationUpdateSQL)).
		WithArgs("renombrada", (*string)(nil), "https://example.com/r.jpg", 5).
		WillReturnError(dbErr)

	err := repo.Update(context.Background(), 5, domain.Organization{Title: "renombrada", Image: "https://example.com/r.jpg"})
	if !errors.Is(err, dbErr) {
		t.Errorf("Update() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestOrganizationRepository_Delete_ExecutesWithID(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrganizationRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(organizationDeleteSQL)).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
}

func TestOrganizationRepository_Delete_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrganizationRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectExec(regexp.QuoteMeta(organizationDeleteSQL)).
		WithArgs(7).
		WillReturnError(dbErr)

	err := repo.Delete(context.Background(), 7)
	if !errors.Is(err, dbErr) {
		t.Errorf("Delete() error = %v, want it to wrap %v", err, dbErr)
	}
}
