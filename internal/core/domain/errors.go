package domain

import "errors"

// 领域错误哨兵：全部收敛在这里，业务层用 errors.Is 判断。
// HTTP 状态码映射不在这里（那是 HTTP 层的事）。
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionNotBookable = errors.New("session is not bookable")
	ErrSeatNotAvailable   = errors.New("seat not available")
	ErrSeatLockConflict   = errors.New("seat already locked")
	ErrCouponNotAvailable = errors.New("coupon not available")

	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderExpired      = errors.New("order expired")
	ErrInvalidTransition = errors.New("invalid order status transition")
	ErrMoneyInvalid      = errors.New("invalid money")

	ErrPaymentNotFound       = errors.New("payment transaction not found")
	ErrPaymentAmountMismatch = errors.New("payment amount mismatch")

	ErrAdminNotFound      = errors.New("admin not found")
	ErrRoleNotFound       = errors.New("role not found")
	ErrInvalidCredentials = errors.New("invalid username or password")

	ErrMovieNotFound     = errors.New("movie not found")
	ErrMovieInvalid      = errors.New("movie invalid")
	ErrHallNotFound      = errors.New("hall not found")
	ErrHallInvalid       = errors.New("hall invalid")
	ErrSeatLayoutInvalid = errors.New("seat layout invalid")

	ErrSessionTimeConflict    = errors.New("session time conflict")
	ErrSessionLockedForChange = errors.New("session locked for change")
	ErrSessionInvalid         = errors.New("session invalid")

	ErrForbidden      = errors.New("forbidden")
	ErrRefundNotFound = errors.New("refund not found")

	ErrBannerNotFound = errors.New("banner not found")
	ErrBannerInvalid  = errors.New("banner invalid")

	ErrInsufficientPoints = errors.New("insufficient points")

	ErrOrderNotRefundable = errors.New("order not refundable")
	ErrRefundExists       = errors.New("refund already exists")
)
