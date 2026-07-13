package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	handlerhttp "github.com/gianluca-v/filas-backend/internal/handler/http"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewRouter_HealthRouteWiredAndReturnsOK(t *testing.T) {
	r := handlerhttp.NewRouter(handlerhttp.RouterDeps{
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
	r := handlerhttp.NewRouter(handlerhttp.RouterDeps{
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

func TestNewRouter_UnknownRouteReturns404(t *testing.T) {
	r := handlerhttp.NewRouter(handlerhttp.RouterDeps{
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
