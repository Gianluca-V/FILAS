package usecase

import (
	"context"
	"fmt"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// FamilyService is consumed by internal/handler/rest.
type FamilyService struct {
	repo domain.FamilyRepository
}

// NewFamilyService wires a FamilyService to its repository.
func NewFamilyService(repo domain.FamilyRepository) *FamilyService {
	return &FamilyService{repo: repo}
}

func (s *FamilyService) List(ctx context.Context) ([]domain.FamilyItem, error) {
	return s.repo.List(ctx)
}

func (s *FamilyService) Get(ctx context.Context, id int) (domain.FamilyItem, error) {
	return s.repo.Get(ctx, id)
}

// validate mirrors legacy family.php's two-step check on both POST and PUT:
// first Body/Category presence (legacy: isset($data->Body) &&
// isset($data->Category) -> 400 "Missing Body, or Category parameter"),
// then the Category enum (legacy: 400 "Invalid category parameter").
// Generalized to reject empty string, not just an absent JSON key, for the
// same reason as GalleryService.Create. Image is NOT required — legacy
// never isset()-checks it, it real_escape_string's whatever is present
// (nil/absent -> effectively empty).
func validateFamilyItem(f domain.FamilyItem) error {
	if f.Body == "" || f.Category == "" {
		return fmt.Errorf("body and category are required: %w", domain.ErrValidation)
	}
	if !domain.ValidFamilyCategory(f.Category) {
		return fmt.Errorf("category must be %q or %q: %w", domain.FamilyCategoryCentroDeDia, domain.FamilyCategoryTallerProtegido, domain.ErrValidation)
	}
	return nil
}

// Create validates Body/Category (see validateFamilyItem), then persists a
// new family item.
func (s *FamilyService) Create(ctx context.Context, f domain.FamilyItem) (domain.FamilyItem, error) {
	if err := validateFamilyItem(f); err != nil {
		return domain.FamilyItem{}, err
	}
	created, err := s.repo.Create(ctx, f)
	if err != nil {
		return domain.FamilyItem{}, fmt.Errorf("usecase: create family item: %w", err)
	}
	return created, nil
}

// Update validates Body/Category (see validateFamilyItem), then overwrites
// Body/Category/Image for the given ID.
func (s *FamilyService) Update(ctx context.Context, id int, f domain.FamilyItem) error {
	if err := validateFamilyItem(f); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, id, f); err != nil {
		return fmt.Errorf("usecase: update family item %d: %w", id, err)
	}
	return nil
}

func (s *FamilyService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("usecase: delete family item %d: %w", id, err)
	}
	return nil
}
