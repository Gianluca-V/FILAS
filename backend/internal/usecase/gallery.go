package usecase

import (
	"context"
	"fmt"

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

// Create validates Image is non-empty, then persists a new gallery image.
// Legacy gallery.php already gated this with isset($data->Image) -> 400
// "Missing Image parameter", but isset() only checks the KEY is present,
// not that the value is non-empty — an explicit `{"Image":""}` body would
// have slipped through legacy and inserted an empty row. This generalizes
// the check to reject empty string too (same "never persist a zero-value
// where a real value is required" principle as PR3's admin password fix),
// a deliberate, documented hardening beyond legacy's literal isset() check.
func (s *GalleryService) Create(ctx context.Context, g domain.GalleryImage) (domain.GalleryImage, error) {
	if g.Image == "" {
		return domain.GalleryImage{}, fmt.Errorf("image is required: %w", domain.ErrValidation)
	}
	created, err := s.repo.Create(ctx, g)
	if err != nil {
		return domain.GalleryImage{}, fmt.Errorf("usecase: create gallery image: %w", err)
	}
	return created, nil
}

// Update validates image is non-empty (same rationale as Create), then
// overwrites Image for the given ID.
func (s *GalleryService) Update(ctx context.Context, id int, image string) error {
	if image == "" {
		return fmt.Errorf("image is required: %w", domain.ErrValidation)
	}
	if err := s.repo.Update(ctx, id, image); err != nil {
		return fmt.Errorf("usecase: update gallery image %d: %w", id, err)
	}
	return nil
}

func (s *GalleryService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("usecase: delete gallery image %d: %w", id, err)
	}
	return nil
}
