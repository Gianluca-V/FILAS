package domain

import "context"

// GalleryImage is a single gallery photo. Image is NOT NULL in the schema,
// so it never needs nullable handling (unlike most other text columns).
type GalleryImage struct {
	ID    int
	Image string
}

// GalleryRepository is implemented by internal/repository/mysql.
type GalleryRepository interface {
	List(ctx context.Context) ([]GalleryImage, error)
	Get(ctx context.Context, id int) (GalleryImage, error)
}
