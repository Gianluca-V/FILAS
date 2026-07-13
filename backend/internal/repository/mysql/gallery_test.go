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

func newGalleryTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return sqlx.NewDb(db, "sqlmock"), mock
}

func TestGalleryRepository_List_MapsRowsToDomain(t *testing.T) {
	db, mock := newGalleryTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Image"}).
		AddRow(1, "assets/galeria-1.jpg").
		AddRow(2, "assets/galeria-2.jpg")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT ID, Image FROM gallery")).WillReturnRows(rows)

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	want := []domain.GalleryImage{{ID: 1, Image: "assets/galeria-1.jpg"}, {ID: 2, Image: "assets/galeria-2.jpg"}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("List() = %+v, want %+v", got, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestGalleryRepository_List_ReturnsEmptySliceWhenNoRows(t *testing.T) {
	db, mock := newGalleryTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Image"})
	mock.ExpectQuery(regexp.QuoteMeta("SELECT ID, Image FROM gallery")).WillReturnRows(rows)

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("List() = %+v, want empty slice", got)
	}
}

func TestGalleryRepository_Get_ReturnsMatchingImage(t *testing.T) {
	db, mock := newGalleryTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Image"}).AddRow(5, "assets/galeria-5.jpg")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT ID, Image FROM gallery WHERE ID = ?")).
		WithArgs(5).
		WillReturnRows(rows)

	got, err := repo.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	want := domain.GalleryImage{ID: 5, Image: "assets/galeria-5.jpg"}
	if got != want {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestGalleryRepository_Get_ReturnsDomainNotFoundWhenNoRows(t *testing.T) {
	db, mock := newGalleryTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT ID, Image FROM gallery WHERE ID = ?")).
		WithArgs(999999).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Image"}))

	_, err := repo.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}
