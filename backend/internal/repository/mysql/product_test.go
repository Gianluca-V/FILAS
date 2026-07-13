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

func newProductTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return sqlx.NewDb(db, "sqlmock"), mock
}

const productSelectCols = "SELECT ID, Name, Price, Stock, Image, Description FROM products"

func TestProductRepository_List_MapsNullAndNonNullDescription(t *testing.T) {
	// Triangulation: legacy seed data has BOTH empty-string and NULL
	// Description rows (see obs #31/#29 characterization). Both must be
	// preserved as distinct values through the domain layer.
	db, mock := newProductTestDB(t)
	repo := mysql.NewProductRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Name", "Price", "Stock", "Image", "Description"}).
		AddRow(1, "Mermelada de pera", 600.0, 112, "assets/default-img.png", "").
		AddRow(18, "Encurtido de ajo", 700.0, 1, "assets/ajo.jpg", nil)
	mock.ExpectQuery(regexp.QuoteMeta(productSelectCols)).WillReturnRows(rows)

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("List() returned %d products, want 2", len(got))
	}
	if got[0].Description == nil || *got[0].Description != "" {
		t.Errorf("got[0].Description = %v, want pointer to empty string", got[0].Description)
	}
	if got[1].Description != nil {
		t.Errorf("got[1].Description = %v, want nil (NULL in DB)", *got[1].Description)
	}
	if got[0].Price != 600.0 || got[0].Stock != 112 {
		t.Errorf("got[0] = %+v, want Price=600 Stock=112", got[0])
	}
}

func TestProductRepository_List_ReturnsEmptySliceWhenNoRows(t *testing.T) {
	db, mock := newProductTestDB(t)
	repo := mysql.NewProductRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(productSelectCols)).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Name", "Price", "Stock", "Image", "Description"}))

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("List() = %+v, want empty slice", got)
	}
}

func TestProductRepository_Get_ReturnsMatchingProduct(t *testing.T) {
	db, mock := newProductTestDB(t)
	repo := mysql.NewProductRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Name", "Price", "Stock", "Image", "Description"}).
		AddRow(1, "Mermelada de pera", 600.0, 112, "assets/default-img.png", "")
	mock.ExpectQuery(regexp.QuoteMeta(productSelectCols + " WHERE ID = ?")).
		WithArgs(1).
		WillReturnRows(rows)

	got, err := repo.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got.ID != 1 || got.Name != "Mermelada de pera" || got.Price != 600.0 {
		t.Errorf("Get() = %+v, want ID=1 Name=Mermelada de pera Price=600", got)
	}
}

func TestProductRepository_Get_ReturnsDomainNotFoundWhenNoRows(t *testing.T) {
	db, mock := newProductTestDB(t)
	repo := mysql.NewProductRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(productSelectCols + " WHERE ID = ?")).
		WithArgs(999999).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Name", "Price", "Stock", "Image", "Description"}))

	_, err := repo.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}
