package domain

import "context"

// FamilyItem is a single "family" workshop entry. Image is nullable in the
// schema; Body and Category are NOT NULL. Category enum validation
// ("Centro de dia" / "Taller protegido") is enforced on writes, added in a
// later PR (Phase 3) — this read-only slice only carries the raw value.
type FamilyItem struct {
	ID       int
	Image    *string
	Body     string
	Category string
}

// FamilyRepository is implemented by internal/repository/mysql.
type FamilyRepository interface {
	List(ctx context.Context) ([]FamilyItem, error)
	Get(ctx context.Context, id int) (FamilyItem, error)
}
