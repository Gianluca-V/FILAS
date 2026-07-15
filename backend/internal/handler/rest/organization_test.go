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

// fakeOrganizationRepoE2E implements domain.OrganizationRepository (not
// rest.OrganizationService) so end-to-end validation tests can wire the
// REAL usecase.OrganizationService through the real HTTP stack, mirroring
// the realSvc pattern established in admin_test.go: a fake at the SERVICE
// layer would bypass usecase-level validation, since
// usecase.OrganizationService.Create/Update is what rejects missing
// Title/Image — not the handler.
type fakeOrganizationRepoE2E struct {
	createCalled bool
	updateCalled bool
}

func (f *fakeOrganizationRepoE2E) List(ctx context.Context) ([]domain.Organization, error) {
	return nil, nil
}

func (f *fakeOrganizationRepoE2E) Get(ctx context.Context, id int) (domain.Organization, error) {
	return domain.Organization{}, domain.ErrNotFound
}

func (f *fakeOrganizationRepoE2E) Create(ctx context.Context, o domain.Organization) (domain.Organization, error) {
	f.createCalled = true
	return o, nil
}

func (f *fakeOrganizationRepoE2E) Update(ctx context.Context, id int, o domain.Organization) error {
	f.updateCalled = true
	return nil
}

func (f *fakeOrganizationRepoE2E) Delete(ctx context.Context, id int) error {
	return nil
}

// fakeOrganizationService uses pointer receivers (unlike the read-only PR2
// fake) so Create/Update/Delete tests can record what was actually passed
// to the service, mirroring the fakeAdminService convention established in
// PR3. This is the new write-path production code for PR4 (task 3.1) —
// organizations had no handler write paths at all before this change.
type fakeOrganizationService struct {
	items []domain.Organization
	byID  map[int]domain.Organization
	err   error

	createErr    error
	createdItem  domain.Organization
	updateErr    error
	updateCalled bool
	updateID     int
	updateItem   domain.Organization
	deleteErr    error
	deleteID     int
}

func (f *fakeOrganizationService) List(ctx context.Context) ([]domain.Organization, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f *fakeOrganizationService) Get(ctx context.Context, id int) (domain.Organization, error) {
	if f.err != nil {
		return domain.Organization{}, f.err
	}
	item, ok := f.byID[id]
	if !ok {
		return domain.Organization{}, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakeOrganizationService) Create(ctx context.Context, o domain.Organization) (domain.Organization, error) {
	if f.createErr != nil {
		return domain.Organization{}, f.createErr
	}
	f.createdItem = o
	o.ID = 9001
	return o, nil
}

func (f *fakeOrganizationService) Update(ctx context.Context, id int, o domain.Organization) error {
	f.updateCalled = true
	f.updateID = id
	f.updateItem = o
	return f.updateErr
}

func (f *fakeOrganizationService) Delete(ctx context.Context, id int) error {
	f.deleteID = id
	return f.deleteErr
}

func newOrganizationTestRouter(svc rest.OrganizationService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewOrganizationHandler(svc)
	r.GET("/api/organizations", h.List)
	r.GET("/api/organizations/:id", h.Get)
	r.POST("/api/organizations", h.Create)
	r.PUT("/api/organizations/:id", h.Update)
	r.DELETE("/api/organizations/:id", h.Delete)
	return r
}

// newAuthedOrganizationTestRouter wires the same write routes behind
// middleware.RequireAuth, mirroring the real router.go wiring (task 3.1).
func newAuthedOrganizationTestRouter(svc rest.OrganizationService, jwtSvc *auth.JWTService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewOrganizationHandler(svc)
	r.GET("/api/organizations", h.List)
	r.GET("/api/organizations/:id", h.Get)
	r.POST("/api/organizations", middleware.RequireAuth(jwtSvc), h.Create)
	r.PUT("/api/organizations/:id", middleware.RequireAuth(jwtSvc), h.Update)
	r.DELETE("/api/organizations/:id", middleware.RequireAuth(jwtSvc), h.Delete)
	return r
}

func TestOrganizationHandler_List_ReturnsArray(t *testing.T) {
	desc := "Descripcion de prueba 1"
	svc := &fakeOrganizationService{items: []domain.Organization{
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
	r := newOrganizationTestRouter(&fakeOrganizationService{items: []domain.Organization{}})

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
	svc := &fakeOrganizationService{byID: map[int]domain.Organization{1: {ID: 1, Title: "organizacions de prueba 1", Image: "https://example.com/1.jpg"}}}
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
	r := newOrganizationTestRouter(&fakeOrganizationService{byID: map[int]domain.Organization{}})

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
	r := newOrganizationTestRouter(&fakeOrganizationService{byID: map[int]domain.Organization{}})

	req := httptest.NewRequest(http.MethodGet, "/api/organizations/abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestOrganizationHandler_List_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newOrganizationTestRouter(&fakeOrganizationService{err: errors.New("db down")})

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
	r := newOrganizationTestRouter(&fakeOrganizationService{err: errors.New("db down")})

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

// --- Create (POST /api/organizations) ---

func TestOrganizationHandler_Create_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAuthedOrganizationTestRouter(&fakeOrganizationService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/organizations", strings.NewReader(`{"Title":"Nueva","Image":"https://example.com/n.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestOrganizationHandler_Create_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeOrganizationService{}
	r := newAuthedOrganizationTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/organizations", strings.NewReader(`{"Title":"Nueva","Image":"https://example.com/n.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Organizations added successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Organizations added successfully"}`)
	}
	if svc.createdItem.Title != "Nueva" {
		t.Errorf("Create() persisted Title = %q, want %q", svc.createdItem.Title, "Nueva")
	}
}

func TestOrganizationHandler_Create_RejectsMissingTitleOrImageWithBadRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeOrganizationRepoE2E{}
	realSvc := usecase.NewOrganizationService(repo)
	r := newAuthedOrganizationTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/organizations", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.createCalled {
		t.Errorf("repository.Create was called for a missing Title/Image — no persistence should happen on validation failure")
	}
}

func TestOrganizationHandler_Create_TreatsMalformedJSONBodyAsEmptyRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeOrganizationRepoE2E{}
	realSvc := usecase.NewOrganizationService(repo)
	r := newAuthedOrganizationTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/organizations", strings.NewReader(`not json`))
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

func TestOrganizationHandler_Create_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeOrganizationService{createErr: errors.New("db down")}
	r := newAuthedOrganizationTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/organizations", strings.NewReader(`{"Title":"Nueva","Image":"https://example.com/n.jpg"}`))
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

// --- Update (PUT /api/organizations/:id) ---

func TestOrganizationHandler_Update_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAuthedOrganizationTestRouter(&fakeOrganizationService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/organizations/1", strings.NewReader(`{"Title":"renombrada","Image":"https://example.com/r.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestOrganizationHandler_Update_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeOrganizationService{}
	r := newAuthedOrganizationTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/organizations/1", strings.NewReader(`{"Title":"renombrada","Image":"https://example.com/r.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Organizations updated successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Organizations updated successfully"}`)
	}
	if svc.updateID != 1 || svc.updateItem.Title != "renombrada" {
		t.Errorf("Update() called with id=%d item=%+v, want id=1 Title=renombrada", svc.updateID, svc.updateItem)
	}
}

func TestOrganizationHandler_Update_RejectsMissingTitleOrImageWithBadRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeOrganizationRepoE2E{}
	realSvc := usecase.NewOrganizationService(repo)
	r := newAuthedOrganizationTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/organizations/1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.updateCalled {
		t.Errorf("repository.Update was called for a missing Title/Image — no persistence should happen on validation failure")
	}
}

func TestOrganizationHandler_Update_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeOrganizationService{updateErr: errors.New("db down")}
	r := newAuthedOrganizationTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/organizations/1", strings.NewReader(`{"Title":"renombrada","Image":"https://example.com/r.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

// --- Delete (DELETE /api/organizations/:id) ---

func TestOrganizationHandler_Delete_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAuthedOrganizationTestRouter(&fakeOrganizationService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/organizations/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestOrganizationHandler_Delete_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeOrganizationService{}
	r := newAuthedOrganizationTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/organizations/7", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	// Locks the lowercase "organizations" quirk (see OrganizationHandler.Delete
	// doc comment): organizations.php's delete branch uses a different
	// literal than create/update.
	if rec.Body.String() != `{"message":"organizations deleted successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"organizations deleted successfully"}`)
	}
	if svc.deleteID != 7 {
		t.Errorf("Delete() called with id=%d, want 7", svc.deleteID)
	}
}

func TestOrganizationHandler_Delete_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeOrganizationService{deleteErr: errors.New("db down")}
	r := newAuthedOrganizationTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/organizations/7", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}
