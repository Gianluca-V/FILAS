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

func newNewsTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return sqlx.NewDb(db, "sqlmock"), mock
}

const newsSelectCols = "SELECT ID, Title, Body, Image FROM news"

func TestNewsRepository_List_MapsRowsToDomain(t *testing.T) {
	db, mock := newNewsTestDB(t)
	repo := mysql.NewNewsRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Title", "Body", "Image"}).
		AddRow(1, "Noticia de prueba 1", "Lorem ipsum", "https://example.com/1.jpg").
		AddRow(2, "Noticia sin cuerpo", nil, nil)
	mock.ExpectQuery(regexp.QuoteMeta(newsSelectCols)).WillReturnRows(rows)

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("List() returned %d items, want 2", len(got))
	}
	if got[0].Body == nil || *got[0].Body != "Lorem ipsum" {
		t.Errorf("got[0].Body = %v, want pointer to %q", got[0].Body, "Lorem ipsum")
	}
	if got[1].Body != nil || got[1].Image != nil {
		t.Errorf("got[1] = %+v, want nil Body/Image (NULL in DB)", got[1])
	}
}

func TestNewsRepository_List_ReturnsEmptySliceWhenNoRows(t *testing.T) {
	db, mock := newNewsTestDB(t)
	repo := mysql.NewNewsRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(newsSelectCols)).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Title", "Body", "Image"}))

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("List() = %+v, want empty slice", got)
	}
}

func TestNewsRepository_Get_ReturnsMatchingItem(t *testing.T) {
	db, mock := newNewsTestDB(t)
	repo := mysql.NewNewsRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Title", "Body", "Image"}).
		AddRow(1, "Noticia de prueba 1", "Lorem ipsum", "https://example.com/1.jpg")
	mock.ExpectQuery(regexp.QuoteMeta(newsSelectCols + " WHERE ID = ?")).
		WithArgs(1).
		WillReturnRows(rows)

	got, err := repo.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got.ID != 1 || got.Title != "Noticia de prueba 1" {
		t.Errorf("Get() = %+v, want ID=1 Title=Noticia de prueba 1", got)
	}
}

func TestNewsRepository_Get_ReturnsDomainNotFoundWhenNoRows(t *testing.T) {
	db, mock := newNewsTestDB(t)
	repo := mysql.NewNewsRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(newsSelectCols + " WHERE ID = ?")).
		WithArgs(999999).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Title", "Body", "Image"}))

	_, err := repo.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}
