package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

type fakeOrganizationRepo struct {
	items []domain.Organization
	byID  map[int]domain.Organization
	err   error
}

func (f fakeOrganizationRepo) List(ctx context.Context) ([]domain.Organization, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f fakeOrganizationRepo) Get(ctx context.Context, id int) (domain.Organization, error) {
	if f.err != nil {
		return domain.Organization{}, f.err
	}
	item, ok := f.byID[id]
	if !ok {
		return domain.Organization{}, domain.ErrNotFound
	}
	return item, nil
}

func TestOrganizationService_List_ReturnsRepositoryItems(t *testing.T) {
	want := []domain.Organization{{ID: 1, Title: "organizacions de prueba 1"}}
	svc := usecase.NewOrganizationService(fakeOrganizationRepo{items: want})

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %+v, want %+v", got, want)
	}
}

func TestOrganizationService_List_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewOrganizationService(fakeOrganizationRepo{err: repoErr})

	_, err := svc.List(context.Background())
	if !errors.Is(err, repoErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestOrganizationService_Get_ReturnsMatchingItem(t *testing.T) {
	want := domain.Organization{ID: 1, Title: "organizacions de prueba 1"}
	svc := usecase.NewOrganizationService(fakeOrganizationRepo{byID: map[int]domain.Organization{1: want}})

	got, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != want {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestOrganizationService_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	svc := usecase.NewOrganizationService(fakeOrganizationRepo{byID: map[int]domain.Organization{}})

	_, err := svc.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}
