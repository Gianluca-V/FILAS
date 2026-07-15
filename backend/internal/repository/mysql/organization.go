package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// OrganizationRepository implements domain.OrganizationRepository on top of
// sqlx with parameterized queries.
type OrganizationRepository struct {
	db *sqlx.DB
}

// NewOrganizationRepository wires an OrganizationRepository to an open
// connection pool.
func NewOrganizationRepository(db *sqlx.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

type organizationRow struct {
	ID          int            `db:"ID"`
	Title       string         `db:"Title"`
	Description sql.NullString `db:"Description"`
	Image       string         `db:"Image"`
}

func (r organizationRow) toDomain() domain.Organization {
	org := domain.Organization{ID: r.ID, Title: r.Title, Image: r.Image}
	if r.Description.Valid {
		desc := r.Description.String
		org.Description = &desc
	}
	return org
}

const organizationSelect = "SELECT ID, Title, Description, Image FROM organizations"

func (r *OrganizationRepository) List(ctx context.Context) ([]domain.Organization, error) {
	var rows []organizationRow
	if err := r.db.SelectContext(ctx, &rows, organizationSelect); err != nil {
		return nil, fmt.Errorf("mysql: list organizations: %w", err)
	}
	items := make([]domain.Organization, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toDomain())
	}
	return items, nil
}

func (r *OrganizationRepository) Get(ctx context.Context, id int) (domain.Organization, error) {
	var row organizationRow
	err := r.db.GetContext(ctx, &row, organizationSelect+" WHERE ID = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Organization{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Organization{}, fmt.Errorf("mysql: get organization %d: %w", id, err)
	}
	return row.toDomain(), nil
}

const organizationInsert = "INSERT INTO organizations (Title, Description, Image) VALUES (?, ?, ?)"
const organizationUpdate = "UPDATE organizations SET Title = ?, Description = ?, Image = ? WHERE ID = ?"
const organizationDelete = "DELETE FROM organizations WHERE ID = ?"

// Create inserts a new organization WITHOUT supplying an ID — the seed
// schema's AUTO_INCREMENT PRIMARY KEY assigns it. The assigned ID is read
// back via LastInsertId() and populated on the returned domain.Organization.
func (r *OrganizationRepository) Create(ctx context.Context, o domain.Organization) (domain.Organization, error) {
	res, err := r.db.ExecContext(ctx, organizationInsert, o.Title, o.Description, o.Image)
	if err != nil {
		return domain.Organization{}, fmt.Errorf("mysql: create organization: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Organization{}, fmt.Errorf("mysql: read new organization id: %w", err)
	}
	o.ID = int(id)
	return o, nil
}

// Update overwrites Title, Description, and Image for the given ID (legacy
// updateOrganization includes all three columns, unlike products' Update).
// Like legacy, it does not check row existence first.
func (r *OrganizationRepository) Update(ctx context.Context, id int, o domain.Organization) error {
	if _, err := r.db.ExecContext(ctx, organizationUpdate, o.Title, o.Description, o.Image, id); err != nil {
		return fmt.Errorf("mysql: update organization %d: %w", id, err)
	}
	return nil
}

// Delete removes the organization with the given ID, same
// no-existence-check quirk as Update.
func (r *OrganizationRepository) Delete(ctx context.Context, id int) error {
	if _, err := r.db.ExecContext(ctx, organizationDelete, id); err != nil {
		return fmt.Errorf("mysql: delete organization %d: %w", id, err)
	}
	return nil
}
