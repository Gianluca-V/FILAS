package mysql_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/repository/mysql"
)

// orderJoinSelect mirrors the repository's unexported query constant: a
// single INNER JOIN across orders/orderproduct/products, characterized
// live against orders.php's GROUP_CONCAT-built response (see
// backend/docs/legacy-quirks.md §15). List/Get share this shape; Get adds
// a `WHERE o.ID = ?` clause.
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

var orderJoinCols = []string{
	"orderID", "orderTotal", "orderStartDate", "orderFinishDate", "orderState",
	"orderName", "orderPhone", "lineID", "productID", "productQuantity",
	"orderPrice", "productName", "productPrice",
}

func TestOrderRepository_List_GroupsLineItemsByOrder(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	start := time.Date(2023, 11, 20, 18, 30, 2, 0, time.UTC)
	rows := sqlmock.NewRows(orderJoinCols).
		AddRow(27, 3000.0, start, nil, "pending", "Turco agustin", "32142432546", 69, 3, 5, 3000.0, "Mermelada de tomate", 600.0).
		AddRow(28, 3000.0, start, nil, "canceled", "Gianluca Vespe", "352435234234", 60, 1, 5, 3000.0, "Mermelada de pera", 600.0)
	mock.ExpectQuery(regexp.QuoteMeta(orderJoinSelect + " ORDER BY o.ID, op.ID")).WillReturnRows(rows)

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("List() returned %d orders, want 2", len(got))
	}
	if got[0].ID != 27 || len(got[0].Products) != 1 {
		t.Errorf("got[0] = %+v, want ID=27 with 1 product line", got[0])
	}
	if got[0].Products[0].ProductName != "Mermelada de tomate" || got[0].Products[0].ProductPrice != 600.0 {
		t.Errorf("got[0].Products[0] = %+v, want ProductName=Mermelada de tomate ProductPrice=600 (LIVE product price, not the frozen line total)", got[0].Products[0])
	}
}

// TestOrderRepository_List_AggregatesMultipleLinesPerOrder proves the
// grouping logic collects every JOINed row belonging to the SAME order
// into one domain.Order.Products slice, in row order — mirroring legacy's
// GROUP_CONCAT ... GROUP BY o.ID.
func TestOrderRepository_List_AggregatesMultipleLinesPerOrder(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	start := time.Date(2023, 11, 20, 0, 25, 8, 0, time.UTC)
	rows := sqlmock.NewRows(orderJoinCols).
		AddRow(22, 14650.0, start, nil, "pending", "", "", 32, 16, 2, 1700.0, "Pan integral (800gr)", 850.0).
		AddRow(22, 14650.0, start, nil, "pending", "", "", 33, 17, 2, 2400.0, "Tarta de coco y dulce de leche", 1200.0)
	mock.ExpectQuery(regexp.QuoteMeta(orderJoinSelect + " ORDER BY o.ID, op.ID")).WillReturnRows(rows)

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("List() returned %d orders, want 1 (both rows belong to order 22)", len(got))
	}
	if len(got[0].Products) != 2 {
		t.Fatalf("got[0].Products = %+v, want 2 lines", got[0].Products)
	}
}

func TestOrderRepository_List_ReturnsEmptySliceWhenNoRows(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(orderJoinSelect + " ORDER BY o.ID, op.ID")).
		WillReturnRows(sqlmock.NewRows(orderJoinCols))

	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("List() = %+v, want empty slice", got)
	}
}

func TestOrderRepository_List_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(orderJoinSelect + " ORDER BY o.ID, op.ID")).WillReturnError(dbErr)

	_, err := repo.List(context.Background())
	if !errors.Is(err, dbErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestOrderRepository_Get_ReturnsMatchingOrderWithFinishDate(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	start := time.Date(2023, 11, 18, 20, 9, 58, 0, time.UTC)
	finish := time.Date(2023, 11, 18, 20, 42, 33, 0, time.UTC)
	rows := sqlmock.NewRows(orderJoinCols).
		AddRow(14, 3250.0, start, finish, "finished", "", "", 25, 1, 2, 1200.0, "Mermelada de pera", 600.0)
	mock.ExpectQuery(regexp.QuoteMeta(orderJoinSelect + " WHERE o.ID = ? ORDER BY o.ID, op.ID")).
		WithArgs(14).
		WillReturnRows(rows)

	got, err := repo.Get(context.Background(), 14)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got.ID != 14 || got.State != domain.OrderStateFinished {
		t.Errorf("Get() = %+v, want ID=14 State=finished", got)
	}
	if got.FinishDate == nil || !got.FinishDate.Equal(finish) {
		t.Errorf("Get().FinishDate = %v, want %v", got.FinishDate, finish)
	}
}

// TestOrderRepository_Get_ReturnsNotFoundForOrderWithoutLineItems locks the
// INNER JOIN quirk documented on domain.OrderRepository.Get: an order that
// EXISTS in `orders` but has zero matching `orderproduct` rows never
// appears in the JOINed result set, so Get cannot distinguish it from a
// genuinely missing order — both surface domain.ErrNotFound, mirroring
// legacy's num_rows>0 check on the JOINed result (see
// backend/docs/legacy-quirks.md §15).
func TestOrderRepository_Get_ReturnsNotFoundForOrderWithoutLineItems(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(orderJoinSelect + " WHERE o.ID = ? ORDER BY o.ID, op.ID")).
		WithArgs(999999).
		WillReturnRows(sqlmock.NewRows(orderJoinCols))

	_, err := repo.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestOrderRepository_Get_PropagatesQueryError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	dbErr := errors.New("connection refused")
	mock.ExpectQuery(regexp.QuoteMeta(orderJoinSelect + " WHERE o.ID = ? ORDER BY o.ID, op.ID")).
		WithArgs(14).
		WillReturnError(dbErr)

	_, err := repo.Get(context.Background(), 14)
	if !errors.Is(err, dbErr) {
		t.Errorf("Get() error = %v, want it to wrap %v", err, dbErr)
	}
}

// --- Create: ATOMIC order+lines+stock (domain.OrderRepository.Create contract) ---

const orderInsertSQL = "INSERT INTO orders (name, phone, total) VALUES (?, ?, ?)"
const orderProductInsertSQL = "INSERT INTO orderproduct (orderID, productID, productQuantity, orderPrice) VALUES (?, ?, ?, ?)"
const stockAdjustSQL = "UPDATE products SET Stock = Stock + ? WHERE ID = ?"

// stockDecrementSQL is Create's stock-decrement query only: a FLOOR-GUARDED
// conditional UPDATE mirroring orderTransitionSQL's CAS idiom. stockAdjustSQL
// (unconditional) remains TransitionState's stock-RESTORE query — a positive
// delta never needs the floor guard.
const stockDecrementSQL = "UPDATE products SET Stock = Stock + ? WHERE ID = ? AND Stock + ? >= 0"

func TestOrderRepository_Create_InsertsOrderLinesAndAppliesDeltasInOneTx(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	order := domain.Order{
		Name:  "Turco agustin",
		Phone: "32142432546",
		Total: 3000,
		State: domain.OrderStatePending,
		Products: []domain.OrderProduct{
			{ProductID: 1, Quantity: 2, Price: 1200},
			{ProductID: 2, Quantity: 3, Price: 1800},
		},
	}
	deltas := []domain.StockAdjustment{{ProductID: 1, Delta: -2}, {ProductID: 2, Delta: -3}}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(orderInsertSQL)).
		WithArgs("Turco agustin", "32142432546", 3000.0).
		WillReturnResult(sqlmock.NewResult(9001, 1))
	mock.ExpectExec(regexp.QuoteMeta(orderProductInsertSQL)).
		WithArgs(9001, 1, 2, 1200.0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(orderProductInsertSQL)).
		WithArgs(9001, 2, 3, 1800.0).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec(regexp.QuoteMeta(stockDecrementSQL)).
		WithArgs(-2, 1, -2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(stockDecrementSQL)).
		WithArgs(-3, 2, -3).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	got, err := repo.Create(context.Background(), order, deltas)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (LastInsertId)", got.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestOrderRepository_Create_RollsBackWhenLineInsertFails is the atomicity
// contract's core proof (domain.OrderRepository.Create doc comment): if
// ANY step fails, the order insert must NOT survive — no orphan order with
// partial or missing line items/stock changes.
func TestOrderRepository_Create_RollsBackWhenLineInsertFails(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	order := domain.Order{
		Name: "Cliente", Phone: "123", Total: 1200,
		Products: []domain.OrderProduct{{ProductID: 1, Quantity: 2, Price: 1200}},
	}
	dbErr := errors.New("constraint violation")

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(orderInsertSQL)).
		WithArgs("Cliente", "123", 1200.0).
		WillReturnResult(sqlmock.NewResult(9002, 1))
	mock.ExpectExec(regexp.QuoteMeta(orderProductInsertSQL)).
		WithArgs(9002, 1, 2, 1200.0).
		WillReturnError(dbErr)
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), order, nil)
	if !errors.Is(err, dbErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, dbErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (rollback must be called): %v", err)
	}
}

// TestOrderRepository_Create_RollsBackWhenStockAdjustmentFails proves the
// stock-decrement step is in the SAME transaction as the order/line
// inserts — a failure there must also roll back the already-issued order
// insert, not leave an order with no matching stock change.
func TestOrderRepository_Create_RollsBackWhenStockAdjustmentFails(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	order := domain.Order{
		Name: "Cliente", Phone: "123", Total: 1200,
		Products: []domain.OrderProduct{{ProductID: 1, Quantity: 2, Price: 1200}},
	}
	dbErr := errors.New("deadlock")

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(orderInsertSQL)).
		WithArgs("Cliente", "123", 1200.0).
		WillReturnResult(sqlmock.NewResult(9003, 1))
	mock.ExpectExec(regexp.QuoteMeta(orderProductInsertSQL)).
		WithArgs(9003, 1, 2, 1200.0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(stockDecrementSQL)).
		WithArgs(-2, 1, -2).
		WillReturnError(dbErr)
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), order, []domain.StockAdjustment{{ProductID: 1, Delta: -2}})
	if !errors.Is(err, dbErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, dbErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (rollback must be called): %v", err)
	}
}

// TestOrderRepository_Create_RollsBackWhenStockDecrementHitsFloor is the
// TOCTOU-fix corrective's core proof: the usecase's aggregated-demand check
// reads stock in a separate, earlier transaction, so two concurrent
// checkouts for the same product could both pass it before either decrements
// — the repository closes that gap by making the decrement itself a
// FLOOR-GUARDED conditional UPDATE (stockDecrementSQL), mirroring
// TransitionState's CAS idiom (orderTransitionSQL). Zero rows affected means
// the decrement would have driven stock negative: the whole order rolls
// back and domain.ErrInsufficientStock is returned, regardless of what the
// usecase's earlier pre-check concluded.
func TestOrderRepository_Create_RollsBackWhenStockDecrementHitsFloor(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	order := domain.Order{
		Name: "Cliente", Phone: "123", Total: 1200,
		Products: []domain.OrderProduct{{ProductID: 1, Quantity: 2, Price: 1200}},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(orderInsertSQL)).
		WithArgs("Cliente", "123", 1200.0).
		WillReturnResult(sqlmock.NewResult(9004, 1))
	mock.ExpectExec(regexp.QuoteMeta(orderProductInsertSQL)).
		WithArgs(9004, 1, 2, 1200.0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(stockDecrementSQL)).
		WithArgs(-2, 1, -2).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), order, []domain.StockAdjustment{{ProductID: 1, Delta: -2}})
	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, domain.ErrInsufficientStock)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (rollback must be called): %v", err)
	}
}

func TestOrderRepository_Create_PropagatesBeginTxError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	dbErr := errors.New("too many connections")
	mock.ExpectBegin().WillReturnError(dbErr)

	_, err := repo.Create(context.Background(), domain.Order{Name: "A", Phone: "1"}, nil)
	if !errors.Is(err, dbErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, dbErr)
	}
}

// TestOrderRepository_Create_RollsBackWhenLastInsertIdFails proves a driver
// that fails to report the new order's ID (e.g. no AUTO_INCREMENT support)
// aborts the whole transaction instead of proceeding with a zero-value ID.
func TestOrderRepository_Create_RollsBackWhenLastInsertIdFails(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	order := domain.Order{Name: "Cliente", Phone: "123", Total: 1200}
	dbErr := errors.New("driver does not support LastInsertId")

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(orderInsertSQL)).
		WithArgs("Cliente", "123", 1200.0).
		WillReturnResult(sqlmock.NewErrorResult(dbErr))
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), order, nil)
	if !errors.Is(err, dbErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, dbErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (rollback must be called): %v", err)
	}
}

// TestOrderRepository_Create_PropagatesCommitError proves a commit failure
// (e.g. connection lost after all statements succeeded) surfaces as an
// error to the caller instead of silently reporting success.
func TestOrderRepository_Create_PropagatesCommitError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	order := domain.Order{
		Name: "Cliente", Phone: "123", Total: 1200,
		Products: []domain.OrderProduct{{ProductID: 1, Quantity: 2, Price: 1200}},
	}
	dbErr := errors.New("connection lost")

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(orderInsertSQL)).
		WithArgs("Cliente", "123", 1200.0).
		WillReturnResult(sqlmock.NewResult(9005, 1))
	mock.ExpectExec(regexp.QuoteMeta(orderProductInsertSQL)).
		WithArgs(9005, 1, 2, 1200.0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(stockDecrementSQL)).
		WithArgs(-2, 1, -2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(dbErr)

	_, err := repo.Create(context.Background(), order, []domain.StockAdjustment{{ProductID: 1, Delta: -2}})
	if !errors.Is(err, dbErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, dbErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- TransitionState: CAS-conditional UPDATE + finishDate/deltas in one tx ---

const orderTransitionSQL = "UPDATE orders SET state = ?, finishDate = ? WHERE ID = ? AND state = 'pending'"

func TestOrderRepository_TransitionState_FinishedStampsFinishDateNoDeltas(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	finish := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(orderTransitionSQL)).
		WithArgs(string(domain.OrderStateFinished), finish, 5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ok, err := repo.TransitionState(context.Background(), 5, domain.OrderStateFinished, &finish, nil)
	if err != nil {
		t.Fatalf("TransitionState() error = %v, want nil", err)
	}
	if !ok {
		t.Error("TransitionState() ok = false, want true (1 row affected)")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestOrderRepository_TransitionState_CanceledAppliesRestoreDeltas(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	deltas := []domain.StockAdjustment{{ProductID: 1, Delta: 2}, {ProductID: 2, Delta: 3}}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(orderTransitionSQL)).
		WithArgs(string(domain.OrderStateCanceled), nil, 22).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(stockAdjustSQL)).WithArgs(2, 1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(stockAdjustSQL)).WithArgs(3, 2).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ok, err := repo.TransitionState(context.Background(), 22, domain.OrderStateCanceled, nil, deltas)
	if err != nil {
		t.Fatalf("TransitionState() error = %v, want nil", err)
	}
	if !ok {
		t.Error("TransitionState() ok = false, want true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestOrderRepository_TransitionState_ReturnsFalseWhenNotPendingAppliesNoDeltas
// is the CAS contract's core proof (domain.OrderRepository.TransitionState
// doc comment): zero rows affected by the conditional UPDATE means the
// order was not pending — ok=false, err=nil, and NEITHER finishDate NOR
// any stock delta may be applied. The transaction still rolls back (no
// partial write), even though the conditional UPDATE itself "succeeded"
// (0 rows is not a database/sql error).
func TestOrderRepository_TransitionState_ReturnsFalseWhenNotPendingAppliesNoDeltas(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	deltas := []domain.StockAdjustment{{ProductID: 1, Delta: 2}}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(orderTransitionSQL)).
		WithArgs(string(domain.OrderStateCanceled), nil, 5).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	ok, err := repo.TransitionState(context.Background(), 5, domain.OrderStateCanceled, nil, deltas)
	if err != nil {
		t.Fatalf("TransitionState() error = %v, want nil (CAS miss is not an error)", err)
	}
	if ok {
		t.Error("TransitionState() ok = true, want false (0 rows affected — order was not pending)")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (deltas must NOT be applied on a CAS miss, and the tx must roll back): %v", err)
	}
}

func TestOrderRepository_TransitionState_RollsBackWhenDeltaApplicationFails(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	dbErr := errors.New("deadlock")
	deltas := []domain.StockAdjustment{{ProductID: 1, Delta: 2}}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(orderTransitionSQL)).
		WithArgs(string(domain.OrderStateCanceled), nil, 22).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(stockAdjustSQL)).WithArgs(2, 1).WillReturnError(dbErr)
	mock.ExpectRollback()

	ok, err := repo.TransitionState(context.Background(), 22, domain.OrderStateCanceled, nil, deltas)
	if !errors.Is(err, dbErr) {
		t.Errorf("TransitionState() error = %v, want it to wrap %v", err, dbErr)
	}
	if ok {
		t.Error("TransitionState() ok = true, want false on error")
	}
}

func TestOrderRepository_TransitionState_PropagatesBeginTxError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	dbErr := errors.New("too many connections")
	mock.ExpectBegin().WillReturnError(dbErr)

	_, err := repo.TransitionState(context.Background(), 5, domain.OrderStateFinished, nil, nil)
	if !errors.Is(err, dbErr) {
		t.Errorf("TransitionState() error = %v, want it to wrap %v", err, dbErr)
	}
}

// --- ReplaceProducts: delete-then-reinsert + total update (legacy updateOrder parity) ---

const orderProductDeleteSQL = "DELETE FROM orderproduct WHERE orderID = ?"
const orderTotalUpdateSQL = "UPDATE orders SET total = ? WHERE ID = ?"

func TestOrderRepository_ReplaceProducts_DeletesReinsertsAndUpdatesTotal(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	items := []domain.OrderProduct{{ProductID: 3, Quantity: 5, Price: 3000}}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(orderProductDeleteSQL)).WithArgs(27).WillReturnResult(sqlmock.NewResult(0, 9))
	mock.ExpectExec(regexp.QuoteMeta(orderProductInsertSQL)).
		WithArgs(27, 3, 5, 3000.0).
		WillReturnResult(sqlmock.NewResult(70, 1))
	mock.ExpectExec(regexp.QuoteMeta(orderTotalUpdateSQL)).WithArgs(3000.0, 27).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.ReplaceProducts(context.Background(), 27, items, 3000); err != nil {
		t.Fatalf("ReplaceProducts() error = %v, want nil", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestOrderRepository_ReplaceProducts_RollsBackWhenInsertFails(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	dbErr := errors.New("constraint violation")
	items := []domain.OrderProduct{{ProductID: 3, Quantity: 5, Price: 3000}}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(orderProductDeleteSQL)).WithArgs(27).WillReturnResult(sqlmock.NewResult(0, 9))
	mock.ExpectExec(regexp.QuoteMeta(orderProductInsertSQL)).
		WithArgs(27, 3, 5, 3000.0).
		WillReturnError(dbErr)
	mock.ExpectRollback()

	err := repo.ReplaceProducts(context.Background(), 27, items, 3000)
	if !errors.Is(err, dbErr) {
		t.Errorf("ReplaceProducts() error = %v, want it to wrap %v", err, dbErr)
	}
}

func TestOrderRepository_ReplaceProducts_PropagatesBeginTxError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := mysql.NewOrderRepository(db)

	dbErr := errors.New("too many connections")
	mock.ExpectBegin().WillReturnError(dbErr)

	err := repo.ReplaceProducts(context.Background(), 27, []domain.OrderProduct{{ProductID: 1, Quantity: 1, Price: 1}}, 1)
	if !errors.Is(err, dbErr) {
		t.Errorf("ReplaceProducts() error = %v, want it to wrap %v", err, dbErr)
	}
}
