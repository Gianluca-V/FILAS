package rest

import "strconv"

// parseLegacyID mirrors PHP's intval() semantics for the narrow case these
// handlers need: a non-numeric path segment becomes 0 (an ID that never
// exists), which the legacy PHP code also effectively resolves to "not
// found" rather than a distinct 400. Unlike intval, this does not parse a
// numeric prefix out of a mixed string (e.g. "5abc"); that edge case is not
// exercised by any real client and is documented as a deliberate
// simplification in sdd/migrate-go-vue/apply-progress.
func parseLegacyID(raw string) int {
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return id
}
