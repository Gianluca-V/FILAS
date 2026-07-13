package usecase

import (
	"context"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// ProductService is consumed by internal/handler/rest.
type ProductService struct {
	repo domain.ProductRepository
}

// NewProductService wires a ProductService to its repository.
func NewProductService(repo domain.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List(ctx context.Context) ([]domain.Product, error) {
	return s.repo.List(ctx)
}

func (s *ProductService) Get(ctx context.Context, id int) (domain.Product, error) {
	return s.repo.Get(ctx, id)
}
