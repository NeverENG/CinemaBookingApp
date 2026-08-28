package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

// UserCouponRepo 用户优惠券仓储接口。
// LockForOrder 用条件 UPDATE（status='UNUSED'）保证并发下只有一单能锁到券。
type UserCouponRepo interface {
	GetByCouponNo(ctx context.Context, couponNo string) (*domain.UserCoupon, error)
	GetTemplateByID(ctx context.Context, templateID int64) (*domain.CouponTemplate, error)
	CreateTemplate(ctx context.Context, template *domain.CouponTemplate) error
	ListTemplates(ctx context.Context) ([]domain.CouponTemplate, error)
	SetTemplateStatus(ctx context.Context, templateID int64, status string) error
	ListRedeemableTemplates(ctx context.Context) ([]domain.CouponTemplate, error)
	CreateInstance(ctx context.Context, coupon *domain.UserCoupon) error
	LockForOrder(ctx context.Context, couponNo, orderNo string) error
	UnlockByOrderNo(ctx context.Context, orderNo string) error
	MarkUsedByOrderNo(ctx context.Context, orderNo string) error
}
