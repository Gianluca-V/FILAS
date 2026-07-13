package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

type fakeNewsRepo struct {
	items []domain.NewsItem
	byID  map[int]domain.NewsItem
	err   error
}

func (f fakeNewsRepo) List(ctx context.Context) ([]domain.NewsItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f fakeNewsRepo) Get(ctx context.Context, id int) (domain.NewsItem, error) {
	if f.err != nil {
		return domain.NewsItem{}, f.err
	}
	item, ok := f.byID[id]
	if !ok {
		return domain.NewsItem{}, domain.ErrNotFound
	}
	return item, nil
}

func TestNewsService_List_ReturnsRepositoryItems(t *testing.T) {
	want := []domain.NewsItem{{ID: 1, Title: "Noticia de prueba 1"}}
	svc := usecase.NewNewsService(fakeNewsRepo{items: want})

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %+v, want %+v", got, want)
	}
}

func TestNewsService_List_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewNewsService(fakeNewsRepo{err: repoErr})

	_, err := svc.List(context.Background())
	if !errors.Is(err, repoErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestNewsService_Get_ReturnsMatchingItem(t *testing.T) {
	want := domain.NewsItem{ID: 1, Title: "Noticia de prueba 1"}
	svc := usecase.NewNewsService(fakeNewsRepo{byID: map[int]domain.NewsItem{1: want}})

	got, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != want {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestNewsService_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	svc := usecase.NewNewsService(fakeNewsRepo{byID: map[int]domain.NewsItem{}})

	_, err := svc.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestNewsService_Get_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewNewsService(fakeNewsRepo{err: repoErr})

	_, err := svc.Get(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, repoErr)
	}
}
