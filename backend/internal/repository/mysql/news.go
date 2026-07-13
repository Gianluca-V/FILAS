package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// NewsRepository implements domain.NewsRepository on top of sqlx with
// parameterized queries.
type NewsRepository struct {
	db *sqlx.DB
}

// NewNewsRepository wires a NewsRepository to an open connection pool.
func NewNewsRepository(db *sqlx.DB) *NewsRepository {
	return &NewsRepository{db: db}
}

type newsRow struct {
	ID    int            `db:"ID"`
	Title string         `db:"Title"`
	Body  sql.NullString `db:"Body"`
	Image sql.NullString `db:"Image"`
}

func (r newsRow) toDomain() domain.NewsItem {
	item := domain.NewsItem{ID: r.ID, Title: r.Title}
	if r.Body.Valid {
		body := r.Body.String
		item.Body = &body
	}
	if r.Image.Valid {
		image := r.Image.String
		item.Image = &image
	}
	return item
}

const newsSelect = "SELECT ID, Title, Body, Image FROM news"

func (r *NewsRepository) List(ctx context.Context) ([]domain.NewsItem, error) {
	var rows []newsRow
	if err := r.db.SelectContext(ctx, &rows, newsSelect); err != nil {
		return nil, fmt.Errorf("mysql: list news: %w", err)
	}
	items := make([]domain.NewsItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toDomain())
	}
	return items, nil
}

func (r *NewsRepository) Get(ctx context.Context, id int) (domain.NewsItem, error) {
	var row newsRow
	err := r.db.GetContext(ctx, &row, newsSelect+" WHERE ID = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NewsItem{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.NewsItem{}, fmt.Errorf("mysql: get news %d: %w", id, err)
	}
	return row.toDomain(), nil
}
