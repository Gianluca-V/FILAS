package middleware_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/handler/rest/middleware"
)

func TestRequestLogger_LogsMethodPathAndStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestLogger())
	r.GET("/widgets/42", func(c *gin.Context) { c.Status(http.StatusOK) })

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	req := httptest.NewRequest(http.MethodGet, "/widgets/42", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	logged := buf.String()
	if !strings.Contains(logged, http.MethodGet) {
		t.Errorf("log output = %q, want it to contain method %q", logged, http.MethodGet)
	}
	if !strings.Contains(logged, "/widgets/42") {
		t.Errorf("log output = %q, want it to contain path %q", logged, "/widgets/42")
	}
	if !strings.Contains(logged, "200") {
		t.Errorf("log output = %q, want it to contain status %q", logged, "200")
	}
}

func TestRequestLogger_LogsErrorStatusForFailedRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestLogger())
	r.GET("/boom", func(c *gin.Context) { c.Status(http.StatusInternalServerError) })

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	logged := buf.String()
	if !strings.Contains(logged, "500") {
		t.Errorf("log output = %q, want it to contain status %q", logged, "500")
	}
}
