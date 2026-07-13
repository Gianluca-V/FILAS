package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/handler/rest/middleware"
)

func newCORSRouter(allowedOrigins []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CORS(allowedOrigins))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func TestCORS_PreflightReturnsNoContentWithHeaders(t *testing.T) {
	r := newCORSRouter([]string{"http://localhost:5173"})

	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("Access-Control-Allow-Methods header not set on preflight response")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("Access-Control-Allow-Headers header not set on preflight response")
	}
}

func TestCORS_EchoesAllowedOrigin(t *testing.T) {
	r := newCORSRouter([]string{"http://localhost:5173"})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:5173")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestCORS_OmitsHeaderForDisallowedOrigin(t *testing.T) {
	r := newCORSRouter([]string{"http://localhost:5173"})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://evil.example")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty for disallowed origin", got)
	}
}

func TestCORS_FailsClosedWhenNoOriginsConfigured(t *testing.T) {
	// fix #8 from the PR1 gate review: an unset CORS_ALLOWED_ORIGINS must
	// fail CLOSED (no ACAO header at all), never fall back to a wildcard.
	r := newCORSRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://anything.example")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty when no origins are configured", got)
	}
}

func TestCORS_WildcardOnlyWhenExplicitlyConfigured(t *testing.T) {
	// Wildcard is opt-in: only when the operator explicitly sets
	// CORS_ALLOWED_ORIGINS=* does any origin get echoed back.
	r := newCORSRouter([]string{"*"})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://anything.example")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "*")
	}
}
