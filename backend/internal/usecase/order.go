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
//
// The usecase computes a PLAN (prices, totals, aggregated stock deltas,
// finishDate) and hands it to domain.OrderRepository as a SINGLE atomic
// composite call per mutation (Create, PatchState's transition) — it never
// orchestrates persistence as separate order-write-then-stock-loop calls.
// That split was the PR6 gate-review blocker: a persist-then-loop from the
// usecase cannot be transactional, risks an orphan order if the loop fails
// partway, and cannot express "reject the write" as a single atomic
// decision. PR7's mysql.OrderRepository implements each composite call in
// one transaction.
type OrderService struct {
	orders   domain.OrderRepository
	products domain.ProductRepository
	now      func() time.Time
}

// NewOrderService wires an OrderService to its order and product
// repositories, using the real wall clock for finishDate. Product access is
// needed to read price/stock when computing order lines (ported
// before_orderProduct_insert/update); PatchState's stock restore does NOT
// need product access — it only needs the order's own line-item quantities
// (see PatchState).
func NewOrderService(orders domain.OrderRepository, products domain.ProductRepository) *OrderService {
	return &OrderService{orders: orders, products: products, now: time.Now}
}

// NewOrderServiceWithClock is NewOrderService with an injected clock,
// primarily for tests that need an EXACT, assertable finishDate instead of
// a before/after time-window check.
func NewOrderServiceWithClock(orders domain.OrderRepository, products domain.ProductRepository, now func() time.Time) *OrderService {
	return &OrderService{orders: orders, products: products, now: now}
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
//     despite its own "Quantity can not be less than 1" message;
//  2. computes orderPrice = product.Price * quantity (ported
//     before_orderProduct_insert trigger), caching each distinct product's
//     lookup so a duplicate ProductID across lines only calls
//     products.Get once.
//
// Stock availability is then checked PER DISTINCT PRODUCT against the
// AGGREGATED demand across all lines referencing it — NOT per line. Two
// lines for the same product must be validated against their COMBINED
// quantity: validating each independently against the same un-mutated
// stock snapshot would let both "pass" and then apply the delta twice from
// a stale read (a lost update — e.g. two {Qty:3} lines vs Stock:5 would
// each individually look fine, but 3+3=6 > 5). See
// TestOrderService_Create_RejectsInsufficientStockForCombinedDuplicateDemand.
// A product whose aggregated demand exceeds its stock rejects the WHOLE
// order with domain.ErrInsufficientStock — a DELIBERATE divergence from
// legacy, which blindly subtracted with no floor check (orders.php:211),
// which is how the seed ended up with negative stock/totals. See
// backend/docs/legacy-quirks.md §14 for both divergences.
//
// The order total (ported after_orderProduct_insert trigger:
// SUM(orderPrice)) is the sum of all line prices. Persistence AND the
// per-product stock decrement happen in ONE call to
// domain.OrderRepository.Create — see that method's doc comment for the
// atomicity contract PR7 must implement.
func (s *OrderService) Create(ctx context.Context, name, phone string, items []domain.OrderProduct) (domain.Order, error) {
	if name == "" || phone == "" || len(items) == 0 {
		return domain.Order{}, fmt.Errorf("orderProducts, name, and phone are required: %w", domain.ErrValidation)
	}

	products := make(map[int]domain.Product, len(items))
	computed := make([]domain.OrderProduct, len(items))
	var total float64
	for i, item := range items {
		if item.Quantity <= 0 {
			return domain.Order{}, fmt.Errorf("quantity must be greater than 0: %w", domain.ErrValidation)
		}
		product, ok := products[item.ProductID]
		if !ok {
			var err error
			product, err = s.products.Get(ctx, item.ProductID)
			if err != nil {
				return domain.Order{}, fmt.Errorf("usecase: create order: load product %d: %w", item.ProductID, err)
			}
			products[item.ProductID] = product
		}
		price := product.Price * float64(item.Quantity)
		computed[i] = domain.OrderProduct{ProductID: item.ProductID, Quantity: item.Quantity, Price: price}
		total += price
	}

	// Aggregate demand PER PRODUCT (see doc comment above), then check
	// availability and build the NET decrement deltas the repository will
	// apply atomically.
	demand := aggregateStockDeltas(computed, 1)
	deltas := make([]domain.StockAdjustment, 0, len(demand))
	for _, d := range demand {
		if products[d.ProductID].Stock < d.Delta {
			return domain.Order{}, fmt.Errorf("product %d has insufficient stock: %w", d.ProductID, domain.ErrInsufficientStock)
		}
		deltas = append(deltas, domain.StockAdjustment{ProductID: d.ProductID, Delta: -d.Delta})
	}

	order := domain.Order{
		Name:     name,
		Phone:    phone,
		Total:    total,
		State:    domain.OrderStatePending,
		Products: computed,
	}
	created, err := s.orders.Create(ctx, order, deltas)
	if err != nil {
		return domain.Order{}, fmt.Errorf("usecase: create order: %w", err)
	}
	return created, nil
}

// Update recomputes each line's orderPrice and the order total (ported
// before/after_orderProduct_insert triggers, same computation as Create,
// including the per-product lookup cache), then replaces all line items
// for orderID — mirroring legacy updateOrder's delete-then-reinsert
// (FilasServer/orders.php:232). Quantity <= 0 is rejected for consistency
// with Create's invariant.
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

	products := make(map[int]domain.Product, len(items))
	computed := make([]domain.OrderProduct, len(items))
	var total float64
	for i, item := range items {
		if item.Quantity <= 0 {
			return fmt.Errorf("quantity must be greater than 0: %w", domain.ErrValidation)
		}
		product, ok := products[item.ProductID]
		if !ok {
			var err error
			product, err = s.products.Get(ctx, item.ProductID)
			if err != nil {
				return fmt.Errorf("usecase: update order %d: load product %d: %w", orderID, item.ProductID, err)
			}
			products[item.ProductID] = product
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
// (FilasServer/orders.php:261). The target state MUST be "finished" or
// "canceled" (domain.ValidOrderTransitionTarget), else domain.ErrValidation
// (400) — "invalid state". This is checked BEFORE any repository call.
//
// The order is then loaded once via Get — to surface domain.ErrNotFound
// for a missing order, and (for a ->canceled target) to read its line
// items for the stock-restore deltas below. This Get result's State field
// is NOT used to authorize the transition: the "must currently be pending"
// gate is enforced by domain.OrderRepository.TransitionState's atomic
// conditional update (CAS semantics), closing the read-then-write race a
// usecase-side `if order.State != pending` check would have (two
// concurrent PATCH calls could both observe "pending" from a stale Get and
// both proceed). TransitionState returning ok=false (order was not
// pending at the moment of the atomic update) maps to domain.ErrConflict
// (409) — legacy's "order not pending" (orders.php:273-277).
//
// On a pending->finished transition, finishDate is stamped via s.now()
// (ported before_update_orders trigger; injectable for tests, see
// NewOrderServiceWithClock) and no stock deltas are sent. On a
// pending->canceled transition, each DISTINCT product's line-item
// quantities are aggregated into ONE net positive delta (ported
// patchOrder's inline restore loop, orders.php:293-325 — same
// aggregation rule as Create, see aggregateStockDeltas) and no finishDate
// is sent. Both the state flip and its accompanying finishDate/stock
// change happen in the SAME TransitionState call — see that method's doc
// comment for the atomicity contract PR7 must implement.
func (s *OrderService) PatchState(ctx context.Context, orderID int, target domain.OrderState) (domain.Order, error) {
	if !domain.ValidOrderTransitionTarget(target) {
		return domain.Order{}, fmt.Errorf("state must be finished or canceled: %w", domain.ErrValidation)
	}

	order, err := s.orders.Get(ctx, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("usecase: patch order %d state: %w", orderID, err)
	}

	var finishDate *time.Time
	var deltas []domain.StockAdjustment
	if target == domain.OrderStateFinished {
		fd := s.now()
		finishDate = &fd
	} else {
		deltas = aggregateStockDeltas(order.Products, 1)
	}

	ok, err := s.orders.TransitionState(ctx, orderID, target, finishDate, deltas)
	if err != nil {
		return domain.Order{}, fmt.Errorf("usecase: patch order %d state: %w", orderID, err)
	}
	if !ok {
		return domain.Order{}, fmt.Errorf("order %d is not pending: %w", orderID, domain.ErrConflict)
	}

	order.State = target
	order.FinishDate = finishDate
	return order, nil
}

// aggregateStockDeltas sums each line's Quantity per DISTINCT ProductID and
// returns one domain.StockAdjustment per product (deterministic order:
// first occurrence in items), with sign applied via the given multiplier
// (1 for a positive aggregate/restore magnitude, used as-is by PatchState's
// cancel path; Create negates it itself to express a decrement). Callers
// MUST aggregate before checking availability or building repository
// deltas — applying one adjustment per LINE instead of one net adjustment
// per PRODUCT lets multiple lines for the same product each act against
// the same stale pre-mutation snapshot (a lost update). See
// TestOrderService_Create_AggregatesDuplicateProductDemandWithinStock and
// TestOrderService_PatchState_AggregatesDuplicateProductLinesOnCancel.
func aggregateStockDeltas(items []domain.OrderProduct, sign int) []domain.StockAdjustment {
	order := make([]int, 0, len(items))
	demand := make(map[int]int, len(items))
	for _, item := range items {
		if _, seen := demand[item.ProductID]; !seen {
			order = append(order, item.ProductID)
		}
		demand[item.ProductID] += item.Quantity
	}
	deltas := make([]domain.StockAdjustment, 0, len(order))
	for _, productID := range order {
		deltas = append(deltas, domain.StockAdjustment{ProductID: productID, Delta: sign * demand[productID]})
	}
	return deltas
}
