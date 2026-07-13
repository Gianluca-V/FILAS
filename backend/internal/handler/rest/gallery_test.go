package rest_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/auth"
	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest"
	"github.com/gianluca-v/filas-backend/internal/handler/rest/middleware"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

// fakeGalleryRepoE2E implements domain.GalleryRepository (not
// rest.GalleryService) so end-to-end validation tests can wire the REAL
// usecase.GalleryService through the real HTTP stack, mirroring the realSvc
// pattern established in admin_test.go: a fake at the SERVICE layer would
// bypass usecase-level validation, since usecase.GalleryService.Create is
// what rejects an empty Image — not the handler.
type fakeGalleryRepoE2E struct {
	createCalled bool
	updateCalled bool
}

func (f *fakeGalleryRepoE2E) List(ctx context.Context) ([]domain.GalleryImage, error) {
	return nil, nil
}

func (f *fakeGalleryRepoE2E) Get(ctx context.Context, id int) (domain.GalleryImage, error) {
	return domain.GalleryImage{}, domain.ErrNotFound
}

func (f *fakeGalleryRepoE2E) Create(ctx context.Context, g domain.GalleryImage) (domain.GalleryImage, error) {
	f.createCalled = true
	return g, nil
}

func (f *fakeGalleryRepoE2E) Update(ctx context.Context, id int, image string) error {
	f.updateCalled = true
	return nil
}

func (f *fakeGalleryRepoE2E) Delete(ctx context.Context, id int) error {
	return nil
}

// fakeGalleryService uses pointer receivers (unlike the read-only PR2 fake)
// so Create/Update/Delete tests can record what was actually passed to the
// service, mirroring the fakeAdminService convention established in PR3.
type fakeGalleryService struct {
	images []domain.GalleryImage
	byID   map[int]domain.GalleryImage
	err    error

	createErr    error
	createdImage domain.GalleryImage
	updateErr    error
	updateCalled bool
	updateID     int
	updateImage  string
	deleteErr    error
	deleteID     int
}

func (f *fakeGalleryService) List(ctx context.Context) ([]domain.GalleryImage, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.images, nil
}

func (f *fakeGalleryService) Get(ctx context.Context, id int) (domain.GalleryImage, error) {
	if f.err != nil {
		return domain.GalleryImage{}, f.err
	}
	img, ok := f.byID[id]
	if !ok {
		return domain.GalleryImage{}, domain.ErrNotFound
	}
	return img, nil
}

func (f *fakeGalleryService) Create(ctx context.Context, g domain.GalleryImage) (domain.GalleryImage, error) {
	if f.createErr != nil {
		return domain.GalleryImage{}, f.createErr
	}
	f.createdImage = g
	g.ID = 9001
	return g, nil
}

func (f *fakeGalleryService) Update(ctx context.Context, id int, image string) error {
	f.updateCalled = true
	f.updateID = id
	f.updateImage = image
	return f.updateErr
}

func (f *fakeGalleryService) Delete(ctx context.Context, id int) error {
	f.deleteID = id
	return f.deleteErr
}

func newGalleryTestRouter(svc rest.GalleryService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewGalleryHandler(svc)
	r.GET("/api/gallery", h.List)
	r.GET("/api/gallery/:id", h.Get)
	r.POST("/api/gallery", h.Create)
	r.PUT("/api/gallery/:id", h.Update)
	r.DELETE("/api/gallery/:id", h.Delete)
	return r
}

// newAuthedGalleryTestRouter wires the same write routes behind
// middleware.RequireAuth, mirroring the real router.go wiring (task 3.1).
func newAuthedGalleryTestRouter(svc rest.GalleryService, jwtSvc *auth.JWTService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewGalleryHandler(svc)
	r.GET("/api/gallery", h.List)
	r.GET("/api/gallery/:id", h.Get)
	r.POST("/api/gallery", middleware.RequireAuth(jwtSvc), h.Create)
	r.PUT("/api/gallery/:id", middleware.RequireAuth(jwtSvc), h.Update)
	r.DELETE("/api/gallery/:id", middleware.RequireAuth(jwtSvc), h.Delete)
	return r
}

func TestGalleryHandler_List_ReturnsArrayEvenWhenPopulated(t *testing.T) {
	svc := &fakeGalleryService{images: []domain.GalleryImage{
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
	r := newGalleryTestRouter(&fakeGalleryService{images: []domain.GalleryImage{}})

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
	svc := &fakeGalleryService{byID: map[int]domain.GalleryImage{5: {ID: 5, Image: "assets/galeria-5.jpg"}}}
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
	r := newGalleryTestRouter(&fakeGalleryService{byID: map[int]domain.GalleryImage{}})

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
	r := newGalleryTestRouter(&fakeGalleryService{byID: map[int]domain.GalleryImage{}})

	req := httptest.NewRequest(http.MethodGet, "/api/gallery/abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestGalleryHandler_List_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newGalleryTestRouter(&fakeGalleryService{err: errors.New("db down")})

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
	r := newGalleryTestRouter(&fakeGalleryService{err: errors.New("db down")})

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

// --- Create (POST /api/gallery) ---

func TestGalleryHandler_Create_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAuthedGalleryTestRouter(&fakeGalleryService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/gallery", strings.NewReader(`{"Image":"assets/nueva.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestGalleryHandler_Create_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeGalleryService{}
	r := newAuthedGalleryTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/gallery", strings.NewReader(`{"Image":"assets/nueva.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Image added successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Image added successfully"}`)
	}
	if svc.createdImage.Image != "assets/nueva.jpg" {
		t.Errorf("Create() persisted Image = %q, want %q", svc.createdImage.Image, "assets/nueva.jpg")
	}
}

func TestGalleryHandler_Create_RejectsEmptyImageWithBadRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeGalleryRepoE2E{}
	realSvc := usecase.NewGalleryService(repo)
	r := newAuthedGalleryTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/gallery", strings.NewReader(`{"Image":""}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.createCalled {
		t.Errorf("repository.Create was called for an empty Image — no persistence should happen on validation failure")
	}
}

func TestGalleryHandler_Create_TreatsMalformedJSONBodyAsEmptyRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeGalleryRepoE2E{}
	realSvc := usecase.NewGalleryService(repo)
	r := newAuthedGalleryTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/gallery", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.createCalled {
		t.Errorf("repository.Create was called for a malformed body")
	}
}

func TestGalleryHandler_Create_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeGalleryService{createErr: errors.New("db down")}
	r := newAuthedGalleryTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/gallery", strings.NewReader(`{"Image":"assets/nueva.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"internal server error"}` {
		t.Errorf("body = %s, want the generic ErrorHandler body, not a leaked error", rec.Body.String())
	}
}

// --- Update (PUT /api/gallery/:id) ---

func TestGalleryHandler_Update_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAuthedGalleryTestRouter(&fakeGalleryService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/gallery/5", strings.NewReader(`{"Image":"assets/actualizada.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestGalleryHandler_Update_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeGalleryService{}
	r := newAuthedGalleryTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/gallery/5", strings.NewReader(`{"Image":"assets/actualizada.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.updateID != 5 || svc.updateImage != "assets/actualizada.jpg" {
		t.Errorf("Update() called with id=%d image=%q, want id=5 image=assets/actualizada.jpg", svc.updateID, svc.updateImage)
	}
}

func TestGalleryHandler_Update_RejectsEmptyImageWithBadRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeGalleryRepoE2E{}
	realSvc := usecase.NewGalleryService(repo)
	r := newAuthedGalleryTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/gallery/5", strings.NewReader(`{"Image":""}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.updateCalled {
		t.Errorf("repository.Update was called for an empty Image — no persistence should happen on validation failure")
	}
}

func TestGalleryHandler_Update_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeGalleryService{updateErr: errors.New("db down")}
	r := newAuthedGalleryTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/gallery/5", strings.NewReader(`{"Image":"assets/actualizada.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

// --- Delete (DELETE /api/gallery/:id) ---

func TestGalleryHandler_Delete_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAuthedGalleryTestRouter(&fakeGalleryService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/gallery/5", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestGalleryHandler_Delete_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeGalleryService{}
	r := newAuthedGalleryTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/gallery/7", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.deleteID != 7 {
		t.Errorf("Delete() called with id=%d, want 7", svc.deleteID)
	}
}

func TestGalleryHandler_Delete_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeGalleryService{deleteErr: errors.New("db down")}
	r := newAuthedGalleryTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/gallery/7", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}
