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

type fakeFamilyService struct {
	items []domain.FamilyItem
	byID  map[int]domain.FamilyItem
	err   error
}

func (f fakeFamilyService) List(ctx context.Context) ([]domain.FamilyItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f fakeFamilyService) Get(ctx context.Context, id int) (domain.FamilyItem, error) {
	if f.err != nil {
		return domain.FamilyItem{}, f.err
	}
	item, ok := f.byID[id]
	if !ok {
		return domain.FamilyItem{}, domain.ErrNotFound
	}
	return item, nil
}

func newFamilyTestRouter(svc rest.FamilyService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewFamilyHandler(svc)
	r.GET("/api/family", h.List)
	r.GET("/api/family/:id", h.Get)
	return r
}

func TestFamilyHandler_List_ReturnsArray(t *testing.T) {
	svc := fakeFamilyService{items: []domain.FamilyItem{{ID: 3, Body: "aqui va la descripcion 3", Category: "Taller protegido"}}}
	r := newFamilyTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/family", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `[{"ID":"3","Image":null,"Body":"aqui va la descripcion 3","Category":"Taller protegido"}]`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestFamilyHandler_List_ReturnsNotFoundWhenEmpty(t *testing.T) {
	r := newFamilyTestRouter(fakeFamilyService{items: []domain.FamilyItem{}})

	req := httptest.NewRequest(http.MethodGet, "/api/family", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"No workshops found"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"No workshops found"}`)
	}
}

func TestFamilyHandler_Get_WrapsSingleResultInArray(t *testing.T) {
	svc := fakeFamilyService{byID: map[int]domain.FamilyItem{3: {ID: 3, Body: "aqui va la descripcion 3", Category: "Taller protegido"}}}
	r := newFamilyTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/family/3", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `[{"ID":"3","Image":null,"Body":"aqui va la descripcion 3","Category":"Taller protegido"}]`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestFamilyHandler_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	r := newFamilyTestRouter(fakeFamilyService{byID: map[int]domain.FamilyItem{}})

	req := httptest.NewRequest(http.MethodGet, "/api/family/999999", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"No workshops found"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"No workshops found"}`)
	}
}

func TestFamilyHandler_Get_TreatsNonNumericIDAsNotFound(t *testing.T) {
	// Mirrors legacy PHP intval("abc") == 0 -> lookup of ID 0 -> not found.
	r := newFamilyTestRouter(fakeFamilyService{byID: map[int]domain.FamilyItem{}})

	req := httptest.NewRequest(http.MethodGet, "/api/family/abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestFamilyHandler_List_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newFamilyTestRouter(fakeFamilyService{err: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/family", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"internal server error"}` {
		t.Errorf("body = %s, want the generic ErrorHandler body, not a leaked error", rec.Body.String())
	}
}

func TestFamilyHandler_Get_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newFamilyTestRouter(fakeFamilyService{err: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/family/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"internal server error"}` {
		t.Errorf("body = %s, want the generic ErrorHandler body, not a leaked error", rec.Body.String())
	}
}
