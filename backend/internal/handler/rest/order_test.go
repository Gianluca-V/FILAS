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
)

// fakeOrderService uses pointer receivers so Create/Update/PatchState
// tests can record what was actually passed to the service, mirroring the
// fakeNewsService/fakeProductService convention established in PR4/PR5.
type fakeOrderService struct {
	orders []domain.Order
	byID   map[int]domain.Order
	err    error

	createErr   error
	createName  string
	createPhone string
	createItems []domain.OrderProduct
	createOrder domain.Order

	updateErr   error
	updateID    int
	updateItems []domain.OrderProduct

	patchErr    error
	patchID     int
	patchTarget domain.OrderState
	patchOrder  domain.Order
}

func (f *fakeOrderService) List(ctx context.Context) ([]domain.Order, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.orders, nil
}

func (f *fakeOrderService) Get(ctx context.Context, id int) (domain.Order, error) {
	if f.err != nil {
		return domain.Order{}, f.err
	}
	o, ok := f.byID[id]
	if !ok {
		return domain.Order{}, domain.ErrNotFound
	}
	return o, nil
}

func (f *fakeOrderService) Create(ctx context.Context, name, phone string, items []domain.OrderProduct) (domain.Order, error) {
	f.createName, f.createPhone, f.createItems = name, phone, items
	if f.createErr != nil {
		return domain.Order{}, f.createErr
	}
	f.createOrder = domain.Order{ID: 9001, Name: name, Phone: phone, Products: items}
	return f.createOrder, nil
}

func (f *fakeOrderService) Update(ctx context.Context, orderID int, items []domain.OrderProduct) error {
	f.updateID, f.updateItems = orderID, items
	return f.updateErr
}

func (f *fakeOrderService) PatchState(ctx context.Context, orderID int, target domain.OrderState) (domain.Order, error) {
	f.patchID, f.patchTarget = orderID, target
	if f.patchErr != nil {
		return domain.Order{}, f.patchErr
	}
	f.patchOrder = domain.Order{ID: orderID, State: target}
	return f.patchOrder, nil
}

// newOrderTestRouter mirrors the REAL router.go wiring for orders (PR7):
// POST is PUBLIC, GET/PUT/PATCH require a valid JWT.
func newOrderTestRouter(svc rest.OrderService, jwtSvc *auth.JWTService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := rest.NewOrderHandler(svc)
	r.POST("/api/orders", h.Create)
	r.GET("/api/orders", middleware.RequireAuth(jwtSvc), h.List)
	r.GET("/api/orders/:id", middleware.RequireAuth(jwtSvc), h.Get)
	r.PUT("/api/orders/:id", middleware.RequireAuth(jwtSvc), h.Update)
	r.PATCH("/api/orders/:id", middleware.RequireAuth(jwtSvc), h.PatchState)
	return r
}

// --- List (GET /api/orders, auth-gated) ---

func TestOrderHandler_List_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newOrderTestRouter(&fakeOrderService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestOrderHandler_List_ReturnsEmptyArrayNotNotFound(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	r := newOrderTestRouter(&fakeOrderService{orders: []domain.Order{}}, jwtSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != "[]" {
		t.Errorf("body = %s, want [] (orders' empty-list is 200, NOT 404 — unlike news/gallery/family/organizations)", rec.Body.String())
	}
}

func TestOrderHandler_List_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	r := newOrderTestRouter(&fakeOrderService{err: errors.New("db down")}, jwtSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

// --- Get (GET /api/orders/:id, auth-gated) ---

func TestOrderHandler_Get_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	r := newOrderTestRouter(&fakeOrderService{}, jwtSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestOrderHandler_Get_ReturnsNotFoundForMissingOrder(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	r := newOrderTestRouter(&fakeOrderService{byID: map[int]domain.Order{}}, jwtSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/999999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Order not found"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Order not found"}`)
	}
}

// --- Create (POST /api/orders, PUBLIC — the one unauthenticated write) ---

func TestOrderHandler_Create_SucceedsWithoutAnyJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	svc := &fakeOrderService{}
	r := newOrderTestRouter(svc, jwtSvc)

	body := `{"orderProducts":[{"productID":1,"quantity":2}],"name":"Cliente","phone":"555"}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (public checkout, no auth header sent, deliberate CHANGE from legacy's bare 200), body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Order created successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Order created successfully"}`)
	}
	if svc.createName != "Cliente" || svc.createPhone != "555" {
		t.Errorf("Create() called with name=%q phone=%q, want Cliente/555", svc.createName, svc.createPhone)
	}
	if len(svc.createItems) != 1 || svc.createItems[0].ProductID != 1 || svc.createItems[0].Quantity != 2 {
		t.Errorf("Create() items = %+v, want [{ProductID:1 Quantity:2}]", svc.createItems)
	}
}

func TestOrderHandler_Create_PropagatesValidationErrorAsBadRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	svc := &fakeOrderService{createErr: domain.ErrValidation}
	r := newOrderTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestOrderHandler_Create_PropagatesInsufficientStockAsConflict(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	svc := &fakeOrderService{createErr: domain.ErrInsufficientStock}
	r := newOrderTestRouter(svc, jwtSvc)

	body := `{"orderProducts":[{"productID":1,"quantity":999}],"name":"Cliente","phone":"555"}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestOrderHandler_Create_ReturnsInternalServerErrorOnServiceFailure(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	svc := &fakeOrderService{createErr: errors.New("db down")}
	r := newOrderTestRouter(svc, jwtSvc)

	body := `{"orderProducts":[{"productID":1,"quantity":1}],"name":"Cliente","phone":"555"}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

// --- Update (PUT /api/orders/:id, auth-gated) ---

func TestOrderHandler_Update_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	svc := &fakeOrderService{}
	r := newOrderTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/orders/27", strings.NewReader(`{"orderProducts":[{"productID":1,"quantity":1}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if svc.updateID != 0 {
		t.Errorf("Update() should not have been called without a JWT")
	}
}

func TestOrderHandler_Update_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeOrderService{}
	r := newOrderTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/orders/27", strings.NewReader(`{"orderProducts":[{"productID":3,"quantity":5}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Order updated successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Order updated successfully"}`)
	}
	if svc.updateID != 27 || len(svc.updateItems) != 1 || svc.updateItems[0].ProductID != 3 {
		t.Errorf("Update() called with id=%d items=%+v, want id=27 items=[{ProductID:3 Quantity:5}]", svc.updateID, svc.updateItems)
	}
}

func TestOrderHandler_Update_PropagatesValidationErrorAsBadRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeOrderService{updateErr: domain.ErrValidation}
	r := newOrderTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPut, "/api/orders/27", strings.NewReader(`{"orderProducts":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

// --- PatchState (PATCH /api/orders/:id, auth-gated) ---

func TestOrderHandler_PatchState_RejectsMissingJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	svc := &fakeOrderService{}
	r := newOrderTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPatch, "/api/orders/5", strings.NewReader(`{"state":"finished"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if svc.patchID != 0 {
		t.Errorf("PatchState() should not have been called without a JWT")
	}
}

func TestOrderHandler_PatchState_SucceedsWithValidJWT(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeOrderService{}
	r := newOrderTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPatch, "/api/orders/5", strings.NewReader(`{"state":"finished"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Order patched successfully"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Order patched successfully"}`)
	}
	if svc.patchID != 5 || svc.patchTarget != domain.OrderStateFinished {
		t.Errorf("PatchState() called with id=%d target=%q, want id=5 target=finished", svc.patchID, svc.patchTarget)
	}
}

func TestOrderHandler_PatchState_PropagatesInvalidStateAsBadRequest(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeOrderService{patchErr: domain.ErrValidation}
	r := newOrderTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPatch, "/api/orders/5", strings.NewReader(`{"state":"bogus"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestOrderHandler_PatchState_PropagatesConflictAsConflictStatus(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeOrderService{patchErr: domain.ErrConflict}
	r := newOrderTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPatch, "/api/orders/5", strings.NewReader(`{"state":"canceled"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d (mirrors legacy patchOrder's 409 'order not pending'), body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestOrderHandler_PatchState_ReturnsNotFoundForMissingOrder(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, err := jwtSvc.Generate(1543)
	if err != nil {
		t.Fatalf("jwtSvc.Generate() error = %v", err)
	}
	svc := &fakeOrderService{patchErr: domain.ErrNotFound}
	r := newOrderTestRouter(svc, jwtSvc)

	req := httptest.NewRequest(http.MethodPatch, "/api/orders/999999", strings.NewReader(`{"state":"finished"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}
