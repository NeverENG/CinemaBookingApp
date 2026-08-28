package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

func changeFixture() (
	*fakeUserRepo,
	*fakeSessionRepo,
	*fakeSeatRepo,
	*fakeOrderRepo,
	*fakePaymentRepo,
	*fakeSeatLockRepo,
	*fakeCouponRepo,
	*fakeRefundRepo,
	*fakePointsRepo,
	*fakeBoxOfficeRepo,
) {
	now := time.Now()
	users := &fakeUserRepo{users: map[int64]*domain.User{1: {ID: 1, Status: "ACTIVE"}}}
	sessions := &fakeSessionRepo{sessions: map[int64]*domain.ShowSession{
		10: {ID: 10, CinemaID: 100, HallID: 1000, MovieID: 9, StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour), BasePriceCents: 5000, Status: domain.SessionOpen},
		11: {ID: 11, CinemaID: 100, HallID: 1000, MovieID: 9, StartTime: now.Add(5 * time.Hour), EndTime: now.Add(7 * time.Hour), BasePriceCents: 6000, Status: domain.SessionOpen},
	}}
	seats := &fakeSeatRepo{seats: map[int64]domain.Seat{
		1: {ID: 1, HallID: 1000, SeatNo: "A1", Status: domain.SeatEnabled},
		2: {ID: 2, HallID: 1000, SeatNo: "A2", Status: domain.SeatEnabled},
	}}
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{
		"O1": {OrderNo: "O1", UserID: 1, SessionID: 10, CinemaID: 100, MovieID: 9, Status: domain.OrderPaid, PaidCents: 5000, Version: 2},
	}}
	payments := &fakePaymentRepo{txns: map[string]*domain.PaymentTransaction{
		"T1": {TransactionNo: "T1", OrderNo: "O1", Status: domain.PaymentSuccess, Version: 1},
	}}
	locks := &fakeSeatLockRepo{}
	coupons := &fakeCouponRepo{}
	refunds := &fakeRefundRepo{}
	points := &fakePointsRepo{}
	box := &fakeBoxOfficeRepo{}
	return users, sessions, seats, orders, payments, locks, coupons, refunds, points, box
}

func newChangeTestSvc(
	users *fakeUserRepo,
	sessions *fakeSessionRepo,
	seats *fakeSeatRepo,
	orders *fakeOrderRepo,
	payments *fakePaymentRepo,
	locks *fakeSeatLockRepo,
	coupons *fakeCouponRepo,
	refunds *fakeRefundRepo,
	points *fakePointsRepo,
	box *fakeBoxOfficeRepo,
) *ChangeTicketSvc {
	orderSvc := NewOrderSvc(fakeTxManager{}, users, sessions, seats, locks, coupons, orders)
	paymentSvc := NewPaymentSvc(fakeTxManager{}, payments, &fakeCallbackRepo{}, orders, locks, coupons, points, box, &fakeMembershipRepo{})
	refundSvc := NewRefundSvc(fakeTxManager{}, orders, refunds, payments, locks, points, sessions, box)
	return NewChangeTicketSvc(orders, sessions, orderSvc, paymentSvc, refundSvc)
}

func TestChangeTicketHappy(t *testing.T) {
	users, sessions, seats, orders, payments, locks, coupons, refunds, points, box := changeFixture()
	svc := newChangeTestSvc(users, sessions, seats, orders, payments, locks, coupons, refunds, points, box)

	result, err := svc.Change(context.Background(), 1, "O1", 11, []int64{1})
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if result.NewOrderNo == "" || result.NewPaidCents != 6000 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if orders.orders["O1"].Status != domain.OrderRefunded {
		t.Fatalf("expected original REFUNDED, got %s", orders.orders["O1"].Status)
	}
	newOrder := orders.orders[result.NewOrderNo]
	if newOrder == nil || newOrder.Status != domain.OrderPaid {
		t.Fatalf("expected new order PAID, got %+v", newOrder)
	}
	if len(refunds.refunds) != 1 {
		t.Fatal("expected original refund created")
	}
}

func TestChangeTicketMovieMismatch(t *testing.T) {
	users, sessions, seats, orders, payments, locks, coupons, refunds, points, box := changeFixture()
	sessions.sessions[11].MovieID = 8
	svc := newChangeTestSvc(users, sessions, seats, orders, payments, locks, coupons, refunds, points, box)

	_, err := svc.Change(context.Background(), 1, "O1", 11, []int64{1})
	if !errors.Is(err, domain.ErrChangeMovieMismatch) {
		t.Fatalf("expected ErrChangeMovieMismatch, got %v", err)
	}
	if orders.orders["O1"].Status != domain.OrderPaid {
		t.Fatal("original order should stay PAID")
	}
}

func TestChangeTicketOriginalNotRefundable(t *testing.T) {
	users, sessions, seats, orders, payments, locks, coupons, refunds, points, box := changeFixture()
	sessions.sessions[10].StartTime = time.Now().Add(-time.Hour)
	svc := newChangeTestSvc(users, sessions, seats, orders, payments, locks, coupons, refunds, points, box)

	_, err := svc.Change(context.Background(), 1, "O1", 11, []int64{1})
	if !errors.Is(err, domain.ErrOrderNotRefundable) {
		t.Fatalf("expected ErrOrderNotRefundable, got %v", err)
	}
}
