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

const familyInsert = "INSERT INTO family (Body, Category, Image) VALUES (?, ?, ?)"
const familyUpdate = "UPDATE family SET Body = ?, Category = ?, Image = ? WHERE ID = ?"
const familyDelete = "DELETE FROM family WHERE ID = ?"

// Create inserts a new family item WITHOUT supplying an ID — the seed
// schema's AUTO_INCREMENT PRIMARY KEY assigns it. The assigned ID is read
// back via LastInsertId() and populated on the returned domain.FamilyItem.
func (r *FamilyRepository) Create(ctx context.Context, f domain.FamilyItem) (domain.FamilyItem, error) {
	res, err := r.db.ExecContext(ctx, familyInsert, f.Body, f.Category, f.Image)
	if err != nil {
		return domain.FamilyItem{}, fmt.Errorf("mysql: create family item: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.FamilyItem{}, fmt.Errorf("mysql: read new family item id: %w", err)
	}
	f.ID = int(id)
	return f, nil
}

// Update overwrites Body, Category, and Image for the given ID (legacy
// updateFamily includes all three columns, unlike products' Update). Like
// legacy, it does not check row existence first.
func (r *FamilyRepository) Update(ctx context.Context, id int, f domain.FamilyItem) error {
	if _, err := r.db.ExecContext(ctx, familyUpdate, f.Body, f.Category, f.Image, id); err != nil {
		return fmt.Errorf("mysql: update family item %d: %w", id, err)
	}
	return nil
}

// Delete removes the family item with the given ID, same
// no-existence-check quirk as Update.
func (r *FamilyRepository) Delete(ctx context.Context, id int) error {
	if _, err := r.db.ExecContext(ctx, familyDelete, id); err != nil {
		return fmt.Errorf("mysql: delete family item %d: %w", id, err)
	}
	return nil
}
