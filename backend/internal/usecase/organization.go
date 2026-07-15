package usecase

import (
	"context"
	"fmt"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// OrganizationService is consumed by internal/handler/rest.
type OrganizationService struct {
	repo domain.OrganizationRepository
}

// NewOrganizationService wires an OrganizationService to its repository.
func NewOrganizationService(repo domain.OrganizationRepository) *OrganizationService {
	return &OrganizationService{repo: repo}
}

func (s *OrganizationService) List(ctx context.Context) ([]domain.Organization, error) {
	return s.repo.List(ctx)
}

func (s *OrganizationService) Get(ctx context.Context, id int) (domain.Organization, error) {
	return s.repo.Get(ctx, id)
}

// validateOrganization mirrors legacy organizations.php's check on both
// POST and PUT: isset($data->Title) && isset($data->Image) -> 400
// "Missing Title, or Image parameter". Generalized to reject empty string,
// not just an absent JSON key (same reasoning as FamilyService). Description
// is NOT required — legacy never isset()-checks it.
func validateOrganization(o domain.Organization) error {
	if o.Title == "" || o.Image == "" {
		return fmt.Errorf("title and image are required: %w", domain.ErrValidation)
	}
	return nil
}

// Create validates Title/Image (see validateOrganization), then persists a
// new organization.
func (s *OrganizationService) Create(ctx context.Context, o domain.Organization) (domain.Organization, error) {
	if err := validateOrganization(o); err != nil {
		return domain.Organization{}, err
	}
	created, err := s.repo.Create(ctx, o)
	if err != nil {
		return domain.Organization{}, fmt.Errorf("usecase: create organization: %w", err)
	}
	return created, nil
}

// Update validates Title/Image (see validateOrganization), then overwrites
// Title/Description/Image for the given ID.
func (s *OrganizationService) Update(ctx context.Context, id int, o domain.Organization) error {
	if err := validateOrganization(o); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, id, o); err != nil {
		return fmt.Errorf("usecase: update organization %d: %w", id, err)
	}
	return nil
}

func (s *OrganizationService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("usecase: delete organization %d: %w", id, err)
	}
	return nil
}
