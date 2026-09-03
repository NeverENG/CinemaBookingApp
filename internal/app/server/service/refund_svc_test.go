package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

func refundFixture() (*fakeOrderRepo, *fakeSessionRepo) {
	now := time.Now()
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{
		"O1": {OrderNo: "O1", UserID: 1, SessionID: 10, Status: domain.OrderPaid, PaidCents: 5000, Version: 2},
	}}
	sessions := &fakeSessionRepo{sessions: map[int64]*domain.ShowSession{
		10: {ID: 10, HallID: 1000, MovieID: 9, StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour), Status: domain.SessionOpen},
	}}
	return orders, sessions
}

func newRefundTestSvc(orders *fakeOrderRepo, sessions *fakeSessionRepo, refunds *fakeRefundRepo, payments *fakePaymentRepo, locks *fakeSeatLockRepo, points *fakePointsRepo, box *fakeBoxOfficeRepo) *RefundSvc {
	return NewRefundSvc(fakeTxManager{}, orders, refunds, payments, locks, points, sessions, box)
}

func TestApplyRefundHappy(t *testing.T) {
	orders, sessions := refundFixture()
	refunds := &fakeRefundRepo{}
	svc := newRefundTestSvc(orders, sessions, refunds, &fakePaymentRepo{}, &fakeSeatLockRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{})

	refund, err := svc.ApplyRefund(context.Background(), 1, ApplyRefundInput{OrderNo: "O1", Reason: "计划有变"})
	if err != nil {
		t.Fatalf("apply refund: %v", err)
	}
	if refund.Status != domain.RefundPending || refund.AmountCents != 5000 {
		t.Fatalf("unexpected refund: %+v", refund)
	}
	if orders.orders["O1"].Status != domain.OrderRefunding {
		t.Fatalf("expected order REFUNDING, got %s", orders.orders["O1"].Status)
	}
}

func TestApplyRefundOwnership(t *testing.T) {
	orders, sessions := refundFixture()
	svc := newRefundTestSvc(orders, sessions, &fakeRefundRepo{}, &fakePaymentRepo{}, &fakeSeatLockRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{})

	_, err := svc.ApplyRefund(context.Background(), 999, ApplyRefundInput{OrderNo: "O1"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestApplyRefundNotRefundable(t *testing.T) {
	orders, sessions := refundFixture()
	sessions.sessions[10].StartTime = time.Now().Add(-time.Hour)
	svc := newRefundTestSvc(orders, sessions, &fakeRefundRepo{}, &fakePaymentRepo{}, &fakeSeatLockRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{})

	_, err := svc.ApplyRefund(context.Background(), 1, ApplyRefundInput{OrderNo: "O1"})
	if !errors.Is(err, domain.ErrOrderNotRefundable) {
		t.Fatalf("expected ErrOrderNotRefundable, got %v", err)
	}
}

func TestApplyRefundIdempotent(t *testing.T) {
	orders, sessions := refundFixture()
	refunds := &fakeRefundRepo{}
	svc := newRefundTestSvc(orders, sessions, refunds, &fakePaymentRepo{}, &fakeSeatLockRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{})

	r1, err := svc.ApplyRefund(context.Background(), 1, ApplyRefundInput{OrderNo: "O1"})
	if err != nil {
		t.Fatalf("first apply: %v", err)
	}
	r2, err := svc.ApplyRefund(context.Background(), 1, ApplyRefundInput{OrderNo: "O1"})
	if err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if r1.RefundNo != r2.RefundNo {
		t.Fatal("expected same refund on retry")
	}
}

func TestHandleRefundCallbackHappy(t *testing.T) {
	orders, sessions := refundFixture()
	orders.orders["O1"].Status = domain.OrderRefunding
	refunds := &fakeRefundRepo{}
	refund := &domain.Refund{
		RefundNo:    "RF1",
		OrderNo:     "O1",
		UserID:      1,
		AmountCents: 5000,
		Status:      domain.RefundPending,
	}
	if err := refunds.Create(context.Background(), refund); err != nil {
		t.Fatalf("seed refund: %v", err)
	}
	payments := &fakePaymentRepo{txns: map[string]*domain.PaymentTransaction{
		"T1": {TransactionNo: "T1", OrderNo: "O1", Status: domain.PaymentSuccess, Version: 1},
	}}
	locks := &fakeSeatLockRepo{}
	points := &fakePointsRepo{}
	box := &fakeBoxOfficeRepo{}
	svc := newRefundTestSvc(orders, sessions, refunds, payments, locks, points, box)

	if err := svc.HandleMockCallback(context.Background(), "RF1"); err != nil {
		t.Fatalf("callback: %v", err)
	}
	if orders.orders["O1"].Status != domain.OrderRefunded {
		t.Fatalf("expected order REFUNDED, got %s", orders.orders["O1"].Status)
	}
	if payments.txns["T1"].Status != domain.PaymentRefunded {
		t.Fatal("expected payment REFUNDED")
	}
	if len(locks.releasedOrders) != 1 {
		t.Fatal("expected locks released")
	}
	if len(points.reclaimed) != 1 {
		t.Fatal("expected points reclaimed")
	}
	if refunds.refunds[refund.RefundNo].Status != domain.RefundSuccess {
		t.Fatal("expected refund SUCCESS")
	}
	if len(box.events) != 1 || box.events[0].BizType != domain.BoxOrderRefund || box.events[0].RefundDelta != 5000 {
		t.Fatalf("expected refund box event, got %+v", box.events)
	}
}

func TestHandleRefundCallbackDuplicate(t *testing.T) {
	orders, sessions := refundFixture()
	orders.orders["O1"].Status = domain.OrderRefunding
	refunds := &fakeRefundRepo{}
	if err := refunds.Create(context.Background(), &domain.Refund{
		RefundNo:    "RF1",
		OrderNo:     "O1",
		UserID:      1,
		AmountCents: 5000,
		Status:      domain.RefundPending,
	}); err != nil {
		t.Fatalf("seed refund: %v", err)
	}
	svc := newRefundTestSvc(orders, sessions, refunds, &fakePaymentRepo{}, &fakeSeatLockRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{})

	if err := svc.HandleMockCallback(context.Background(), "RF1"); err != nil {
		t.Fatalf("first callback: %v", err)
	}
	if err := svc.HandleMockCallback(context.Background(), "RF1"); err != nil {
		t.Fatalf("duplicate callback should be idempotent, got %v", err)
	}
}

func TestHandleRefundCallbackForUserChecksOwnership(t *testing.T) {
	orders, sessions := refundFixture()
	orders.orders["O1"].Status = domain.OrderRefunding
	refunds := &fakeRefundRepo{}
	if err := refunds.Create(context.Background(), &domain.Refund{RefundNo: "RF-owner", OrderNo: "O1", UserID: 1, AmountCents: 5000, Status: domain.RefundPending}); err != nil {
		t.Fatal(err)
	}
	svc := newRefundTestSvc(orders, sessions, refunds, &fakePaymentRepo{}, &fakeSeatLockRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{})
	if err := svc.HandleMockCallbackForUser(context.Background(), 2, "RF-owner"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
