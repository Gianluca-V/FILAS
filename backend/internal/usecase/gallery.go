package usecase

import (
	"context"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// GalleryService is consumed by internal/handler/rest.
type GalleryService struct {
	repo domain.GalleryRepository
}

// NewGalleryService wires a GalleryService to its repository.
func NewGalleryService(repo domain.GalleryRepository) *GalleryService {
	return &GalleryService{repo: repo}
}

func (s *GalleryService) List(ctx context.Context) ([]domain.GalleryImage, error) {
	return s.repo.List(ctx)
}

func (s *GalleryService) Get(ctx context.Context, id int) (domain.GalleryImage, error) {
	return s.repo.Get(ctx, id)
}
