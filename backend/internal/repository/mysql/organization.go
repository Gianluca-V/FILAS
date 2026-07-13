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
