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
	// Create inserts a new gallery image and returns it with ID populated
	// from the database's AUTO_INCREMENT. Unlike legacy, the caller never
	// supplies an ID.
	Create(ctx context.Context, g GalleryImage) (GalleryImage, error)
	// Update overwrites Image for the given ID. Mirrors legacy: does not
	// check row existence first.
	Update(ctx context.Context, id int, image string) error
	// Delete removes the gallery image with the given ID. Same
	// no-existence-check quirk as Update.
	Delete(ctx context.Context, id int) error
}
