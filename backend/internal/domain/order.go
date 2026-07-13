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

// OrderRepository is implemented by internal/repository/mysql (PR7 — the
// transactional order+orderproduct+stock write path). The usecase has
// already computed OrderProduct.Price, Order.Total, and FinishDate (ported
// trigger logic, see usecase.OrderService) before calling any of these
// methods, so the repository is a pure persistence boundary — it does not
// recompute anything.
type OrderRepository interface {
	List(ctx context.Context) ([]Order, error)
	Get(ctx context.Context, id int) (Order, error)
	// Create persists a new order and its line items and returns it with
	// IDs populated from the database's AUTO_INCREMENT. Mirrors legacy
	// createOrder's insert-order-then-insert-orderproducts sequence
	// (FilasServer/orders.php:168), except Total and each line's Price
	// arrive pre-computed instead of being derived by triggers.
	Create(ctx context.Context, o Order) (Order, error)
	// ReplaceProducts replaces all line items for an existing order and
	// updates orders.total, mirroring legacy updateOrder's
	// delete-then-reinsert (FilasServer/orders.php:232). Does not touch
	// orders.state, finishDate, or product stock.
	ReplaceProducts(ctx context.Context, orderID int, items []OrderProduct, total float64) error
	// UpdateState transitions orders.state and persists the finishDate
	// already computed by the usecase (nil unless transitioning to
	// "finished"). Mirrors legacy patchOrder (FilasServer/orders.php:261).
	UpdateState(ctx context.Context, orderID int, state OrderState, finishDate *time.Time) error
}
