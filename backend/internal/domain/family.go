package domain

import "context"

// FamilyItem is a single "family" workshop entry. Image is nullable in the
// schema; Body and Category are NOT NULL. Category enum validation
// ("Centro de dia" / "Taller protegido") is enforced on writes by
// usecase.FamilyService via ValidFamilyCategory below (PR4, task 3.1).
type FamilyItem struct {
	ID       int
	Image    *string
	Body     string
	Category string
}

// Valid FamilyItem.Category values, matching legacy family.php's inline
// string comparison (`$category === "Centro de dia" || $category ===
// "Taller protegido"`) exactly, including the deliberately unaccented
// "dia" (not "día").
const (
	FamilyCategoryCentroDeDia     = "Centro de dia"
	FamilyCategoryTallerProtegido = "Taller protegido"
)

// ValidFamilyCategory reports whether category is one of the two enum
// values legacy accepts. Used by usecase.FamilyService.Create/Update to
// reject any other value with domain.ErrValidation before persisting.
func ValidFamilyCategory(category string) bool {
	return category == FamilyCategoryCentroDeDia || category == FamilyCategoryTallerProtegido
}

// FamilyRepository is implemented by internal/repository/mysql.
type FamilyRepository interface {
	List(ctx context.Context) ([]FamilyItem, error)
	Get(ctx context.Context, id int) (FamilyItem, error)
	// Create inserts a new family item and returns it with ID populated
	// from the database's AUTO_INCREMENT. Unlike legacy, the caller never
	// supplies an ID.
	Create(ctx context.Context, f FamilyItem) (FamilyItem, error)
	// Update overwrites Body, Category, and Image for the given ID (legacy
	// updateFamily includes all three columns, unlike products' Update).
	// Mirrors legacy: does not check row existence first.
	Update(ctx context.Context, id int, f FamilyItem) error
	// Delete removes the family item with the given ID. Same
	// no-existence-check quirk as Update.
	Delete(ctx context.Context, id int) error
}
