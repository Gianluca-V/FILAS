package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

// fakeGalleryRepo uses pointer receivers so Create/Update/Delete tests can
// record what was actually passed to the repository, mirroring the
// fakeAdminRepo convention established in PR3.
type fakeGalleryRepo struct {
	images []domain.GalleryImage
	byID   map[int]domain.GalleryImage
	err    error

	createErr    error
	createdImage domain.GalleryImage
	updateErr    error
	updateCalled bool
	updateID     int
	updateImage  string
	deleteErr    error
	deleteID     int
}

func (f *fakeGalleryRepo) List(ctx context.Context) ([]domain.GalleryImage, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.images, nil
}

func (f *fakeGalleryRepo) Get(ctx context.Context, id int) (domain.GalleryImage, error) {
	if f.err != nil {
		return domain.GalleryImage{}, f.err
	}
	img, ok := f.byID[id]
	if !ok {
		return domain.GalleryImage{}, domain.ErrNotFound
	}
	return img, nil
}

func (f *fakeGalleryRepo) Create(ctx context.Context, g domain.GalleryImage) (domain.GalleryImage, error) {
	if f.createErr != nil {
		return domain.GalleryImage{}, f.createErr
	}
	f.createdImage = g
	g.ID = 9001
	return g, nil
}

func (f *fakeGalleryRepo) Update(ctx context.Context, id int, image string) error {
	f.updateCalled = true
	f.updateID = id
	f.updateImage = image
	return f.updateErr
}

func (f *fakeGalleryRepo) Delete(ctx context.Context, id int) error {
	f.deleteID = id
	return f.deleteErr
}

func TestGalleryService_List_ReturnsRepositoryImages(t *testing.T) {
	want := []domain.GalleryImage{{ID: 1, Image: "assets/galeria-1.jpg"}, {ID: 2, Image: "assets/galeria-2.jpg"}}
	svc := usecase.NewGalleryService(&fakeGalleryRepo{images: want})

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
	svc := usecase.NewGalleryService(&fakeGalleryRepo{err: repoErr})

	_, err := svc.List(context.Background())
	if !errors.Is(err, repoErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestGalleryService_Get_ReturnsMatchingImage(t *testing.T) {
	want := domain.GalleryImage{ID: 5, Image: "assets/galeria-5.jpg"}
	svc := usecase.NewGalleryService(&fakeGalleryRepo{byID: map[int]domain.GalleryImage{5: want}})

	got, err := svc.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != want {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestGalleryService_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	svc := usecase.NewGalleryService(&fakeGalleryRepo{byID: map[int]domain.GalleryImage{}})

	_, err := svc.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestGalleryService_Get_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewGalleryService(&fakeGalleryRepo{err: repoErr})

	_, err := svc.Get(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestGalleryService_Create_PersistsAndReturnsRepoAssignedID(t *testing.T) {
	repo := &fakeGalleryRepo{}
	svc := usecase.NewGalleryService(repo)

	got, err := svc.Create(context.Background(), domain.GalleryImage{Image: "assets/nueva.jpg"})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (repo-assigned via LastInsertId)", got.ID)
	}
	if repo.createdImage.Image != "assets/nueva.jpg" {
		t.Errorf("Create() persisted Image = %q, want %q", repo.createdImage.Image, "assets/nueva.jpg")
	}
}

func TestGalleryService_Create_RejectsEmptyImage(t *testing.T) {
	repo := &fakeGalleryRepo{}
	svc := usecase.NewGalleryService(repo)

	_, err := svc.Create(context.Background(), domain.GalleryImage{Image: ""})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Create() error = %v, want %v", err, domain.ErrValidation)
	}
	if repo.createdImage.Image != "" {
		t.Errorf("Create() should not have called the repository, but persisted %+v", repo.createdImage)
	}
}

func TestGalleryService_Create_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeGalleryRepo{createErr: repoErr}
	svc := usecase.NewGalleryService(repo)

	_, err := svc.Create(context.Background(), domain.GalleryImage{Image: "assets/nueva.jpg"})
	if !errors.Is(err, repoErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestGalleryService_Update_PersistsImage(t *testing.T) {
	repo := &fakeGalleryRepo{}
	svc := usecase.NewGalleryService(repo)

	if err := svc.Update(context.Background(), 5, "assets/actualizada.jpg"); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if repo.updateID != 5 {
		t.Errorf("Update() called repo with id = %d, want 5", repo.updateID)
	}
	if repo.updateImage != "assets/actualizada.jpg" {
		t.Errorf("Update() called repo with Image = %q, want %q", repo.updateImage, "assets/actualizada.jpg")
	}
}

func TestGalleryService_Update_RejectsEmptyImage(t *testing.T) {
	repo := &fakeGalleryRepo{}
	svc := usecase.NewGalleryService(repo)

	err := svc.Update(context.Background(), 5, "")
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Update() error = %v, want %v", err, domain.ErrValidation)
	}
	if repo.updateCalled {
		t.Errorf("Update() called the repository with an empty Image — no persistence should happen on validation failure")
	}
}

func TestGalleryService_Update_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeGalleryRepo{updateErr: repoErr}
	svc := usecase.NewGalleryService(repo)

	err := svc.Update(context.Background(), 5, "assets/actualizada.jpg")
	if !errors.Is(err, repoErr) {
		t.Errorf("Update() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestGalleryService_Delete_DelegatesToRepository(t *testing.T) {
	repo := &fakeGalleryRepo{}
	svc := usecase.NewGalleryService(repo)

	if err := svc.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
	if repo.deleteID != 7 {
		t.Errorf("Delete() called repo with id = %d, want 7", repo.deleteID)
	}
}

func TestGalleryService_Delete_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeGalleryRepo{deleteErr: repoErr}
	svc := usecase.NewGalleryService(repo)

	err := svc.Delete(context.Background(), 7)
	if !errors.Is(err, repoErr) {
		t.Errorf("Delete() error = %v, want it to wrap %v", err, repoErr)
	}
}
