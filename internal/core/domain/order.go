package domain

import (
	"errors"
	"time"
)

type OrderStatus string

const (
	OrderPayPending      OrderStatus = "PENDING_PAYMENT"
	OrderPaid            OrderStatus = "PAID"
	OrderCompleted       OrderStatus = "COMPLETED"
	OrderFail            OrderStatus = "FAIL"
	OrderCanceled        OrderStatus = "CANCELED"
	OrderExpire          OrderStatus = "EXPIRE"
	OrderRefund          OrderStatus = "REFUNDING"
	OrderRefundCompleted OrderStatus = "ReFUNDED"
)

var ErrInvalidTransition = errors.New("invalid order status transition")
var ErrMoneyInvalid = errors.New("invalid money")

var allowedTransitions = map[OrderStatus]map[OrderStatus]bool{
	OrderPayPending: {OrderPaid: true, OrderExpire: true, OrderCanceled: true},
	OrderPaid:       {OrderCompleted: true, OrderFail: true},
	OrderCompleted:  {OrderRefund: true},
	OrderRefund:     {OrderRefundCompleted: true},
}

func (s OrderStatus) CanTransition(to OrderStatus) bool {
	return allowedTransitions[s][to]
}

type Order struct {
	OrderNo       string
	UserID        int64
	SessionID     int64
	Status        OrderStatus
	TotalCents    int64
	DiscountCents int64
	CouponCents   int64
	PaidCents     int64
	ExpireAt      time.Time
	Version       int32
}

func (o *Order) CanMarkPaid() bool {
	return o.Status.CanTransition(OrderPaid)
}

func (o *Order) IsExpired(now time.Time) bool {
	return o.Status == OrderPayPending && now.After(o.ExpireAt)
}

func (o *Order) Settle(total, discount, coupon int64) error {
	if total < 0 || discount < 0 || coupon < 0 || total-discount-coupon < 0 {
		return ErrMoneyInvalid
	}
	o.TotalCents, o.DiscountCents, o.CouponCents = total, discount, coupon
	o.PaidCents = total - discount - coupon
	return nil
}

type Money int64

func (m Money) Cents() int64      { return int64(m) }
func (m Money) Add(n Money) Money { return m + n }
