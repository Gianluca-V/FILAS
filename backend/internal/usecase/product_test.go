package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

type fakeProductRepo struct {
	products []domain.Product
	byID     map[int]domain.Product
	err      error
}

func (f fakeProductRepo) List(ctx context.Context) ([]domain.Product, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.products, nil
}

func (f fakeProductRepo) Get(ctx context.Context, id int) (domain.Product, error) {
	if f.err != nil {
		return domain.Product{}, f.err
	}
	p, ok := f.byID[id]
	if !ok {
		return domain.Product{}, domain.ErrNotFound
	}
	return p, nil
}

func TestProductService_List_ReturnsRepositoryProducts(t *testing.T) {
	image := "assets/default-img.png"
	want := []domain.Product{{ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112, Image: &image}}
	svc := usecase.NewProductService(fakeProductRepo{products: want})

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
	svc := usecase.NewProductService(fakeProductRepo{err: repoErr})

	_, err := svc.List(context.Background())
	if !errors.Is(err, repoErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestProductService_Get_ReturnsMatchingProduct(t *testing.T) {
	want := domain.Product{ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112}
	svc := usecase.NewProductService(fakeProductRepo{byID: map[int]domain.Product{1: want}})

	got, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != want {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestProductService_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	svc := usecase.NewProductService(fakeProductRepo{byID: map[int]domain.Product{}})

	_, err := svc.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestProductService_Get_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewProductService(fakeProductRepo{err: repoErr})

	_, err := svc.Get(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, repoErr)
	}
}
