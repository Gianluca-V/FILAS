package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// OrderService is consumed by internal/handler/rest. It replaces the 5
// legacy MySQL triggers (after_orderProduct_insert/update,
// before_orderProduct_insert/update, before_update_orders — see
// backend/db/init/01-schema.sql) with application-layer logic, per design
// decision #2 (ADR-3). The triggers themselves stay in the seed until PR7's
// gate drops them once this port's tests are green.
type OrderService struct {
	orders   domain.OrderRepository
	products domain.ProductRepository
}

// NewOrderService wires an OrderService to its order and product
// repositories. Product access is needed to read price/stock when
// computing order lines (ported before_orderProduct_insert/update) and to
// decrement/restore stock (ported inline PHP logic, see Create/PatchState).
func NewOrderService(orders domain.OrderRepository, products domain.ProductRepository) *OrderService {
	return &OrderService{orders: orders, products: products}
}

func (s *OrderService) List(ctx context.Context) ([]domain.Order, error) {
	return s.orders.List(ctx)
}

func (s *OrderService) Get(ctx context.Context, id int) (domain.Order, error) {
	return s.orders.Get(ctx, id)
}

// Create validates orderProducts/name/phone (mirrors legacy createOrder's
// isset check, FilasServer/orders.php:172 — 400 "Parameters missing
// (OrderProducts, name or phone)"), then for each line:
//  1. rejects quantity <= 0 — a DELIBERATE divergence from legacy, which
//     only rejected quantity < 0 (orders.php:192), letting 0 through
//     despite its own "Quantity can not be less than 1" message (see
//     backend/docs/legacy-quirks.md §14);
//  2. computes orderPrice = product.Price * quantity (ported
//     before_orderProduct_insert trigger);
//  3. rejects the order if it would drive stock below zero
//     (domain.ErrInsufficientStock) — a DELIBERATE divergence from legacy,
//     which blindly subtracted with no floor check (orders.php:211), which
//     is how the seed ended up with negative stock/totals (see
//     legacy-quirks.md §14).
//
// The order total (ported after_orderProduct_insert trigger:
// SUM(orderPrice)) is computed as the sum of all line prices. Once
// persisted, stock is decremented per line — ported from createOrder's
// inline stock update (orders.php:198-217), now guarded by the invariant
// above instead of applied unconditionally.
func (s *OrderService) Create(ctx context.Context, name, phone string, items []domain.OrderProduct) (domain.Order, error) {
	if name == "" || phone == "" || len(items) == 0 {
		return domain.Order{}, fmt.Errorf("orderProducts, name, and phone are required: %w", domain.ErrValidation)
	}

	type line struct {
		item    domain.OrderProduct
		product domain.Product
	}
	lines := make([]line, len(items))
	var total float64
	for i, item := range items {
		if item.Quantity <= 0 {
			return domain.Order{}, fmt.Errorf("quantity must be greater than 0: %w", domain.ErrValidation)
		}
		product, err := s.products.Get(ctx, item.ProductID)
		if err != nil {
			return domain.Order{}, fmt.Errorf("usecase: create order: load product %d: %w", item.ProductID, err)
		}
		if product.Stock < item.Quantity {
			return domain.Order{}, fmt.Errorf("product %d has insufficient stock: %w", item.ProductID, domain.ErrInsufficientStock)
		}
		price := product.Price * float64(item.Quantity)
		lines[i] = line{
			item:    domain.OrderProduct{ProductID: item.ProductID, Quantity: item.Quantity, Price: price},
			product: product,
		}
		total += price
	}

	computed := make([]domain.OrderProduct, len(lines))
	for i, l := range lines {
		computed[i] = l.item
	}

	order := domain.Order{
		Name:     name,
		Phone:    phone,
		Total:    total,
		State:    domain.OrderStatePending,
		Products: computed,
	}
	created, err := s.orders.Create(ctx, order)
	if err != nil {
		return domain.Order{}, fmt.Errorf("usecase: create order: %w", err)
	}

	for _, l := range lines {
		updated := l.product
		updated.Stock -= l.item.Quantity
		if err := s.products.Update(ctx, l.item.ProductID, updated); err != nil {
			return domain.Order{}, fmt.Errorf("usecase: create order: decrement stock for product %d: %w", l.item.ProductID, err)
		}
	}

	return created, nil
}

// Update recomputes each line's orderPrice and the order total (ported
// before/after_orderProduct_insert triggers, same computation as Create),
// then replaces all line items for orderID — mirroring legacy updateOrder's
// delete-then-reinsert (FilasServer/orders.php:232). Quantity <= 0 is
// rejected for consistency with Create's invariant.
//
// Unlike Create, this does NOT touch product stock: legacy updateOrder
// never adjusted stock when line items changed (a pre-existing gap), and
// reconciling stock here would require reading the order's OLD line items
// atomically with the new write — that needs the transactional repository
// landing in PR7, not the usecase alone. See backend/docs/legacy-quirks.md
// §14 for the scope decision.
func (s *OrderService) Update(ctx context.Context, orderID int, items []domain.OrderProduct) error {
	if len(items) == 0 {
		return fmt.Errorf("orderProducts is required: %w", domain.ErrValidation)
	}

	computed := make([]domain.OrderProduct, len(items))
	var total float64
	for i, item := range items {
		if item.Quantity <= 0 {
			return fmt.Errorf("quantity must be greater than 0: %w", domain.ErrValidation)
		}
		product, err := s.products.Get(ctx, item.ProductID)
		if err != nil {
			return fmt.Errorf("usecase: update order %d: load product %d: %w", orderID, item.ProductID, err)
		}
		price := product.Price * float64(item.Quantity)
		computed[i] = domain.OrderProduct{OrderID: orderID, ProductID: item.ProductID, Quantity: item.Quantity, Price: price}
		total += price
	}

	if err := s.orders.ReplaceProducts(ctx, orderID, computed, total); err != nil {
		return fmt.Errorf("usecase: update order %d: %w", orderID, err)
	}
	return nil
}

// PatchState transitions an order's state, mirroring legacy patchOrder
// (FilasServer/orders.php:261):
//   - the target state MUST be "finished" or "canceled", else
//     domain.ErrValidation (400) — "invalid state";
//   - the order MUST currently be "pending", else domain.ErrConflict (409)
//     — "order not pending". This gate also guarantees finishDate is never
//     re-stamped: an already-finished (or canceled) order can never reach
//     the mutation below.
//
// On a pending->finished transition, finishDate is stamped to now (ported
// before_update_orders trigger) — and left nil for any other transition
// (pending->canceled), matching the trigger's `IF NEW.state = 'finished'`
// guard.
//
// On a pending->canceled transition, each line's quantity is added back to
// product stock — ported from patchOrder's inline stock-restore loop
// (orders.php:293-325), symmetric with Create's guarded decrement.
func (s *OrderService) PatchState(ctx context.Context, orderID int, target domain.OrderState) (domain.Order, error) {
	if target != domain.OrderStateFinished && target != domain.OrderStateCanceled {
		return domain.Order{}, fmt.Errorf("state must be finished or canceled: %w", domain.ErrValidation)
	}

	order, err := s.orders.Get(ctx, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("usecase: patch order %d state: %w", orderID, err)
	}
	if order.State != domain.OrderStatePending {
		return domain.Order{}, fmt.Errorf("order %d is not pending: %w", orderID, domain.ErrConflict)
	}

	var finishDate *time.Time
	if target == domain.OrderStateFinished {
		now := time.Now()
		finishDate = &now
	}

	if err := s.orders.UpdateState(ctx, orderID, target, finishDate); err != nil {
		return domain.Order{}, fmt.Errorf("usecase: patch order %d state: %w", orderID, err)
	}

	if target == domain.OrderStateCanceled {
		for _, item := range order.Products {
			product, err := s.products.Get(ctx, item.ProductID)
			if err != nil {
				return domain.Order{}, fmt.Errorf("usecase: patch order %d state: restore stock for product %d: %w", orderID, item.ProductID, err)
			}
			product.Stock += item.Quantity
			if err := s.products.Update(ctx, item.ProductID, product); err != nil {
				return domain.Order{}, fmt.Errorf("usecase: patch order %d state: restore stock for product %d: %w", orderID, item.ProductID, err)
			}
		}
	}

	order.State = target
	order.FinishDate = finishDate
	return order, nil
}
