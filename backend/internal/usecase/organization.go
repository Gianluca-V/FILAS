package usecase

import (
	"context"

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
