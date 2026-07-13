package usecase

import (
	"context"

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
