package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// ProductRepository implements domain.ProductRepository on top of sqlx with
// parameterized queries.
type ProductRepository struct {
	db *sqlx.DB
}

// NewProductRepository wires a ProductRepository to an open connection pool.
func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

type productRow struct {
	ID          int            `db:"ID"`
	Name        string         `db:"Name"`
	Price       float64        `db:"Price"`
	Stock       int            `db:"Stock"`
	Image       sql.NullString `db:"Image"`
	Description sql.NullString `db:"Description"`
}

func (r productRow) toDomain() domain.Product {
	p := domain.Product{ID: r.ID, Name: r.Name, Price: r.Price, Stock: r.Stock}
	if r.Image.Valid {
		image := r.Image.String
		p.Image = &image
	}
	if r.Description.Valid {
		desc := r.Description.String
		p.Description = &desc
	}
	return p
}

const productSelect = "SELECT ID, Name, Price, Stock, Image, Description FROM products"

func (r *ProductRepository) List(ctx context.Context) ([]domain.Product, error) {
	var rows []productRow
	if err := r.db.SelectContext(ctx, &rows, productSelect); err != nil {
		return nil, fmt.Errorf("mysql: list products: %w", err)
	}
	products := make([]domain.Product, 0, len(rows))
	for _, row := range rows {
		products = append(products, row.toDomain())
	}
	return products, nil
}

func (r *ProductRepository) Get(ctx context.Context, id int) (domain.Product, error) {
	var row productRow
	err := r.db.GetContext(ctx, &row, productSelect+" WHERE ID = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Product{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Product{}, fmt.Errorf("mysql: get product %d: %w", id, err)
	}
	return row.toDomain(), nil
}

const productInsert = "INSERT INTO products (Name, Price, Stock, Image, Description) VALUES (?, ?, ?, ?, ?)"

// productUpdate deliberately excludes Description — see
// domain.ProductRepository.Update's doc comment for why this mirrors
// legacy's updateProduct SQL exactly.
const productUpdate = "UPDATE products SET Name = ?, Price = ?, Stock = ?, Image = ? WHERE ID = ?"

const productDelete = "DELETE FROM products WHERE ID = ?"

// Create inserts a new product WITHOUT supplying an ID — the seed schema's
// AUTO_INCREMENT PRIMARY KEY assigns it. The assigned ID is read back via
// LastInsertId() and populated on the returned domain.Product.
func (r *ProductRepository) Create(ctx context.Context, p domain.Product) (domain.Product, error) {
	res, err := r.db.ExecContext(ctx, productInsert, p.Name, p.Price, p.Stock, p.Image, p.Description)
	if err != nil {
		return domain.Product{}, fmt.Errorf("mysql: create product: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Product{}, fmt.Errorf("mysql: read new product id: %w", err)
	}
	p.ID = int(id)
	return p, nil
}

// Update overwrites Name, Price, Stock, and Image for the given ID.
// Description is NOT touched (see productUpdate). Like legacy, it does not
// check row existence first — a nonexistent ID is a no-op UPDATE that
// still reports success.
func (r *ProductRepository) Update(ctx context.Context, id int, p domain.Product) error {
	if _, err := r.db.ExecContext(ctx, productUpdate, p.Name, p.Price, p.Stock, p.Image, id); err != nil {
		return fmt.Errorf("mysql: update product %d: %w", id, err)
	}
	return nil
}

// Delete removes the product with the given ID, same no-existence-check
// quirk as Update.
func (r *ProductRepository) Delete(ctx context.Context, id int) error {
	if _, err := r.db.ExecContext(ctx, productDelete, id); err != nil {
		return fmt.Errorf("mysql: delete product %d: %w", id, err)
	}
	return nil
}
