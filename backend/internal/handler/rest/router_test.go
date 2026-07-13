package rest_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewRouter_HealthRouteWiredAndReturnsOK(t *testing.T) {
	r := rest.NewRouter(rest.RouterDeps{
		CORSAllowedOrigins: []string{"http://localhost:5173"},
		HealthDB:           fakePinger{err: nil},
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestNewRouter_AppliesCORSHeaders(t *testing.T) {
	r := rest.NewRouter(rest.RouterDeps{
		CORSAllowedOrigins: []string{"http://localhost:5173"},
		HealthDB:           fakePinger{err: nil},
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:5173")
	}
}

func TestNewRouter_LogsRequests(t *testing.T) {
	// fix #6 from the PR1 gate review: request logging middleware must be
	// wired into every request the router handles, not just an isolated
	// unit test of the middleware in a vacuum.
	r := rest.NewRouter(rest.RouterDeps{
		CORSAllowedOrigins: nil,
		HealthDB:           fakePinger{err: nil},
	})

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	logged := buf.String()
	if !strings.Contains(logged, "/health") {
		t.Errorf("log output = %q, want it to contain the request path %q", logged, "/health")
	}
	if !strings.Contains(logged, "200") {
		t.Errorf("log output = %q, want it to contain the response status %q", logged, "200")
	}
}

func TestNewRouter_UnknownRouteReturns404(t *testing.T) {
	r := rest.NewRouter(rest.RouterDeps{
		CORSAllowedOrigins: nil,
		HealthDB:           fakePinger{err: nil},
	})

	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestNewRouter_WiresPublicResourceReadRoutes(t *testing.T) {
	// PR2: the 5 public GET resources must be reachable through the real
	// router wiring, not just via ad-hoc gin.New() routers in each
	// resource's own test file.
	r := rest.NewRouter(rest.RouterDeps{
		HealthDB:            fakePinger{err: nil},
		ProductService:      fakeProductService{products: []domain.Product{}},
		NewsService:         fakeNewsService{items: []domain.NewsItem{{ID: 1, Title: "t"}}},
		GalleryService:      fakeGalleryService{images: []domain.GalleryImage{{ID: 1, Image: "i"}}},
		FamilyService:       fakeFamilyService{items: []domain.FamilyItem{{ID: 1, Body: "b", Category: "Centro de dia"}}},
		OrganizationService: fakeOrganizationService{items: []domain.Organization{{ID: 1, Title: "t", Image: "i"}}},
	})

	cases := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/api/products", http.StatusOK},
		{http.MethodGet, "/api/products/1", http.StatusNotFound},
		{http.MethodGet, "/api/news", http.StatusOK},
		{http.MethodGet, "/api/gallery", http.StatusOK},
		{http.MethodGet, "/api/family", http.StatusOK},
		{http.MethodGet, "/api/organizations", http.StatusOK},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("%s %s -> status = %d, want %d, body=%s", tc.method, tc.path, rec.Code, tc.want, rec.Body.String())
		}
	}
}
