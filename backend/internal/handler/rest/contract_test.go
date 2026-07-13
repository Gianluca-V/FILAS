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

// Contract/characterization tests: lock the Go handlers' JSON output against
// golden fixtures captured LIVE from the legacy PHP (FilasServer/*.php)
// running against a fresh copy of the seeded Docker DB (see
// sdd/migrate-go-vue/apply-progress obs #31 for the exact capture
// procedure and raw curl output). Fixtures live in backend/testdata/contract.
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
		ProductService: fakeProductService{
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
		ProductService: fakeProductService{products: []domain.Product{}},
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
		NewsService: fakeNewsService{
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
		NewsService: fakeNewsService{items: []domain.NewsItem{}},
	})
	assertJSONMatchesFixture(t, r, http.MethodGet, "/api/news", http.StatusNotFound, "news_not_found.json")
}

func TestContract_Gallery(t *testing.T) {
	r := rest.NewRouter(rest.RouterDeps{
		HealthDB: fakePinger{},
		GalleryService: fakeGalleryService{
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
		FamilyService: fakeFamilyService{
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
		OrganizationService: fakeOrganizationService{
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
// synthetic admin (see sdd/migrate-go-vue/apply-progress for the exact
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
