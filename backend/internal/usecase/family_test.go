package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

// fakeFamilyRepo uses pointer receivers so Create/Update/Delete tests can
// record what was actually passed to the repository, mirroring the
// fakeAdminRepo convention established in PR3.
type fakeFamilyRepo struct {
	items []domain.FamilyItem
	byID  map[int]domain.FamilyItem
	err   error

	createErr    error
	createdItem  domain.FamilyItem
	updateErr    error
	updateCalled bool
	updateID     int
	updateItem   domain.FamilyItem
	deleteErr    error
	deleteID     int
}

func (f *fakeFamilyRepo) List(ctx context.Context) ([]domain.FamilyItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f *fakeFamilyRepo) Get(ctx context.Context, id int) (domain.FamilyItem, error) {
	if f.err != nil {
		return domain.FamilyItem{}, f.err
	}
	item, ok := f.byID[id]
	if !ok {
		return domain.FamilyItem{}, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakeFamilyRepo) Create(ctx context.Context, item domain.FamilyItem) (domain.FamilyItem, error) {
	if f.createErr != nil {
		return domain.FamilyItem{}, f.createErr
	}
	f.createdItem = item
	item.ID = 9001
	return item, nil
}

func (f *fakeFamilyRepo) Update(ctx context.Context, id int, item domain.FamilyItem) error {
	f.updateCalled = true
	f.updateID = id
	f.updateItem = item
	return f.updateErr
}

func (f *fakeFamilyRepo) Delete(ctx context.Context, id int) error {
	f.deleteID = id
	return f.deleteErr
}

func TestFamilyService_List_ReturnsRepositoryItems(t *testing.T) {
	want := []domain.FamilyItem{{ID: 3, Body: "aqui va la descripcion 3", Category: "Taller protegido"}}
	svc := usecase.NewFamilyService(&fakeFamilyRepo{items: want})

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
	svc := usecase.NewFamilyService(&fakeFamilyRepo{err: repoErr})

	_, err := svc.List(context.Background())
	if !errors.Is(err, repoErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestFamilyService_Get_ReturnsMatchingItem(t *testing.T) {
	want := domain.FamilyItem{ID: 3, Body: "aqui va la descripcion 3", Category: "Taller protegido"}
	svc := usecase.NewFamilyService(&fakeFamilyRepo{byID: map[int]domain.FamilyItem{3: want}})

	got, err := svc.Get(context.Background(), 3)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != want {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestFamilyService_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	svc := usecase.NewFamilyService(&fakeFamilyRepo{byID: map[int]domain.FamilyItem{}})

	_, err := svc.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestFamilyService_Get_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewFamilyService(&fakeFamilyRepo{err: repoErr})

	_, err := svc.Get(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestFamilyService_Create_PersistsAndReturnsRepoAssignedID(t *testing.T) {
	repo := &fakeFamilyRepo{}
	svc := usecase.NewFamilyService(repo)

	got, err := svc.Create(context.Background(), domain.FamilyItem{Body: "cuerpo", Category: domain.FamilyCategoryCentroDeDia})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (repo-assigned via LastInsertId)", got.ID)
	}
	if repo.createdItem.Body != "cuerpo" {
		t.Errorf("Create() persisted Body = %q, want %q", repo.createdItem.Body, "cuerpo")
	}
}

func TestFamilyService_Create_RejectsMissingBodyOrCategory(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		category string
	}{
		{"missing body", "", domain.FamilyCategoryCentroDeDia},
		{"missing category", "cuerpo", ""},
		{"missing both", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeFamilyRepo{}
			svc := usecase.NewFamilyService(repo)

			_, err := svc.Create(context.Background(), domain.FamilyItem{Body: tt.body, Category: tt.category})
			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Create() error = %v, want %v", err, domain.ErrValidation)
			}
			if repo.createdItem.Body != "" {
				t.Errorf("Create() should not have called the repository, but persisted %+v", repo.createdItem)
			}
		})
	}
}

// TestFamilyService_Create_RejectsInvalidCategory locks the Category enum
// invariant from the design/spec (family.php's inline comparison against
// exactly "Centro de dia" / "Taller protegido").
func TestFamilyService_Create_RejectsInvalidCategory(t *testing.T) {
	repo := &fakeFamilyRepo{}
	svc := usecase.NewFamilyService(repo)

	_, err := svc.Create(context.Background(), domain.FamilyItem{Body: "cuerpo", Category: "Invalid"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Create() error = %v, want %v", err, domain.ErrValidation)
	}
	if repo.createdItem.Body != "" {
		t.Errorf("Create() should not have called the repository, but persisted %+v", repo.createdItem)
	}
}

func TestFamilyService_Create_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeFamilyRepo{createErr: repoErr}
	svc := usecase.NewFamilyService(repo)

	_, err := svc.Create(context.Background(), domain.FamilyItem{Body: "cuerpo", Category: domain.FamilyCategoryCentroDeDia})
	if !errors.Is(err, repoErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestFamilyService_Update_PersistsFields(t *testing.T) {
	repo := &fakeFamilyRepo{}
	svc := usecase.NewFamilyService(repo)

	item := domain.FamilyItem{Body: "renombrado", Category: domain.FamilyCategoryTallerProtegido}
	if err := svc.Update(context.Background(), 5, item); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if repo.updateID != 5 {
		t.Errorf("Update() called repo with id = %d, want 5", repo.updateID)
	}
	if repo.updateItem.Body != "renombrado" {
		t.Errorf("Update() called repo with Body = %q, want %q", repo.updateItem.Body, "renombrado")
	}
}

func TestFamilyService_Update_RejectsInvalidCategory(t *testing.T) {
	repo := &fakeFamilyRepo{}
	svc := usecase.NewFamilyService(repo)

	err := svc.Update(context.Background(), 5, domain.FamilyItem{Body: "cuerpo", Category: "Invalid"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Update() error = %v, want %v", err, domain.ErrValidation)
	}
	if repo.updateCalled {
		t.Errorf("Update() called the repository with an invalid category — no persistence should happen on validation failure")
	}
}

func TestFamilyService_Update_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeFamilyRepo{updateErr: repoErr}
	svc := usecase.NewFamilyService(repo)

	err := svc.Update(context.Background(), 5, domain.FamilyItem{Body: "cuerpo", Category: domain.FamilyCategoryCentroDeDia})
	if !errors.Is(err, repoErr) {
		t.Errorf("Update() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestFamilyService_Delete_DelegatesToRepository(t *testing.T) {
	repo := &fakeFamilyRepo{}
	svc := usecase.NewFamilyService(repo)

	if err := svc.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
	if repo.deleteID != 7 {
		t.Errorf("Delete() called repo with id = %d, want 7", repo.deleteID)
	}
}

func TestFamilyService_Delete_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &fakeFamilyRepo{deleteErr: repoErr}
	svc := usecase.NewFamilyService(repo)

	err := svc.Delete(context.Background(), 7)
	if !errors.Is(err, repoErr) {
		t.Errorf("Delete() error = %v, want it to wrap %v", err, repoErr)
	}
}
