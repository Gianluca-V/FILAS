package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest/dto"
)

// OrderService is the usecase-layer contract this handler depends on,
// satisfied by *usecase.OrderService (PR6/PR7). Create/Update take the
// usecase's own parameter shape (not a domain.Order) because the
// usecase — not this handler — computes orderPrice/total/stock deltas;
// see usecase.OrderService's doc comments for the ported-trigger logic.
type OrderService interface {
	List(ctx context.Context) ([]domain.Order, error)
	Get(ctx context.Context, id int) (domain.Order, error)
	Create(ctx context.Context, name, phone string, items []domain.OrderProduct) (domain.Order, error)
	Update(ctx context.Context, orderID int, items []domain.OrderProduct) error
	PatchState(ctx context.Context, orderID int, target domain.OrderState) (domain.Order, error)
}

// OrderHandler serves /api/orders, reproducing orders.php's contract
// (characterized live — see backend/docs/legacy-quirks.md §15):
//   - POST (createOrder) is PUBLIC — no auth (router.go). Every other
//     method requires a valid JWT (router.go wires middleware.RequireAuth
//     on GET/PUT/PATCH, unified to 401 same as every other gated resource
//     — legacy distinguishes 400-missing-header from 401-invalid-token
//     here too, same unification precedent as §5).
//   - GET's empty-list case is 200 `[]`, NOT 404 (unlike news/gallery/
//     family/organizations): legacy's `$conn->query($sql)` on a
//     zero-matching-rows SELECT returns a valid (non-false) mysqli_result,
//     so the "$result === false" 404 branch never fires — same reasoning
//     as products'/admins' 200-empty-list behavior (§8). Live-confirmed:
//     truncating orders/orderproduct and calling GET returned 200 `[]`.
//   - A by-ID GET wraps the single result in a one-element array, same
//     quirk as news/gallery/family/organizations (§8), and 404s with
//     `{"message":"Order not found"}` when no matching order exists —
//     including an order that EXISTS but has zero line items (INNER JOIN
//     quirk, see domain.OrderRepository.Get's doc comment).
type OrderHandler struct {
	svc OrderService
}

func NewOrderHandler(svc OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) List(c *gin.Context) {
	orders, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	resp := make([]dto.OrderResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, dto.NewOrderResponse(o))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) Get(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	order, err := h.svc.Get(c.Request.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
		return
	}
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, []dto.OrderResponse{dto.NewOrderResponse(order)})
}

// orderProductRequest is one inbound line item, matching legacy's
// `$product->productID`/`$product->quantity` field names
// (orders.php:190-191/245-246).
type orderProductRequest struct {
	ProductID int `json:"productID"`
	Quantity  int `json:"quantity"`
}

func toDomainOrderProducts(items []orderProductRequest) []domain.OrderProduct {
	out := make([]domain.OrderProduct, 0, len(items))
	for _, item := range items {
		out = append(out, domain.OrderProduct{ProductID: item.ProductID, Quantity: item.Quantity})
	}
	return out
}

// createOrderRequest is the inbound shape for POST /api/orders, matching
// legacy createOrder's `$data->orderProducts`/`$data->name`/`$data->phone`
// (orders.php:172-179).
type createOrderRequest struct {
	OrderProducts []orderProductRequest `json:"orderProducts"`
	Name          string                `json:"name"`
	Phone         string                `json:"phone"`
}

// Create handles POST /api/orders — PUBLIC (no auth, see router.go and
// this handler's doc comment). Legacy's success status is a bare 200 (no
// explicit http_response_code call — orders.php:224-229); the Go endpoint
// returns 201 instead, a deliberate CHANGE mandated by the spec ("Public
// Checkout with Tightened Validation" — THEN 201, stock decremented).
// usecase.OrderService.Create enforces field presence, quantity>0 (a
// DELIBERATE divergence from legacy's quantity<0-only check), and stock
// availability (ErrInsufficientStock -> 409 via middleware.ErrorHandler).
func (h *OrderHandler) Create(c *gin.Context) {
	var req createOrderRequest
	_ = c.ShouldBindJSON(&req)

	if _, err := h.svc.Create(c.Request.Context(), req.Name, req.Phone, toDomainOrderProducts(req.OrderProducts)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Order created successfully"})
}

// updateOrderRequest is the inbound shape for PUT /api/orders/:id,
// matching legacy updateOrder's `$data->orderProducts` (orders.php:237).
type updateOrderRequest struct {
	OrderProducts []orderProductRequest `json:"orderProducts"`
}

// Update handles PUT /api/orders/:id (auth-gated via router.go).
// usecase.OrderService.Update recomputes orderPrice/total and replaces all
// line items (ReplaceProducts); it deliberately does NOT touch product
// stock — see that method's doc comment.
func (h *OrderHandler) Update(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	var req updateOrderRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.svc.Update(c.Request.Context(), id, toDomainOrderProducts(req.OrderProducts)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order updated successfully"})
}

// patchOrderRequest is the inbound shape for PATCH /api/orders/:id,
// matching legacy patchOrder's `$data->state` (orders.php:264).
type patchOrderRequest struct {
	State string `json:"state"`
}

// PatchState handles PATCH /api/orders/:id (auth-gated via router.go).
// usecase.OrderService.PatchState enforces target is "finished" or
// "canceled" (domain.ErrValidation -> 400) and the pending-only CAS gate
// (domain.ErrConflict -> 409, mirroring legacy's "order not pending"
// check, orders.php:273-277, but now race-free — see
// domain.OrderRepository.TransitionState's doc comment).
func (h *OrderHandler) PatchState(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	var req patchOrderRequest
	_ = c.ShouldBindJSON(&req)

	if _, err := h.svc.PatchState(c.Request.Context(), id, domain.OrderState(req.State)); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order patched successfully"})
}
