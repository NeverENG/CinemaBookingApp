package service

import (
	"context"
	"testing"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

func TestExpireOverdueOrders(t *testing.T) {
	now := time.Now()
	couponID := int64(7)
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{
		"O1": {OrderNo: "O1", SessionID: 10, Status: domain.OrderPendingPayment, ExpireAt: now.Add(-time.Minute), Version: 1, CouponInstanceID: &couponID},
		"O2": {OrderNo: "O2", SessionID: 10, Status: domain.OrderPendingPayment, ExpireAt: now.Add(time.Minute), Version: 1},
	}}
	locks := &fakeSeatLockRepo{}
	coupons := &fakeCouponRepo{}
	svc := newTestOrderSvc(&fakeUserRepo{}, &fakeSessionRepo{}, &fakeSeatRepo{}, locks, coupons, orders)

	n, err := svc.ExpireOverdueOrders(context.Background(), now)
	if err != nil {
		t.Fatalf("expire: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 expired, got %d", n)
	}
	if orders.orders["O1"].Status != domain.OrderExpired {
		t.Fatalf("expected O1 EXPIRED, got %s", orders.orders["O1"].Status)
	}
	if orders.orders["O2"].Status != domain.OrderPendingPayment {
		t.Fatal("O2 should remain pending")
	}
	if len(locks.releasedOrders) != 1 || locks.releasedOrders[0] != "O1" {
		t.Fatalf("expected O1 locks released, got %v", locks.releasedOrders)
	}
	if len(coupons.unlocked) != 1 || coupons.unlocked[0] != "O1" {
		t.Fatalf("expected O1 coupon unlocked, got %v", coupons.unlocked)
	}
}

func TestRetryCallbacks(t *testing.T) {
	// 有效回调：可重试成功
	order := paidOrderFixture()
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{"O1": order}}
	payments := &fakePaymentRepo{}
	callbacks := &fakeCallbackRepo{}
	svc := newPaymentTestSvc(orders, payments, callbacks, &fakeSeatLockRepo{}, &fakeCouponRepo{}, &fakePointsRepo{}, &fakeBoxOfficeRepo{})

	tx, _ := svc.CreatePayment(context.Background(), CreatePaymentInput{OrderNo: "O1"})
	callbacks.callbacks = map[string]*domain.PaymentCallback{
		"E1": {EventID: "E1", TransactionNo: tx.TransactionNo, AmountCents: 10000, Status: domain.CallbackReceived},
		"E2": {EventID: "E2", TransactionNo: "T-NOT-EXIST", AmountCents: 10000, Status: domain.CallbackReceived},
	}

	failed, err := svc.RetryCallbacks(context.Background(), 10)
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	if failed != 1 {
		t.Fatalf("expected 1 failed, got %d", failed)
	}
	if order.Status != domain.OrderPaid {
		t.Fatalf("expected order PAID after retry, got %s", order.Status)
	}
	if callbacks.callbacks["E2"].RetryCount != 1 {
		t.Fatalf("expected E2 retry count 1, got %d", callbacks.callbacks["E2"].RetryCount)
	}
}
