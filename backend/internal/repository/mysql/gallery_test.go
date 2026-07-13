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

const gallerySelectCols = "SELECT ID, Image FROM gallery"

func TestGalleryRepository_List_MapsRowsToDomain(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Image"}).
		AddRow(1, "assets/galeria-1.jpg").
		AddRow(2, "assets/galeria-2.jpg")
	mock.ExpectQuery(regexp.QuoteMeta(gallerySelectCols)).WillReturnRows(rows)

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
	db, mock := newTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Image"})
	mock.ExpectQuery(regexp.QuoteMeta(gallerySelectCols)).WillReturnRows(rows)

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("List() = %+v, want empty slice", got)
	}
}

func TestGalleryRepository_Get_ReturnsMatchingImage(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Image"}).AddRow(5, "assets/galeria-5.jpg")
	mock.ExpectQuery(regexp.QuoteMeta(gallerySelectCols + " WHERE ID = ?")).
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
	db, mock := newTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(gallerySelectCols + " WHERE ID = ?")).
		WithArgs(999999).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Image"}))

	_, err := repo.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestGalleryRepository_List_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(gallerySelectCols)).WillReturnError(dbErr)

	_, err := repo.List(context.Background())
	if !errors.Is(err, dbErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestGalleryRepository_Get_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(gallerySelectCols + " WHERE ID = ?")).
		WithArgs(5).
		WillReturnError(dbErr)

	_, err := repo.Get(context.Background(), 5)
	if !errors.Is(err, dbErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, dbErr)
	}
}

const galleryInsertSQL = "INSERT INTO gallery (Image) VALUES (?)"
const galleryUpdateSQL = "UPDATE gallery SET Image = ? WHERE ID = ?"
const galleryDeleteSQL = "DELETE FROM gallery WHERE ID = ?"

func TestGalleryRepository_Create_AssignsAutoIncrementID(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(galleryInsertSQL)).
		WithArgs("assets/nueva.jpg").
		WillReturnResult(sqlmock.NewResult(9001, 1))

	got, err := repo.Create(context.Background(), domain.GalleryImage{Image: "assets/nueva.jpg"})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (from LastInsertId)", got.ID)
	}
}

func TestGalleryRepository_Create_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	dbErr := errors.New("duplicate entry")
	mock.ExpectExec(regexp.QuoteMeta(galleryInsertSQL)).
		WithArgs("assets/nueva.jpg").
		WillReturnError(dbErr)

	_, err := repo.Create(context.Background(), domain.GalleryImage{Image: "assets/nueva.jpg"})
	if !errors.Is(err, dbErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestGalleryRepository_Update_ExecutesWithImage(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(galleryUpdateSQL)).
		WithArgs("assets/actualizada.jpg", 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), 5, "assets/actualizada.jpg"); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
}

func TestGalleryRepository_Update_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectExec(regexp.QuoteMeta(galleryUpdateSQL)).
		WithArgs("assets/actualizada.jpg", 5).
		WillReturnError(dbErr)

	err := repo.Update(context.Background(), 5, "assets/actualizada.jpg")
	if !errors.Is(err, dbErr) {
		t.Errorf("Update() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestGalleryRepository_Delete_ExecutesWithID(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(galleryDeleteSQL)).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
}

func TestGalleryRepository_Delete_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewGalleryRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectExec(regexp.QuoteMeta(galleryDeleteSQL)).
		WithArgs(7).
		WillReturnError(dbErr)

	err := repo.Delete(context.Background(), 7)
	if !errors.Is(err, dbErr) {
		t.Errorf("Delete() error = %v, want it to wrap %v", err, dbErr)
	}
}
