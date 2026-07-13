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
	Image       string         `db:"Image"`
	Description sql.NullString `db:"Description"`
}

func (r productRow) toDomain() domain.Product {
	p := domain.Product{ID: r.ID, Name: r.Name, Price: r.Price, Stock: r.Stock, Image: r.Image}
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
