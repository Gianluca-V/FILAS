package rest_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest"
	"github.com/gianluca-v/filas-backend/internal/handler/rest/middleware"
)

type fakeOrganizationService struct {
	items []domain.Organization
	byID  map[int]domain.Organization
	err   error
}

func (f fakeOrganizationService) List(ctx context.Context) ([]domain.Organization, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f fakeOrganizationService) Get(ctx context.Context, id int) (domain.Organization, error) {
	if f.err != nil {
		return domain.Organization{}, f.err
	}
	item, ok := f.byID[id]
	if !ok {
		return domain.Organization{}, domain.ErrNotFound
	}
	return item, nil
}

func newOrganizationTestRouter(svc rest.OrganizationService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewOrganizationHandler(svc)
	r.GET("/api/organizations", h.List)
	r.GET("/api/organizations/:id", h.Get)
	return r
}

func TestOrganizationHandler_List_ReturnsArray(t *testing.T) {
	desc := "Descripcion de prueba 1"
	svc := fakeOrganizationService{items: []domain.Organization{
		{ID: 1, Title: "organizacions de prueba 1", Description: &desc, Image: "https://example.com/1.jpg"},
	}}
	r := newOrganizationTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/organizations", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `[{"ID":"1","Title":"organizacions de prueba 1","Description":"Descripcion de prueba 1","Image":"https://example.com/1.jpg"}]`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestOrganizationHandler_List_ReturnsNotFoundWhenEmpty(t *testing.T) {
	r := newOrganizationTestRouter(fakeOrganizationService{items: []domain.Organization{}})

	req := httptest.NewRequest(http.MethodGet, "/api/organizations", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"No organizations found"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"No organizations found"}`)
	}
}

func TestOrganizationHandler_Get_WrapsSingleResultInArray(t *testing.T) {
	svc := fakeOrganizationService{byID: map[int]domain.Organization{1: {ID: 1, Title: "organizacions de prueba 1", Image: "https://example.com/1.jpg"}}}
	r := newOrganizationTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/organizations/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `[{"ID":"1","Title":"organizacions de prueba 1","Description":null,"Image":"https://example.com/1.jpg"}]`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestOrganizationHandler_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	r := newOrganizationTestRouter(fakeOrganizationService{byID: map[int]domain.Organization{}})

	req := httptest.NewRequest(http.MethodGet, "/api/organizations/999999", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"No organizations found"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"No organizations found"}`)
	}
}

func TestOrganizationHandler_Get_TreatsNonNumericIDAsNotFound(t *testing.T) {
	// Mirrors legacy PHP intval("abc") == 0 -> lookup of ID 0 -> not found.
	r := newOrganizationTestRouter(fakeOrganizationService{byID: map[int]domain.Organization{}})

	req := httptest.NewRequest(http.MethodGet, "/api/organizations/abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestOrganizationHandler_List_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newOrganizationTestRouter(fakeOrganizationService{err: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/organizations", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"internal server error"}` {
		t.Errorf("body = %s, want the generic ErrorHandler body, not a leaked error", rec.Body.String())
	}
}

func TestOrganizationHandler_Get_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newOrganizationTestRouter(fakeOrganizationService{err: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/organizations/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"internal server error"}` {
		t.Errorf("body = %s, want the generic ErrorHandler body, not a leaked error", rec.Body.String())
	}
}
