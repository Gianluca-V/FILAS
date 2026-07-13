package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// GalleryRepository implements domain.GalleryRepository on top of sqlx with
// parameterized queries.
type GalleryRepository struct {
	db *sqlx.DB
}

// NewGalleryRepository wires a GalleryRepository to an open connection pool.
func NewGalleryRepository(db *sqlx.DB) *GalleryRepository {
	return &GalleryRepository{db: db}
}

type galleryRow struct {
	ID    int    `db:"ID"`
	Image string `db:"Image"`
}

func (r galleryRow) toDomain() domain.GalleryImage {
	return domain.GalleryImage{ID: r.ID, Image: r.Image}
}

func (r *GalleryRepository) List(ctx context.Context) ([]domain.GalleryImage, error) {
	var rows []galleryRow
	if err := r.db.SelectContext(ctx, &rows, "SELECT ID, Image FROM gallery"); err != nil {
		return nil, fmt.Errorf("mysql: list gallery: %w", err)
	}
	images := make([]domain.GalleryImage, 0, len(rows))
	for _, row := range rows {
		images = append(images, row.toDomain())
	}
	return images, nil
}

func (r *GalleryRepository) Get(ctx context.Context, id int) (domain.GalleryImage, error) {
	var row galleryRow
	err := r.db.GetContext(ctx, &row, "SELECT ID, Image FROM gallery WHERE ID = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.GalleryImage{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.GalleryImage{}, fmt.Errorf("mysql: get gallery %d: %w", id, err)
	}
	return row.toDomain(), nil
}
