package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

type fakeFamilyRepo struct {
	items []domain.FamilyItem
	byID  map[int]domain.FamilyItem
	err   error
}

func (f fakeFamilyRepo) List(ctx context.Context) ([]domain.FamilyItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f fakeFamilyRepo) Get(ctx context.Context, id int) (domain.FamilyItem, error) {
	if f.err != nil {
		return domain.FamilyItem{}, f.err
	}
	item, ok := f.byID[id]
	if !ok {
		return domain.FamilyItem{}, domain.ErrNotFound
	}
	return item, nil
}

func TestFamilyService_List_ReturnsRepositoryItems(t *testing.T) {
	want := []domain.FamilyItem{{ID: 3, Body: "aqui va la descripcion 3", Category: "Taller protegido"}}
	svc := usecase.NewFamilyService(fakeFamilyRepo{items: want})

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %+v, want %+v", got, want)
	}
}

func TestFamilyService_List_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewFamilyService(fakeFamilyRepo{err: repoErr})

	_, err := svc.List(context.Background())
	if !errors.Is(err, repoErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestFamilyService_Get_ReturnsMatchingItem(t *testing.T) {
	want := domain.FamilyItem{ID: 3, Body: "aqui va la descripcion 3", Category: "Taller protegido"}
	svc := usecase.NewFamilyService(fakeFamilyRepo{byID: map[int]domain.FamilyItem{3: want}})

	got, err := svc.Get(context.Background(), 3)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != want {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestFamilyService_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	svc := usecase.NewFamilyService(fakeFamilyRepo{byID: map[int]domain.FamilyItem{}})

	_, err := svc.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}
