package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

// fakeNewsRepo uses pointer receivers (unlike the read-only PR2 fake) so
// Create/Update/Delete tests can record what was actually passed to the
// repository, mirroring the fakeProductRepo convention established in PR4.
type fakeNewsRepo struct {
	items []domain.NewsItem
	byID  map[int]domain.NewsItem
	err   error

	createErr    error
	createdItem  domain.NewsItem
	updateErr    error
	updateCalled bool
	updateID     int
	updateItem   domain.NewsItem
	deleteErr    error
	deleteID     int
}

func (f *fakeNewsRepo) List(ctx context.Context) ([]domain.NewsItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f *fakeNewsRepo) Get(ctx context.Context, id int) (domain.NewsItem, error) {
	if f.err != nil {
		return domain.NewsItem{}, f.err
	}
	item, ok := f.byID[id]
	if !ok {
		return domain.NewsItem{}, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakeNewsRepo) Create(ctx context.Context, n domain.NewsItem) (domain.NewsItem, error) {
	if f.createErr != nil {
		return domain.NewsItem{}, f.createErr
	}
	f.createdItem = n
	n.ID = 9001
	return n, nil
}

func (f *fakeNewsRepo) Update(ctx context.Context, id int, n domain.NewsItem) error {
	f.updateCalled = true
	f.updateID = id
	f.updateItem = n
	return f.updateErr
}

func (f *fakeNewsRepo) Delete(ctx context.Context, id int) error {
	f.deleteID = id
	return f.deleteErr
}

func TestNewsService_List_ReturnsRepositoryItems(t *testing.T) {
	want := []domain.NewsItem{{ID: 1, Title: "Noticia de prueba 1"}}
	svc := usecase.NewNewsService(&fakeNewsRepo{items: want})

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
	svc := usecase.NewNewsService(&fakeNewsRepo{err: repoErr})

	_, err := svc.List(context.Background())
	if !errors.Is(err, repoErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestNewsService_Get_ReturnsMatchingItem(t *testing.T) {
	want := domain.NewsItem{ID: 1, Title: "Noticia de prueba 1"}
	svc := usecase.NewNewsService(&fakeNewsRepo{byID: map[int]domain.NewsItem{1: want}})

	got, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != want {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestNewsService_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	svc := usecase.NewNewsService(&fakeNewsRepo{byID: map[int]domain.NewsItem{}})

	_, err := svc.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestNewsService_Get_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewNewsService(&fakeNewsRepo{err: repoErr})

	_, err := svc.Get(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, repoErr)
	}
}

// --- Create ---

func TestNewsService_Create_PersistsAndReturnsRepoAssignedID(t *testing.T) {
	repo := &fakeNewsRepo{}
	svc := usecase.NewNewsService(repo)

	body := "Lorem ipsum"
	image := "https://example.com/1.jpg"
	got, err := svc.Create(context.Background(), domain.NewsItem{Title: "Noticia nueva", Body: &body, Image: &image})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (repo-assigned via LastInsertId)", got.ID)
	}
	if repo.createdItem.Title != "Noticia nueva" {
		t.Errorf("Create() persisted Title = %q, want %q", repo.createdItem.Title, "Noticia nueva")
	}
}

// TestNewsService_Create_RejectsMissingTitleBodyOrImage locks legacy
// news.php's POST validation (isset($data->Title) && isset($data->Body) &&
// isset($data->Image) -> 400 "Missing Title, Body, or Image parameter"),
// generalized to reject empty string too (not just an absent JSON key),
// same "never persist a zero-value where a real value is required"
// principle applied to gallery/family/product writes in PR4.
func TestNewsService_Create_RejectsMissingTitleBodyOrImage(t *testing.T) {
	body := "Lorem ipsum"
	image := "https://example.com/1.jpg"
	cases := []struct {
		name string
		item domain.NewsItem
	}{
		{"missing Title", domain.NewsItem{Title: "", Body: &body, Image: &image}},
		{"nil Body", domain.NewsItem{Title: "Noticia", Body: nil, Image: &image}},
		{"empty Body", domain.NewsItem{Title: "Noticia", Body: strPtr(""), Image: &image}},
		{"nil Image", domain.NewsItem{Title: "Noticia", Body: &body, Image: nil}},
		{"empty Image", domain.NewsItem{Title: "Noticia", Body: &body, Image: strPtr("")}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeNewsRepo{}
			svc := usecase.NewNewsService(repo)

			_, err := svc.Create(context.Background(), tc.item)
			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Create() error = %v, want %v", err, domain.ErrValidation)
			}
			if repo.createdItem.Title != "" {
				t.Errorf("Create() should not have called the repository, but persisted %+v", repo.createdItem)
			}
		})
	}
}

func TestNewsService_Create_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeNewsRepo{createErr: repoErr}
	svc := usecase.NewNewsService(repo)

	body := "Lorem ipsum"
	image := "https://example.com/1.jpg"
	_, err := svc.Create(context.Background(), domain.NewsItem{Title: "Noticia nueva", Body: &body, Image: &image})
	if !errors.Is(err, repoErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, repoErr)
	}
}

// --- Update ---

func TestNewsService_Update_PersistsFields(t *testing.T) {
	repo := &fakeNewsRepo{}
	svc := usecase.NewNewsService(repo)

	body := "Cuerpo actualizado"
	image := "https://example.com/2.jpg"
	if err := svc.Update(context.Background(), 5, domain.NewsItem{Title: "Renombrada", Body: &body, Image: &image}); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if repo.updateID != 5 {
		t.Errorf("Update() called repo with id = %d, want 5", repo.updateID)
	}
	if repo.updateItem.Title != "Renombrada" {
		t.Errorf("Update() called repo with Title = %q, want %q", repo.updateItem.Title, "Renombrada")
	}
}

func TestNewsService_Update_RejectsMissingTitleBodyOrImage(t *testing.T) {
	repo := &fakeNewsRepo{}
	svc := usecase.NewNewsService(repo)

	err := svc.Update(context.Background(), 5, domain.NewsItem{Title: ""})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Update() error = %v, want %v", err, domain.ErrValidation)
	}
	if repo.updateCalled {
		t.Errorf("Update() called the repository with an invalid item — no persistence should happen on validation failure")
	}
}

func TestNewsService_Update_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeNewsRepo{updateErr: repoErr}
	svc := usecase.NewNewsService(repo)

	body := "Cuerpo"
	image := "https://example.com/2.jpg"
	err := svc.Update(context.Background(), 5, domain.NewsItem{Title: "Renombrada", Body: &body, Image: &image})
	if !errors.Is(err, repoErr) {
		t.Errorf("Update() error = %v, want it to wrap %v", err, repoErr)
	}
}

// --- Delete ---

func TestNewsService_Delete_DelegatesToRepository(t *testing.T) {
	repo := &fakeNewsRepo{}
	svc := usecase.NewNewsService(repo)

	if err := svc.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
	if repo.deleteID != 7 {
		t.Errorf("Delete() called repo with id = %d, want 7", repo.deleteID)
	}
}

func TestNewsService_Delete_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeNewsRepo{deleteErr: repoErr}
	svc := usecase.NewNewsService(repo)

	err := svc.Delete(context.Background(), 7)
	if !errors.Is(err, repoErr) {
		t.Errorf("Delete() error = %v, want it to wrap %v", err, repoErr)
	}
}

func strPtr(s string) *string { return &s }
