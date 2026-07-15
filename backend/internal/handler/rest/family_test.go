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

// fakeFamilyRepoE2E implements domain.FamilyRepository (not
// rest.FamilyService) so end-to-end validation tests can wire the REAL
// usecase.FamilyService through the real HTTP stack, mirroring the realSvc
// pattern established in admin_test.go: a fake at the SERVICE layer would
// bypass usecase-level validation, since usecase.FamilyService.Create/Update
// is what rejects missing Body/Category and the Category enum — not the
// handler.
type fakeFamilyRepoE2E struct {
	createCalled bool
	updateCalled bool
}

func (f *fakeFamilyRepoE2E) List(ctx context.Context) ([]domain.FamilyItem, error) {
	return nil, nil
}

func (f *fakeFamilyRepoE2E) Get(ctx context.Context, id int) (domain.FamilyItem, error) {
	return domain.FamilyItem{}, domain.ErrNotFound
}

func (f *fakeFamilyRepoE2E) Create(ctx context.Context, item domain.FamilyItem) (domain.FamilyItem, error) {
	f.createCalled = true
	return item, nil
}

func (f *fakeFamilyRepoE2E) Update(ctx context.Context, id int, item domain.FamilyItem) error {
	f.updateCalled = true
	return nil
}

func (f *fakeFamilyRepoE2E) Delete(ctx context.Context, id int) error {
	return nil
}

// fakeFamilyService uses pointer receivers (unlike the read-only PR2 fake)
// so Create/Update/Delete tests can record what was actually passed to the
// service, mirroring the fakeAdminService convention established in PR3.
type fakeFamilyService struct {
	items []domain.FamilyItem
	byID  map[int]domain.FamilyItem
	err   error

	createErr    error
	createdItem  domain.FamilyItem
	updateErr    error
	updateCalled bool
	updateID     int
	updateItem   domain.FamilyItem
	deleteErr    error
	deleteID     int
}

func (f *fakeFamilyService) List(ctx context.Context) ([]domain.FamilyItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f *fakeFamilyService) Get(ctx context.Context, id int) (domain.FamilyItem, error) {
	if f.err != nil {
		return domain.FamilyItem{}, f.err
	}
	item, ok := f.byID[id]
	if !ok {
		return domain.FamilyItem{}, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakeFamilyService) Create(ctx context.Context, item domain.FamilyItem) (domain.FamilyItem, error) {
	if f.createErr != nil {
		return domain.FamilyItem{}, f.createErr
	}
	f.createdItem = item
	item.ID = 9001
	return item, nil
}

func (f *fakeFamilyService) Update(ctx context.Context, id int, item domain.FamilyItem) error {
	f.updateCalled = true
	f.updateID = id
	f.updateItem = item
	return f.updateErr
}

func (f *fakeFamilyService) Delete(ctx context.Context, id int) error {
	f.deleteID = id
	return f.deleteErr
}

func newFamilyTestRouter(svc rest.FamilyService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewFamilyHandler(svc)
	r.GET("/api/family", h.List)
	r.GET("/api/family/:id", h.Get)
	r.POST("/api/family", h.Create)
	r.PUT("/api/family/:id", h.Update)
	r.DELETE("/api/family/:id", h.Delete)
	return r
}

// newAuthedFamilyTestRouter wires the same write routes behind
// middleware.RequireAuth, mirroring the real router.go wiring (task 3.1).
func newAuthedFamilyTestRouter(svc rest.FamilyService, jwtSvc *auth.JWTService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewFamilyHandler(svc)
	r.GET("/api/family", h.List)
	r.GET("/api/family/:id", h.Get)
	r.POST("/api/family", middleware.RequireAuth(jwtSvc), h.Create)
	r.PUT("/api/family/:id", middleware.RequireAuth(jwtSvc), h.Update)
	r.DELETE("/api/family/:id", middleware.RequireAuth(jwtSvc), h.Delete)
	return r
}

func TestFamilyHandler_List_ReturnsArray(t *testing.T) {
	svc := &fakeFamilyService{items: []domain.FamilyItem{{ID: 3, Body: "aqui va la descripcion 3", Category: "Taller protegido"}}}
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
	r := newFamilyTestRouter(&fakeFamilyService{items: []domain.FamilyItem{}})

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
	svc := &fakeFamilyService{byID: map[int]domain.FamilyItem{3: {ID: 3, Body: "aqui va la descripcion 3", Category: "Taller protegido"}}}
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
	r := newFamilyTestRouter(&fakeFamilyService{byID: map[int]domain.FamilyItem{}})

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
	r := newFamilyTestRouter(&fakeFamilyService{byID: map[int]domain.FamilyItem{}})

	req := httptest.NewRequest(http.MethodGet, "/api/family/abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestFamilyHandler_List_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newFamilyTestRouter(&fakeFamilyService{err: errors.New("db down")})

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
	r := newFamilyTestRouter(&fakeFamilyService{err: errors.New("db down")})

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

// --- Create (POST /api/family) ---

func TestFamilyHandler_Create_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAuthedFamilyTestRouter(&fakeFamilyService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/family", strings.NewReader(`{"Body":"cuerpo","Category":"Centro de dia"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestFamilyHandler_Create_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeFamilyService{}
	r := newAuthedFamilyTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/family", strings.NewReader(`{"Body":"cuerpo","Category":"Centro de dia"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Workshops added successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Workshops added successfully"}`)
	}
	if svc.createdItem.Body != "cuerpo" {
		t.Errorf("Create() persisted Body = %q, want %q", svc.createdItem.Body, "cuerpo")
	}
}

func TestFamilyHandler_Create_RejectsInvalidCategoryWithBadRequest(t *testing.T) {
	// Spec requirement: Category MUST be "Centro de dia" or "Taller
	// protegido"; anything else -> 400, no record created. Wires the REAL
	// usecase (not the fake service) so this actually exercises the
	// Category enum validation through the full HTTP stack.
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeFamilyRepoE2E{}
	realSvc := usecase.NewFamilyService(repo)
	r := newAuthedFamilyTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/family", strings.NewReader(`{"Body":"cuerpo","Category":"Invalid"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.createCalled {
		t.Errorf("repository.Create was called for an invalid category — no persistence should happen on validation failure")
	}
}

func TestFamilyHandler_Create_RejectsMissingBodyOrCategoryWithBadRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeFamilyRepoE2E{}
	realSvc := usecase.NewFamilyService(repo)
	r := newAuthedFamilyTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/family", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.createCalled {
		t.Errorf("repository.Create was called for a missing Body/Category — no persistence should happen on validation failure")
	}
}

func TestFamilyHandler_Create_TreatsMalformedJSONBodyAsEmptyRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeFamilyRepoE2E{}
	realSvc := usecase.NewFamilyService(repo)
	r := newAuthedFamilyTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/family", strings.NewReader(`not json`))
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

func TestFamilyHandler_Create_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeFamilyService{createErr: errors.New("db down")}
	r := newAuthedFamilyTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/family", strings.NewReader(`{"Body":"cuerpo","Category":"Centro de dia"}`))
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

// --- Update (PUT /api/family/:id) ---

func TestFamilyHandler_Update_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAuthedFamilyTestRouter(&fakeFamilyService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/family/3", strings.NewReader(`{"Body":"renombrado","Category":"Taller protegido"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestFamilyHandler_Update_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeFamilyService{}
	r := newAuthedFamilyTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/family/3", strings.NewReader(`{"Body":"renombrado","Category":"Taller protegido"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.updateID != 3 || svc.updateItem.Body != "renombrado" {
		t.Errorf("Update() called with id=%d item=%+v, want id=3 Body=renombrado", svc.updateID, svc.updateItem)
	}
}

func TestFamilyHandler_Update_RejectsInvalidCategoryWithBadRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeFamilyRepoE2E{}
	realSvc := usecase.NewFamilyService(repo)
	r := newAuthedFamilyTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/family/3", strings.NewReader(`{"Body":"cuerpo","Category":"Invalid"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.updateCalled {
		t.Errorf("repository.Update was called for an invalid category — no persistence should happen on validation failure")
	}
}

func TestFamilyHandler_Update_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeFamilyService{updateErr: errors.New("db down")}
	r := newAuthedFamilyTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/family/3", strings.NewReader(`{"Body":"cuerpo","Category":"Centro de dia"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

// --- Delete (DELETE /api/family/:id) ---

func TestFamilyHandler_Delete_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAuthedFamilyTestRouter(&fakeFamilyService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/family/3", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestFamilyHandler_Delete_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeFamilyService{}
	r := newAuthedFamilyTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/family/7", nil)
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

func TestFamilyHandler_Delete_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeFamilyService{deleteErr: errors.New("db down")}
	r := newAuthedFamilyTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/family/7", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}
