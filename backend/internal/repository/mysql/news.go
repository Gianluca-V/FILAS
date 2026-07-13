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

const newsInsert = "INSERT INTO news (Title, Body, Image) VALUES (?, ?, ?)"
const newsUpdate = "UPDATE news SET Title = ?, Body = ?, Image = ? WHERE ID = ?"
const newsDelete = "DELETE FROM news WHERE ID = ?"

// Create inserts a new news item WITHOUT supplying an ID — the seed
// schema's AUTO_INCREMENT PRIMARY KEY assigns it. The assigned ID is read
// back via LastInsertId() and populated on the returned domain.NewsItem.
func (r *NewsRepository) Create(ctx context.Context, n domain.NewsItem) (domain.NewsItem, error) {
	res, err := r.db.ExecContext(ctx, newsInsert, n.Title, n.Body, n.Image)
	if err != nil {
		return domain.NewsItem{}, fmt.Errorf("mysql: create news item: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.NewsItem{}, fmt.Errorf("mysql: read new news item id: %w", err)
	}
	n.ID = int(id)
	return n, nil
}

// Update overwrites Title, Body, and Image for the given ID. Like legacy
// (mysqli::query() returns TRUE even when zero rows match — see
// backend/docs/legacy-quirks.md §10), it does not check row existence
// first — a nonexistent ID is a no-op UPDATE that still reports success.
func (r *NewsRepository) Update(ctx context.Context, id int, n domain.NewsItem) error {
	if _, err := r.db.ExecContext(ctx, newsUpdate, n.Title, n.Body, n.Image, id); err != nil {
		return fmt.Errorf("mysql: update news item %d: %w", id, err)
	}
	return nil
}

// Delete removes the news item with the given ID, same
// no-existence-check quirk as Update.
func (r *NewsRepository) Delete(ctx context.Context, id int) error {
	if _, err := r.db.ExecContext(ctx, newsDelete, id); err != nil {
		return fmt.Errorf("mysql: delete news item %d: %w", id, err)
	}
	return nil
}
