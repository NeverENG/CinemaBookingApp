package service

import (
	"context"
	"errors"
	"testing"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

func TestAdminCouponCreateTemplateInvalid(t *testing.T) {
	coupons := &fakeCouponRepo{}
	svc := NewAdminCouponSvc(fakeTxManager{}, coupons, &fakeUserRepo{}, &fakeOperationLogRepo{})

	_, err := svc.CreateTemplate(context.Background(), 1, CouponTemplateInput{Name: "券", Type: "UNKNOWN"})
	if !errors.Is(err, domain.ErrCouponNotAvailable) {
		t.Fatalf("expected ErrCouponNotAvailable, got %v", err)
	}
}

func TestAdminCouponCreateAndIssue(t *testing.T) {
	users := &fakeUserRepo{users: map[int64]*domain.User{
		1: {ID: 1, Username: "a@b.com", Status: "ACTIVE"},
	}}
	coupons := &fakeCouponRepo{}
	logs := &fakeOperationLogRepo{}
	svc := NewAdminCouponSvc(fakeTxManager{}, coupons, users, logs)

	tpl, err := svc.CreateTemplate(context.Background(), 1, CouponTemplateInput{
		Name: "10元券", Type: domain.CouponTypeFixed, ValueCents: 1000, TotalQty: 100,
	})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}
	if tpl.ID == 0 || tpl.Status != "ACTIVE" || tpl.ValidDays != 30 {
		t.Fatalf("unexpected template: %+v", tpl)
	}

	coupon, err := svc.IssueToUser(context.Background(), 1, 1, tpl.ID)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if coupon.CouponNo == "" || coupon.Status != domain.CouponUnused {
		t.Fatalf("unexpected coupon: %+v", coupon)
	}
	if len(coupons.instances) != 1 {
		t.Fatal("expected instance created")
	}
}
