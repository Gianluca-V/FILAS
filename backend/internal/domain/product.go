package domain

import "context"

// Product is a sellable item. Description and Image are both nullable in
// the schema (`text DEFAULT NULL`); some legacy rows store NULL, others an
// empty string — both are preserved as distinct values through the stack,
// see dto.ProductResponse.
type Product struct {
	ID          int
	Name        string
	Price       float64
	Stock       int
	Image       *string
	Description *string
}

// ProductRepository is implemented by internal/repository/mysql.
type ProductRepository interface {
	List(ctx context.Context) ([]Product, error)
	Get(ctx context.Context, id int) (Product, error)
	// Create inserts a new product and returns it with ID populated from
	// the database's AUTO_INCREMENT (see backend/db/init/01-schema.sql).
	// Unlike legacy createProduct, the caller never supplies an ID.
	Create(ctx context.Context, p Product) (Product, error)
	// Update overwrites Name, Price, Stock, and Image for the given ID.
	// Description is DELIBERATELY excluded: legacy updateProduct's SQL
	// never included the Description column (characterized live, see
	// backend/docs/legacy-quirks.md) — a PUT can never change a product's
	// Description, only its value from Create persists. Preserved here for
	// contract fidelity, not an oversight.
	Update(ctx context.Context, id int, p Product) error
	// Delete removes the product with the given ID. Mirrors legacy
	// deleteProduct: does not check row existence first.
	Delete(ctx context.Context, id int) error
}
