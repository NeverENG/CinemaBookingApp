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

func (f *fakeUserRepo) Create(ctx context.Context, user *domain.User) error {
	if f.users == nil {
		f.users = make(map[int64]*domain.User)
	}
	user.ID = int64(len(f.users) + 1)
	f.users[user.ID] = user
	return nil
}

func (f *fakeUserRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	if u, ok := f.users[userID]; ok {
		u.PasswordHash = passwordHash
		return nil
	}
	return domain.ErrUserNotFound
}

type fakeSessionRepo struct {
	sessions map[int64]*domain.ShowSession
	recalc   []int64
}

func (f *fakeSessionRepo) GetSessionByID(ctx context.Context, id int64) (*domain.ShowSession, error) {
	if s, ok := f.sessions[id]; ok {
		return s, nil
	}
	return nil, domain.ErrSessionNotFound
}

func (f *fakeSessionRepo) Create(ctx context.Context, session *domain.ShowSession) error {
	if f.sessions == nil {
		f.sessions = make(map[int64]*domain.ShowSession)
	}
	session.ID = int64(len(f.sessions) + 1)
	f.sessions[session.ID] = session
	return nil
}

func (f *fakeSessionRepo) UpdatePrice(ctx context.Context, id int64, basePriceCents int64, priceRulesJSON string) error {
	if s, ok := f.sessions[id]; ok {
		s.BasePriceCents = basePriceCents
		return nil
	}
	return domain.ErrSessionNotFound
}

func (f *fakeSessionRepo) Cancel(ctx context.Context, id int64) error {
	if s, ok := f.sessions[id]; ok {
		s.Status = domain.SessionCanceled
		return nil
	}
	return domain.ErrSessionNotFound
}

func (f *fakeSessionRepo) ListOverlapping(ctx context.Context, hallID int64, start, end time.Time) ([]domain.ShowSession, error) {
	out := make([]domain.ShowSession, 0)
	for _, s := range f.sessions {
		if s.HallID == hallID &&
			s.Status != domain.SessionCanceled && s.Status != domain.SessionClosed &&
			s.StartTime.Before(end) && s.EndTime.After(start) {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (f *fakeSessionRepo) ListByFilter(ctx context.Context, movieID, cinemaID int64) ([]domain.ShowSession, error) {
	out := make([]domain.ShowSession, 0)
	for _, s := range f.sessions {
		if (movieID == 0 || s.MovieID == movieID) &&
			(cinemaID == 0 || s.CinemaID == cinemaID) &&
			(s.Status == domain.SessionOpen || s.Status == domain.SessionSoldOut) {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (f *fakeSessionRepo) RecalcStatus(ctx context.Context, sessionID int64) error {
	f.recalc = append(f.recalc, sessionID)
	return nil
}

type fakeSeatRepo struct {
	seats  map[int64]domain.Seat
	synced [][]domain.Seat
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

func (f *fakeSeatRepo) ListByHallID(ctx context.Context, hallID int64) ([]domain.Seat, error) {
	out := make([]domain.Seat, 0)
	for _, s := range f.seats {
		if s.HallID == hallID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakeSeatRepo) SyncSeats(ctx context.Context, hallID int64, seats []domain.Seat) error {
	f.synced = append(f.synced, seats)
	if f.seats == nil {
		f.seats = make(map[int64]domain.Seat)
	}
	for _, s := range seats {
		s.ID = int64(len(f.seats) + 1)
		s.HallID = hallID
		f.seats[s.ID] = s
	}
	return nil
}

type fakeSeatLockRepo struct {
	lockErr        error
	locked         []domain.SeatLock
	booked         []string
	released       []int64
	releasedOrders []string
	active         []domain.SeatLock
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
	f.releasedOrders = append(f.releasedOrders, orderNo)
	return nil
}

func (f *fakeSeatLockRepo) ReleaseBySessionID(ctx context.Context, sessionID int64, status domain.SeatLockStatus) error {
	f.released = append(f.released, sessionID)
	return nil
}

func (f *fakeSeatLockRepo) ListActiveBySessionID(ctx context.Context, sessionID int64) ([]domain.SeatLock, error) {
	out := make([]domain.SeatLock, 0)
	for _, l := range f.active {
		if l.SessionID == sessionID {
			out = append(out, l)
		}
	}
	return out, nil
}

type fakeCouponRepo struct {
	coupons   map[string]*domain.UserCoupon
	templates map[int64]*domain.CouponTemplate
	instances map[string]*domain.UserCoupon
	lockErr   error
	unlocked  []string
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

func (f *fakeCouponRepo) CreateTemplate(ctx context.Context, template *domain.CouponTemplate) error {
	if f.templates == nil {
		f.templates = make(map[int64]*domain.CouponTemplate)
	}
	template.ID = int64(len(f.templates) + 1)
	f.templates[template.ID] = template
	return nil
}

func (f *fakeCouponRepo) ListTemplates(ctx context.Context) ([]domain.CouponTemplate, error) {
	out := make([]domain.CouponTemplate, 0, len(f.templates))
	for _, t := range f.templates {
		out = append(out, *t)
	}
	return out, nil
}

func (f *fakeCouponRepo) SetTemplateStatus(ctx context.Context, templateID int64, status string) error {
	if t, ok := f.templates[templateID]; ok {
		t.Status = status
		return nil
	}
	return domain.ErrCouponNotAvailable
}

func (f *fakeCouponRepo) ListRedeemableTemplates(ctx context.Context) ([]domain.CouponTemplate, error) {
	out := make([]domain.CouponTemplate, 0)
	for _, t := range f.templates {
		if t.Redeemable {
			out = append(out, *t)
		}
	}
	return out, nil
}

func (f *fakeCouponRepo) CreateInstance(ctx context.Context, coupon *domain.UserCoupon) error {
	if f.instances == nil {
		f.instances = make(map[string]*domain.UserCoupon)
	}
	coupon.ID = int64(len(f.instances) + 1)
	f.instances[coupon.CouponNo] = coupon
	return nil
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

func (f *fakeCouponRepo) UnlockByOrderNo(ctx context.Context, orderNo string) error {
	f.unlocked = append(f.unlocked, orderNo)
	return nil
}

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
	o := f.orders[orderNo]
	if o == nil {
		return domain.ErrOrderNotFound
	}
	// 与真实 DB 语义一致：条件更新由 SQL WHERE 保证，这里直接更新存储对象
	o.Status = to
	o.Version++
	return nil
}

func (f *fakeOrderRepo) ListExpiredPending(ctx context.Context, now time.Time) ([]domain.Order, error) {
	out := make([]domain.Order, 0)
	for _, o := range f.orders {
		if o.Status == domain.OrderPendingPayment && o.ExpireAt.Before(now) {
			out = append(out, *o)
		}
	}
	return out, nil
}

func (f *fakeOrderRepo) ExpirePendingBySessionID(ctx context.Context, sessionID int64) ([]string, error) {
	var nos []string
	for _, o := range f.orders {
		if o.SessionID == sessionID && o.Status == domain.OrderPendingPayment {
			o.Status = domain.OrderExpired
			nos = append(nos, o.OrderNo)
		}
	}
	return nos, nil
}

func (f *fakeOrderRepo) ListPaidBySessionID(ctx context.Context, sessionID int64) ([]domain.Order, error) {
	out := make([]domain.Order, 0)
	for _, o := range f.orders {
		if o.SessionID == sessionID && o.Status == domain.OrderPaid {
			out = append(out, *o)
		}
	}
	return out, nil
}

func (f *fakeOrderRepo) CountPaidByMovieIDs(ctx context.Context, movieIDs []int64) (map[int64]int64, error) {
	counts := make(map[int64]int64)
	for _, o := range f.orders {
		if o.Status == domain.OrderPaid {
			counts[o.MovieID]++
		}
	}
	return counts, nil
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
	if order.Items[0].TicketNo == "" || order.Items[1].TicketNo == "" {
		t.Fatal("expected ticket_no generated at order creation")
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

func TestGetOrderOwnership(t *testing.T) {
	users := &fakeUserRepo{users: map[int64]*domain.User{1: {ID: 1, Status: "ACTIVE"}}}
	sessions := &fakeSessionRepo{}
	seats := &fakeSeatRepo{}
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{
		"O1": {OrderNo: "O1", UserID: 1, Status: domain.OrderPaid},
	}}
	svc := newTestOrderSvc(users, sessions, seats, &fakeSeatLockRepo{}, &fakeCouponRepo{}, orders)

	if _, err := svc.GetOrder(context.Background(), 1, "O1"); err != nil {
		t.Fatalf("owner query: %v", err)
	}
	if _, err := svc.GetOrder(context.Background(), 999, "O1"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
