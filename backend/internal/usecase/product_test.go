package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

// fakeProductRepo uses pointer receivers (unlike the read-only PR2 fakes) so
// Create/Update/Delete tests can record what was actually passed to the
// repository, mirroring the fakeAdminRepo convention established in PR3.
type fakeProductRepo struct {
	products []domain.Product
	byID     map[int]domain.Product
	err      error

	createErr      error
	createdProduct domain.Product
	updateErr      error
	updateCalled   bool
	updateID       int
	updateProduct  domain.Product
	deleteErr      error
	deleteID       int
}

func (f *fakeProductRepo) List(ctx context.Context) ([]domain.Product, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.products, nil
}

func (f *fakeProductRepo) Get(ctx context.Context, id int) (domain.Product, error) {
	if f.err != nil {
		return domain.Product{}, f.err
	}
	p, ok := f.byID[id]
	if !ok {
		return domain.Product{}, domain.ErrNotFound
	}
	return p, nil
}

func (f *fakeProductRepo) Create(ctx context.Context, p domain.Product) (domain.Product, error) {
	if f.createErr != nil {
		return domain.Product{}, f.createErr
	}
	f.createdProduct = p
	p.ID = 9001
	return p, nil
}

func (f *fakeProductRepo) Update(ctx context.Context, id int, p domain.Product) error {
	f.updateCalled = true
	f.updateID = id
	f.updateProduct = p
	return f.updateErr
}

func (f *fakeProductRepo) Delete(ctx context.Context, id int) error {
	f.deleteID = id
	return f.deleteErr
}

func TestProductService_List_ReturnsRepositoryProducts(t *testing.T) {
	image := "assets/default-img.png"
	want := []domain.Product{{ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112, Image: &image}}
	svc := usecase.NewProductService(&fakeProductRepo{products: want})

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %+v, want %+v", got, want)
	}
}

func TestProductService_List_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewProductService(&fakeProductRepo{err: repoErr})

	_, err := svc.List(context.Background())
	if !errors.Is(err, repoErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestProductService_Get_ReturnsMatchingProduct(t *testing.T) {
	want := domain.Product{ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112}
	svc := usecase.NewProductService(&fakeProductRepo{byID: map[int]domain.Product{1: want}})

	got, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != want {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestProductService_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	svc := usecase.NewProductService(&fakeProductRepo{byID: map[int]domain.Product{}})

	_, err := svc.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestProductService_Get_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewProductService(&fakeProductRepo{err: repoErr})

	_, err := svc.Get(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestProductService_Create_PersistsAndReturnsRepoAssignedID(t *testing.T) {
	repo := &fakeProductRepo{}
	svc := usecase.NewProductService(repo)

	got, err := svc.Create(context.Background(), domain.Product{Name: "Nuevo producto", Price: 100, Stock: 5})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (repo-assigned via LastInsertId)", got.ID)
	}
	if repo.createdProduct.Name != "Nuevo producto" {
		t.Errorf("Create() persisted Name = %q, want %q", repo.createdProduct.Name, "Nuevo producto")
	}
}

// TestProductService_Create_RejectsEmptyName generalizes the "never persist
// a zero-value where a real value is required" principle established for
// admins' password field (PR3, gate #36 blocker) to product Name: legacy
// createProduct has NO validation at all, but a nameless product is a
// broken storefront entry (see usecase.ProductService.Create doc comment).
func TestProductService_Create_RejectsEmptyName(t *testing.T) {
	repo := &fakeProductRepo{}
	svc := usecase.NewProductService(repo)

	_, err := svc.Create(context.Background(), domain.Product{Name: "", Price: 100, Stock: 5})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Create() error = %v, want %v", err, domain.ErrValidation)
	}
	if repo.createdProduct.Name != "" || repo.createdProduct.Price != 0 {
		t.Errorf("Create() should not have called the repository, but persisted %+v", repo.createdProduct)
	}
}

func TestProductService_Create_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeProductRepo{createErr: repoErr}
	svc := usecase.NewProductService(repo)

	_, err := svc.Create(context.Background(), domain.Product{Name: "Nuevo producto"})
	if !errors.Is(err, repoErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, repoErr)
	}
}

// TestProductService_Create_AcceptsNegativePriceAndStock locks the
// intentional decision (PR4 corrective, see ProductService.Create's doc
// comment) that Price/Stock range validation is explicitly OUT of scope:
// legacy never range-checked them either, and the seed already has
// negative-Stock rows. This test exists so a future "fix" that silently
// adds range validation breaks a test instead of shipping unreviewed.
func TestProductService_Create_AcceptsNegativePriceAndStock(t *testing.T) {
	repo := &fakeProductRepo{}
	svc := usecase.NewProductService(repo)

	got, err := svc.Create(context.Background(), domain.Product{Name: "Producto con stock negativo", Price: -50, Stock: -3})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil (negative Price/Stock are intentionally accepted)", err)
	}
	if got.Price != -50 || got.Stock != -3 {
		t.Errorf("Create() = %+v, want Price=-50 Stock=-3 preserved", got)
	}
	if repo.createdProduct.Price != -50 || repo.createdProduct.Stock != -3 {
		t.Errorf("Create() persisted Price=%v Stock=%v, want the negative values passed through unchanged", repo.createdProduct.Price, repo.createdProduct.Stock)
	}
}

func TestProductService_Update_PersistsFields(t *testing.T) {
	repo := &fakeProductRepo{}
	svc := usecase.NewProductService(repo)

	if err := svc.Update(context.Background(), 5, domain.Product{Name: "Renombrado", Price: 200, Stock: 10}); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if repo.updateID != 5 {
		t.Errorf("Update() called repo with id = %d, want 5", repo.updateID)
	}
	if repo.updateProduct.Name != "Renombrado" {
		t.Errorf("Update() called repo with Name = %q, want %q", repo.updateProduct.Name, "Renombrado")
	}
}

func TestProductService_Update_RejectsEmptyName(t *testing.T) {
	repo := &fakeProductRepo{}
	svc := usecase.NewProductService(repo)

	err := svc.Update(context.Background(), 5, domain.Product{Name: ""})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Update() error = %v, want %v", err, domain.ErrValidation)
	}
	if repo.updateCalled {
		t.Errorf("Update() called the repository with an empty Name — no persistence should happen on validation failure")
	}
}

func TestProductService_Update_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeProductRepo{updateErr: repoErr}
	svc := usecase.NewProductService(repo)

	err := svc.Update(context.Background(), 5, domain.Product{Name: "Renombrado"})
	if !errors.Is(err, repoErr) {
		t.Errorf("Update() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestProductService_Delete_DelegatesToRepository(t *testing.T) {
	repo := &fakeProductRepo{}
	svc := usecase.NewProductService(repo)

	if err := svc.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
	if repo.deleteID != 7 {
		t.Errorf("Delete() called repo with id = %d, want 7", repo.deleteID)
	}
}

func TestProductService_Delete_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeProductRepo{deleteErr: repoErr}
	svc := usecase.NewProductService(repo)

	err := svc.Delete(context.Background(), 7)
	if !errors.Is(err, repoErr) {
		t.Errorf("Delete() error = %v, want it to wrap %v", err, repoErr)
	}
}
