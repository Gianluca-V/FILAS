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

// fakeProductRepoE2E implements domain.ProductRepository (not
// rest.ProductService) so end-to-end validation tests can wire the REAL
// usecase.ProductService through the real HTTP stack, mirroring the
// realSvc pattern established in admin_test.go: a fake at the SERVICE layer
// would bypass usecase-level validation entirely, since usecase.ProductService.Create
// is what rejects an empty Name — not the handler.
type fakeProductRepoE2E struct {
	createCalled bool
	updateCalled bool
}

func (f *fakeProductRepoE2E) List(ctx context.Context) ([]domain.Product, error) {
	return nil, nil
}

func (f *fakeProductRepoE2E) Get(ctx context.Context, id int) (domain.Product, error) {
	return domain.Product{}, domain.ErrNotFound
}

func (f *fakeProductRepoE2E) Create(ctx context.Context, p domain.Product) (domain.Product, error) {
	f.createCalled = true
	return p, nil
}

func (f *fakeProductRepoE2E) Update(ctx context.Context, id int, p domain.Product) error {
	f.updateCalled = true
	return nil
}

func (f *fakeProductRepoE2E) Delete(ctx context.Context, id int) error {
	return nil
}

// fakeProductService uses pointer receivers (unlike the read-only PR2 fake)
// so Create/Update/Delete tests can record what was actually passed to the
// service, mirroring the fakeAdminService convention established in PR3.
type fakeProductService struct {
	products []domain.Product
	byID     map[int]domain.Product
	err      error

	createErr      error
	createdProduct domain.Product
	updateErr      error
	updateCalled   bool
	updateID       int
	updateProduct  domain.Product
	deleteErr      error
	deleteID       int
}

func (f *fakeProductService) List(ctx context.Context) ([]domain.Product, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.products, nil
}

func (f *fakeProductService) Get(ctx context.Context, id int) (domain.Product, error) {
	if f.err != nil {
		return domain.Product{}, f.err
	}
	p, ok := f.byID[id]
	if !ok {
		return domain.Product{}, domain.ErrNotFound
	}
	return p, nil
}

func (f *fakeProductService) Create(ctx context.Context, p domain.Product) (domain.Product, error) {
	if f.createErr != nil {
		return domain.Product{}, f.createErr
	}
	f.createdProduct = p
	p.ID = 9001
	return p, nil
}

func (f *fakeProductService) Update(ctx context.Context, id int, p domain.Product) error {
	f.updateCalled = true
	f.updateID = id
	f.updateProduct = p
	return f.updateErr
}

func (f *fakeProductService) Delete(ctx context.Context, id int) error {
	f.deleteID = id
	return f.deleteErr
}

func newProductTestRouter(svc rest.ProductService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewProductHandler(svc)
	r.GET("/api/products", h.List)
	r.GET("/api/products/:id", h.Get)
	r.POST("/api/products", h.Create)
	r.PUT("/api/products/:id", h.Update)
	r.DELETE("/api/products/:id", h.Delete)
	return r
}

// newAuthedProductTestRouter wires the same write routes behind
// middleware.RequireAuth, mirroring the real router.go wiring (task 3.1),
// for tests that need to prove the 401-without-JWT / 2xx-with-JWT contract.
func newAuthedProductTestRouter(svc rest.ProductService, jwtSvc *auth.JWTService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewProductHandler(svc)
	r.GET("/api/products", h.List)
	r.GET("/api/products/:id", h.Get)
	r.POST("/api/products", middleware.RequireAuth(jwtSvc), h.Create)
	r.PUT("/api/products/:id", middleware.RequireAuth(jwtSvc), h.Update)
	r.DELETE("/api/products/:id", middleware.RequireAuth(jwtSvc), h.Delete)
	return r
}

func TestProductHandler_List_ReturnsArray(t *testing.T) {
	image := "assets/default-img.png"
	svc := &fakeProductService{products: []domain.Product{
		{ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112, Image: &image},
	}}
	r := newProductTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `[{"ID":"1","Name":"Mermelada de pera","Price":"600","Stock":"112","Image":"assets/default-img.png","Description":null}]`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestProductHandler_List_ReturnsEmptyArrayWhenNoProducts(t *testing.T) {
	// Legacy quirk (characterized live, obs #29/#31): unlike the other four
	// resources, products.php returns 200 [] on an empty table, NOT 404.
	// products.php has dedicated getProducts()/getProduct() functions that
	// don't share the generic "num_rows==0 -> 404" code path the others do.
	r := newProductTestRouter(&fakeProductService{products: []domain.Product{}})

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `[]` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `[]`)
	}
}

func TestProductHandler_Get_ReturnsBareObjectNotArray(t *testing.T) {
	// Legacy quirk: unlike news/gallery/family/organizations, a found
	// product is a bare JSON object, not wrapped in an array.
	image := "assets/default-img.png"
	svc := &fakeProductService{byID: map[int]domain.Product{1: {ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112, Image: &image}}}
	r := newProductTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/products/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `{"ID":"1","Name":"Mermelada de pera","Price":"600","Stock":"112","Image":"assets/default-img.png","Description":null}`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestProductHandler_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	r := newProductTestRouter(&fakeProductService{byID: map[int]domain.Product{}})

	req := httptest.NewRequest(http.MethodGet, "/api/products/999999", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Product not found"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Product not found"}`)
	}
}

func TestProductHandler_Get_TreatsNonNumericIDAsNotFound(t *testing.T) {
	r := newProductTestRouter(&fakeProductService{byID: map[int]domain.Product{}})

	req := httptest.NewRequest(http.MethodGet, "/api/products/abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestProductHandler_List_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newProductTestRouter(&fakeProductService{err: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"internal server error"}` {
		t.Errorf("body = %s, want the generic ErrorHandler body, not a leaked error", rec.Body.String())
	}
}

func TestProductHandler_Get_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	r := newProductTestRouter(&fakeProductService{err: errors.New("db down")})

	req := httptest.NewRequest(http.MethodGet, "/api/products/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"internal server error"}` {
		t.Errorf("body = %s, want the generic ErrorHandler body, not a leaked error", rec.Body.String())
	}
}

// --- Create (POST /api/products) ---

func TestProductHandler_Create_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAuthedProductTestRouter(&fakeProductService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(`{"Name":"Nuevo","Price":100,"Stock":5}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestProductHandler_Create_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeProductService{}
	r := newAuthedProductTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(`{"Name":"Nuevo","Price":100,"Stock":5}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Product created successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Product created successfully"}`)
	}
	if svc.createdProduct.Name != "Nuevo" {
		t.Errorf("Create() persisted Name = %q, want %q", svc.createdProduct.Name, "Nuevo")
	}
}

func TestProductHandler_Create_RejectsMissingNameWithBadRequest(t *testing.T) {
	// Generalizes the "never persist a zero-value where a real value is
	// required" principle (PR3 gate #36 blocker) to product Name; legacy
	// createProduct has NO validation at all, but usecase.ProductService.Create
	// rejects an empty Name BEFORE it reaches the repository. Wires the REAL
	// usecase (not the fake service) so this actually exercises validation
	// through the full HTTP stack, mirroring admin_test.go's realSvc pattern.
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeProductRepoE2E{}
	realSvc := usecase.NewProductService(repo)
	r := newAuthedProductTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(`{"Name":"","Price":100,"Stock":5}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.createCalled {
		t.Errorf("repository.Create was called for an empty Name — no persistence should happen on validation failure")
	}
}

func TestProductHandler_Create_TreatsMalformedJSONBodyAsEmptyRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeProductRepoE2E{}
	realSvc := usecase.NewProductService(repo)
	r := newAuthedProductTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(`not json`))
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

func TestProductHandler_Create_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeProductService{createErr: errors.New("db down")}
	r := newAuthedProductTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(`{"Name":"Nuevo","Price":100,"Stock":5}`))
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

// --- Update (PUT /api/products/:id) ---

func TestProductHandler_Update_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAuthedProductTestRouter(&fakeProductService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/products/1", strings.NewReader(`{"Name":"Renombrado","Price":200,"Stock":10}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestProductHandler_Update_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeProductService{}
	r := newAuthedProductTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/products/1", strings.NewReader(`{"Name":"Renombrado","Price":200,"Stock":10}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.updateID != 1 || svc.updateProduct.Name != "Renombrado" {
		t.Errorf("Update() called with id=%d product=%+v, want id=1 Name=Renombrado", svc.updateID, svc.updateProduct)
	}
}

func TestProductHandler_Update_RejectsMissingNameWithBadRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	repo := &fakeProductRepoE2E{}
	realSvc := usecase.NewProductService(repo)
	r := newAuthedProductTestRouter(realSvc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/products/1", strings.NewReader(`{"Name":"","Price":200,"Stock":10}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if repo.updateCalled {
		t.Errorf("repository.Update was called for an empty Name — no persistence should happen on validation failure")
	}
}

func TestProductHandler_Update_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeProductService{updateErr: errors.New("db down")}
	r := newAuthedProductTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/products/1", strings.NewReader(`{"Name":"Renombrado","Price":200,"Stock":10}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

// --- Delete (DELETE /api/products/:id) ---

func TestProductHandler_Delete_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newAuthedProductTestRouter(&fakeProductService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/products/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestProductHandler_Delete_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeProductService{}
	r := newAuthedProductTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/products/7", nil)
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

func TestProductHandler_Delete_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeProductService{deleteErr: errors.New("db down")}
	r := newAuthedProductTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/products/7", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}
