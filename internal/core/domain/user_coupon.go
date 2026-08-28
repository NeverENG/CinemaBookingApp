package domain

import "time"

type UserCouponStatus string

const (
	CouponUnused  UserCouponStatus = "UNUSED"
	CouponLocked  UserCouponStatus = "LOCKED"
	CouponUsed    UserCouponStatus = "USED"
	CouponExpired UserCouponStatus = "EXPIRED"
)

// UserCoupon 用户持有的优惠券实例。
type UserCoupon struct {
	ID         int64
	CouponNo   string
	UserID     int64
	TemplateID int64
	Status     UserCouponStatus
	OrderNo    string
	ExpireAt   time.Time
}
