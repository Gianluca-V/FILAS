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
}
