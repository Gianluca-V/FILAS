package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

type fakeGalleryRepo struct {
	images []domain.GalleryImage
	byID   map[int]domain.GalleryImage
	err    error
}

func (f fakeGalleryRepo) List(ctx context.Context) ([]domain.GalleryImage, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.images, nil
}

func (f fakeGalleryRepo) Get(ctx context.Context, id int) (domain.GalleryImage, error) {
	if f.err != nil {
		return domain.GalleryImage{}, f.err
	}
	img, ok := f.byID[id]
	if !ok {
		return domain.GalleryImage{}, domain.ErrNotFound
	}
	return img, nil
}

func TestGalleryService_List_ReturnsRepositoryImages(t *testing.T) {
	want := []domain.GalleryImage{{ID: 1, Image: "assets/galeria-1.jpg"}, {ID: 2, Image: "assets/galeria-2.jpg"}}
	svc := usecase.NewGalleryService(fakeGalleryRepo{images: want})

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %+v, want %+v", got, want)
	}
}

func TestGalleryService_List_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewGalleryService(fakeGalleryRepo{err: repoErr})

	_, err := svc.List(context.Background())
	if !errors.Is(err, repoErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestGalleryService_Get_ReturnsMatchingImage(t *testing.T) {
	want := domain.GalleryImage{ID: 5, Image: "assets/galeria-5.jpg"}
	svc := usecase.NewGalleryService(fakeGalleryRepo{byID: map[int]domain.GalleryImage{5: want}})

	got, err := svc.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != want {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestGalleryService_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	svc := usecase.NewGalleryService(fakeGalleryRepo{byID: map[int]domain.GalleryImage{}})

	_, err := svc.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestGalleryService_Get_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewGalleryService(fakeGalleryRepo{err: repoErr})

	_, err := svc.Get(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, repoErr)
	}
}
