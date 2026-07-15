package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// OrderRepository implements domain.OrderRepository on top of sqlx with
// parameterized queries. Unlike the other resource repositories, every
// write method here runs inside an explicit transaction — see the
// atomicity/CAS contract documented on domain.OrderRepository.Create and
// .TransitionState (PR6 corrective), which this repository is the first
// implementation of.
type OrderRepository struct {
	db *sqlx.DB
}

// NewOrderRepository wires an OrderRepository to an open connection pool.
func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// orderJoinRow is one row of the flat, denormalized read-path query: one
// row per order LINE ITEM (an order with N lines produces N rows sharing
// the same order* columns). This mirrors legacy getOrders/getOrder's
// GROUP_CONCAT approach (FilasServer/orders.php:69/118) without SQL-level
// string concatenation — the grouping happens in Go (groupOrderRows)
// instead, which sidesteps the GROUP_CONCAT/manual-JSON fragility
// characterized live (see backend/docs/legacy-quirks.md §15): a product
// name containing a quote or comma cannot corrupt this shape the way it
// corrupts legacy's hand-built JSON string.
type orderJoinRow struct {
	OrderID      int             `db:"orderID"`
	Total        sql.NullFloat64 `db:"orderTotal"`
	StartDate    sql.NullTime    `db:"orderStartDate"`
	FinishDate   sql.NullTime    `db:"orderFinishDate"`
	State        string          `db:"orderState"`
	Name         string          `db:"orderName"`
	Phone        string          `db:"orderPhone"`
	LineID       int             `db:"lineID"`
	ProductID    int             `db:"productID"`
	Quantity     int             `db:"productQuantity"`
	OrderPrice   float64         `db:"orderPrice"`
	ProductName  string          `db:"productName"`
	ProductPrice float64         `db:"productPrice"`
}

// orderJoinSelect is the shared List/Get query: an INNER JOIN across
// orders -> orderproduct -> products. Being an INNER JOIN (not LEFT JOIN),
// an order with ZERO orderproduct rows produces NO output rows for that
// order at all — see the domain.OrderRepository.List/Get doc comment and
// backend/docs/legacy-quirks.md §15 for why this is deliberate parity, not
// an oversight. `p.Name`/`p.Price` are the product's CURRENT values (a
// live join), NOT the frozen `op.orderPrice` — see
// domain.OrderProduct.ProductName's doc comment for why the response
// deliberately does NOT show what was actually billed.
const orderJoinSelect = `SELECT
	o.ID AS orderID,
	o.total AS orderTotal,
	o.startDate AS orderStartDate,
	o.finishDate AS orderFinishDate,
	o.state AS orderState,
	o.name AS orderName,
	o.phone AS orderPhone,
	op.ID AS lineID,
	op.productID AS productID,
	op.productQuantity AS productQuantity,
	op.orderPrice AS orderPrice,
	p.Name AS productName,
	p.Price AS productPrice
FROM orders o
JOIN orderproduct op ON o.ID = op.orderID
JOIN products p ON op.productID = p.ID`

const orderJoinOrderBy = " ORDER BY o.ID, op.ID"

// groupOrderRows folds the flat per-line orderJoinRow slice into
// domain.Order values, one per distinct orderID, preserving first-seen
// order (matches the query's `ORDER BY o.ID, op.ID`) — the Go equivalent
// of legacy's `GROUP BY o.ID` + row-collector loop.
func groupOrderRows(rows []orderJoinRow) []domain.Order {
	orders := make([]domain.Order, 0, len(rows))
	index := make(map[int]int, len(rows))
	for _, row := range rows {
		i, ok := index[row.OrderID]
		if !ok {
			o := domain.Order{
				ID:    row.OrderID,
				Total: row.Total.Float64,
				State: domain.OrderState(row.State),
				Name:  row.Name,
				Phone: row.Phone,
			}
			if row.StartDate.Valid {
				o.StartDate = row.StartDate.Time
			}
			if row.FinishDate.Valid {
				fd := row.FinishDate.Time
				o.FinishDate = &fd
			}
			orders = append(orders, o)
			i = len(orders) - 1
			index[row.OrderID] = i
		}
		orders[i].Products = append(orders[i].Products, domain.OrderProduct{
			ID:           row.LineID,
			OrderID:      row.OrderID,
			ProductID:    row.ProductID,
			Quantity:     row.Quantity,
			Price:        row.OrderPrice,
			ProductName:  row.ProductName,
			ProductPrice: row.ProductPrice,
		})
	}
	return orders
}

func (r *OrderRepository) List(ctx context.Context) ([]domain.Order, error) {
	var rows []orderJoinRow
	if err := r.db.SelectContext(ctx, &rows, orderJoinSelect+orderJoinOrderBy); err != nil {
		return nil, fmt.Errorf("mysql: list orders: %w", err)
	}
	return groupOrderRows(rows), nil
}

// Get returns domain.ErrNotFound both for a genuinely missing order ID AND
// for an order that exists but has zero line items (INNER JOIN quirk, see
// orderJoinSelect's doc comment) — indistinguishable by design, exactly
// like legacy's num_rows>0 check on the same JOINed result set.
func (r *OrderRepository) Get(ctx context.Context, id int) (domain.Order, error) {
	var rows []orderJoinRow
	if err := r.db.SelectContext(ctx, &rows, orderJoinSelect+" WHERE o.ID = ?"+orderJoinOrderBy, id); err != nil {
		return domain.Order{}, fmt.Errorf("mysql: get order %d: %w", id, err)
	}
	orders := groupOrderRows(rows)
	if len(orders) == 0 {
		return domain.Order{}, domain.ErrNotFound
	}
	return orders[0], nil
}

const orderInsert = "INSERT INTO orders (name, phone, total) VALUES (?, ?, ?)"
const orderProductInsert = "INSERT INTO orderproduct (orderID, productID, productQuantity, orderPrice) VALUES (?, ?, ?, ?)"
const stockAdjust = "UPDATE products SET Stock = Stock + ? WHERE ID = ?"

// Create implements the atomic contract documented on
// domain.OrderRepository.Create: order insert, every line-item insert, and
// every stock delta all happen in ONE transaction, rolling back on the
// first failure — no orphan order, no partial stock decrement. Mirrors
// legacy createOrder's 3 columns (name, phone, total; state/startDate use
// the schema's DEFAULT 'pending'/CURRENT_TIMESTAMP, same as legacy —
// FilasServer/orders.php:182), but unlike legacy, total is the
// usecase-computed final value, never NULL-then-trigger-patched.
func (r *OrderRepository) Create(ctx context.Context, o domain.Order, deltas []domain.StockAdjustment) (domain.Order, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return domain.Order{}, fmt.Errorf("mysql: create order: begin tx: %w", err)
	}

	res, err := tx.ExecContext(ctx, orderInsert, o.Name, o.Phone, o.Total)
	if err != nil {
		_ = tx.Rollback()
		return domain.Order{}, fmt.Errorf("mysql: create order: insert order: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return domain.Order{}, fmt.Errorf("mysql: create order: read new order id: %w", err)
	}
	o.ID = int(id)

	for i := range o.Products {
		o.Products[i].OrderID = o.ID
		line := o.Products[i]
		if _, err := tx.ExecContext(ctx, orderProductInsert, o.ID, line.ProductID, line.Quantity, line.Price); err != nil {
			_ = tx.Rollback()
			return domain.Order{}, fmt.Errorf("mysql: create order: insert line item for product %d: %w", line.ProductID, err)
		}
	}

	for _, d := range deltas {
		if _, err := tx.ExecContext(ctx, stockAdjust, d.Delta, d.ProductID); err != nil {
			_ = tx.Rollback()
			return domain.Order{}, fmt.Errorf("mysql: create order: adjust stock for product %d: %w", d.ProductID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.Order{}, fmt.Errorf("mysql: create order: commit: %w", err)
	}
	return o, nil
}

const orderProductDelete = "DELETE FROM orderproduct WHERE orderID = ?"
const orderTotalUpdate = "UPDATE orders SET total = ? WHERE ID = ?"

// ReplaceProducts implements domain.OrderRepository.ReplaceProducts:
// delete-then-reinsert line items, then update the order's total, all in
// one transaction — mirroring legacy updateOrder's delete-then-reinsert
// (FilasServer/orders.php:232), plus the total update legacy's own PUT
// never did (a pre-existing gap; here it is a single atomic follow-up
// instead of relying on a trigger). Does not touch state/finishDate/stock
// — see usecase.OrderService.Update's doc comment for that scope decision.
func (r *OrderRepository) ReplaceProducts(ctx context.Context, orderID int, items []domain.OrderProduct, total float64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mysql: replace products for order %d: begin tx: %w", orderID, err)
	}

	if _, err := tx.ExecContext(ctx, orderProductDelete, orderID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("mysql: replace products for order %d: delete existing lines: %w", orderID, err)
	}
	for _, item := range items {
		if _, err := tx.ExecContext(ctx, orderProductInsert, orderID, item.ProductID, item.Quantity, item.Price); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("mysql: replace products for order %d: insert line for product %d: %w", orderID, item.ProductID, err)
		}
	}
	if _, err := tx.ExecContext(ctx, orderTotalUpdate, total, orderID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("mysql: replace products for order %d: update total: %w", orderID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("mysql: replace products for order %d: commit: %w", orderID, err)
	}
	return nil
}

const orderTransition = "UPDATE orders SET state = ?, finishDate = ? WHERE ID = ? AND state = 'pending'"

// TransitionState implements the CAS contract documented on
// domain.OrderRepository.TransitionState: a single CONDITIONAL UPDATE
// (`... WHERE state = 'pending'`) decides atomically whether the
// transition is legal — no separate read-then-write from the caller. Zero
// rows affected means the order was not pending (or does not exist):
// ok=false, err=nil, the transaction rolls back, and NEITHER finishDate
// NOR any stock delta is applied. On success, finishDate/deltas apply in
// the SAME transaction as the state flip.
func (r *OrderRepository) TransitionState(ctx context.Context, orderID int, target domain.OrderState, finishDate *time.Time, deltas []domain.StockAdjustment) (bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("mysql: transition order %d state: begin tx: %w", orderID, err)
	}

	res, err := tx.ExecContext(ctx, orderTransition, string(target), finishDate, orderID)
	if err != nil {
		_ = tx.Rollback()
		return false, fmt.Errorf("mysql: transition order %d state: conditional update: %w", orderID, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return false, fmt.Errorf("mysql: transition order %d state: read affected rows: %w", orderID, err)
	}
	if affected == 0 {
		_ = tx.Rollback()
		return false, nil
	}

	for _, d := range deltas {
		if _, err := tx.ExecContext(ctx, stockAdjust, d.Delta, d.ProductID); err != nil {
			_ = tx.Rollback()
			return false, fmt.Errorf("mysql: transition order %d state: restore stock for product %d: %w", orderID, d.ProductID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("mysql: transition order %d state: commit: %w", orderID, err)
	}
	return true, nil
}
