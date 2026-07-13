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

const productSelectCols = "SELECT ID, Name, Price, Stock, Image, Description FROM products"

func TestProductRepository_List_MapsNullAndNonNullDescription(t *testing.T) {
	// Triangulation: legacy seed data has BOTH empty-string and NULL
	// Description rows (see obs #31/#29 characterization). Both must be
	// preserved as distinct values through the domain layer.
	db, mock := newTestDB(t)
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

func TestProductRepository_List_MapsNullImage(t *testing.T) {
	// BLOCKER (gate #34): products.Image is `text DEFAULT NULL` in the
	// schema, but productRow.Image was a plain Go string, so a single NULL
	// Image row made sqlx's row scan fail and turned the ENTIRE list into a
	// 500. Legacy PHP emits "Image":null for a NULL column, so the domain
	// layer must preserve NULL vs "" exactly as it already does for
	// Description.
	db, mock := newTestDB(t)
	repo := mysql.NewProductRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Name", "Price", "Stock", "Image", "Description"}).
		AddRow(18, "Encurtido de ajo", 700.0, 1, nil, nil)
	mock.ExpectQuery(regexp.QuoteMeta(productSelectCols)).WillReturnRows(rows)

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil (NULL Image must not fail the scan)", err)
	}
	if len(got) != 1 {
		t.Fatalf("List() returned %d products, want 1", len(got))
	}
	if got[0].Image != nil {
		t.Errorf("got[0].Image = %v, want nil (NULL in DB)", *got[0].Image)
	}
}

func TestProductRepository_Get_MapsNullImage(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewProductRepository(db)

	rows := sqlmock.NewRows([]string{"ID", "Name", "Price", "Stock", "Image", "Description"}).
		AddRow(18, "Encurtido de ajo", 700.0, 1, nil, nil)
	mock.ExpectQuery(regexp.QuoteMeta(productSelectCols + " WHERE ID = ?")).
		WithArgs(18).
		WillReturnRows(rows)

	got, err := repo.Get(context.Background(), 18)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil (NULL Image must not fail the scan)", err)
	}
	if got.Image != nil {
		t.Errorf("got.Image = %v, want nil (NULL in DB)", *got.Image)
	}
}

func TestProductRepository_List_ReturnsEmptySliceWhenNoRows(t *testing.T) {
	db, mock := newTestDB(t)
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
	db, mock := newTestDB(t)
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
	db, mock := newTestDB(t)
	repo := mysql.NewProductRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(productSelectCols + " WHERE ID = ?")).
		WithArgs(999999).
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Name", "Price", "Stock", "Image", "Description"}))

	_, err := repo.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestProductRepository_List_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewProductRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(productSelectCols)).WillReturnError(dbErr)

	_, err := repo.List(context.Background())
	if !errors.Is(err, dbErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestProductRepository_Get_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewProductRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(productSelectCols + " WHERE ID = ?")).
		WithArgs(1).
		WillReturnError(dbErr)

	_, err := repo.Get(context.Background(), 1)
	if !errors.Is(err, dbErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, dbErr)
	}
}

const productInsertSQL = "INSERT INTO products (Name, Price, Stock, Image, Description) VALUES (?, ?, ?, ?, ?)"
const productUpdateSQL = "UPDATE products SET Name = ?, Price = ?, Stock = ?, Image = ? WHERE ID = ?"
const productDeleteSQL = "DELETE FROM products WHERE ID = ?"

func TestProductRepository_Create_AssignsAutoIncrementID(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewProductRepository(db)

	image := "assets/nuevo.png"
	mock.ExpectExec(regexp.QuoteMeta(productInsertSQL)).
		WithArgs("Nuevo producto", 150.0, 20, &image, (*string)(nil)).
		WillReturnResult(sqlmock.NewResult(9001, 1))

	got, err := repo.Create(context.Background(), domain.Product{Name: "Nuevo producto", Price: 150.0, Stock: 20, Image: &image})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (from LastInsertId)", got.ID)
	}
}

func TestProductRepository_Create_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewProductRepository(db)

	dbErr := errors.New("duplicate entry")
	mock.ExpectExec(regexp.QuoteMeta(productInsertSQL)).
		WithArgs("Nuevo producto", 150.0, 20, (*string)(nil), (*string)(nil)).
		WillReturnError(dbErr)

	_, err := repo.Create(context.Background(), domain.Product{Name: "Nuevo producto", Price: 150.0, Stock: 20})
	if !errors.Is(err, dbErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestProductRepository_Update_ExcludesDescription(t *testing.T) {
	// Locks the deliberate legacy quirk documented on
	// domain.ProductRepository.Update: Description is never part of the
	// UPDATE statement.
	db, mock := newTestDB(t)
	repo := mysql.NewProductRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(productUpdateSQL)).
		WithArgs("Renombrado", 200.0, 10, (*string)(nil), 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), 5, domain.Product{Name: "Renombrado", Price: 200.0, Stock: 10}); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
}

func TestProductRepository_Update_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewProductRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectExec(regexp.QuoteMeta(productUpdateSQL)).
		WithArgs("Renombrado", 200.0, 10, (*string)(nil), 5).
		WillReturnError(dbErr)

	err := repo.Update(context.Background(), 5, domain.Product{Name: "Renombrado", Price: 200.0, Stock: 10})
	if !errors.Is(err, dbErr) {
		t.Errorf("Update() error = %v, want it to wrap %v", err, dbErr)
	}
}

// TestProductRepository_Update_SucceedsWithZeroRowsAffected is a PR4
// corrective reliability lock: legacy updateProduct only checks
// `$conn->query($sql) === TRUE`, and mysqli::query() returns TRUE for ANY
// successful DML — including an UPDATE that matches zero rows (a missing
// or non-numeric ID). Go's database/sql has the exact same semantics: a
// zero-rows-affected Exec is not an error. sqlmock.NewResult(0, 0) proves
// Update does not treat "0 rows matched" as a failure — see
// backend/docs/legacy-quirks.md §10.
func TestProductRepository_Update_SucceedsWithZeroRowsAffected(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewProductRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(productUpdateSQL)).
		WithArgs("Fantasma", 1.0, 1, (*string)(nil), 999999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := repo.Update(context.Background(), 999999, domain.Product{Name: "Fantasma", Price: 1.0, Stock: 1}); err != nil {
		t.Fatalf("Update() error = %v, want nil (zero rows affected is a success, matching legacy mysqli parity)", err)
	}
}

func TestProductRepository_Delete_ExecutesWithID(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewProductRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(productDeleteSQL)).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
}

func TestProductRepository_Delete_PropagatesExecError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewProductRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectExec(regexp.QuoteMeta(productDeleteSQL)).
		WithArgs(7).
		WillReturnError(dbErr)

	err := repo.Delete(context.Background(), 7)
	if !errors.Is(err, dbErr) {
		t.Errorf("Delete() error = %v, want it to wrap %v", err, dbErr)
	}
}

// TestProductRepository_Delete_SucceedsWithZeroRowsAffected mirrors
// TestProductRepository_Update_SucceedsWithZeroRowsAffected for DELETE —
// see backend/docs/legacy-quirks.md §10.
func TestProductRepository_Delete_SucceedsWithZeroRowsAffected(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewProductRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(productDeleteSQL)).
		WithArgs(999999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := repo.Delete(context.Background(), 999999); err != nil {
		t.Fatalf("Delete() error = %v, want nil (zero rows affected is a success, matching legacy mysqli parity)", err)
	}
}
