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

const newsSelectCols = "SELECT ID, Title, Body, Image FROM news"

func TestNewsRepository_List_MapsRowsToDomain(t *testing.T) {
	db, mock := newTestDB(t)
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
	db, mock := newTestDB(t)
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
	db, mock := newTestDB(t)
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
	db, mock := newTestDB(t)
	repo := mysql.NewNewsRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(newsSelectCols + " WHERE ID = ?")).
		WithArgs(999999).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Title", "Body", "Image"}))

	_, err := repo.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestNewsRepository_List_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewNewsRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(newsSelectCols)).WillReturnError(dbErr)

	_, err := repo.List(context.Background())
	if !errors.Is(err, dbErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestNewsRepository_Get_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewNewsRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(newsSelectCols + " WHERE ID = ?")).
		WithArgs(1).
		WillReturnError(dbErr)

	_, err := repo.Get(context.Background(), 1)
	if !errors.Is(err, dbErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, dbErr)
	}
}

const newsInsertSQL = "INSERT INTO news (Title, Body, Image) VALUES (?, ?, ?)"
const newsUpdateSQL = "UPDATE news SET Title = ?, Body = ?, Image = ? WHERE ID = ?"
const newsDeleteSQL = "DELETE FROM news WHERE ID = ?"

func TestNewsRepository_Create_AssignsAutoIncrementID(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewNewsRepository(db)

	body := "Lorem ipsum"
	image := "https://example.com/1.jpg"
	mock.ExpectExec(regexp.QuoteMeta(newsInsertSQL)).
		WithArgs("Noticia nueva", &body, &image).
		WillReturnResult(sqlmock.NewResult(9001, 1))

	got, err := repo.Create(context.Background(), domain.NewsItem{Title: "Noticia nueva", Body: &body, Image: &image})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (from LastInsertId)", got.ID)
	}
}

func TestNewsRepository_Create_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewNewsRepository(db)

	body := "Lorem ipsum"
	image := "https://example.com/1.jpg"
	dbErr := errors.New("duplicate entry")
	mock.ExpectExec(regexp.QuoteMeta(newsInsertSQL)).
		WithArgs("Noticia nueva", &body, &image).
		WillReturnError(dbErr)

	_, err := repo.Create(context.Background(), domain.NewsItem{Title: "Noticia nueva", Body: &body, Image: &image})
	if !errors.Is(err, dbErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestNewsRepository_Update_ExecutesWithFields(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewNewsRepository(db)

	body := "Cuerpo actualizado"
	image := "https://example.com/2.jpg"
	mock.ExpectExec(regexp.QuoteMeta(newsUpdateSQL)).
		WithArgs("Renombrada", &body, &image, 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), 5, domain.NewsItem{Title: "Renombrada", Body: &body, Image: &image}); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
}

func TestNewsRepository_Update_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewNewsRepository(db)

	body := "Cuerpo"
	image := "https://example.com/2.jpg"
	dbErr := errors.New("connection refused")
	mock.ExpectExec(regexp.QuoteMeta(newsUpdateSQL)).
		WithArgs("Renombrada", &body, &image, 5).
		WillReturnError(dbErr)

	err := repo.Update(context.Background(), 5, domain.NewsItem{Title: "Renombrada", Body: &body, Image: &image})
	if !errors.Is(err, dbErr) {
		t.Errorf("Update() error = %v, want it to wrap %v", err, dbErr)
	}
}

// TestNewsRepository_Update_SucceedsWithZeroRowsAffected locks the same
// legacy mysqli parity as products/family (PR4 corrective, see
// backend/docs/legacy-quirks.md §10): a zero-rows-affected UPDATE (missing
// or non-numeric ID) is not an error.
func TestNewsRepository_Update_SucceedsWithZeroRowsAffected(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewNewsRepository(db)

	body := "Fantasma"
	image := "https://example.com/x.jpg"
	mock.ExpectExec(regexp.QuoteMeta(newsUpdateSQL)).
		WithArgs("Fantasma", &body, &image, 999999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := repo.Update(context.Background(), 999999, domain.NewsItem{Title: "Fantasma", Body: &body, Image: &image}); err != nil {
		t.Fatalf("Update() error = %v, want nil (zero rows affected is a success, matching legacy mysqli parity)", err)
	}
}

func TestNewsRepository_Delete_ExecutesWithID(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewNewsRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(newsDeleteSQL)).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
}

func TestNewsRepository_Delete_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewNewsRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectExec(regexp.QuoteMeta(newsDeleteSQL)).
		WithArgs(7).
		WillReturnError(dbErr)

	err := repo.Delete(context.Background(), 7)
	if !errors.Is(err, dbErr) {
		t.Errorf("Delete() error = %v, want it to wrap %v", err, dbErr)
	}
}

// TestNewsRepository_Delete_SucceedsWithZeroRowsAffected mirrors
// TestNewsRepository_Update_SucceedsWithZeroRowsAffected for DELETE — see
// backend/docs/legacy-quirks.md §10.
func TestNewsRepository_Delete_SucceedsWithZeroRowsAffected(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewNewsRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(newsDeleteSQL)).
		WithArgs(999999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := repo.Delete(context.Background(), 999999); err != nil {
		t.Fatalf("Delete() error = %v, want nil (zero rows affected is a success, matching legacy mysqli parity)", err)
	}
}
