package service

import (
	"context"
	"errors"
	"testing"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

func TestExchangeHappyPath(t *testing.T) {
	points := &fakePointsRepo{balance: 2000}
	coupons := &fakeCouponRepo{templates: map[int64]*domain.CouponTemplate{
		5: {ID: 5, Name: "10元券", Redeemable: true, RedeemPoints: 1000, ValidDays: 30},
	}}
	svc := NewPointsSvc(fakeTxManager{}, points, coupons)

	res, err := svc.Exchange(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}
	if res.CouponNo == "" || res.BalanceAfter != 1000 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if len(coupons.instances) != 1 {
		t.Fatal("expected coupon instance created")
	}
	for _, c := range coupons.instances {
		if c.Status != domain.CouponUnused || c.TemplateID != 5 {
			t.Fatalf("unexpected instance: %+v", c)
		}
	}
}

func TestExchangeInsufficientPoints(t *testing.T) {
	points := &fakePointsRepo{balance: 500}
	coupons := &fakeCouponRepo{templates: map[int64]*domain.CouponTemplate{
		5: {ID: 5, Name: "10元券", Redeemable: true, RedeemPoints: 1000},
	}}
	svc := NewPointsSvc(fakeTxManager{}, points, coupons)

	_, err := svc.Exchange(context.Background(), 1, 5)
	if !errors.Is(err, domain.ErrInsufficientPoints) {
		t.Fatalf("expected ErrInsufficientPoints, got %v", err)
	}
	if len(coupons.instances) != 0 {
		t.Fatal("no coupon should be issued")
	}
}

func TestExchangeNotRedeemable(t *testing.T) {
	points := &fakePointsRepo{balance: 2000}
	coupons := &fakeCouponRepo{templates: map[int64]*domain.CouponTemplate{
		5: {ID: 5, Name: "普通券", Redeemable: false},
	}}
	svc := NewPointsSvc(fakeTxManager{}, points, coupons)

	_, err := svc.Exchange(context.Background(), 1, 5)
	if !errors.Is(err, domain.ErrCouponNotAvailable) {
		t.Fatalf("expected ErrCouponNotAvailable, got %v", err)
	}
}
