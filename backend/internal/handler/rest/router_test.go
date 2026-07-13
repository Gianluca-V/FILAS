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
