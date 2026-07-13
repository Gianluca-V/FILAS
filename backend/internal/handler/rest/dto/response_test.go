package dto_test

import (
	"encoding/json"
	"testing"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest/dto"
)

// TestNewProductResponse_EncodesNumericFieldsAsJSONStrings locks the #1
// legacy contract quirk: mysqli (via this codebase's PHP driver
// configuration) returns numeric columns as PHP strings, so the legacy JSON
// has "ID":"1" not "ID":1. The Go DTO must reproduce this exactly or the
// legacy frontend's string-based rendering breaks silently. Characterized
// live against FilasServer/products.php + the seeded DB (obs #29/#31).
func TestNewProductResponse_EncodesNumericFieldsAsJSONStrings(t *testing.T) {
	desc := ""
	p := domain.Product{ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112, Image: "assets/default-img.png", Description: &desc}

	resp := dto.NewProductResponse(p)
	body, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if _, ok := raw["ID"].(string); !ok {
		t.Errorf(`ID = %v (%T), want a JSON string like "1"`, raw["ID"], raw["ID"])
	}
	if raw["ID"] != "1" {
		t.Errorf("ID = %v, want %q", raw["ID"], "1")
	}
	if raw["Price"] != "600" {
		t.Errorf("Price = %v, want %q", raw["Price"], "600")
	}
	if raw["Stock"] != "112" {
		t.Errorf("Stock = %v, want %q", raw["Stock"], "112")
	}
}

func TestNewProductResponse_FormatsFractionalPriceWithoutTrailingZeros(t *testing.T) {
	// Triangulation: no fractional price exists in the legacy seed data, but
	// MySQL's double->text formatting drops unnecessary trailing zeros
	// (e.g. 199.5, not 199.50 or 199.5000000001). strconv.FormatFloat with
	// precision -1 (shortest round-trip) mirrors that.
	p := domain.Product{ID: 2, Name: "Fractional", Price: 199.5, Stock: 1, Image: ""}

	resp := dto.NewProductResponse(p)
	if resp.Price != "199.5" {
		t.Errorf("Price = %q, want %q", resp.Price, "199.5")
	}
}

func TestNewProductResponse_PreservesNullDescriptionAsJSONNull(t *testing.T) {
	p := domain.Product{ID: 18, Name: "Encurtido de ajo", Price: 700, Stock: 1, Image: "assets/ajo.jpg", Description: nil}

	resp := dto.NewProductResponse(p)
	body, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if got := string(body); !jsonContains(got, `"Description":null`) {
		t.Errorf("body = %s, want it to contain %q", got, `"Description":null`)
	}
}

func jsonContains(body, substr string) bool {
	return len(body) >= len(substr) && (func() bool {
		for i := 0; i+len(substr) <= len(body); i++ {
			if body[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}
