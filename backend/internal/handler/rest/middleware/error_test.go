package middleware_test

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest/middleware"
)

func newErrorRouter(handlerErr error) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.GET("/boom", func(c *gin.Context) {
		if handlerErr != nil {
			_ = c.Error(handlerErr)
		}
	})
	return r
}

func TestErrorHandler_MapsDomainErrorsToStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"not found", domain.ErrNotFound, http.StatusNotFound},
		{"unauthorized", domain.ErrUnauthorized, http.StatusUnauthorized},
		{"validation", domain.ErrValidation, http.StatusBadRequest},
		{"conflict", domain.ErrConflict, http.StatusConflict},
		{"insufficient stock", domain.ErrInsufficientStock, http.StatusConflict},
		{"unmapped unknown error", errors.New("wrap: "), http.StatusInternalServerError}, // unknown -> 500
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newErrorRouter(tt.err)

			req := httptest.NewRequest(http.MethodGet, "/boom", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body=%s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestErrorHandler_WrappedDomainErrorStillMaps(t *testing.T) {
	wrapped := fmt.Errorf("get product: %w", domain.ErrNotFound)

	r := newErrorRouter(wrapped)

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d (body=%s)", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestErrorHandler_LogsUnderlyingErrorWithMethodAndPath(t *testing.T) {
	// fix #6 from the PR1 gate review: the 500 branch must log the real
	// error server-side (method + path + error) even though the JSON
	// response body stays generic.
	underlying := errors.New("boom: unexpected nil pointer")
	r := newErrorRouter(underlying)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	logged := buf.String()
	if !strings.Contains(logged, "boom: unexpected nil pointer") {
		t.Errorf("log output = %q, want it to contain the underlying error", logged)
	}
	if !strings.Contains(logged, http.MethodGet) {
		t.Errorf("log output = %q, want it to contain the request method %q", logged, http.MethodGet)
	}
	if !strings.Contains(logged, "/boom") {
		t.Errorf("log output = %q, want it to contain the request path %q", logged, "/boom")
	}
	if strings.Contains(rec.Body.String(), "boom: unexpected nil pointer") {
		t.Errorf("response body = %q, must stay generic and not leak the underlying error", rec.Body.String())
	}
}

func TestErrorHandler_DoesNotOverwriteAlreadyWrittenResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}
