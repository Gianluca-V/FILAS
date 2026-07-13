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

// fakeNewsService uses pointer receivers (unlike the read-only PR2 fake) so
// Create/Update/Delete tests can record what was actually passed to the
// service, mirroring the fakeProductService convention established in PR4.
type fakeNewsService struct {
	items []domain.NewsItem
	byID  map[int]domain.NewsItem
	err   error

	createErr   error
	createdItem domain.NewsItem
	updateErr   error
	updateID    int
	updateItem  domain.NewsItem
	deleteErr   error
	deleteID    int
}

func (f *fakeNewsService) List(ctx context.Context) ([]domain.NewsItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f *fakeNewsService) Get(ctx context.Context, id int) (domain.NewsItem, error) {
	if f.err != nil {
		return domain.NewsItem{}, f.err
	}
	item, ok := f.byID[id]
	if !ok {
		return domain.NewsItem{}, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakeNewsService) Create(ctx context.Context, n domain.NewsItem) (domain.NewsItem, error) {
	if f.createErr != nil {
		return domain.NewsItem{}, f.createErr
	}
	f.createdItem = n
	n.ID = 9001
	return n, nil
}

func (f *fakeNewsService) Update(ctx context.Context, id int, n domain.NewsItem) error {
	f.updateID = id
	f.updateItem = n
	return f.updateErr
}

func (f *fakeNewsService) Delete(ctx context.Context, id int) error {
	f.deleteID = id
	return f.deleteErr
}

// fakeNewsRepoE2E implements domain.NewsRepository (not rest.NewsService)
// so end-to-end validation tests can wire the REAL usecase.NewsService
// through the real HTTP stack, mirroring fakeProductRepoE2E (PR4): a fake
// at the SERVICE layer would bypass usecase-level validation entirely,
// since usecase.NewsService.Create/Update is what rejects a missing
// Title/Body/Image — not the handler.
type fakeNewsRepoE2E struct {
	createCalled bool
	updateCalled bool
	updateID     int
	deleteCalled bool
	deleteID     int
}

func (f *fakeNewsRepoE2E) List(ctx context.Context) ([]domain.NewsItem, error) {
	return nil, nil
}

func (f *fakeNewsRepoE2E) Get(ctx context.Context, id int) (domain.NewsItem, error) {
	return domain.NewsItem{}, domain.ErrNotFound
}

func (f *fakeNewsRepoE2E) Create(ctx context.Context, n domain.NewsItem) (domain.NewsItem, error) {
	f.createCalled = true
	return n, nil
}

func (f *fakeNewsRepoE2E) Update(ctx context.Context, id int, n domain.NewsItem) error {
	f.updateCalled = true
	f.updateID = id
	// Matches legacy mysqli parity (backend/docs/legacy-quirks.md §10): a
	// zero-rows-affected UPDATE (e.g. a missing or non-numeric ID) is NOT
	// an error, so this fake never returns one here regardless of id.
	return nil
}

func (f *fakeNewsRepoE2E) Delete(ctx context.Context, id int) error {
	f.deleteCalled = true
	f.deleteID = id
	// Same zero-rows-affected-is-not-an-error parity as Update.
	return nil
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

// newAuthedNewsTestRouter wires the same write routes behind
// middleware.RequireAuth, mirroring the real router.go wiring (task 4.1),
// for tests that need to prove the 401-without-JWT / 2xx-with-JWT contract
// — THE headline fix of PR5: legacy news.php only validated the JWT if the
// Authorization header happened to be present at all (see news.php
// lines 30-33/61-64/94-97); a request with no header skipped validation
// entirely and the write proceeded unauthenticated. Go always enforces it.
func newAuthedNewsTestRouter(svc rest.NewsService, jwtSvc *auth.JWTService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewNewsHandler(svc)
	r.GET("/api/news", h.List)
	r.GET("/api/news/:id", h.Get)
	r.POST("/api/news", middleware.RequireAuth(jwtSvc), h.Create)
	r.PUT("/api/news/:id", middleware.RequireAuth(jwtSvc), h.Update)
	r.DELETE("/api/news/:id", middleware.RequireAuth(jwtSvc), h.Delete)
	return r
}

func TestNewsHandler_List_ReturnsArray(t *testing.T) {
	body := "Lorem ipsum"
	svc := &fakeNewsService{items: []domain.NewsItem{{ID: 1, Title: "Noticia de prueba 1", Body: &body}}}
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
	r := newNewsTestRouter(&fakeNewsService{items: []domain.NewsItem{}})

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
	svc := &fakeNewsService{byID: map[int]domain.NewsItem{1: {ID: 1, Title: "Noticia de prueba 1"}}}
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
	r := newNewsTestRouter(&fakeNewsService{byID: map[int]domain.NewsItem{}})

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
	r := newNewsTestRouter(&fakeNewsService{byID: map[int]domain.NewsItem{}})

	req := httptest.NewRequest(http.MethodGet, "/api/news/abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestNewsHandler_List_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newNewsTestRouter(&fakeNewsService{err: errors.New("db down")})

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
	r := newNewsTestRouter(&fakeNewsService{err: errors.New("db down")})

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

// --- Create (POST /api/news) ---
//
// THE headline behavior of PR5 (task 4.1): legacy news.php's auth check is
// gated behind `if(isset($headers['Authorization']))` (news.php lines
// 30-33) — a request with NO Authorization header skips TokenValidationResponse
// entirely and the write proceeds UNAUTHENTICATED. The Go handler fixes
// this: middleware.RequireAuth runs unconditionally, so a missing OR
// invalid token always returns 401 and never reaches the usecase/repo.

func TestNewsHandler_Create_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	svc := &fakeNewsService{}
	r := newAuthedNewsTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/news", strings.NewReader(`{"Title":"Nueva","Body":"Cuerpo","Image":"i.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if svc.createdItem.Title != "" {
		t.Errorf("Create() should not have been called without a JWT, but persisted %+v", svc.createdItem)
	}
}

func TestNewsHandler_Create_RejectsInvalidOrExpiredJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", -time.Hour) // already-expired token
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeNewsService{}
	r := newAuthedNewsTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/news", strings.NewReader(`{"Title":"Nueva","Body":"Cuerpo","Image":"i.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if svc.createdItem.Title != "" {
		t.Errorf("Create() should not have been called with an expired JWT, but persisted %+v", svc.createdItem)
	}
}

func TestNewsHandler_Create_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeNewsService{}
	r := newAuthedNewsTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/news", strings.NewReader(`{"Title":"Nueva","Body":"Cuerpo","Image":"i.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"News added successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"News added successfully"}`)
	}
	if svc.createdItem.Title != "Nueva" {
		t.Errorf("Create() persisted Title = %q, want %q", svc.createdItem.Title, "Nueva")
	}
}

func TestNewsHandler_Create_RejectsMissingFieldWithBadRequest(t *testing.T) {
	// Wires the REAL usecase (not the fake service) so this actually
	// exercises usecase.NewsService.Create's validation through the full
	// HTTP stack, mirroring product_test.go's realSvc pattern (PR4).
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeNewsRepoE2E{}
	realSvc := usecase.NewNewsService(repo)
	r := newAuthedNewsTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/news", strings.NewReader(`{"Title":"Nueva","Body":"Cuerpo"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Missing Title, Body, or Image parameter"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Missing Title, Body, or Image parameter"}`)
	}
	if repo.createCalled {
		t.Errorf("repository.Create was called for a missing Image — no persistence should happen on validation failure")
	}
}

func TestNewsHandler_Create_TreatsMalformedJSONBodyAsEmptyRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeNewsRepoE2E{}
	realSvc := usecase.NewNewsService(repo)
	r := newAuthedNewsTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/news", strings.NewReader(`not json`))
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

func TestNewsHandler_Create_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeNewsService{createErr: errors.New("db down")}
	r := newAuthedNewsTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/news", strings.NewReader(`{"Title":"Nueva","Body":"Cuerpo","Image":"i.jpg"}`))
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

// --- Update (PUT /api/news/:id) ---

func TestNewsHandler_Update_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	svc := &fakeNewsService{}
	r := newAuthedNewsTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/news/1", strings.NewReader(`{"Title":"Renombrada","Body":"Cuerpo","Image":"i.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if svc.updateItem.Title != "" {
		t.Errorf("Update() should not have been called without a JWT, but persisted %+v", svc.updateItem)
	}
}

func TestNewsHandler_Update_RejectsInvalidOrExpiredJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", -time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeNewsService{}
	r := newAuthedNewsTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/news/1", strings.NewReader(`{"Title":"Renombrada","Body":"Cuerpo","Image":"i.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if svc.updateItem.Title != "" {
		t.Errorf("Update() should not have been called with an expired JWT, but persisted %+v", svc.updateItem)
	}
}

func TestNewsHandler_Update_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeNewsService{}
	r := newAuthedNewsTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/news/1", strings.NewReader(`{"Title":"Renombrada","Body":"Cuerpo","Image":"i.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"News updated successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"News updated successfully"}`)
	}
	if svc.updateID != 1 || svc.updateItem.Title != "Renombrada" {
		t.Errorf("Update() called with id=%d item=%+v, want id=1 Title=Renombrada", svc.updateID, svc.updateItem)
	}
}

func TestNewsHandler_Update_RejectsMissingFieldWithBadRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeNewsRepoE2E{}
	realSvc := usecase.NewNewsService(repo)
	r := newAuthedNewsTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/news/1", strings.NewReader(`{"Title":"","Body":"Cuerpo","Image":"i.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Missing Title, Body, or Image parameter"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Missing Title, Body, or Image parameter"}`)
	}
	if repo.updateCalled {
		t.Errorf("repository.Update was called for a missing Title — no persistence should happen on validation failure")
	}
}

func TestNewsHandler_Update_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeNewsService{updateErr: errors.New("db down")}
	r := newAuthedNewsTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/news/1", strings.NewReader(`{"Title":"Renombrada","Body":"Cuerpo","Image":"i.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

// TestNewsHandler_Update_TreatsNonNumericIDAsNoOpSuccess mirrors
// TestProductHandler_Update_TreatsNonNumericIDAsNoOpSuccess (PR4 corrective)
// for news — see backend/docs/legacy-quirks.md §10.
func TestNewsHandler_Update_TreatsNonNumericIDAsNoOpSuccess(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeNewsRepoE2E{}
	realSvc := usecase.NewNewsService(repo)
	r := newAuthedNewsTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/news/abc", strings.NewReader(`{"Title":"Fantasma","Body":"Cuerpo","Image":"i.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"News updated successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"News updated successfully"}`)
	}
	if !repo.updateCalled || repo.updateID != 0 {
		t.Errorf("repository.Update called=%v with id=%d, want called=true id=0 (parseLegacyID(\"abc\") == 0)", repo.updateCalled, repo.updateID)
	}
}

// --- Delete (DELETE /api/news/:id) ---

func TestNewsHandler_Delete_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	svc := &fakeNewsService{}
	r := newAuthedNewsTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/news/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if svc.deleteID != 0 {
		t.Errorf("Delete() should not have been called without a JWT")
	}
}

func TestNewsHandler_Delete_RejectsInvalidOrExpiredJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", -time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeNewsService{}
	r := newAuthedNewsTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/news/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if svc.deleteID != 0 {
		t.Errorf("Delete() should not have been called with an expired JWT")
	}
}

func TestNewsHandler_Delete_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeNewsService{}
	r := newAuthedNewsTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/news/7", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"News deleted successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"News deleted successfully"}`)
	}
	if svc.deleteID != 7 {
		t.Errorf("Delete() called with id=%d, want 7", svc.deleteID)
	}
}

func TestNewsHandler_Delete_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeNewsService{deleteErr: errors.New("db down")}
	r := newAuthedNewsTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/news/7", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

// TestNewsHandler_Delete_TreatsNonNumericIDAsNoOpSuccess mirrors
// TestProductHandler_Delete_TreatsNonNumericIDAsNoOpSuccess (PR4
// corrective) for news — see backend/docs/legacy-quirks.md §10.
func TestNewsHandler_Delete_TreatsNonNumericIDAsNoOpSuccess(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeNewsRepoE2E{}
	realSvc := usecase.NewNewsService(repo)
	r := newAuthedNewsTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/news/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"News deleted successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"News deleted successfully"}`)
	}
	if !repo.deleteCalled || repo.deleteID != 0 {
		t.Errorf("repository.Delete called=%v with id=%d, want called=true id=0 (parseLegacyID(\"abc\") == 0)", repo.deleteCalled, repo.deleteID)
	}
}
