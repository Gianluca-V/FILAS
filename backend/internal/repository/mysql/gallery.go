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

const gallerySelect = "SELECT ID, Image FROM gallery"

func (r *GalleryRepository) List(ctx context.Context) ([]domain.GalleryImage, error) {
	var rows []galleryRow
	if err := r.db.SelectContext(ctx, &rows, gallerySelect); err != nil {
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
	err := r.db.GetContext(ctx, &row, gallerySelect+" WHERE ID = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.GalleryImage{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.GalleryImage{}, fmt.Errorf("mysql: get gallery %d: %w", id, err)
	}
	return row.toDomain(), nil
}

const galleryInsert = "INSERT INTO gallery (Image) VALUES (?)"
const galleryUpdate = "UPDATE gallery SET Image = ? WHERE ID = ?"
const galleryDelete = "DELETE FROM gallery WHERE ID = ?"

// Create inserts a new gallery image WITHOUT supplying an ID — the seed
// schema's AUTO_INCREMENT PRIMARY KEY assigns it. The assigned ID is read
// back via LastInsertId() and populated on the returned domain.GalleryImage.
func (r *GalleryRepository) Create(ctx context.Context, g domain.GalleryImage) (domain.GalleryImage, error) {
	res, err := r.db.ExecContext(ctx, galleryInsert, g.Image)
	if err != nil {
		return domain.GalleryImage{}, fmt.Errorf("mysql: create gallery image: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.GalleryImage{}, fmt.Errorf("mysql: read new gallery image id: %w", err)
	}
	g.ID = int(id)
	return g, nil
}

// Update overwrites Image for the given ID. Like legacy, it does not check
// row existence first.
func (r *GalleryRepository) Update(ctx context.Context, id int, image string) error {
	if _, err := r.db.ExecContext(ctx, galleryUpdate, image, id); err != nil {
		return fmt.Errorf("mysql: update gallery image %d: %w", id, err)
	}
	return nil
}

// Delete removes the gallery image with the given ID, same
// no-existence-check quirk as Update.
func (r *GalleryRepository) Delete(ctx context.Context, id int) error {
	if _, err := r.db.ExecContext(ctx, galleryDelete, id); err != nil {
		return fmt.Errorf("mysql: delete gallery image %d: %w", id, err)
	}
	return nil
}
