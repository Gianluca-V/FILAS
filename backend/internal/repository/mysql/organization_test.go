package mysql_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/repository/mysql"
)

func newOrganizationTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return sqlx.NewDb(db, "sqlmock"), mock
}

const organizationSelectCols = "SELECT ID, Title, Description, Image FROM organizations"

func TestOrganizationRepository_List_MapsRowsToDomain(t *testing.T) {
	db, mock := newOrganizationTestDB(t)
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
	db, mock := newOrganizationTestDB(t)
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
	db, mock := newOrganizationTestDB(t)
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
	db, mock := newOrganizationTestDB(t)
	repo := mysql.NewOrganizationRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(organizationSelectCols + " WHERE ID = ?")).
		WithArgs(999999).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Title", "Description", "Image"}))

	_, err := repo.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}
