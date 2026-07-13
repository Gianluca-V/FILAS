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

type fakeAdminService struct {
	admins       []domain.Admin
	byID         map[int]domain.Admin
	err          error
	createErr    error
	createdAdmin domain.Admin
	updateErr    error
	updateCalled bool
	updateID     int
	updateUser   string
	updatePass   string
	deleteErr    error
	deleteID     int
}

func (f *fakeAdminService) List(ctx context.Context) ([]domain.Admin, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.admins, nil
}

func (f *fakeAdminService) Get(ctx context.Context, id int) (domain.Admin, error) {
	if f.err != nil {
		return domain.Admin{}, f.err
	}
	a, ok := f.byID[id]
	if !ok {
		return domain.Admin{}, domain.ErrNotFound
	}
	return a, nil
}

func (f *fakeAdminService) Create(ctx context.Context, username, password string) (domain.Admin, error) {
	if f.createErr != nil {
		return domain.Admin{}, f.createErr
	}
	f.createdAdmin = domain.Admin{Username: username, Password: password}
	return domain.Admin{ID: 9001, Username: username}, nil
}

func (f *fakeAdminService) Update(ctx context.Context, id int, username, password string) error {
	f.updateCalled = true
	f.updateID, f.updateUser, f.updatePass = id, username, password
	return f.updateErr
}

func (f *fakeAdminService) Delete(ctx context.Context, id int) error {
	f.deleteID = id
	return f.deleteErr
}

type fakeAuthService struct {
	token            string
	err              error
	gotUser, gotPass string
}

func (f *fakeAuthService) Login(ctx context.Context, username, password string) (string, error) {
	f.gotUser, f.gotPass = username, password
	if f.err != nil {
		return "", f.err
	}
	return f.token, nil
}

func newAdminTestRouter(adminSvc rest.AdminService, authSvc rest.AuthService, jwtSvc *auth.JWTService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewAdminHandler(adminSvc, authSvc, jwtSvc)
	r.POST("/api/admins", h.Create)
	r.GET("/api/admins", middleware.RequireAuth(jwtSvc), h.List)
	r.GET("/api/admins/:id", middleware.RequireAuth(jwtSvc), h.Get)
	r.PUT("/api/admins/:id", middleware.RequireAuth(jwtSvc), h.Update)
	r.DELETE("/api/admins/:id", middleware.RequireAuth(jwtSvc), h.Delete)
	return r
}

func validToken(t *testing.T, jwtSvc *auth.JWTService) string {
	t.Helper()
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	return token
}

// --- Login (POST with login:true) ---

func TestAdminHandler_Create_Login_ReturnsTokenOnSuccess(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	authSvc := &fakeAuthService{token: "signed-jwt"}
	r := newAdminTestRouter(&fakeAdminService{}, authSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/admins", strings.NewReader(`{"login":true,"username":"FilasAdmin","password":"filas-local-dev-2026"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `{"token":"signed-jwt"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"token":"signed-jwt"}`)
	}
	if authSvc.gotUser != "FilasAdmin" || authSvc.gotPass != "filas-local-dev-2026" {
		t.Errorf("Login called with (%q, %q), want (FilasAdmin, filas-local-dev-2026)", authSvc.gotUser, authSvc.gotPass)
	}
}

func TestAdminHandler_Create_Login_ReturnsUnauthorizedOnFailure(t *testing.T) {
	// Legacy contract, characterized live: wrong credentials -> 401
	// {"message":"Login failed"}, no token.
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	authSvc := &fakeAuthService{err: domain.ErrUnauthorized}
	r := newAdminTestRouter(&fakeAdminService{}, authSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/admins", strings.NewReader(`{"login":true,"username":"FilasAdmin","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Login failed"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Login failed"}`)
	}
}

func TestAdminHandler_Create_Login_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	authSvc := &fakeAuthService{err: errors.New("db down")}
	r := newAdminTestRouter(&fakeAdminService{}, authSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/admins", strings.NewReader(`{"login":true,"username":"x","password":"y"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

// --- Create (POST without login:true) ---

func TestAdminHandler_Create_RejectsMissingAuthorizationHeader(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAdminTestRouter(&fakeAdminService{}, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/admins", strings.NewReader(`{"username":"newadmin","password":"pw"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestAdminHandler_Create_RejectsInvalidToken(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAdminTestRouter(&fakeAdminService{}, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/admins", strings.NewReader(`{"username":"newadmin","password":"pw"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestAdminHandler_Create_RejectsMissingUsernameOrPassword(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	r := newAdminTestRouter(&fakeAdminService{}, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/admins", strings.NewReader(`{"username":"newadmin"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Missing username or password parameters"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Missing username or password parameters"}`)
	}
}

func TestAdminHandler_Create_ReturnsSuccessMessageWithValidTokenAndFields(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	svc := &fakeAdminService{}
	r := newAdminTestRouter(svc, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/admins", strings.NewReader(`{"username":"newadmin","password":"raw-pw"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"User created successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"User created successfully"}`)
	}
	if svc.createdAdmin.Username != "newadmin" {
		t.Errorf("Create called with username = %q, want %q", svc.createdAdmin.Username, "newadmin")
	}
}

func TestAdminHandler_Create_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	// Non-login branch: adminSvc.Create failing (e.g. DB down), NOT
	// authSvc.Login. Restores the PR2-sibling convention of a 500-path test
	// per handler method (gate #36 reliability finding).
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	svc := &fakeAdminService{createErr: errors.New("db down")}
	r := newAdminTestRouter(svc, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/admins", strings.NewReader(`{"username":"newadmin","password":"raw-pw"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

func TestAdminHandler_Create_TreatsMalformedJSONBodyAsEmptyRequest(t *testing.T) {
	// Legacy: json_decode(...) === null on malformed input, then the
	// null-safe isset() checks fall through to the same 400/401 paths a
	// well-formed-but-incomplete body would take. c.ShouldBindJSON's error
	// is deliberately discarded (see AdminHandler.Create doc comment) so a
	// genuinely invalid/truncated body degrades the same way: no token ->
	// 401 (login=false zero value, non-login branch, auth checked first).
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAdminTestRouter(&fakeAdminService{}, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/admins", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestAdminHandler_Create_TreatsTruncatedJSONBodyAsEmptyRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	svc := &fakeAdminService{}
	r := newAdminTestRouter(svc, &fakeAuthService{}, jwtSvc)

	// Truncated mid-object: not valid JSON.
	req := httptest.NewRequest(http.MethodPost, "/api/admins", strings.NewReader(`{"username":"newadmin"`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Auth passes (valid token present), body decodes to zero-value req ->
	// falls through to the missing-fields 400, same as a well-formed empty
	// body would.
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if svc.createdAdmin.Username != "" {
		t.Errorf("Create should not have been called for a truncated body, got %+v", svc.createdAdmin)
	}
}

// --- List (GET /api/admins) ---

func TestAdminHandler_List_RejectsMissingAuthorizationHeader(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAdminTestRouter(&fakeAdminService{}, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/admins", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestAdminHandler_List_ExcludesPasswordAndSaltWithValidToken(t *testing.T) {
	// Legacy leaked password+salt via SELECT * — characterized live against
	// FilasServer/admins.php (see backend/docs/legacy-quirks.md §1). Go
	// deliberately fixes this: AdminResponse never carries them.
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	svc := &fakeAdminService{admins: []domain.Admin{
		{ID: 1543, Username: "FilasAdmin", Password: "should-never-appear", Salt: "should-never-appear-either"},
	}}
	r := newAdminTestRouter(svc, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/admins", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `[{"ID":"1543","username":"FilasAdmin"}]`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
	if strings.Contains(rec.Body.String(), "should-never-appear") {
		t.Errorf("body leaked password/salt: %s", rec.Body.String())
	}
}

func TestAdminHandler_List_ReturnsEmptyArrayWhenNoAdmins(t *testing.T) {
	// Legacy getAdmins(): num_rows==0 -> 200 [] (NOT 404, unlike
	// news/gallery/family/organizations — same pattern as products.php).
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	r := newAdminTestRouter(&fakeAdminService{admins: []domain.Admin{}}, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/admins", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `[]` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `[]`)
	}
}

func TestAdminHandler_List_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	r := newAdminTestRouter(&fakeAdminService{err: errors.New("db down")}, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/admins", nil)
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

// --- Get (GET /api/admins/:id) ---

func TestAdminHandler_Get_ReturnsBareObjectExcludingPasswordAndSalt(t *testing.T) {
	// Legacy getUser() returns a bare object (own dedicated function, not
	// the array-wrap quirk shared by news/gallery/family/organizations).
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	svc := &fakeAdminService{byID: map[int]domain.Admin{
		1543: {ID: 1543, Username: "FilasAdmin", Password: "should-never-appear", Salt: "nope"},
	}}
	r := newAdminTestRouter(svc, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/admins/1543", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `{"ID":"1543","username":"FilasAdmin"}`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestAdminHandler_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	r := newAdminTestRouter(&fakeAdminService{byID: map[int]domain.Admin{}}, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/admins/999999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"User not found"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"User not found"}`)
	}
}

func TestAdminHandler_Get_RejectsMissingAuthorizationHeader(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAdminTestRouter(&fakeAdminService{}, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/admins/1543", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestAdminHandler_Get_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	r := newAdminTestRouter(&fakeAdminService{err: errors.New("db down")}, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/admins/1543", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

// --- Update (PUT /api/admins/:id) ---

func TestAdminHandler_Update_RejectsMissingAuthorizationHeader(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAdminTestRouter(&fakeAdminService{}, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/admins/5", strings.NewReader(`{"username":"x","password":"y"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestAdminHandler_Update_ReturnsSuccessMessageWithValidToken(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	svc := &fakeAdminService{}
	r := newAdminTestRouter(svc, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/admins/5", strings.NewReader(`{"username":"renamed","password":"newpw"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"User updated successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"User updated successfully"}`)
	}
	if svc.updateID != 5 || svc.updateUser != "renamed" || svc.updatePass != "newpw" {
		t.Errorf("Update called with (%d, %q, %q), want (5, renamed, newpw)", svc.updateID, svc.updateUser, svc.updatePass)
	}
}

func TestAdminHandler_Update_ForwardsValidationErrorAs400(t *testing.T) {
	// Confirms the handler forwards usecase.AdminService's blocker-fix
	// validation error (domain.ErrValidation) through the generic
	// middleware.ErrorHandler mapping, rather than swallowing it as a 500.
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	svc := &fakeAdminService{updateErr: domain.ErrValidation}
	r := newAdminTestRouter(svc, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/admins/5", strings.NewReader(`{"username":"renamed","password":""}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestAdminHandler_Update_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	svc := &fakeAdminService{updateErr: errors.New("db down")}
	r := newAdminTestRouter(svc, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/admins/5", strings.NewReader(`{"username":"renamed","password":"newpw"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

// fakeAdminRepoE2E is a minimal domain.AdminRepository used ONLY to prove
// the blocker fix end-to-end: a real usecase.AdminService (not the
// fakeAdminService test double used everywhere else in this file) wired
// through the real router, so the validation logic that actually lives in
// the usecase layer is exercised by an HTTP request.
type fakeAdminRepoE2E struct {
	updateCalled         bool
	updatePasswordCalled bool
}

func (f *fakeAdminRepoE2E) List(ctx context.Context) ([]domain.Admin, error) { return nil, nil }
func (f *fakeAdminRepoE2E) Get(ctx context.Context, id int) (domain.Admin, error) {
	return domain.Admin{}, nil
}
func (f *fakeAdminRepoE2E) GetByUsername(ctx context.Context, username string) (domain.Admin, error) {
	return domain.Admin{}, domain.ErrNotFound
}
func (f *fakeAdminRepoE2E) Create(ctx context.Context, a domain.Admin) (domain.Admin, error) {
	return a, nil
}
func (f *fakeAdminRepoE2E) Update(ctx context.Context, id int, username, passwordHash string) error {
	f.updateCalled = true
	return nil
}
func (f *fakeAdminRepoE2E) Delete(ctx context.Context, id int) error { return nil }
func (f *fakeAdminRepoE2E) UpdatePassword(ctx context.Context, id int, oldHash, newHash string) error {
	f.updatePasswordCalled = true
	return nil
}

// TestAdminHandler_Update_EndToEnd_RejectsOmittedPasswordWithoutPersisting
// is the blocker fix (obs #36) proven through the full HTTP stack: PUT
// /api/admins/:id with a body that OMITS the password field must be
// rejected 400 and must NEVER reach the repository — closing the
// account-takeover path where auth.HashPassword("") would otherwise
// produce a valid, persistable bcrypt hash of the empty string.
func TestAdminHandler_Update_EndToEnd_RejectsOmittedPasswordWithoutPersisting(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	repo := &fakeAdminRepoE2E{}
	realSvc := usecase.NewAdminService(repo)
	r := newAdminTestRouter(realSvc, &fakeAuthService{}, jwtSvc)

	// Body omits "password" entirely (not even an empty string key).
	req := httptest.NewRequest(http.MethodPut, "/api/admins/5", strings.NewReader(`{"username":"renamed"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.updateCalled {
		t.Errorf("repository.Update was called for an omitted-password request — account-takeover path is NOT closed")
	}
	if repo.updatePasswordCalled {
		t.Errorf("repository.UpdatePassword was called for an omitted-password request")
	}
}

func TestAdminHandler_Update_EndToEnd_RejectsExplicitEmptyPasswordWithoutPersisting(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	repo := &fakeAdminRepoE2E{}
	realSvc := usecase.NewAdminService(repo)
	r := newAdminTestRouter(realSvc, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/admins/5", strings.NewReader(`{"username":"renamed","password":""}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.updateCalled {
		t.Errorf("repository.Update was called for an explicit empty-password request")
	}
}

// TestAdminHandler_InlineAuthAndMiddlewareAuth_ProduceIdenticalErrorBodies
// locks the shared bearer-auth helper (gate #36 readability finding): POST
// /api/admins (non-login) uses AdminHandler.requireAuthInline while
// GET/PUT/DELETE use middleware.RequireAuth via route-group middleware —
// two call sites that used to duplicate the header-extraction/prefix-trim
// logic and the two 401 message strings by hand. This test proves both
// call sites now emit BYTE-IDENTICAL bodies for the same failure, which is
// only possible if they share one implementation (middleware.
// AuthenticateRequest) — a hand-copied duplicate would be free to drift.
func TestAdminHandler_InlineAuthAndMiddlewareAuth_ProduceIdenticalErrorBodies(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAdminTestRouter(&fakeAdminService{}, &fakeAuthService{}, jwtSvc)

	t.Run("missing header", func(t *testing.T) {
		inlineReq := httptest.NewRequest(http.MethodPost, "/api/admins", strings.NewReader(`{"username":"x","password":"y"}`))
		inlineReq.Header.Set("Content-Type", "application/json")
		inlineRec := httptest.NewRecorder()
		r.ServeHTTP(inlineRec, inlineReq)

		middlewareReq := httptest.NewRequest(http.MethodGet, "/api/admins", nil)
		middlewareRec := httptest.NewRecorder()
		r.ServeHTTP(middlewareRec, middlewareReq)

		if inlineRec.Code != middlewareRec.Code {
			t.Errorf("status mismatch: inline=%d middleware=%d", inlineRec.Code, middlewareRec.Code)
		}
		if inlineRec.Body.String() != middlewareRec.Body.String() {
			t.Errorf("body mismatch: inline=%q middleware=%q", inlineRec.Body.String(), middlewareRec.Body.String())
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		inlineReq := httptest.NewRequest(http.MethodPost, "/api/admins", strings.NewReader(`{"username":"x","password":"y"}`))
		inlineReq.Header.Set("Content-Type", "application/json")
		inlineReq.Header.Set("Authorization", "Bearer garbage.garbage.garbage")
		inlineRec := httptest.NewRecorder()
		r.ServeHTTP(inlineRec, inlineReq)

		middlewareReq := httptest.NewRequest(http.MethodGet, "/api/admins", nil)
		middlewareReq.Header.Set("Authorization", "Bearer garbage.garbage.garbage")
		middlewareRec := httptest.NewRecorder()
		r.ServeHTTP(middlewareRec, middlewareReq)

		if inlineRec.Code != middlewareRec.Code {
			t.Errorf("status mismatch: inline=%d middleware=%d", inlineRec.Code, middlewareRec.Code)
		}
		if inlineRec.Body.String() != middlewareRec.Body.String() {
			t.Errorf("body mismatch: inline=%q middleware=%q", inlineRec.Body.String(), middlewareRec.Body.String())
		}
	})
}

// --- Delete (DELETE /api/admins/:id) ---

func TestAdminHandler_Delete_RejectsMissingAuthorizationHeader(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAdminTestRouter(&fakeAdminService{}, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/admins/5", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestAdminHandler_Delete_ReturnsSuccessMessageWithValidToken(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	svc := &fakeAdminService{}
	r := newAdminTestRouter(svc, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/admins/7", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"User deleted successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"User deleted successfully"}`)
	}
	if svc.deleteID != 7 {
		t.Errorf("Delete called with id = %d, want 7", svc.deleteID)
	}
}

func TestAdminHandler_Delete_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token := validToken(t, jwtSvc)
	svc := &fakeAdminService{deleteErr: errors.New("db down")}
	r := newAdminTestRouter(svc, &fakeAuthService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/admins/7", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}
