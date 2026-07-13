package usecase

import (
	"context"

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
