package usecase

import (
	"context"
	"fmt"

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

// Create validates Name is present, then persists a new product. Unlike
// legacy createProduct (which has NO validation whatsoever — real_escape_string
// happily persists an empty Name), this rejects an empty Name before it
// reaches the repository: a nameless product is a broken storefront entry,
// the same "never persist a zero-value where a real value is required"
// principle applied to admins' password field in PR3 (gate #36 blocker
// fix), generalized here as new, deliberate hardening — not a legacy
// contract requirement. Price and Stock are NOT required to be non-zero:
// both are legitimate values (a free promotional item, an out-of-stock
// listing), unlike Name which is the product's identity.
func (s *ProductService) Create(ctx context.Context, p domain.Product) (domain.Product, error) {
	if p.Name == "" {
		return domain.Product{}, fmt.Errorf("name is required: %w", domain.ErrValidation)
	}
	created, err := s.repo.Create(ctx, p)
	if err != nil {
		return domain.Product{}, fmt.Errorf("usecase: create product: %w", err)
	}
	return created, nil
}

// Update validates Name is present (same rationale as Create), then
// overwrites Name/Price/Stock/Image for the given ID. Description is
// deliberately never touched (see domain.ProductRepository.Update).
func (s *ProductService) Update(ctx context.Context, id int, p domain.Product) error {
	if p.Name == "" {
		return fmt.Errorf("name is required: %w", domain.ErrValidation)
	}
	if err := s.repo.Update(ctx, id, p); err != nil {
		return fmt.Errorf("usecase: update product %d: %w", id, err)
	}
	return nil
}

func (s *ProductService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("usecase: delete product %d: %w", id, err)
	}
	return nil
}
