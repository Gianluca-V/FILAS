package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/auth"
	"github.com/gianluca-v/filas-backend/internal/handler/http/middleware"
)

func newProtectedRouter(svc *auth.JWTService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", middleware.RequireAuth(svc), func(c *gin.Context) {
		adminID, _ := c.Get("adminID")
		c.JSON(http.StatusOK, gin.H{"adminID": adminID})
	})
	return r
}

func TestRequireAuth_RejectsMissingToken(t *testing.T) {
	svc := auth.NewJWTService("secret", time.Hour)
	r := newProtectedRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_RejectsInvalidToken(t *testing.T) {
	svc := auth.NewJWTService("secret", time.Hour)
	r := newProtectedRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_PassesValidTokenWithBearerPrefix(t *testing.T) {
	svc := auth.NewJWTService("secret", time.Hour)
	token, err := svc.Generate(1543)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	r := newProtectedRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if want := `"adminID":1543`; !strings.Contains(rec.Body.String(), want) {
		t.Errorf("body = %q, want it to contain %q", rec.Body.String(), want)
	}
}

func TestRequireAuth_PassesValidTokenWithoutBearerPrefix(t *testing.T) {
	// Tolerates the raw token (no "Bearer " prefix) for parity with the
	// legacy client, which sends the token unprefixed.
	svc := auth.NewJWTService("secret", time.Hour)
	token, err := svc.Generate(7)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	r := newProtectedRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if want := `"adminID":7`; !strings.Contains(rec.Body.String(), want) {
		t.Errorf("body = %q, want it to contain %q", rec.Body.String(), want)
	}
}
