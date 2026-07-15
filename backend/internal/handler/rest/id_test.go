package rest

import "testing"

// TestParseLegacyID exercises parseLegacyID directly (pure function, no
// mocks needed) since it previously had no dedicated coverage — only its
// effect was indirectly exercised through a subset of the handler
// non-numeric-ID tests.
func TestParseLegacyID(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want int
	}{
		{name: "non-numeric string becomes 0", raw: "abc", want: 0},
		{
			name: "numeric-prefix string becomes 0 (documented intval divergence)",
			// PHP's intval("5abc") == 5 (parses a leading numeric prefix out
			// of a mixed string). parseLegacyID does NOT replicate that —
			// it treats any non-fully-numeric string as a parse failure and
			// returns 0. This is a deliberate, documented simplification
			// (see the parseLegacyID doc comment and
			// backend/docs/legacy-quirks.md §9): no real client sends mixed
			// numeric/alpha IDs.
			raw:  "5abc",
			want: 0,
		},
		{name: "zero string stays zero", raw: "0", want: 0},
		{name: "negative numeric string parses through", raw: "-1", want: -1},
		{name: "positive numeric string parses through", raw: "42", want: 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLegacyID(tt.raw)
			if got != tt.want {
				t.Errorf("parseLegacyID(%q) = %d, want %d", tt.raw, got, tt.want)
			}
		})
	}
}
