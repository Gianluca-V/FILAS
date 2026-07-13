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
	// Create inserts a new organization and returns it with ID populated
	// from the database's AUTO_INCREMENT. Unlike legacy, the caller never
	// supplies an ID.
	Create(ctx context.Context, o Organization) (Organization, error)
	// Update overwrites Title, Description, and Image for the given ID
	// (legacy updateOrganization includes all three columns, unlike
	// products' Update). Mirrors legacy: does not check row existence
	// first.
	Update(ctx context.Context, id int, o Organization) error
	// Delete removes the organization with the given ID. Same
	// no-existence-check quirk as Update.
	Delete(ctx context.Context, id int) error
}
