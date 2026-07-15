package domain

import "context"

// NewsItem is a single news post. Body and Image are nullable in the schema.
type NewsItem struct {
	ID    int
	Title string
	Body  *string
	Image *string
}

// NewsRepository is implemented by internal/repository/mysql.
type NewsRepository interface {
	List(ctx context.Context) ([]NewsItem, error)
	Get(ctx context.Context, id int) (NewsItem, error)
	// Create inserts a new news item and returns it with ID populated from
	// the database's AUTO_INCREMENT. Unlike legacy, the caller never
	// supplies an ID.
	Create(ctx context.Context, n NewsItem) (NewsItem, error)
	// Update overwrites Title, Body, and Image for the given ID. Mirrors
	// legacy: does not check row existence first — a missing/non-numeric ID
	// is a no-op that still reports success (see
	// backend/docs/legacy-quirks.md §10, same parity already locked for
	// products/family in PR4).
	Update(ctx context.Context, id int, n NewsItem) error
	// Delete removes the news item with the given ID. Same
	// no-existence-check quirk as Update.
	Delete(ctx context.Context, id int) error
}
