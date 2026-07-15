package rest_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gianluca-v/filas-backend/internal/auth"
	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest"
)

// orderPtr is a small helper for building a *time.Time inline in test data.
func orderTimePtr(t time.Time) *time.Time { return &t }

// Contract/characterization tests: lock the Go handlers' JSON output against
// golden fixtures captured LIVE from the legacy PHP (FilasServer/*.php)
// running against a fresh copy of the seeded Docker DB (see
// backend/docs/legacy-quirks.md for the capture methodology). Fixtures
// live in backend/testdata/contract.
//
// One deliberate, documented divergence: fixture text uses correctly
// decoded UTF-8 (e.g. plain URLs, no mojibake), whereas the raw legacy
// output showed double-encoded UTF-8 for some accented characters. That
// corruption comes from the legacy PHP code never calling
// mysqli::set_charset("utf8mb4"), is environment/driver-dependent (not a
// stable, testable contract), and Go's driver decodes correctly by default.
// Per the design's own precedent (reproduce the SHAPE, not a fragility bug
// — see ADR on the orders GROUP_CONCAT quirk), this is not reproduced.

func loadFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "..", "testdata", "contract", name)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}
	return strings.TrimRight(string(body), "\n")
}

func TestContract_Products(t *testing.T) {
	desc := ""
	image1 := "assets/default-img.png"
	image18 := `assets\ajo.jpg`
	r := rest.NewRouter(rest.RouterDeps{
		HealthDB: fakePinger{},
		ProductService: &fakeProductService{
			products: []domain.Product{
				{ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112, Image: &image1, Description: &desc},
				{ID: 18, Name: "Encurtido de ajo", Price: 700, Stock: 1, Image: &image18, Description: nil},
			},
			byID: map[int]domain.Product{
				1: {ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112, Image: &image1, Description: &desc},
			},
		},
	})

	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/products", http.StatusOK, "products_list.json")
	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/products/1", http.StatusOK, "products_get.json")
	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/products/999999", http.StatusNotFound, "products_not_found.json")
}

func TestContract_Products_EmptyListIsOKNotNotFound(t *testing.T) {
	r := rest.NewRouter(rest.RouterDeps{
		HealthDB:       fakePinger{},
		ProductService: &fakeProductService{products: []domain.Product{}},
	})
	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/products", http.StatusOK, "products_empty_list.json")
}

func TestContract_News(t *testing.T) {
	// Triangulation (fix #6 from the gate review): seed 2 items so list vs
	// get-by-id actually diverge, matching the family/organizations
	// convention below instead of list and get being byte-identical.
	body := "Lorem ipsum dolor sit amet"
	image := "https://example.com/1.jpg"
	r := rest.NewRouter(rest.RouterDeps{
		HealthDB: fakePinger{},
		NewsService: &fakeNewsService{
			items: []domain.NewsItem{
				{ID: 1, Title: "Noticia de prueba 1", Body: &body, Image: &image},
				{ID: 2, Title: "Noticia de prueba 2", Body: nil, Image: nil},
			},
			byID: map[int]domain.NewsItem{1: {ID: 1, Title: "Noticia de prueba 1", Body: &body, Image: &image}},
		},
	})

	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/news", http.StatusOK, "news_list.json")
	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/news/1", http.StatusOK, "news_get.json")
	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/news/999999", http.StatusNotFound, "news_not_found.json")
}

func TestContract_News_EmptyListIsNotFound(t *testing.T) {
	r := rest.NewRouter(rest.RouterDeps{
		HealthDB:    fakePinger{},
		NewsService: &fakeNewsService{items: []domain.NewsItem{}},
	})
	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/news", http.StatusNotFound, "news_not_found.json")
}

func TestContract_Gallery(t *testing.T) {
	r := rest.NewRouter(rest.RouterDeps{
		HealthDB: fakePinger{},
		GalleryService: &fakeGalleryService{
			images: []domain.GalleryImage{{ID: 1, Image: "assets/galeria-1.jpg"}, {ID: 2, Image: "assets/galeria-2.jpg"}},
			byID:   map[int]domain.GalleryImage{1: {ID: 1, Image: "assets/galeria-1.jpg"}},
		},
	})

	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/gallery", http.StatusOK, "gallery_list.json")
	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/gallery/1", http.StatusOK, "gallery_get.json")
	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/gallery/999999", http.StatusNotFound, "gallery_not_found.json")
}

func TestContract_Family(t *testing.T) {
	image := "https://example.com/3.jpg"
	r := rest.NewRouter(rest.RouterDeps{
		HealthDB: fakePinger{},
		FamilyService: &fakeFamilyService{
			items: []domain.FamilyItem{
				{ID: 3, Image: &image, Body: "aqui va la descripcion 3", Category: "Taller protegido"},
				{ID: 6, Image: nil, Body: "Taller de musica", Category: "Centro de dia"},
			},
			byID: map[int]domain.FamilyItem{3: {ID: 3, Image: &image, Body: "aqui va la descripcion 3", Category: "Taller protegido"}},
		},
	})

	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/family", http.StatusOK, "family_list.json")
	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/family/3", http.StatusOK, "family_get.json")
	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/family/999999", http.StatusNotFound, "family_not_found.json")
}

func TestContract_Organizations(t *testing.T) {
	desc1 := "Descripcion de prueba 1"
	r := rest.NewRouter(rest.RouterDeps{
		HealthDB: fakePinger{},
		OrganizationService: &fakeOrganizationService{
			items: []domain.Organization{
				{ID: 1, Title: "organizacions de prueba 1", Description: &desc1, Image: "https://example.com/1.jpg"},
				{ID: 2, Title: "sin descripcion", Description: nil, Image: "https://example.com/2.jpg"},
			},
			byID: map[int]domain.Organization{1: {ID: 1, Title: "organizacions de prueba 1", Description: &desc1, Image: "https://example.com/1.jpg"}},
		},
	})

	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/organizations", http.StatusOK, "organizations_list.json")
	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/organizations/1", http.StatusOK, "organizations_get.json")
	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/organizations/999999", http.StatusNotFound, "organizations_not_found.json")
}

// TestContract_Admins locks the admins list/get/not-found shapes against
// fixtures captured LIVE from FilasServer/admins.php against the seeded
// synthetic admin (see backend/docs/legacy-quirks.md §1 for the exact
// characterization procedure). Unlike the fixtures above, these are NOT
// byte-for-byte copies of the legacy output: legacy leaked password/salt
// via SELECT * on both routes (confirmed live), which the Go DTO
// deliberately excludes (see dto.AdminResponse doc comment) — the fixture
// files record the FIXED shape. Login/create/update/delete are message-only
// (or carry a dynamically-signed token) and are exact-string-asserted in
// admin_test.go instead of fixture files, since a signed JWT is not a
// stable byte-for-byte fixture value.
func TestContract_Admins(t *testing.T) {
	jwtSvc := auth.NewJWTService("contract-test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	r := rest.NewRouter(rest.RouterDeps{
		HealthDB: fakePinger{},
		AdminService: &fakeAdminService{
			admins: []domain.Admin{{ID: 1543, Username: "FilasAdmin", Password: "leaked-in-legacy", Salt: "leaked-in-legacy"}},
			byID:   map[int]domain.Admin{1543: {ID: 1543, Username: "FilasAdmin", Password: "leaked-in-legacy", Salt: "leaked-in-legacy"}},
		},
		AuthService: &fakeAuthService{},
		JWTService:  jwtSvc,
	})

	assertAuthedJSONMatchesFixture(t, r, http.MethodGet, "/api/admins", token, http.StatusOK, "admins_list.json")
	assertAuthedJSONMatchesFixture(t, r, http.MethodGet, "/api/admins/1543", token, http.StatusOK, "admins_get.json")
	assertAuthedJSONMatchesFixture(t, r, http.MethodGet, "/api/admins/999999", token, http.StatusNotFound, "admins_not_found.json")
}

// TestContract_Orders locks the GET /api/orders[/:id] response shape
// against fixtures captured LIVE from a scratch copy of FilasServer/
// orders.php against the seeded Docker db (see
// backend/docs/legacy-quirks.md §15 for the full capture methodology and
// the CRITICAL characterization: `products[].productPrice` is the
// product's CURRENT price via a live JOIN, NOT the frozen per-line price
// actually billed). GET is auth-gated (router.go PR7 wiring), unlike the
// PR2 public-read resources above.
func TestContract_Orders(t *testing.T) {
	jwtSvc := auth.NewJWTService("contract-test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}

	start10 := time.Date(2023, 11, 18, 20, 9, 58, 0, time.UTC)
	finish10 := time.Date(2023, 11, 18, 20, 41, 43, 0, time.UTC)
	start27 := time.Date(2023, 11, 20, 18, 30, 2, 0, time.UTC)

	order10 := domain.Order{
		ID: 10, Total: 3600, State: domain.OrderStateCanceled,
		StartDate: start10, FinishDate: orderTimePtr(finish10),
		Products: []domain.OrderProduct{
			{ProductID: 1, Quantity: 2, ProductName: "Mermelada de pera", ProductPrice: 600},
			{ProductID: 3, Quantity: 3, ProductName: "Mermelada de tomate", ProductPrice: 600},
		},
	}
	order27 := domain.Order{
		ID: 27, Total: 13100, State: domain.OrderStatePending,
		StartDate: start27, Name: "Turco agustin", Phone: "32142432546",
		Products: []domain.OrderProduct{
			{ProductID: 23, Quantity: 1, ProductName: "Galletas (6 u.)", ProductPrice: 850},
		},
	}

	r := rest.NewRouter(rest.RouterDeps{
		HealthDB: fakePinger{},
		OrderService: &fakeOrderService{
			orders: []domain.Order{order10, order27},
			byID:   map[int]domain.Order{27: order27},
		},
		AuthService: &fakeAuthService{},
		JWTService:  jwtSvc,
	})

	assertAuthedJSONMatchesFixture(t, r, http.MethodGet, "/api/orders", token, http.StatusOK, "orders_list.json")
	assertAuthedJSONMatchesFixture(t, r, http.MethodGet, "/api/orders/27", token, http.StatusOK, "orders_get.json")
	assertAuthedJSONMatchesFixture(t, r, http.MethodGet, "/api/orders/999999", token, http.StatusNotFound, "orders_not_found.json")
}

// TestContract_Orders_EmptyListIsOKNotNotFound locks the orders-specific
// empty-list quirk: unlike news/gallery/family/organizations (404 on
// empty), orders' underlying legacy query never returns `false` on zero
// rows, so the "$result === false" 404 branch never fires — same
// reasoning as products'/admins' 200-empty-list behavior (see §8,
// live-confirmed by truncating orders/orderproduct — see §15).
func TestContract_Orders_EmptyListIsOKNotNotFound(t *testing.T) {
	jwtSvc := auth.NewJWTService("contract-test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	r := rest.NewRouter(rest.RouterDeps{
		HealthDB:     fakePinger{},
		OrderService: &fakeOrderService{orders: []domain.Order{}},
		AuthService:  &fakeAuthService{},
		JWTService:   jwtSvc,
	})
	assertAuthedJSONMatchesFixture(t, r, http.MethodGet, "/api/orders", token, http.StatusOK, "orders_empty_list.json")
}

// TestContract_Orders_ProductNameWithQuoteAndCommaDoesNotBreakTheResponse
// is the design's own flagged risk (§9 "Architectural risks"), turned into
// a regression lock: legacy's manual CONCAT-built JSON string breaks on a
// product Name containing `"`/`,` (json_decode silently returns null,
// live-confirmed — see backend/docs/legacy-quirks.md §15), but the Go DTO
// builds the array structurally via encoding/json, so it can NEVER
// corrupt regardless of the product name's content — proving the SHAPE is
// reproduced, not the bug's fragility.
func TestContract_Orders_ProductNameWithQuoteAndCommaDoesNotBreakTheResponse(t *testing.T) {
	jwtSvc := auth.NewJWTService("contract-test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	start := time.Date(2026, 7, 13, 22, 43, 51, 0, time.UTC)
	order35 := domain.Order{
		ID: 35, Total: 100, State: domain.OrderStatePending,
		StartDate: start, Name: "Quote Test", Phone: "555",
		Products: []domain.OrderProduct{
			{ProductID: 25, Quantity: 1, ProductName: `Dulce "Especial", con comas`, ProductPrice: 100},
		},
	}
	r := rest.NewRouter(rest.RouterDeps{
		HealthDB:     fakePinger{},
		OrderService: &fakeOrderService{byID: map[int]domain.Order{35: order35}},
		AuthService:  &fakeAuthService{},
		JWTService:   jwtSvc,
	})
	assertAuthedJSONMatchesFixture(t, r, http.MethodGet, "/api/orders/35", token, http.StatusOK, "orders_get_quote_robustness.json")
}

func assertAuthedJSONMatchesFixture(t *testing.T, r http.Handler, method, path, token string, wantStatus int, fixture string) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != wantStatus {
		t.Errorf("%s %s -> status = %d, want %d, body=%s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	want := loadFixture(t, fixture)
	got := strings.TrimRight(rec.Body.String(), "\n")
	if got != want {
		t.Errorf("%s %s -> body = %s, want (from %s) %s", method, path, got, fixture, want)
	}
}

func assertJSONMatchesFixture(t *testing.T, r http.Handler, method, path string, wantStatus int, fixture string) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != wantStatus {
		t.Errorf("%s %s -> status = %d, want %d, body=%s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	want := loadFixture(t, fixture)
	got := strings.TrimRight(rec.Body.String(), "\n")
	if got != want {
		t.Errorf("%s %s -> body = %s, want (from %s) %s", method, path, got, fixture, want)
	}
}
