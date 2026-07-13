package domain

import "context"

// Organization is a partner organization entry. Description is nullable in
// the schema; Title and Image are NOT NULL.
type Organization struct {
	ID          int
	Title       string
	Description *string
	Image       string
}

// OrganizationRepository is implemented by internal/repository/mysql.
type OrganizationRepository interface {
	List(ctx context.Context) ([]Organization, error)
	Get(ctx context.Context, id int) (Organization, error)
}
