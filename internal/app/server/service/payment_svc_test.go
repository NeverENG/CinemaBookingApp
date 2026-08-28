package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type fakePaymentRepo struct {
	txns map[string]*domain.PaymentTransaction
}

func (f *fakePaymentRepo) CreateTransaction(ctx context.Context, tx *domain.PaymentTransaction) error {
	if f.txns == nil {
		f.txns = make(map[string]*domain.PaymentTransaction)
	}
	f.txns[tx.TransactionNo] = tx
	return nil
}

func (f *fakePaymentRepo) GetByTransactionNo(ctx context.Context, transactionNo string) (*domain.PaymentTransaction, error) {
	if tx, ok := f.txns[transactionNo]; ok {
		return tx, nil
	}
	return nil, domain.ErrPaymentNotFound
}

func (f *fakePaymentRepo) GetByOrderNo(ctx context.Context, orderNo string) (*domain.PaymentTransaction, error) {
	for _, tx := range f.txns {
		if tx.OrderNo == orderNo {
			return tx, nil
		}
	}
	return nil, domain.ErrPaymentNotFound
}

func (f *fakePaymentRepo) ListPendingOlderThan(ctx context.Context, before time.Time, limit int) ([]domain.PaymentTransaction, error) {
	out := make([]domain.PaymentTransaction, 0)
	for _, tx := range f.txns {
		if tx.Status == domain.PaymentPending && tx.CreatedAt.Before(before) {
			out = append(out, *tx)
		}
	}
	return out, nil
}

func (f *fakePaymentRepo) Transition(ctx context.Context, transactionNo string, from, to domain.PaymentStatus, version int32) error {
	tx := f.txns[transactionNo]
	if tx == nil {
		return domain.ErrPaymentNotFound
	}
	// 与真实 DB 语义一致：条件更新由 SQL 的 WHERE 保证，不依赖内存状态
	tx.Status = to
	tx.Version++
	return nil
}

func (f *fakePaymentRepo) ListOrderPaymentMismatches(ctx context.Context) ([]string, error) {
	return nil, nil
}

type fakeCallbackRepo struct {
	callbacks map[string]*domain.PaymentCallback
}

func (f *fakeCallbackRepo) InsertIfAbsent(ctx context.Context, cb *domain.PaymentCallback) (bool, error) {
	if f.callbacks == nil {
		f.callbacks = make(map[string]*domain.PaymentCallback)
	}
	if _, ok := f.callbacks[cb.EventID]; ok {
		return false, nil
	}
	cb.Status = domain.CallbackReceived
	f.callbacks[cb.EventID] = cb
	return true, nil
}

func (f *fakeCallbackRepo) MarkProcessed(ctx context.Context, eventID string) error {
	if cb, ok := f.callbacks[eventID]; ok {
		cb.Status = domain.CallbackProcessed
	}
	return nil
}

func (f *fakeCallbackRepo) MarkFailed(ctx context.Context, eventID, reason string) error {
	if cb, ok := f.callbacks[eventID]; ok {
		cb.Status = domain.CallbackFailed
	}
	return nil
}

func (f *fakeCallbackRepo) GetByEventID(ctx context.Context, eventID string) (*domain.PaymentCallback, error) {
	if cb, ok := f.callbacks[eventID]; ok {
		return cb, nil
	}
	return nil, domain.ErrPaymentNotFound
}

func (f *fakeCallbackRepo) ListPending(ctx context.Context, limit int) ([]domain.PaymentCallback, error) {
	out := make([]domain.PaymentCallback, 0)
	for _, cb := range f.callbacks {
		if (cb.Status == domain.CallbackReceived || cb.Status == domain.CallbackFailed) && cb.RetryCount < 5 {
			out = append(out, *cb)
		}
	}
	return out, nil
}

func (f *fakeCallbackRepo) IncrementRetry(ctx context.Context, eventID string) error {
	if cb, ok := f.callbacks[eventID]; ok {
		cb.RetryCount++
	}
	return nil
}

type fakePointsRepo struct {
	granted   []string
	reclaimed []string
	balance   int
	ledger    []domain.PointsLedger
}

func (f *fakePointsRepo) GrantOnPaid(ctx context.Context, userID int64, paidCents int64, orderNo string) error {
	f.granted = append(f.granted, orderNo)
	return nil
}

func (f *fakePointsRepo) ReclaimOnRefund(ctx context.Context, userID int64, refundCents int64, refundNo string) error {
	f.reclaimed = append(f.reclaimed, refundNo)
	return nil
}

func (f *fakePointsRepo) Exchange(ctx context.Context, userID int64, points int, exchangeNo string) (int, error) {
	if f.balance < points {
		return 0, domain.ErrInsufficientPoints
	}
	f.balance -= points
	return f.balance, nil
}

func (f *fakePointsRepo) GetBalance(ctx context.Context, userID int64) (int, error) {
	return f.balance, nil
}

func (f *fakePointsRepo) GetRecentLedger(ctx context.Context, userID int64, limit int) ([]domain.PointsLedger, error) {
	return f.ledger, nil
}

func (f *fakePointsRepo) ListBalanceMismatches(ctx context.Context) ([]int64, error) {
	return nil, nil
}

func paidOrderFixture() *domain.Order {
	return &domain.Order{
		OrderNo:    "O1",
		UserID:     1,
		SessionID:  10,
		CinemaID:   100,
		MovieID:    9,
		Status:     domain.OrderPendingPayment,
		TotalCents: 10000,
		PaidCents:  10000,
		Version:    1,
		ExpireAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:  time.Now(),
		Items:      []domain.OrderItem{{OrderNo: "O1", SeatID: 1, SeatNo: "A1", PriceCents: 10000}},
	}
}

func newPaymentTestSvc(
	orders *fakeOrderRepo,
	payments *fakePaymentRepo,
	callbacks *fakeCallbackRepo,
	locks *fakeSeatLockRepo,
	coupons *fakeCouponRepo,
	points *fakePointsRepo,
	box *fakeBoxOfficeRepo,
	members *fakeMembershipRepo,
) *PaymentSvc {
	return NewPaymentSvc(fakeTxManager{}, payments, callbacks, orders, locks, coupons, points, box, members)
}

func TestCreatePaymentIdempotent(t *testing.T) {
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{"O1": paidOrderFixture()}}
	payments := &fakePaymentRepo{}
	svc := newPaymentTestSvc(orders, payments, &fakeCallbackRepo{}, &fakeSeatLockRepo{}, &fakeCouponRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{}, &fakeMembershipRepo{})

	tx, err := svc.CreatePayment(context.Background(), CreatePaymentInput{OrderNo: "O1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Status != domain.PaymentPending || tx.AmountCents != 10000 {
		t.Fatalf("unexpected transaction: %+v", tx)
	}

	tx2, err := svc.CreatePayment(context.Background(), CreatePaymentInput{OrderNo: "O1"})
	if err != nil {
		t.Fatalf("unexpected error on retry: %v", err)
	}
	if tx2.TransactionNo != tx.TransactionNo {
		t.Fatal("retry should return the same transaction")
	}
}

func TestCreatePaymentExpiredOrder(t *testing.T) {
	order := paidOrderFixture()
	order.ExpireAt = time.Now().Add(-time.Minute)
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{"O1": order}}
	svc := newPaymentTestSvc(orders, &fakePaymentRepo{}, &fakeCallbackRepo{}, &fakeSeatLockRepo{}, &fakeCouponRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{}, &fakeMembershipRepo{})

	_, err := svc.CreatePayment(context.Background(), CreatePaymentInput{OrderNo: "O1"})
	if !errors.Is(err, domain.ErrOrderExpired) {
		t.Fatalf("expected ErrOrderExpired, got %v", err)
	}
}

func TestHandleMockCallbackHappyPath(t *testing.T) {
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{"O1": paidOrderFixture()}}
	payments := &fakePaymentRepo{}
	callbacks := &fakeCallbackRepo{}
	locks := &fakeSeatLockRepo{}
	points := &fakePointsRepo{}
	box := &fakeBoxOfficeRepo{}
	members := &fakeMembershipRepo{}
	coupons := &fakeCouponRepo{
		coupons: map[string]*domain.UserCoupon{
			"C001": {ID: 1, CouponNo: "C001", UserID: 1, Status: domain.CouponLocked, OrderNo: "O1"},
		},
	}
	svc := newPaymentTestSvc(orders, payments, callbacks, locks, coupons, points, box, members)

	tx, err := svc.CreatePayment(context.Background(), CreatePaymentInput{OrderNo: "O1"})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}

	err = svc.HandleMockCallback(context.Background(), MockCallbackInput{
		EventID:       "E1",
		TransactionNo: tx.TransactionNo,
		AmountCents:   10000,
	})
	if err != nil {
		t.Fatalf("handle callback: %v", err)
	}

	if orders.orders["O1"].Status != domain.OrderPaid {
		t.Fatalf("expected order PAID, got %s", orders.orders["O1"].Status)
	}
	if payments.txns[tx.TransactionNo].Status != domain.PaymentSuccess {
		t.Fatal("expected payment SUCCESS")
	}
	if len(locks.booked) != 1 || locks.booked[0] != "O1" {
		t.Fatalf("expected locks booked, got %v", locks.booked)
	}
	if coupons.coupons["C001"].Status != domain.CouponUsed {
		t.Fatal("expected coupon USED")
	}
	if callbacks.callbacks["E1"].Status != domain.CallbackProcessed {
		t.Fatal("expected callback PROCESSED")
	}
	if len(points.granted) != 1 || points.granted[0] != "O1" {
		t.Fatalf("expected points granted for O1, got %v", points.granted)
	}
	if len(box.events) != 1 || box.events[0].BizType != domain.BoxOrderPaid || box.events[0].GrossDelta != 10000 {
		t.Fatalf("expected paid box event, got %+v", box.events)
	}
	if members.upgrades != 1 {
		t.Fatalf("expected membership upgrade check, got %d", members.upgrades)
	}
}

func TestHandleMockCallbackDuplicateEvent(t *testing.T) {
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{"O1": paidOrderFixture()}}
	payments := &fakePaymentRepo{}
	callbacks := &fakeCallbackRepo{}
	svc := newPaymentTestSvc(orders, payments, callbacks, &fakeSeatLockRepo{}, &fakeCouponRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{}, &fakeMembershipRepo{})

	tx, _ := svc.CreatePayment(context.Background(), CreatePaymentInput{OrderNo: "O1"})
	in := MockCallbackInput{EventID: "E1", TransactionNo: tx.TransactionNo, AmountCents: 10000}

	if err := svc.HandleMockCallback(context.Background(), in); err != nil {
		t.Fatalf("first callback: %v", err)
	}
	if err := svc.HandleMockCallback(context.Background(), in); err != nil {
		t.Fatalf("duplicate callback should be idempotent, got %v", err)
	}
	if len(callbacks.callbacks) != 1 {
		t.Fatalf("expected 1 callback record, got %d", len(callbacks.callbacks))
	}
}

func TestHandleMockCallbackAmountMismatch(t *testing.T) {
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{"O1": paidOrderFixture()}}
	payments := &fakePaymentRepo{}
	svc := newPaymentTestSvc(orders, payments, &fakeCallbackRepo{}, &fakeSeatLockRepo{}, &fakeCouponRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{}, &fakeMembershipRepo{})

	tx, _ := svc.CreatePayment(context.Background(), CreatePaymentInput{OrderNo: "O1"})
	err := svc.HandleMockCallback(context.Background(), MockCallbackInput{
		EventID:       "E1",
		TransactionNo: tx.TransactionNo,
		AmountCents:   1,
	})
	if !errors.Is(err, domain.ErrPaymentAmountMismatch) {
		t.Fatalf("expected ErrPaymentAmountMismatch, got %v", err)
	}
	if orders.orders["O1"].Status != domain.OrderPendingPayment {
		t.Fatal("order should remain PENDING_PAYMENT")
	}
}

func TestMockPay(t *testing.T) {
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{"O1": paidOrderFixture()}}
	payments := &fakePaymentRepo{}
	callbacks := &fakeCallbackRepo{}
	svc := newPaymentTestSvc(orders, payments, callbacks, &fakeSeatLockRepo{}, &fakeCouponRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{}, &fakeMembershipRepo{})

	tx, _ := svc.CreatePayment(context.Background(), CreatePaymentInput{OrderNo: "O1"})
	if err := svc.MockPay(context.Background(), 1, tx.TransactionNo); err != nil {
		t.Fatalf("mock pay: %v", err)
	}
	if orders.orders["O1"].Status != domain.OrderPaid {
		t.Fatal("expected order PAID after mock pay")
	}
	if len(callbacks.callbacks) != 1 {
		t.Fatal("expected one callback record")
	}
}

func TestMockPayWrongUser(t *testing.T) {
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{"O1": paidOrderFixture()}}
	payments := &fakePaymentRepo{}
	svc := newPaymentTestSvc(orders, payments, &fakeCallbackRepo{}, &fakeSeatLockRepo{}, &fakeCouponRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{}, &fakeMembershipRepo{})

	tx, _ := svc.CreatePayment(context.Background(), CreatePaymentInput{OrderNo: "O1"})
	if err := svc.MockPay(context.Background(), 999, tx.TransactionNo); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestGetByOrderOwnership(t *testing.T) {
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{"O1": paidOrderFixture()}}
	payments := &fakePaymentRepo{}
	svc := newPaymentTestSvc(orders, payments, &fakeCallbackRepo{}, &fakeSeatLockRepo{}, &fakeCouponRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{}, &fakeMembershipRepo{})

	tx, _ := svc.CreatePayment(context.Background(), CreatePaymentInput{OrderNo: "O1"})
	if _, err := svc.GetByOrder(context.Background(), 1, "O1"); err != nil {
		t.Fatalf("owner query: %v", err)
	}
	if _, err := svc.GetByOrder(context.Background(), 999, "O1"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if tx == nil {
		t.Fatal("unexpected nil tx")
	}
}

func TestCloseExpiredPending(t *testing.T) {
	now := time.Now()
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{
		"O1": {OrderNo: "O1", Status: domain.OrderPendingPayment, Version: 1},
		"O2": {OrderNo: "O2", Status: domain.OrderPaid, Version: 1},
	}}
	payments := &fakePaymentRepo{txns: map[string]*domain.PaymentTransaction{
		"T1": {TransactionNo: "T1", OrderNo: "O1", Status: domain.PaymentPending, Version: 1, CreatedAt: now.Add(-20 * time.Minute)},
		"T2": {TransactionNo: "T2", OrderNo: "O2", Status: domain.PaymentPending, Version: 1, CreatedAt: now.Add(-20 * time.Minute)},
	}}
	svc := newPaymentTestSvc(orders, payments, &fakeCallbackRepo{}, &fakeSeatLockRepo{}, &fakeCouponRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{}, &fakeMembershipRepo{})

	n, err := svc.CloseExpiredPending(context.Background(), 10)
	if err != nil {
		t.Fatalf("close expired: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 closed, got %d", n)
	}
	if payments.txns["T1"].Status != domain.PaymentClosed {
		t.Fatal("expected T1 CLOSED")
	}
	if payments.txns["T2"].Status != domain.PaymentPending {
		t.Fatal("T2 must not be closed (order paid)")
	}
}
