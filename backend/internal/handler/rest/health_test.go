package rest_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/handler/rest"
)

type fakePinger struct {
	err error
}

func (f fakePinger) PingContext(ctx context.Context) error {
	return f.err
}

func newHealthRouter(pinger interface {
	PingContext(ctx context.Context) error
}) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", rest.NewHealthHandler(pinger))
	return r
}

func TestHealthHandler_ReturnsOKWhenDBReachable(t *testing.T) {
	r := newHealthRouter(fakePinger{err: nil})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Errorf("body = %q, want it to contain %q", rec.Body.String(), `"status":"ok"`)
	}
}

func TestHealthHandler_ReturnsServiceUnavailableWhenDBUnreachable(t *testing.T) {
	r := newHealthRouter(fakePinger{err: errors.New("connection refused")})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(rec.Body.String(), `"status":"down"`) {
		t.Errorf("body = %q, want it to contain %q", rec.Body.String(), `"status":"down"`)
	}
}

func TestHealthHandler_DoesNotLeakDBErrorToCaller(t *testing.T) {
	// Unauthenticated callers must never see raw DB error text; the real
	// error is logged server-side instead (fix #7 from the PR1 gate review).
	r := newHealthRouter(fakePinger{err: errors.New("connection refused: dial tcp 10.0.0.5:3306")})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if strings.Contains(rec.Body.String(), "connection refused") {
		t.Errorf("body = %q, must not leak the underlying DB error to unauthenticated callers", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "10.0.0.5") {
		t.Errorf("body = %q, must not leak internal host details", rec.Body.String())
	}
}
