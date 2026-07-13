package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

// fakeOrganizationRepo uses pointer receivers so Create/Update/Delete tests
// can record what was actually passed to the repository, mirroring the
// fakeAdminRepo convention established in PR3.
type fakeOrganizationRepo struct {
	items []domain.Organization
	byID  map[int]domain.Organization
	err   error

	createErr    error
	createdItem  domain.Organization
	updateErr    error
	updateCalled bool
	updateID     int
	updateItem   domain.Organization
	deleteErr    error
	deleteID     int
}

func (f *fakeOrganizationRepo) List(ctx context.Context) ([]domain.Organization, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f *fakeOrganizationRepo) Get(ctx context.Context, id int) (domain.Organization, error) {
	if f.err != nil {
		return domain.Organization{}, f.err
	}
	item, ok := f.byID[id]
	if !ok {
		return domain.Organization{}, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakeOrganizationRepo) Create(ctx context.Context, o domain.Organization) (domain.Organization, error) {
	if f.createErr != nil {
		return domain.Organization{}, f.createErr
	}
	f.createdItem = o
	o.ID = 9001
	return o, nil
}

func (f *fakeOrganizationRepo) Update(ctx context.Context, id int, o domain.Organization) error {
	f.updateCalled = true
	f.updateID = id
	f.updateItem = o
	return f.updateErr
}

func (f *fakeOrganizationRepo) Delete(ctx context.Context, id int) error {
	f.deleteID = id
	return f.deleteErr
}

func TestOrganizationService_List_ReturnsRepositoryItems(t *testing.T) {
	want := []domain.Organization{{ID: 1, Title: "organizacions de prueba 1"}}
	svc := usecase.NewOrganizationService(&fakeOrganizationRepo{items: want})

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
	svc := usecase.NewOrganizationService(&fakeOrganizationRepo{err: repoErr})

	_, err := svc.List(context.Background())
	if !errors.Is(err, repoErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestOrganizationService_Get_ReturnsMatchingItem(t *testing.T) {
	want := domain.Organization{ID: 1, Title: "organizacions de prueba 1"}
	svc := usecase.NewOrganizationService(&fakeOrganizationRepo{byID: map[int]domain.Organization{1: want}})

	got, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != want {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestOrganizationService_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	svc := usecase.NewOrganizationService(&fakeOrganizationRepo{byID: map[int]domain.Organization{}})

	_, err := svc.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestOrganizationService_Get_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewOrganizationService(&fakeOrganizationRepo{err: repoErr})

	_, err := svc.Get(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestOrganizationService_Create_PersistsAndReturnsRepoAssignedID(t *testing.T) {
	repo := &fakeOrganizationRepo{}
	svc := usecase.NewOrganizationService(repo)

	got, err := svc.Create(context.Background(), domain.Organization{Title: "Nueva organizacion", Image: "https://example.com/n.jpg"})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (repo-assigned via LastInsertId)", got.ID)
	}
	if repo.createdItem.Title != "Nueva organizacion" {
		t.Errorf("Create() persisted Title = %q, want %q", repo.createdItem.Title, "Nueva organizacion")
	}
}

func TestOrganizationService_Create_RejectsMissingTitleOrImage(t *testing.T) {
	tests := []struct {
		name  string
		title string
		image string
	}{
		{"missing title", "", "https://example.com/n.jpg"},
		{"missing image", "Nueva organizacion", ""},
		{"missing both", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeOrganizationRepo{}
			svc := usecase.NewOrganizationService(repo)

			_, err := svc.Create(context.Background(), domain.Organization{Title: tt.title, Image: tt.image})
			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Create() error = %v, want %v", err, domain.ErrValidation)
			}
			if repo.createdItem.Title != "" {
				t.Errorf("Create() should not have called the repository, but persisted %+v", repo.createdItem)
			}
		})
	}
}

func TestOrganizationService_Create_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeOrganizationRepo{createErr: repoErr}
	svc := usecase.NewOrganizationService(repo)

	_, err := svc.Create(context.Background(), domain.Organization{Title: "Nueva organizacion", Image: "https://example.com/n.jpg"})
	if !errors.Is(err, repoErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestOrganizationService_Update_PersistsFields(t *testing.T) {
	repo := &fakeOrganizationRepo{}
	svc := usecase.NewOrganizationService(repo)

	item := domain.Organization{Title: "renombrada", Image: "https://example.com/r.jpg"}
	if err := svc.Update(context.Background(), 5, item); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if repo.updateID != 5 {
		t.Errorf("Update() called repo with id = %d, want 5", repo.updateID)
	}
	if repo.updateItem.Title != "renombrada" {
		t.Errorf("Update() called repo with Title = %q, want %q", repo.updateItem.Title, "renombrada")
	}
}

func TestOrganizationService_Update_RejectsMissingTitleOrImage(t *testing.T) {
	repo := &fakeOrganizationRepo{}
	svc := usecase.NewOrganizationService(repo)

	err := svc.Update(context.Background(), 5, domain.Organization{Title: "", Image: ""})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Update() error = %v, want %v", err, domain.ErrValidation)
	}
	if repo.updateCalled {
		t.Errorf("Update() called the repository with missing Title/Image — no persistence should happen on validation failure")
	}
}

func TestOrganizationService_Update_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeOrganizationRepo{updateErr: repoErr}
	svc := usecase.NewOrganizationService(repo)

	err := svc.Update(context.Background(), 5, domain.Organization{Title: "renombrada", Image: "https://example.com/r.jpg"})
	if !errors.Is(err, repoErr) {
		t.Errorf("Update() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestOrganizationService_Delete_DelegatesToRepository(t *testing.T) {
	repo := &fakeOrganizationRepo{}
	svc := usecase.NewOrganizationService(repo)

	if err := svc.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
	if repo.deleteID != 7 {
		t.Errorf("Delete() called repo with id = %d, want 7", repo.deleteID)
	}
}

func TestOrganizationService_Delete_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeOrganizationRepo{deleteErr: repoErr}
	svc := usecase.NewOrganizationService(repo)

	err := svc.Delete(context.Background(), 7)
	if !errors.Is(err, repoErr) {
		t.Errorf("Delete() error = %v, want it to wrap %v", err, repoErr)
	}
}
