package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type fakeTxManager struct{}

func (fakeTxManager) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fakeUserRepo struct {
	users map[int64]*domain.User
}

func (f *fakeUserRepo) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (f *fakeUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	for _, u := range f.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

type fakeSessionRepo struct {
	sessions map[int64]*domain.ShowSession
}

func (f *fakeSessionRepo) GetSessionByID(ctx context.Context, id int64) (*domain.ShowSession, error) {
	if s, ok := f.sessions[id]; ok {
		return s, nil
	}
	return nil, domain.ErrSessionNotFound
}

type fakeSeatRepo struct {
	seats map[int64]domain.Seat
}

func (f *fakeSeatRepo) ListSeatsByIDs(ctx context.Context, ids []int64) ([]domain.Seat, error) {
	out := make([]domain.Seat, 0, len(ids))
	for _, id := range ids {
		if s, ok := f.seats[id]; ok {
			out = append(out, s)
		}
	}
	return out, nil
}

type fakeSeatLockRepo struct {
	lockErr error
	locked  []domain.SeatLock
	booked  []string
}

func (f *fakeSeatLockRepo) CreateLocks(ctx context.Context, locks []domain.SeatLock) error {
	if f.lockErr != nil {
		return f.lockErr
	}
	f.locked = append(f.locked, locks...)
	return nil
}

func (f *fakeSeatLockRepo) MarkBookedByOrderNo(ctx context.Context, orderNo string) error {
	f.booked = append(f.booked, orderNo)
	return nil
}

func (f *fakeSeatLockRepo) ReleaseByOrderNo(ctx context.Context, orderNo string, status domain.SeatLockStatus) error {
	return nil
}

type fakeCouponRepo struct {
	coupons   map[string]*domain.UserCoupon
	templates map[int64]*domain.CouponTemplate
	lockErr   error
}

func (f *fakeCouponRepo) GetByCouponNo(ctx context.Context, couponNo string) (*domain.UserCoupon, error) {
	if c, ok := f.coupons[couponNo]; ok {
		return c, nil
	}
	return nil, domain.ErrCouponNotAvailable
}

func (f *fakeCouponRepo) GetTemplateByID(ctx context.Context, templateID int64) (*domain.CouponTemplate, error) {
	if t, ok := f.templates[templateID]; ok {
		return t, nil
	}
	return nil, domain.ErrCouponNotAvailable
}

func (f *fakeCouponRepo) LockForOrder(ctx context.Context, couponNo, orderNo string) error {
	if f.lockErr != nil {
		return f.lockErr
	}
	c := f.coupons[couponNo]
	if c == nil || c.Status != domain.CouponUnused {
		return domain.ErrCouponNotAvailable
	}
	c.Status = domain.CouponLocked
	c.OrderNo = orderNo
	return nil
}

func (f *fakeCouponRepo) UnlockByOrderNo(ctx context.Context, orderNo string) error { return nil }

func (f *fakeCouponRepo) MarkUsedByOrderNo(ctx context.Context, orderNo string) error {
	for _, c := range f.coupons {
		if c.OrderNo == orderNo && c.Status == domain.CouponLocked {
			c.Status = domain.CouponUsed
		}
	}
	return nil
}

type fakeOrderRepo struct {
	orders map[string]*domain.Order
}

func (f *fakeOrderRepo) CreateOrder(ctx context.Context, order *domain.Order) error {
	if f.orders == nil {
		f.orders = make(map[string]*domain.Order)
	}
	if _, exists := f.orders[order.OrderNo]; exists {
		return domain.ErrInvalidTransition
	}
	f.orders[order.OrderNo] = order
	return nil
}

func (f *fakeOrderRepo) GetOrderByNo(ctx context.Context, orderNo string) (*domain.Order, error) {
	if o, ok := f.orders[orderNo]; ok {
		return o, nil
	}
	return nil, domain.ErrOrderNotFound
}

func (f *fakeOrderRepo) Transition(ctx context.Context, orderNo string, from, to domain.OrderStatus, version int32) error {
	return nil
}

func (f *fakeOrderRepo) IssueTickets(ctx context.Context, orderNo string, tickets []domain.OrderItem) error {
	o := f.orders[orderNo]
	if o == nil {
		return domain.ErrOrderNotFound
	}
	bySeat := make(map[int64]string, len(tickets))
	for _, t := range tickets {
		bySeat[t.SeatID] = t.TicketNo
	}
	for i := range o.Items {
		if no, ok := bySeat[o.Items[i].SeatID]; ok {
			o.Items[i].TicketNo = no
		}
	}
	return nil
}

func (f *fakeOrderRepo) ListOrdersByUserID(ctx context.Context, userID int64) ([]domain.Order, error) {
	return nil, nil
}

func newTestOrderSvc(
	users *fakeUserRepo,
	sessions *fakeSessionRepo,
	seats *fakeSeatRepo,
	locks *fakeSeatLockRepo,
	coupons *fakeCouponRepo,
	orders *fakeOrderRepo,
) *OrderSvc {
	return NewOrderSvc(fakeTxManager{}, users, sessions, seats, locks, coupons, orders)
}

func baseTestData() (*fakeUserRepo, *fakeSessionRepo, *fakeSeatRepo) {
	users := &fakeUserRepo{users: map[int64]*domain.User{1: {ID: 1, Status: "ACTIVE"}}}
	sessions := &fakeSessionRepo{sessions: map[int64]*domain.ShowSession{
		10: {
			ID:             10,
			CinemaID:       100,
			HallID:         1000,
			MovieID:        9,
			StartTime:      time.Now().Add(2 * time.Hour),
			EndTime:        time.Now().Add(4 * time.Hour),
			BasePriceCents: 5000,
			Status:         domain.SessionOpen,
		},
	}}
	seats := &fakeSeatRepo{seats: map[int64]domain.Seat{
		1: {ID: 1, HallID: 1000, SeatNo: "A1", Status: "ENABLED"},
		2: {ID: 2, HallID: 1000, SeatNo: "A2", Status: "ENABLED"},
	}}
	return users, sessions, seats
}

func TestCreateOrderHappyPath(t *testing.T) {
	users, sessions, seats := baseTestData()
	locks := &fakeSeatLockRepo{}
	orders := &fakeOrderRepo{}
	svc := newTestOrderSvc(users, sessions, seats, locks, &fakeCouponRepo{}, orders)

	order, err := svc.CreateOrder(context.Background(), CreateOrderInput{
		UserID:    1,
		SessionID: 10,
		SeatIDs:   []int64{1, 2},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Status != domain.OrderPendingPayment {
		t.Fatalf("expected PENDING_PAYMENT, got %s", order.Status)
	}
	if order.PaidCents != 10000 {
		t.Fatalf("expected paid 10000, got %d", order.PaidCents)
	}
	if len(order.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(order.Items))
	}
	if len(locks.locked) != 2 {
		t.Fatalf("expected 2 locks, got %d", len(locks.locked))
	}
	for _, lock := range locks.locked {
		if lock.OrderNo != order.OrderNo || lock.Status != domain.SeatLockLocked {
			t.Fatalf("lock not bound to order: %+v", lock)
		}
	}
}

func TestCreateOrderWithCoupon(t *testing.T) {
	users, sessions, seats := baseTestData()
	coupons := &fakeCouponRepo{
		coupons: map[string]*domain.UserCoupon{
			"C001": {ID: 1, CouponNo: "C001", UserID: 1, TemplateID: 5, Status: domain.CouponUnused, ExpireAt: time.Now().Add(24 * time.Hour)},
		},
		templates: map[int64]*domain.CouponTemplate{
			5: {ID: 5, Type: domain.CouponTypeFixed, ValueCents: 3000},
		},
	}
	orders := &fakeOrderRepo{}
	svc := newTestOrderSvc(users, sessions, seats, &fakeSeatLockRepo{}, coupons, orders)

	order, err := svc.CreateOrder(context.Background(), CreateOrderInput{
		UserID:    1,
		SessionID: 10,
		SeatIDs:   []int64{1},
		CouponNo:  "C001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.CouponCents != 3000 || order.PaidCents != 2000 {
		t.Fatalf("expected coupon 3000/paid 2000, got %d/%d", order.CouponCents, order.PaidCents)
	}
	if coupons.coupons["C001"].Status != domain.CouponLocked {
		t.Fatal("expected coupon locked")
	}
}

func TestCreateOrderSessionClosed(t *testing.T) {
	users, sessions, seats := baseTestData()
	sessions.sessions[10].Status = domain.SessionClosed
	svc := newTestOrderSvc(users, sessions, seats, &fakeSeatLockRepo{}, &fakeCouponRepo{}, &fakeOrderRepo{})

	_, err := svc.CreateOrder(context.Background(), CreateOrderInput{
		UserID:    1,
		SessionID: 10,
		SeatIDs:   []int64{1},
	})
	if !errors.Is(err, domain.ErrSessionNotBookable) {
		t.Fatalf("expected ErrSessionNotBookable, got %v", err)
	}
}

func TestCreateOrderSeatLockConflict(t *testing.T) {
	users, sessions, seats := baseTestData()
	locks := &fakeSeatLockRepo{lockErr: domain.ErrSeatLockConflict}
	svc := newTestOrderSvc(users, sessions, seats, locks, &fakeCouponRepo{}, &fakeOrderRepo{})

	_, err := svc.CreateOrder(context.Background(), CreateOrderInput{
		UserID:    1,
		SessionID: 10,
		SeatIDs:   []int64{1},
	})
	if !errors.Is(err, domain.ErrSeatLockConflict) {
		t.Fatalf("expected ErrSeatLockConflict, got %v", err)
	}
}

func TestCreateOrderCouponBelongsToOtherUser(t *testing.T) {
	users, sessions, seats := baseTestData()
	coupons := &fakeCouponRepo{
		coupons: map[string]*domain.UserCoupon{
			"C001": {ID: 1, CouponNo: "C001", UserID: 999, TemplateID: 5, Status: domain.CouponUnused, ExpireAt: time.Now().Add(24 * time.Hour)},
		},
		templates: map[int64]*domain.CouponTemplate{5: {ID: 5, Type: domain.CouponTypeFixed, ValueCents: 3000}},
	}
	svc := newTestOrderSvc(users, sessions, seats, &fakeSeatLockRepo{}, coupons, &fakeOrderRepo{})

	_, err := svc.CreateOrder(context.Background(), CreateOrderInput{
		UserID:    1,
		SessionID: 10,
		SeatIDs:   []int64{1},
		CouponNo:  "C001",
	})
	if !errors.Is(err, domain.ErrCouponNotAvailable) {
		t.Fatalf("expected ErrCouponNotAvailable, got %v", err)
	}
}
