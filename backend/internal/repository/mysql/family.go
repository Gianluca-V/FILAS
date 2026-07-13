package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// FamilyRepository implements domain.FamilyRepository on top of sqlx with
// parameterized queries.
type FamilyRepository struct {
	db *sqlx.DB
}

// NewFamilyRepository wires a FamilyRepository to an open connection pool.
func NewFamilyRepository(db *sqlx.DB) *FamilyRepository {
	return &FamilyRepository{db: db}
}

type familyRow struct {
	ID       int            `db:"ID"`
	Image    sql.NullString `db:"Image"`
	Body     string         `db:"Body"`
	Category string         `db:"Category"`
}

func (r familyRow) toDomain() domain.FamilyItem {
	item := domain.FamilyItem{ID: r.ID, Body: r.Body, Category: r.Category}
	if r.Image.Valid {
		image := r.Image.String
		item.Image = &image
	}
	return item
}

const familySelect = "SELECT ID, Image, Body, Category FROM family"

func (r *FamilyRepository) List(ctx context.Context) ([]domain.FamilyItem, error) {
	var rows []familyRow
	if err := r.db.SelectContext(ctx, &rows, familySelect); err != nil {
		return nil, fmt.Errorf("mysql: list family: %w", err)
	}
	items := make([]domain.FamilyItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toDomain())
	}
	return items, nil
}

func (r *FamilyRepository) Get(ctx context.Context, id int) (domain.FamilyItem, error) {
	var row familyRow
	err := r.db.GetContext(ctx, &row, familySelect+" WHERE ID = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.FamilyItem{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.FamilyItem{}, fmt.Errorf("mysql: get family %d: %w", id, err)
	}
	return row.toDomain(), nil
}
