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

type fakeGalleryService struct {
	images []domain.GalleryImage
	byID   map[int]domain.GalleryImage
	err    error
}

func (f fakeGalleryService) List(ctx context.Context) ([]domain.GalleryImage, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.images, nil
}

func (f fakeGalleryService) Get(ctx context.Context, id int) (domain.GalleryImage, error) {
	if f.err != nil {
		return domain.GalleryImage{}, f.err
	}
	img, ok := f.byID[id]
	if !ok {
		return domain.GalleryImage{}, domain.ErrNotFound
	}
	return img, nil
}

func newGalleryTestRouter(svc rest.GalleryService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewGalleryHandler(svc)
	r.GET("/api/gallery", h.List)
	r.GET("/api/gallery/:id", h.Get)
	return r
}

func TestGalleryHandler_List_ReturnsArrayEvenWhenPopulated(t *testing.T) {
	svc := fakeGalleryService{images: []domain.GalleryImage{
		{ID: 1, Image: "assets/galeria-1.jpg"},
		{ID: 2, Image: "assets/galeria-2.jpg"},
	}}
	r := newGalleryTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/gallery", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `[{"ID":"1","Image":"assets/galeria-1.jpg"},{"ID":"2","Image":"assets/galeria-2.jpg"}]`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestGalleryHandler_List_ReturnsNotFoundWhenEmpty(t *testing.T) {
	// Legacy quirk (characterized live, obs #29/#31): gallery.php returns
	// 404 "No images found" when the table is empty, NOT 200 [] like
	// products.php does. Structural, not decorative — preserved exactly.
	r := newGalleryTestRouter(fakeGalleryService{images: []domain.GalleryImage{}})

	req := httptest.NewRequest(http.MethodGet, "/api/gallery", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"No images found"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"No images found"}`)
	}
}

func TestGalleryHandler_Get_WrapsSingleResultInArray(t *testing.T) {
	// Legacy quirk (characterized live): GET /api/gallery/:id reuses the
	// list row-collector, so a found item is wrapped in a 1-element array,
	// not returned as a bare object.
	svc := fakeGalleryService{byID: map[int]domain.GalleryImage{5: {ID: 5, Image: "assets/galeria-5.jpg"}}}
	r := newGalleryTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/gallery/5", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `[{"ID":"5","Image":"assets/galeria-5.jpg"}]`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestGalleryHandler_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	r := newGalleryTestRouter(fakeGalleryService{byID: map[int]domain.GalleryImage{}})

	req := httptest.NewRequest(http.MethodGet, "/api/gallery/999999", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"No images found"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"No images found"}`)
	}
}

func TestGalleryHandler_Get_TreatsNonNumericIDAsNotFound(t *testing.T) {
	// Mirrors legacy PHP intval("abc") == 0 -> lookup of ID 0 -> not found.
	r := newGalleryTestRouter(fakeGalleryService{byID: map[int]domain.GalleryImage{}})

	req := httptest.NewRequest(http.MethodGet, "/api/gallery/abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestGalleryHandler_List_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newGalleryTestRouter(fakeGalleryService{err: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/gallery", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"internal server error"}` {
		t.Errorf("body = %s, want the generic ErrorHandler body, not a leaked error", rec.Body.String())
	}
}

func TestGalleryHandler_Get_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newGalleryTestRouter(fakeGalleryService{err: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/gallery/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"internal server error"}` {
		t.Errorf("body = %s, want the generic ErrorHandler body, not a leaked error", rec.Body.String())
	}
}
