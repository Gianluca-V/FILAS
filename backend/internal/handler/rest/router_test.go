package rest_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/auth"
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
		ProductService:      &fakeProductService{products: []domain.Product{}},
		NewsService:         fakeNewsService{items: []domain.NewsItem{{ID: 1, Title: "t"}}},
		GalleryService:      &fakeGalleryService{images: []domain.GalleryImage{{ID: 1, Image: "i"}}},
		FamilyService:       &fakeFamilyService{items: []domain.FamilyItem{{ID: 1, Body: "b", Category: "Centro de dia"}}},
		OrganizationService: &fakeOrganizationService{items: []domain.Organization{{ID: 1, Title: "t", Image: "i"}}},
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

func TestNewRouter_WiresAdminRoutes(t *testing.T) {
	// PR3: login is public; list/get/create(non-login)/update/delete
	// require a valid JWT through the real router wiring (the first real
	// use of middleware.RequireAuth outside its own unit tests).
	jwtSvc := auth.NewJWTService("router-test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}

	r := rest.NewRouter(rest.RouterDeps{
		HealthDB:     fakePinger{err: nil},
		AdminService: &fakeAdminService{byID: map[int]domain.Admin{1543: {ID: 1543, Username: "FilasAdmin"}}},
		AuthService:  &fakeAuthService{token: "signed-jwt"},
		JWTService:   jwtSvc,
	})

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		token  string
		want   int
	}{
		{"login is public", http.MethodPost, "/api/admins", `{"login":true,"username":"x","password":"y"}`, "", http.StatusOK},
		{"list without token", http.MethodGet, "/api/admins", "", "", http.StatusUnauthorized},
		{"list with token", http.MethodGet, "/api/admins", "", token, http.StatusOK},
		{"get without token", http.MethodGet, "/api/admins/1543", "", "", http.StatusUnauthorized},
		{"get with token", http.MethodGet, "/api/admins/1543", "", token, http.StatusOK},
		{"update without token", http.MethodPut, "/api/admins/1543", `{"username":"x","password":"y"}`, "", http.StatusUnauthorized},
		{"delete without token", http.MethodDelete, "/api/admins/1543", "", "", http.StatusUnauthorized},
	}
	for _, tc := range cases {
		var body *strings.Reader
		if tc.body != "" {
			body = strings.NewReader(tc.body)
		} else {
			body = strings.NewReader("")
		}
		req := httptest.NewRequest(tc.method, tc.path, body)
		req.Header.Set("Content-Type", "application/json")
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("%s: %s %s -> status = %d, want %d, body=%s", tc.name, tc.method, tc.path, rec.Code, tc.want, rec.Body.String())
		}
	}
}

func TestNewRouter_WiresAuthGatedWriteRoutes(t *testing.T) {
	// PR4 (task 3.1): POST/PUT/DELETE for products/gallery/family/organizations
	// must be reachable through the real router wiring, gated behind
	// middleware.RequireAuth exactly like the admin GET/PUT/DELETE routes
	// (PR3). GETs stay public. News writes are wired separately in PR5.
	jwtSvc := auth.NewJWTService("router-test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}

	r := rest.NewRouter(rest.RouterDeps{
		HealthDB:            fakePinger{err: nil},
		ProductService:      &fakeProductService{byID: map[int]domain.Product{1: {ID: 1, Name: "p"}}},
		GalleryService:      &fakeGalleryService{byID: map[int]domain.GalleryImage{1: {ID: 1, Image: "i"}}},
		FamilyService:       &fakeFamilyService{byID: map[int]domain.FamilyItem{1: {ID: 1, Body: "b", Category: "Centro de dia"}}},
		OrganizationService: &fakeOrganizationService{byID: map[int]domain.Organization{1: {ID: 1, Title: "t", Image: "i"}}},
		JWTService:          jwtSvc,
	})

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		token  string
		want   int
	}{
		{"products create without token", http.MethodPost, "/api/products", `{"Name":"n","Price":1,"Stock":1}`, "", http.StatusUnauthorized},
		{"products create with token", http.MethodPost, "/api/products", `{"Name":"n","Price":1,"Stock":1}`, token, http.StatusOK},
		{"products update without token", http.MethodPut, "/api/products/1", `{"Name":"n","Price":1,"Stock":1}`, "", http.StatusUnauthorized},
		{"products delete without token", http.MethodDelete, "/api/products/1", "", "", http.StatusUnauthorized},
		{"gallery create without token", http.MethodPost, "/api/gallery", `{"Image":"i"}`, "", http.StatusUnauthorized},
		{"gallery create with token", http.MethodPost, "/api/gallery", `{"Image":"i"}`, token, http.StatusCreated},
		{"gallery update without token", http.MethodPut, "/api/gallery/1", `{"Image":"i"}`, "", http.StatusUnauthorized},
		{"gallery delete without token", http.MethodDelete, "/api/gallery/1", "", "", http.StatusUnauthorized},
		{"family create without token", http.MethodPost, "/api/family", `{"Body":"b","Category":"Centro de dia"}`, "", http.StatusUnauthorized},
		{"family create with token", http.MethodPost, "/api/family", `{"Body":"b","Category":"Centro de dia"}`, token, http.StatusCreated},
		{"family update without token", http.MethodPut, "/api/family/1", `{"Body":"b","Category":"Centro de dia"}`, "", http.StatusUnauthorized},
		{"family delete without token", http.MethodDelete, "/api/family/1", "", "", http.StatusUnauthorized},
		{"organizations create without token", http.MethodPost, "/api/organizations", `{"Title":"t","Image":"i"}`, "", http.StatusUnauthorized},
		{"organizations create with token", http.MethodPost, "/api/organizations", `{"Title":"t","Image":"i"}`, token, http.StatusCreated},
		{"organizations update without token", http.MethodPut, "/api/organizations/1", `{"Title":"t","Image":"i"}`, "", http.StatusUnauthorized},
		{"organizations delete without token", http.MethodDelete, "/api/organizations/1", "", "", http.StatusUnauthorized},
	}
	for _, tc := range cases {
		var body *strings.Reader
		if tc.body != "" {
			body = strings.NewReader(tc.body)
		} else {
			body = strings.NewReader("")
		}
		req := httptest.NewRequest(tc.method, tc.path, body)
		req.Header.Set("Content-Type", "application/json")
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("%s: %s %s -> status = %d, want %d, body=%s", tc.name, tc.method, tc.path, rec.Code, tc.want, rec.Body.String())
		}
	}
}
