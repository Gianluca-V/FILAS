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

type fakeNewsService struct {
	items []domain.NewsItem
	byID  map[int]domain.NewsItem
	err   error
}

func (f fakeNewsService) List(ctx context.Context) ([]domain.NewsItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f fakeNewsService) Get(ctx context.Context, id int) (domain.NewsItem, error) {
	if f.err != nil {
		return domain.NewsItem{}, f.err
	}
	item, ok := f.byID[id]
	if !ok {
		return domain.NewsItem{}, domain.ErrNotFound
	}
	return item, nil
}

func newNewsTestRouter(svc rest.NewsService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewNewsHandler(svc)
	r.GET("/api/news", h.List)
	r.GET("/api/news/:id", h.Get)
	return r
}

func TestNewsHandler_List_ReturnsArray(t *testing.T) {
	body := "Lorem ipsum"
	svc := fakeNewsService{items: []domain.NewsItem{{ID: 1, Title: "Noticia de prueba 1", Body: &body}}}
	r := newNewsTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/news", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `[{"ID":"1","Title":"Noticia de prueba 1","Body":"Lorem ipsum","Image":null}]`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestNewsHandler_List_ReturnsNotFoundWhenEmpty(t *testing.T) {
	r := newNewsTestRouter(fakeNewsService{items: []domain.NewsItem{}})

	req := httptest.NewRequest(http.MethodGet, "/api/news", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"No news found"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"No news found"}`)
	}
}

func TestNewsHandler_Get_WrapsSingleResultInArray(t *testing.T) {
	svc := fakeNewsService{byID: map[int]domain.NewsItem{1: {ID: 1, Title: "Noticia de prueba 1"}}}
	r := newNewsTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/news/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `[{"ID":"1","Title":"Noticia de prueba 1","Body":null,"Image":null}]`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestNewsHandler_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	r := newNewsTestRouter(fakeNewsService{byID: map[int]domain.NewsItem{}})

	req := httptest.NewRequest(http.MethodGet, "/api/news/999999", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"No news found"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"No news found"}`)
	}
}

func TestNewsHandler_Get_TreatsNonNumericIDAsNotFound(t *testing.T) {
	// Mirrors legacy PHP intval("abc") == 0 -> lookup of ID 0 -> not found.
	r := newNewsTestRouter(fakeNewsService{byID: map[int]domain.NewsItem{}})

	req := httptest.NewRequest(http.MethodGet, "/api/news/abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestNewsHandler_List_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newNewsTestRouter(fakeNewsService{err: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/news", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"internal server error"}` {
		t.Errorf("body = %s, want the generic ErrorHandler body, not a leaked error", rec.Body.String())
	}
}

func TestNewsHandler_Get_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newNewsTestRouter(fakeNewsService{err: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/news/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"internal server error"}` {
		t.Errorf("body = %s, want the generic ErrorHandler body, not a leaked error", rec.Body.String())
	}
}
