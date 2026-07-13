package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

// fakeOrderRepo uses pointer receivers so Create/ReplaceProducts/UpdateState
// tests can record what was actually passed to the repository, mirroring
// the fakeProductRepo/fakeNewsRepo convention established in PR3-PR5.
type fakeOrderRepo struct {
	orders []domain.Order
	byID   map[int]domain.Order
	err    error

	createErr    error
	createdOrder domain.Order
	createReturn domain.Order // if zero-value, Create echoes the input with ID=9001

	replaceErr     error
	replaceCalled  bool
	replaceOrderID int
	replaceItems   []domain.OrderProduct
	replaceTotal   float64

	updateStateErr        error
	updateStateCalled     bool
	updateStateOrderID    int
	updateStateState      domain.OrderState
	updateStateFinishDate *time.Time
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

func (f *fakeOrderRepo) Create(ctx context.Context, o domain.Order) (domain.Order, error) {
	if f.createErr != nil {
		return domain.Order{}, f.createErr
	}
	f.createdOrder = o
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

func (f *fakeOrderRepo) UpdateState(ctx context.Context, orderID int, state domain.OrderState, finishDate *time.Time) error {
	f.updateStateCalled = true
	f.updateStateOrderID = orderID
	f.updateStateState = state
	f.updateStateFinishDate = finishDate
	return f.updateStateErr
}

func newFakeProductCatalog(products ...domain.Product) *fakeProductRepo {
	byID := make(map[int]domain.Product, len(products))
	for _, p := range products {
		byID[p.ID] = p
	}
	return &fakeProductRepo{byID: byID}
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

func TestOrderService_Create_DecrementsStockPerLine(t *testing.T) {
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

	if len(products.updateCalls) != 2 {
		t.Fatalf("Create() called products.Update() %d times, want 2 (one stock decrement per line)", len(products.updateCalls))
	}
	byID := make(map[int]domain.Product, len(products.updateCalls))
	for _, c := range products.updateCalls {
		byID[c.id] = c.product
	}
	if byID[1].Stock != 110 {
		t.Errorf("Create() product 1 stock after decrement = %d, want 110 (112-2)", byID[1].Stock)
	}
	if byID[2].Stock != 0 {
		t.Errorf("Create() product 2 stock after decrement = %d, want 0 (5-5)", byID[2].Stock)
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
	if len(products.updateCalls) != 0 {
		t.Errorf("Create() should not have touched stock, but made %d update calls", len(products.updateCalls))
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

func TestOrderService_Create_PropagatesOrderRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	products := newFakeProductCatalog(domain.Product{ID: 1, Price: 600, Stock: 10})
	orders := &fakeOrderRepo{createErr: repoErr}
	svc := usecase.NewOrderService(orders, products)

	_, err := svc.Create(context.Background(), "Cliente", "123", []domain.OrderProduct{{ProductID: 1, Quantity: 1}})
	if !errors.Is(err, repoErr) {
		t.Errorf("Create() error = %v, want it to wrap %v", err, repoErr)
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
	if len(products.updateCalls) != 0 {
		t.Errorf("Update() should not touch product stock, but made %d update calls", len(products.updateCalls))
	}
}

func TestOrderService_Update_RejectsNonPositiveQuantity(t *testing.T) {
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

// --- PatchState (ports patchOrder, before_update_orders) ---

func TestOrderService_PatchState_SetsFinishDateOnFinished(t *testing.T) {
	existing := domain.Order{ID: 5, State: domain.OrderStatePending}
	orders := &fakeOrderRepo{byID: map[int]domain.Order{5: existing}}
	svc := usecase.NewOrderService(orders, &fakeProductRepo{})

	before := time.Now()
	got, err := svc.PatchState(context.Background(), 5, domain.OrderStateFinished)
	after := time.Now()
	if err != nil {
		t.Fatalf("PatchState() error = %v, want nil", err)
	}
	if !orders.updateStateCalled || orders.updateStateState != domain.OrderStateFinished {
		t.Fatalf("PatchState() should have called UpdateState() with state=finished, got called=%v state=%v", orders.updateStateCalled, orders.updateStateState)
	}
	if orders.updateStateFinishDate == nil {
		t.Fatal("PatchState() should have set finishDate on pending->finished (ported before_update_orders trigger)")
	}
	if orders.updateStateFinishDate.Before(before) || orders.updateStateFinishDate.After(after) {
		t.Errorf("PatchState() finishDate = %v, want between %v and %v", orders.updateStateFinishDate, before, after)
	}
	if got.State != domain.OrderStateFinished || got.FinishDate == nil {
		t.Errorf("PatchState() returned order = %+v, want State=finished and FinishDate set", got)
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
	if orders.updateStateFinishDate != nil {
		t.Errorf("PatchState() finishDate = %v, want nil on pending->canceled (trigger only fires for ->finished)", orders.updateStateFinishDate)
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
	products := newFakeProductCatalog(
		domain.Product{ID: 1, Price: 850, Stock: 5},
		domain.Product{ID: 2, Price: 700, Stock: 0},
	)
	orders := &fakeOrderRepo{byID: map[int]domain.Order{22: existing}}
	svc := usecase.NewOrderService(orders, products)

	if _, err := svc.PatchState(context.Background(), 22, domain.OrderStateCanceled); err != nil {
		t.Fatalf("PatchState() error = %v, want nil", err)
	}

	if len(products.updateCalls) != 2 {
		t.Fatalf("PatchState() made %d stock update calls, want 2 (one restore per line)", len(products.updateCalls))
	}
	byID := make(map[int]domain.Product, len(products.updateCalls))
	for _, c := range products.updateCalls {
		byID[c.id] = c.product
	}
	if byID[1].Stock != 7 {
		t.Errorf("PatchState() product 1 stock after restore = %d, want 7 (5+2)", byID[1].Stock)
	}
	if byID[2].Stock != 3 {
		t.Errorf("PatchState() product 2 stock after restore = %d, want 3 (0+3)", byID[2].Stock)
	}
}

func TestOrderService_PatchState_DoesNotTouchStockOnFinish(t *testing.T) {
	existing := domain.Order{
		ID:       5,
		State:    domain.OrderStatePending,
		Products: []domain.OrderProduct{{ProductID: 1, Quantity: 2}},
	}
	products := newFakeProductCatalog(domain.Product{ID: 1, Price: 600, Stock: 10})
	orders := &fakeOrderRepo{byID: map[int]domain.Order{5: existing}}
	svc := usecase.NewOrderService(orders, products)

	if _, err := svc.PatchState(context.Background(), 5, domain.OrderStateFinished); err != nil {
		t.Fatalf("PatchState() error = %v, want nil", err)
	}
	if len(products.updateCalls) != 0 {
		t.Errorf("PatchState() should not touch stock on ->finished, but made %d update calls", len(products.updateCalls))
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
	if orders.updateStateCalled {
		t.Error("PatchState() should not have called UpdateState() with an invalid target state")
	}
}

// TestOrderService_PatchState_RejectsWhenNotPending mirrors legacy
// patchOrder (FilasServer/orders.php:273): current state must be "pending"
// -> 409 otherwise. This also proves finishDate is never re-set/re-touched
// on an already-finished (or canceled) order, since the gate rejects the
// transition entirely before any state mutation.
func TestOrderService_PatchState_RejectsWhenNotPending(t *testing.T) {
	cases := []domain.OrderState{domain.OrderStateFinished, domain.OrderStateCanceled}
	for _, current := range cases {
		t.Run(string(current), func(t *testing.T) {
			existing := domain.Order{ID: 5, State: current}
			orders := &fakeOrderRepo{byID: map[int]domain.Order{5: existing}}
			svc := usecase.NewOrderService(orders, &fakeProductRepo{})

			_, err := svc.PatchState(context.Background(), 5, domain.OrderStateFinished)
			if !errors.Is(err, domain.ErrConflict) {
				t.Errorf("PatchState() error = %v, want %v", err, domain.ErrConflict)
			}
			if orders.updateStateCalled {
				t.Error("PatchState() should not have called UpdateState() on a non-pending order")
			}
		})
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

func TestOrderService_PatchState_PropagatesUpdateStateError(t *testing.T) {
	repoErr := errors.New("db down")
	existing := domain.Order{ID: 5, State: domain.OrderStatePending}
	orders := &fakeOrderRepo{byID: map[int]domain.Order{5: existing}, updateStateErr: repoErr}
	svc := usecase.NewOrderService(orders, &fakeProductRepo{})

	_, err := svc.PatchState(context.Background(), 5, domain.OrderStateFinished)
	if !errors.Is(err, repoErr) {
		t.Errorf("PatchState() error = %v, want it to wrap %v", err, repoErr)
	}
}
