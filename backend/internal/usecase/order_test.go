package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

// fakeOrderRepo uses pointer receivers so Create/ReplaceProducts/
// TransitionState tests can record what was actually passed to the
// repository, mirroring the fakeProductRepo/fakeNewsRepo convention
// established in PR3-PR5. Unlike the pre-corrective PR6 fake, Create and
// TransitionState now mirror the ATOMIC composite contract on
// domain.OrderRepository: Create takes stock deltas directly (no separate
// stock-update loop from the caller), and TransitionState reports success
// via a boolean (CAS semantics), not by assuming the caller pre-checked
// state.
type fakeOrderRepo struct {
	orders []domain.Order
	byID   map[int]domain.Order
	err    error

	createErr     error
	createdOrder  domain.Order
	createdDeltas []domain.StockAdjustment
	createReturn  domain.Order // if ID != 0, Create echoes this instead of assigning 9001

	replaceErr     error
	replaceCalled  bool
	replaceOrderID int
	replaceItems   []domain.OrderProduct
	replaceTotal   float64

	transitionErr error
	// transitionConflict simulates the conditional UPDATE affecting zero
	// rows (order was not pending at the moment of the atomic update) —
	// TransitionState returns ok=false, err=nil in that case. Defaults to
	// false (success), matching every other fake's implicit-success-unless-
	// configured convention.
	transitionConflict   bool
	transitionCalled     bool
	transitionOrderID    int
	transitionTarget     domain.OrderState
	transitionFinishDate *time.Time
	transitionDeltas     []domain.StockAdjustment
}

func (f *fakeOrderRepo) List(ctx context.Context) ([]domain.Order, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.orders, nil
}

func (f *fakeOrderRepo) Get(ctx context.Context, id int) (domain.Order, error) {
	if f.err != nil {
		return domain.Order{}, f.err
	}
	o, ok := f.byID[id]
	if !ok {
		return domain.Order{}, domain.ErrNotFound
	}
	return o, nil
}

func (f *fakeOrderRepo) Create(ctx context.Context, o domain.Order, deltas []domain.StockAdjustment) (domain.Order, error) {
	if f.createErr != nil {
		return domain.Order{}, f.createErr
	}
	f.createdOrder = o
	f.createdDeltas = deltas
	if f.createReturn.ID != 0 {
		return f.createReturn, nil
	}
	o.ID = 9001
	return o, nil
}

func (f *fakeOrderRepo) ReplaceProducts(ctx context.Context, orderID int, items []domain.OrderProduct, total float64) error {
	f.replaceCalled = true
	f.replaceOrderID = orderID
	f.replaceItems = items
	f.replaceTotal = total
	return f.replaceErr
}

func (f *fakeOrderRepo) TransitionState(ctx context.Context, orderID int, target domain.OrderState, finishDate *time.Time, deltas []domain.StockAdjustment) (bool, error) {
	f.transitionCalled = true
	f.transitionOrderID = orderID
	f.transitionTarget = target
	f.transitionFinishDate = finishDate
	f.transitionDeltas = deltas
	if f.transitionErr != nil {
		return false, f.transitionErr
	}
	if f.transitionConflict {
		return false, nil
	}
	return true, nil
}

func newFakeProductCatalog(products ...domain.Product) *fakeProductRepo {
	byID := make(map[int]domain.Product, len(products))
	for _, p := range products {
		byID[p.ID] = p
	}
	return &fakeProductRepo{byID: byID}
}

func deltaByProductID(deltas []domain.StockAdjustment) map[int]int {
	m := make(map[int]int, len(deltas))
	for _, d := range deltas {
		m[d.ProductID] = d.Delta
	}
	return m
}

// --- List / Get (thin delegation, same pattern as PR2-PR5) ---

func TestOrderService_List_ReturnsRepositoryOrders(t *testing.T) {
	want := []domain.Order{{ID: 1, Name: "Turco agustin"}}
	svc := usecase.NewOrderService(&fakeOrderRepo{orders: want}, &fakeProductRepo{})

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 1 || got[0].ID != 1 {
		t.Errorf("List() = %+v, want %+v", got, want)
	}
}

func TestOrderService_List_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := usecase.NewOrderService(&fakeOrderRepo{err: repoErr}, &fakeProductRepo{})

	_, err := svc.List(context.Background())
	if !errors.Is(err, repoErr) {
		t.Errorf("List() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestOrderService_Get_ReturnsMatchingOrder(t *testing.T) {
	want := domain.Order{ID: 1, Name: "Turco agustin"}
	svc := usecase.NewOrderService(&fakeOrderRepo{byID: map[int]domain.Order{1: want}}, &fakeProductRepo{})

	got, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got.ID != 1 {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestOrderService_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	svc := usecase.NewOrderService(&fakeOrderRepo{byID: map[int]domain.Order{}}, &fakeProductRepo{})

	_, err := svc.Get(context.Background(), 999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, domain.ErrNotFound)
	}
}

// --- Create (ports createOrder, before/after_orderProduct_insert) ---

func TestOrderService_Create_ComputesOrderPriceAndTotal(t *testing.T) {
	products := newFakeProductCatalog(
		domain.Product{ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112},
		domain.Product{ID: 2, Name: "Mermelada de naranja", Price: 600, Stock: 5},
	)
	orders := &fakeOrderRepo{}
	svc := usecase.NewOrderService(orders, products)

	items := []domain.OrderProduct{
		{ProductID: 1, Quantity: 2}, // 600*2 = 1200
		{ProductID: 2, Quantity: 3}, // 600*3 = 1800
	}
	got, err := svc.Create(context.Background(), "Turco agustin", "3214243254", items)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if got.ID != 9001 {
		t.Errorf("Create() ID = %d, want 9001 (repo-assigned via LastInsertId)", got.ID)
	}
	if orders.createdOrder.Total != 3000 {
		t.Errorf("Create() persisted Total = %v, want 3000 (sum of orderPrice, ported after_orderProduct_insert)", orders.createdOrder.Total)
	}
	if len(orders.createdOrder.Products) != 2 {
		t.Fatalf("Create() persisted %d products, want 2", len(orders.createdOrder.Products))
	}
	if orders.createdOrder.Products[0].Price != 1200 {
		t.Errorf("Create() line[0].Price = %v, want 1200 (price*qty, ported before_orderProduct_insert)", orders.createdOrder.Products[0].Price)
	}
	if orders.createdOrder.Products[1].Price != 1800 {
		t.Errorf("Create() line[1].Price = %v, want 1800 (price*qty, ported before_orderProduct_insert)", orders.createdOrder.Products[1].Price)
	}
	if orders.createdOrder.State != domain.OrderStatePending {
		t.Errorf("Create() persisted State = %q, want %q", orders.createdOrder.State, domain.OrderStatePending)
	}
}

// TestOrderService_Create_PassesAggregatedStockDecrementDeltas is the
// corrective's headline fix (gate blocker #2): Create must no longer
// persist the order via a separate products.Update() loop. Instead it
// hands the repository ONE domain.StockAdjustment per distinct product,
// which OrderRepository.Create applies atomically alongside the insert
// (PR7). This test proves the deltas reaching the repository are correct
// and negative (a decrement) for a simple, no-duplicate-product order.
func TestOrderService_Create_PassesAggregatedStockDecrementDeltas(t *testing.T) {
	products := newFakeProductCatalog(
		domain.Product{ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112},
		domain.Product{ID: 2, Name: "Mermelada de naranja", Price: 600, Stock: 5},
	)
	orders := &fakeOrderRepo{}
	svc := usecase.NewOrderService(orders, products)

	items := []domain.OrderProduct{
		{ProductID: 1, Quantity: 2},
		{ProductID: 2, Quantity: 5},
	}
	if _, err := svc.Create(context.Background(), "Turco agustin", "3214243254", items); err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	deltas := deltaByProductID(orders.createdDeltas)
	if len(deltas) != 2 {
		t.Fatalf("Create() passed %d deltas, want 2 (one per distinct product)", len(deltas))
	}
	if deltas[1] != -2 {
		t.Errorf("Create() delta for product 1 = %d, want -2", deltas[1])
	}
	if deltas[2] != -5 {
		t.Errorf("Create() delta for product 2 = %d, want -5", deltas[2])
	}
	// The usecase itself must never call products.Update() directly — stock
	// mutation is delegated entirely to the atomic repo.Create() call.
	if products.updateCalled {
		t.Error("Create() should not call products.Update() directly — stock changes must flow through the atomic repo.Create(deltas) contract")
	}
}

// TestOrderService_Create_AggregatesDuplicateProductDemandWithinStock is
// gate blocker #1's positive case: two line items for the SAME product,
// whose COMBINED demand still fits within stock, must produce a SINGLE net
// delta (not two separate deltas that could each pass an independent,
// stale-snapshot check).
func TestOrderService_Create_AggregatesDuplicateProductDemandWithinStock(t *testing.T) {
	products := newFakeProductCatalog(domain.Product{ID: 1, Name: "Galletas", Price: 850, Stock: 5})
	orders := &fakeOrderRepo{}
	svc := usecase.NewOrderService(orders, products)

	items := []domain.OrderProduct{
		{ProductID: 1, Quantity: 2},
		{ProductID: 1, Quantity: 3},
	}
	if _, err := svc.Create(context.Background(), "Cliente", "123", items); err != nil {
		t.Fatalf("Create() error = %v, want nil (combined demand 2+3=5 fits Stock=5)", err)
	}

	if len(orders.createdDeltas) != 1 {
		t.Fatalf("Create() passed %d deltas, want 1 (one NET adjustment for the single distinct product, not one per line)", len(orders.createdDeltas))
	}
	if orders.createdDeltas[0].ProductID != 1 || orders.createdDeltas[0].Delta != -5 {
		t.Errorf("Create() delta = %+v, want {ProductID:1 Delta:-5} (aggregated 2+3)", orders.createdDeltas[0])
	}
}

// TestOrderService_Create_RejectsInsufficientStockForCombinedDuplicateDemand
// is gate blocker #1's negative case: two lines for the same product, EACH
// individually within stock, but whose COMBINED demand exceeds it, must be
// rejected. Pre-corrective, each line validated against the same
// un-mutated products.Get() snapshot independently and both passed,
// silently overselling (e.g. two {Qty:3} lines vs Stock:5 both "fit"
// individually, but 3+3=6 > 5).
func TestOrderService_Create_RejectsInsufficientStockForCombinedDuplicateDemand(t *testing.T) {
	products := newFakeProductCatalog(domain.Product{ID: 1, Name: "Galletas", Price: 850, Stock: 5})
	orders := &fakeOrderRepo{}
	svc := usecase.NewOrderService(orders, products)

	items := []domain.OrderProduct{
		{ProductID: 1, Quantity: 3},
		{ProductID: 1, Quantity: 3},
	}
	_, err := svc.Create(context.Background(), "Cliente", "123", items)
	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Errorf("Create() error = %v, want %v (combined demand 3+3=6 > Stock=5)", err, domain.ErrInsufficientStock)
	}
	if orders.createdOrder.ID != 0 {
		t.Errorf("Create() should not have persisted an order, but got %+v", orders.createdOrder)
	}
}

// TestOrderService_Create_RejectsInsufficientStock is a DELIBERATE
// divergence from legacy: createOrder (FilasServer/orders.php:211) blindly
// computed `$newStock = $currentStock - $quantity` with no floor check,
// which is how the seed ended up with a -85,000,000 total (see
// 01-schema.sql orders row 31 / orderproduct row 63). The Go usecase
// enforces the Stock >= 0 invariant instead — see
// backend/docs/legacy-quirks.md §14.
func TestOrderService_Create_RejectsInsufficientStock(t *testing.T) {
	products := newFakeProductCatalog(domain.Product{ID: 1, Name: "Pan dulce", Price: 800, Stock: 0})
	orders := &fakeOrderRepo{}
	svc := usecase.NewOrderService(orders, products)

	_, err := svc.Create(context.Background(), "Cliente", "123", []domain.OrderProduct{{ProductID: 1, Quantity: 1}})
	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Errorf("Create() error = %v, want %v", err, domain.ErrInsufficientStock)
	}
	if orders.createdOrder.ID != 0 || orders.createdOrder.Total != 0 {
		t.Errorf("Create() should not have persisted an order, but got %+v", orders.createdOrder)
	}
}

// TestOrderService_Create_RejectsNonPositiveQuantity is a DELIBERATE
// divergence from legacy: createOrder (FilasServer/orders.php:192) only
// rejects `quantity < 0`, so quantity=0 slips through despite the error
// message claiming "Quantity can not be less than 1". The Go usecase
// rejects quantity <= 0 — see backend/docs/legacy-quirks.md §14.
func TestOrderService_Create_RejectsNonPositiveQuantity(t *testing.T) {
	cases := []struct {
		name     string
		quantity int
	}{
		{"zero quantity", 0},
		{"negative quantity", -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			products := newFakeProductCatalog(domain.Product{ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112})
			orders := &fakeOrderRepo{}
			svc := usecase.NewOrderService(orders, products)

			_, err := svc.Create(context.Background(), "Cliente", "123", []domain.OrderProduct{{ProductID: 1, Quantity: tc.quantity}})
			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Create() error = %v, want %v", err, domain.ErrValidation)
			}
			if orders.createdOrder.ID != 0 || orders.createdOrder.Total != 0 {
				t.Errorf("Create() should not have persisted an order, but got %+v", orders.createdOrder)
			}
		})
	}
}

// TestOrderService_Create_RejectsMissingFields mirrors legacy createOrder's
// isset check (FilasServer/orders.php:172): !isset(orderProducts, name,
// phone) -> 400 "Parameters missing (OrderProducts, name or phone)".
func TestOrderService_Create_RejectsMissingFields(t *testing.T) {
	products := newFakeProductCatalog(domain.Product{ID: 1, Price: 600, Stock: 10})
	validItems := []domain.OrderProduct{{ProductID: 1, Quantity: 1}}

	cases := []struct {
		name      string
		orderName string
		phone     string
		items     []domain.OrderProduct
	}{
		{"missing name", "", "123", validItems},
		{"missing phone", "Cliente", "", validItems},
		{"missing orderProducts", "Cliente", "123", nil},
		{"empty orderProducts", "Cliente", "123", []domain.OrderProduct{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orders := &fakeOrderRepo{}
			svc := usecase.NewOrderService(orders, products)

			_, err := svc.Create(context.Background(), tc.orderName, tc.phone, tc.items)
			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Create() error = %v, want %v", err, domain.ErrValidation)
			}
			if orders.createdOrder.ID != 0 {
				t.Errorf("Create() should not have persisted an order, but got %+v", orders.createdOrder)
			}
		})
	}
}

// TestOrderService_Create_PropagatesOrderRepositoryError proves the
// atomic-create failure surfaces cleanly: no order is returned (zero
// value), and the error wraps the repository's error — there is no
// "orphan order" risk because Create is now ONE repository call (order +
// lines + stock deltas together), not a persist-then-loop the usecase
// could partially complete.
func TestOrderService_Create_PropagatesOrderRepositoryError(t *testing.T) {
	repoErr := errors.New("tx rolled back")
	products := newFakeProductCatalog(domain.Product{ID: 1, Price: 600, Stock: 10})
	orders := &fakeOrderRepo{createErr: repoErr}
	svc := usecase.NewOrderService(orders, products)

	got, err := svc.Create(context.Background(), "Cliente", "123", []domain.OrderProduct{{ProductID: 1, Quantity: 1}})
	if !errors.Is(err, repoErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, repoErr)
	}
	if got.ID != 0 || got.Total != 0 || len(got.Products) != 0 {
		t.Errorf("Create() = %+v, want the zero value on failure (no orphan claim)", got)
	}
}

func TestOrderService_Create_PropagatesProductLookupError(t *testing.T) {
	repoErr := errors.New("db down")
	products := &fakeProductRepo{err: repoErr}
	orders := &fakeOrderRepo{}
	svc := usecase.NewOrderService(orders, products)

	_, err := svc.Create(context.Background(), "Cliente", "123", []domain.OrderProduct{{ProductID: 1, Quantity: 1}})
	if !errors.Is(err, repoErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, repoErr)
	}
}

// TestOrderService_Create_PropagatesSpecificProductLookupError proves the
// fake (and by extension the real repo boundary) can fail a SPECIFIC
// product lookup without every other lookup in the same order also
// failing — an order referencing two products where only the SECOND one's
// Get() fails must still surface that specific error, not a coincidental
// pass because the fake's error was global.
func TestOrderService_Create_PropagatesSpecificProductLookupError(t *testing.T) {
	repoErr := errors.New("product 2 lookup failed")
	products := newFakeProductCatalog(domain.Product{ID: 1, Price: 600, Stock: 10})
	products.getErrByID = map[int]error{2: repoErr}
	orders := &fakeOrderRepo{}
	svc := usecase.NewOrderService(orders, products)

	items := []domain.OrderProduct{
		{ProductID: 1, Quantity: 1},
		{ProductID: 2, Quantity: 1},
	}
	_, err := svc.Create(context.Background(), "Cliente", "123", items)
	if !errors.Is(err, repoErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, repoErr)
	}
}

// --- Update / PUT (ports updateOrder, before/after_orderProduct_insert) ---

func TestOrderService_Update_RecomputesOrderPriceAndTotal(t *testing.T) {
	products := newFakeProductCatalog(domain.Product{ID: 1, Price: 600, Stock: 112})
	orders := &fakeOrderRepo{}
	svc := usecase.NewOrderService(orders, products)

	err := svc.Update(context.Background(), 27, []domain.OrderProduct{{ProductID: 1, Quantity: 4}})
	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if !orders.replaceCalled {
		t.Fatal("Update() should have called ReplaceProducts()")
	}
	if orders.replaceOrderID != 27 {
		t.Errorf("Update() called repo with orderID = %d, want 27", orders.replaceOrderID)
	}
	if orders.replaceTotal != 2400 {
		t.Errorf("Update() total = %v, want 2400 (600*4)", orders.replaceTotal)
	}
	if len(orders.replaceItems) != 1 || orders.replaceItems[0].Price != 2400 {
		t.Errorf("Update() items = %+v, want a single line with Price=2400", orders.replaceItems)
	}
}

// TestOrderService_Update_DoesNotTouchStock locks a deliberate PR6 scope
// decision: legacy updateOrder (FilasServer/orders.php:232) never adjusted
// product stock when line items changed. This port preserves that scope —
// stock reconciliation on PUT would require reading the OLD line items
// atomically with the new write, which needs the transactional repository
// (PR7), not the usecase alone. See backend/docs/legacy-quirks.md §14.
func TestOrderService_Update_DoesNotTouchStock(t *testing.T) {
	products := newFakeProductCatalog(domain.Product{ID: 1, Price: 600, Stock: 112})
	orders := &fakeOrderRepo{}
	svc := usecase.NewOrderService(orders, products)

	if err := svc.Update(context.Background(), 27, []domain.OrderProduct{{ProductID: 1, Quantity: 4}}); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if products.updateCalled {
		t.Error("Update() should not touch product stock via products.Update()")
	}
	if orders.transitionCalled {
		t.Error("Update() should not call TransitionState() — that is PatchState's contract")
	}
}

func TestOrderService_Update_RejectsZeroQuantity(t *testing.T) {
	products := newFakeProductCatalog(domain.Product{ID: 1, Price: 600, Stock: 112})
	orders := &fakeOrderRepo{}
	svc := usecase.NewOrderService(orders, products)

	err := svc.Update(context.Background(), 27, []domain.OrderProduct{{ProductID: 1, Quantity: 0}})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Update() error = %v, want %v", err, domain.ErrValidation)
	}
	if orders.replaceCalled {
		t.Error("Update() should not have called ReplaceProducts() on invalid quantity")
	}
}

// TestOrderService_Update_RejectsNegativeQuantity closes the Create/Update
// test-depth asymmetry flagged in the gate review: Create's quantity gate
// was tested for both 0 and negative, Update's only for 0.
func TestOrderService_Update_RejectsNegativeQuantity(t *testing.T) {
	products := newFakeProductCatalog(domain.Product{ID: 1, Price: 600, Stock: 112})
	orders := &fakeOrderRepo{}
	svc := usecase.NewOrderService(orders, products)

	err := svc.Update(context.Background(), 27, []domain.OrderProduct{{ProductID: 1, Quantity: -3}})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Update() error = %v, want %v", err, domain.ErrValidation)
	}
	if orders.replaceCalled {
		t.Error("Update() should not have called ReplaceProducts() on invalid quantity")
	}
}

func TestOrderService_Update_RejectsEmptyItems(t *testing.T) {
	orders := &fakeOrderRepo{}
	svc := usecase.NewOrderService(orders, &fakeProductRepo{})

	err := svc.Update(context.Background(), 27, nil)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Update() error = %v, want %v", err, domain.ErrValidation)
	}
	if orders.replaceCalled {
		t.Error("Update() should not have called ReplaceProducts() on empty items")
	}
}

func TestOrderService_Update_PropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	products := newFakeProductCatalog(domain.Product{ID: 1, Price: 600, Stock: 112})
	orders := &fakeOrderRepo{replaceErr: repoErr}
	svc := usecase.NewOrderService(orders, products)

	err := svc.Update(context.Background(), 27, []domain.OrderProduct{{ProductID: 1, Quantity: 1}})
	if !errors.Is(err, repoErr) {
		t.Errorf("Update() error = %v, want it to wrap %v", err, repoErr)
	}
}

// TestOrderService_Update_PropagatesProductLookupError closes the
// Create/Update test-depth asymmetry flagged in the gate review: Create's
// product-lookup-failure path was tested, Update's was not.
func TestOrderService_Update_PropagatesProductLookupError(t *testing.T) {
	repoErr := errors.New("db down")
	products := &fakeProductRepo{err: repoErr}
	orders := &fakeOrderRepo{}
	svc := usecase.NewOrderService(orders, products)

	err := svc.Update(context.Background(), 27, []domain.OrderProduct{{ProductID: 1, Quantity: 1}})
	if !errors.Is(err, repoErr) {
		t.Errorf("Update() error = %v, want it to wrap %v", err, repoErr)
	}
	if orders.replaceCalled {
		t.Error("Update() should not have called ReplaceProducts() when a product lookup fails")
	}
}

// --- PatchState (ports patchOrder, before_update_orders) ---

// fixedClock returns a deterministic time.Time for finishDate assertions —
// the reliability gate finding: a before/after time-window check is
// weaker than asserting the EXACT value the usecase computed.
func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestOrderService_PatchState_SetsFinishDateOnFinished(t *testing.T) {
	want := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	existing := domain.Order{ID: 5, State: domain.OrderStatePending}
	orders := &fakeOrderRepo{byID: map[int]domain.Order{5: existing}}
	svc := usecase.NewOrderServiceWithClock(orders, &fakeProductRepo{}, fixedClock(want))

	got, err := svc.PatchState(context.Background(), 5, domain.OrderStateFinished)
	if err != nil {
		t.Fatalf("PatchState() error = %v, want nil", err)
	}
	if !orders.transitionCalled || orders.transitionTarget != domain.OrderStateFinished {
		t.Fatalf("PatchState() should have called TransitionState() with target=finished, got called=%v target=%v", orders.transitionCalled, orders.transitionTarget)
	}
	if orders.transitionFinishDate == nil || !orders.transitionFinishDate.Equal(want) {
		t.Errorf("PatchState() finishDate = %v, want exactly %v (ported before_update_orders trigger, via injected clock)", orders.transitionFinishDate, want)
	}
	if len(orders.transitionDeltas) != 0 {
		t.Errorf("PatchState() deltas = %+v, want none on ->finished (no stock change)", orders.transitionDeltas)
	}
	if got.State != domain.OrderStateFinished || got.FinishDate == nil || !got.FinishDate.Equal(want) {
		t.Errorf("PatchState() returned order = %+v, want State=finished and FinishDate=%v", got, want)
	}
}

func TestOrderService_PatchState_DoesNotSetFinishDateOnCanceled(t *testing.T) {
	existing := domain.Order{ID: 5, State: domain.OrderStatePending}
	orders := &fakeOrderRepo{byID: map[int]domain.Order{5: existing}}
	svc := usecase.NewOrderService(orders, &fakeProductRepo{})

	got, err := svc.PatchState(context.Background(), 5, domain.OrderStateCanceled)
	if err != nil {
		t.Fatalf("PatchState() error = %v, want nil", err)
	}
	if orders.transitionFinishDate != nil {
		t.Errorf("PatchState() finishDate = %v, want nil on pending->canceled (trigger only fires for ->finished)", orders.transitionFinishDate)
	}
	if got.FinishDate != nil {
		t.Errorf("PatchState() returned FinishDate = %v, want nil", got.FinishDate)
	}
}

func TestOrderService_PatchState_RestoresStockOnCancel(t *testing.T) {
	existing := domain.Order{
		ID:    22,
		State: domain.OrderStatePending,
		Products: []domain.OrderProduct{
			{ProductID: 1, Quantity: 2},
			{ProductID: 2, Quantity: 3},
		},
	}
	orders := &fakeOrderRepo{byID: map[int]domain.Order{22: existing}}
	svc := usecase.NewOrderService(orders, &fakeProductRepo{})

	if _, err := svc.PatchState(context.Background(), 22, domain.OrderStateCanceled); err != nil {
		t.Fatalf("PatchState() error = %v, want nil", err)
	}

	deltas := deltaByProductID(orders.transitionDeltas)
	if len(deltas) != 2 {
		t.Fatalf("PatchState() passed %d deltas, want 2 (one restore per line)", len(deltas))
	}
	if deltas[1] != 2 {
		t.Errorf("PatchState() delta for product 1 = %d, want +2 (restore)", deltas[1])
	}
	if deltas[2] != 3 {
		t.Errorf("PatchState() delta for product 2 = %d, want +3 (restore)", deltas[2])
	}
}

// TestOrderService_PatchState_AggregatesDuplicateProductLinesOnCancel
// mirrors Create's duplicate-product aggregation (gate blocker #1) on the
// restore path: an order with two lines for the same product must restore
// via ONE net positive delta, not two separate entries.
func TestOrderService_PatchState_AggregatesDuplicateProductLinesOnCancel(t *testing.T) {
	existing := domain.Order{
		ID:    22,
		State: domain.OrderStatePending,
		Products: []domain.OrderProduct{
			{ProductID: 1, Quantity: 2},
			{ProductID: 1, Quantity: 3},
		},
	}
	orders := &fakeOrderRepo{byID: map[int]domain.Order{22: existing}}
	svc := usecase.NewOrderService(orders, &fakeProductRepo{})

	if _, err := svc.PatchState(context.Background(), 22, domain.OrderStateCanceled); err != nil {
		t.Fatalf("PatchState() error = %v, want nil", err)
	}

	if len(orders.transitionDeltas) != 1 {
		t.Fatalf("PatchState() passed %d deltas, want 1 (aggregated)", len(orders.transitionDeltas))
	}
	if orders.transitionDeltas[0].ProductID != 1 || orders.transitionDeltas[0].Delta != 5 {
		t.Errorf("PatchState() delta = %+v, want {ProductID:1 Delta:5} (aggregated 2+3)", orders.transitionDeltas[0])
	}
}

func TestOrderService_PatchState_DoesNotPassDeltasOnFinish(t *testing.T) {
	existing := domain.Order{
		ID:       5,
		State:    domain.OrderStatePending,
		Products: []domain.OrderProduct{{ProductID: 1, Quantity: 2}},
	}
	orders := &fakeOrderRepo{byID: map[int]domain.Order{5: existing}}
	svc := usecase.NewOrderService(orders, &fakeProductRepo{})

	if _, err := svc.PatchState(context.Background(), 5, domain.OrderStateFinished); err != nil {
		t.Fatalf("PatchState() error = %v, want nil", err)
	}
	if len(orders.transitionDeltas) != 0 {
		t.Errorf("PatchState() deltas = %+v, want none on ->finished", orders.transitionDeltas)
	}
}

// TestOrderService_PatchState_RejectsInvalidTargetState mirrors legacy
// patchOrder (FilasServer/orders.php:265): target state must be "finished"
// or "canceled" -> 400 otherwise.
func TestOrderService_PatchState_RejectsInvalidTargetState(t *testing.T) {
	existing := domain.Order{ID: 5, State: domain.OrderStatePending}
	orders := &fakeOrderRepo{byID: map[int]domain.Order{5: existing}}
	svc := usecase.NewOrderService(orders, &fakeProductRepo{})

	_, err := svc.PatchState(context.Background(), 5, domain.OrderStatePending)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("PatchState() error = %v, want %v", err, domain.ErrValidation)
	}
	if orders.transitionCalled {
		t.Error("PatchState() should not have called TransitionState() with an invalid target state")
	}
}

// TestOrderService_PatchState_ReturnsConflictWhenTransitionNotOK proves the
// 409 gate is now driven by the repository's ATOMIC CAS result
// (TransitionState returning ok=false), not a stale read-then-write check
// performed by the usecase — closing the TOCTOU gap the gate review
// flagged. The fake's `existing` order is deliberately left "pending" here
// to prove the usecase does NOT rely on that stale snapshot for the
// conflict decision; only the repo's ok=false return value does.
func TestOrderService_PatchState_ReturnsConflictWhenTransitionNotOK(t *testing.T) {
	existing := domain.Order{ID: 5, State: domain.OrderStatePending}
	orders := &fakeOrderRepo{byID: map[int]domain.Order{5: existing}, transitionConflict: true}
	svc := usecase.NewOrderService(orders, &fakeProductRepo{})

	_, err := svc.PatchState(context.Background(), 5, domain.OrderStateFinished)
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("PatchState() error = %v, want %v", err, domain.ErrConflict)
	}
	if !orders.transitionCalled {
		t.Error("PatchState() should still call TransitionState() — the CAS check happens THERE, not via a usecase-side pre-check")
	}
}

func TestOrderService_PatchState_PropagatesGetError(t *testing.T) {
	repoErr := errors.New("db down")
	orders := &fakeOrderRepo{err: repoErr}
	svc := usecase.NewOrderService(orders, &fakeProductRepo{})

	_, err := svc.PatchState(context.Background(), 5, domain.OrderStateFinished)
	if !errors.Is(err, repoErr) {
		t.Errorf("PatchState() error = %v, want it to wrap %v", err, repoErr)
	}
}

func TestOrderService_PatchState_ReturnsNotFoundForMissingOrder(t *testing.T) {
	orders := &fakeOrderRepo{byID: map[int]domain.Order{}}
	svc := usecase.NewOrderService(orders, &fakeProductRepo{})

	_, err := svc.PatchState(context.Background(), 999999, domain.OrderStateFinished)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("PatchState() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestOrderService_PatchState_PropagatesTransitionStateErrorOnFinish(t *testing.T) {
	repoErr := errors.New("tx rolled back")
	existing := domain.Order{ID: 5, State: domain.OrderStatePending}
	orders := &fakeOrderRepo{byID: map[int]domain.Order{5: existing}, transitionErr: repoErr}
	svc := usecase.NewOrderService(orders, &fakeProductRepo{})

	_, err := svc.PatchState(context.Background(), 5, domain.OrderStateFinished)
	if !errors.Is(err, repoErr) {
		t.Errorf("PatchState() error = %v, want it to wrap %v", err, repoErr)
	}
}

// TestOrderService_PatchState_PropagatesTransitionStateErrorOnCancel is the
// cancel-path counterpart: the atomic state-transition-plus-stock-restore
// call can fail independently of the finish path, and must surface
// cleanly (no partial-cancel claim).
func TestOrderService_PatchState_PropagatesTransitionStateErrorOnCancel(t *testing.T) {
	repoErr := errors.New("tx rolled back")
	existing := domain.Order{
		ID:       22,
		State:    domain.OrderStatePending,
		Products: []domain.OrderProduct{{ProductID: 1, Quantity: 2}},
	}
	orders := &fakeOrderRepo{byID: map[int]domain.Order{22: existing}, transitionErr: repoErr}
	svc := usecase.NewOrderService(orders, &fakeProductRepo{})

	got, err := svc.PatchState(context.Background(), 22, domain.OrderStateCanceled)
	if !errors.Is(err, repoErr) {
		t.Errorf("PatchState() error = %v, want it to wrap %v", err, repoErr)
	}
	if got.ID != 0 || got.State != "" || len(got.Products) != 0 {
		t.Errorf("PatchState() = %+v, want the zero value on failure (no partial-cancel claim)", got)
	}
}
