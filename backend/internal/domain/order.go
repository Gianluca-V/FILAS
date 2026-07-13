package domain

import (
	"context"
	"time"
)

// OrderState is the orders.state enum. Values mirror the legacy MySQL enum
// exactly (see backend/db/init/01-schema.sql, `orders` table).
type OrderState string

const (
	OrderStatePending  OrderState = "pending"
	OrderStateFinished OrderState = "finished"
	OrderStateCanceled OrderState = "canceled"
)

// ValidOrderTransitionTarget reports whether target is a state patchOrder
// may transition an order INTO (mirrors legacy patchOrder's inline check,
// FilasServer/orders.php:265: `$state != "finished" && $state != "canceled"`
// -> 400 "invalid state"). "pending" is deliberately excluded: it is the
// only state an order STARTS in, never a PATCH target. Mirrors the
// ValidFamilyCategory pattern (see family.go) for the same
// "named predicate over an enum-like string" shape.
func ValidOrderTransitionTarget(target OrderState) bool {
	return target == OrderStateFinished || target == OrderStateCanceled
}

// OrderProduct is a single order line item. Price is the ported orderPrice
// column: legacy computed it via the before_orderProduct_insert/update
// triggers (`price * NEW.productQuantity`, see 01-schema.sql); the Go
// usecase computes it instead (see usecase.OrderService), so by the time an
// OrderProduct reaches the repository, Price is already final.
type OrderProduct struct {
	ID        int
	OrderID   int
	ProductID int
	Quantity  int
	Price     float64
}

// Order is a customer order with its line items. Total is the ported
// orders.total column: legacy recomputed it via the
// after_orderProduct_insert/update triggers (`SUM(orderPrice) WHERE
// orderID = ...`); the Go usecase computes it instead, so by the time an
// Order reaches the repository, Total is already final. FinishDate is the
// ported before_update_orders trigger: legacy stamped it only on the
// pending->finished transition; the usecase does the same.
type Order struct {
	ID         int
	Total      float64
	StartDate  time.Time
	FinishDate *time.Time
	State      OrderState
	Name       string
	Phone      string
	Products   []OrderProduct
}

// StockAdjustment is a single product's NET stock delta, computed by the
// usecase and applied by the repository ATOMICALLY alongside an order
// mutation (Create's decrement, PatchState's cancel-restore) in the SAME
// transaction — see OrderRepository.Create/TransitionState doc comments.
// Delta is negative for a decrement (Create), positive for a restore
// (cancel). The usecase aggregates one StockAdjustment per DISTINCT
// ProductID before calling the repository — never one per line item — so
// an order with multiple lines for the same product produces a single net
// adjustment checked/applied against the CURRENT stock value, not a stale
// pre-mutation snapshot applied twice (see
// usecase.OrderService.Create's doc comment for why this matters).
type StockAdjustment struct {
	ProductID int
	Delta     int
}

// OrderRepository is implemented by internal/repository/mysql (PR7 — the
// transactional order+orderproduct+stock write path). The usecase has
// already computed OrderProduct.Price, Order.Total, FinishDate, and
// StockAdjustment deltas (ported trigger logic + aggregated demand, see
// usecase.OrderService) before calling any of these methods, so the
// repository is a pure persistence boundary — it does not recompute
// prices/totals and it does not decide WHETHER a transition is legal
// beyond the atomicity/CAS contract documented per method below.
type OrderRepository interface {
	List(ctx context.Context) ([]Order, error)
	Get(ctx context.Context, id int) (Order, error)
	// Create MUST atomically, in a single transaction: (1) insert the
	// order, (2) insert its line items, and (3) apply every given
	// StockAdjustment (all negative — a decrement) to the corresponding
	// products. The usecase has already validated aggregated per-product
	// demand against current stock (ErrInsufficientStock) before calling
	// this, so the repository does not need to re-check availability, only
	// apply the deltas consistently WITH the insert — if the stock update
	// fails, the order insert MUST roll back too (no orphan order with no
	// matching stock decrement, and vice versa). Returns the created Order
	// with IDs populated from the database's AUTO_INCREMENT. Mirrors
	// legacy createOrder (FilasServer/orders.php:168), which had NO such
	// atomicity guarantee — a partial failure there could leave an order
	// with some, but not all, of its stock decremented.
	Create(ctx context.Context, o Order, deltas []StockAdjustment) (Order, error)
	// ReplaceProducts replaces all line items for an existing order and
	// updates orders.total, mirroring legacy updateOrder's
	// delete-then-reinsert (FilasServer/orders.php:232). Does NOT touch
	// orders.state, finishDate, or product stock — see
	// usecase.OrderService.Update's doc comment for why stock
	// reconciliation is deliberately out of scope here.
	ReplaceProducts(ctx context.Context, orderID int, items []OrderProduct, total float64) error
	// TransitionState MUST atomically, in a single transaction:
	//  1. Transition orders.state from "pending" to target — implemented
	//     as a CONDITIONAL update (e.g. `UPDATE orders SET state=?, ...
	//     WHERE ID=? AND state='pending'`), NOT a read-then-write from the
	//     caller. This closes the check-then-act race a separate
	//     `Get`-then-`Update` would have: two concurrent PATCH calls on the
	//     same order can no longer both observe "pending" and both
	//     "succeed".
	//  2. If the conditional update affected zero rows (order was not
	//     pending, or does not exist), TransitionState MUST return
	//     ok=false, err=nil — the caller distinguishes "order missing"
	//     (via a prior Get) from "order not pending" (via this ok flag) and
	//     maps ok=false to domain.ErrConflict, mirroring legacy patchOrder's
	//     409 "order not pending" (FilasServer/orders.php:273-277).
	//  3. If (and only if) the transition succeeds, apply finishDate
	//     (non-nil only for target=="finished" — ported before_update_orders
	//     trigger) and every given StockAdjustment (non-empty only for
	//     target=="canceled", all positive — a restore; ported patchOrder's
	//     inline restore loop, orders.php:293-325) as part of the SAME
	//     transaction as step 1 — a state flip with a stock restore that
	//     silently fails (or vice versa) is exactly the kind of partial
	//     write this method exists to prevent.
	TransitionState(ctx context.Context, orderID int, target OrderState, finishDate *time.Time, deltas []StockAdjustment) (ok bool, err error)
}
