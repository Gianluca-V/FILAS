package usecase

import (
	"context"
	"fmt"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// NewsService is consumed by internal/handler/rest.
type NewsService struct {
	repo domain.NewsRepository
}

// NewNewsService wires a NewsService to its repository.
func NewNewsService(repo domain.NewsRepository) *NewsService {
	return &NewsService{repo: repo}
}

func (s *NewsService) List(ctx context.Context) ([]domain.NewsItem, error) {
	return s.repo.List(ctx)
}

func (s *NewsService) Get(ctx context.Context, id int) (domain.NewsItem, error) {
	return s.repo.Get(ctx, id)
}

// validateNewsItem mirrors legacy news.php's POST/PUT presence check
// (isset($data->Title) && isset($data->Body) && isset($data->Image) -> 400
// "Missing Title, Body, or Image parameter"), generalized to reject an
// empty string too, not just an absent JSON key — same "never persist a
// zero-value where a real value is required" principle applied to
// gallery/family/product writes in PR4.
func validateNewsItem(n domain.NewsItem) error {
	if n.Title == "" || n.Body == nil || *n.Body == "" || n.Image == nil || *n.Image == "" {
		return fmt.Errorf("title, body, and image are required: %w", domain.ErrValidation)
	}
	return nil
}

// Create validates Title/Body/Image (see validateNewsItem), then persists a
// new news item.
func (s *NewsService) Create(ctx context.Context, n domain.NewsItem) (domain.NewsItem, error) {
	if err := validateNewsItem(n); err != nil {
		return domain.NewsItem{}, err
	}
	created, err := s.repo.Create(ctx, n)
	if err != nil {
		return domain.NewsItem{}, fmt.Errorf("usecase: create news item: %w", err)
	}
	return created, nil
}

// Update validates Title/Body/Image (same rationale as Create), then
// overwrites Title/Body/Image for the given ID.
func (s *NewsService) Update(ctx context.Context, id int, n domain.NewsItem) error {
	if err := validateNewsItem(n); err != nil {
		return err
	}
	if err := s.repo.Update(ctx, id, n); err != nil {
		return fmt.Errorf("usecase: update news item %d: %w", id, err)
	}
	return nil
}

func (s *NewsService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("usecase: delete news item %d: %w", id, err)
	}
	return nil
}
