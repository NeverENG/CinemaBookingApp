package domain

import "time"

type OrderStatus string

const (
	OrderPendingPayment OrderStatus = "PENDING_PAYMENT"
	OrderPaid           OrderStatus = "PAID"
	OrderCompleted      OrderStatus = "COMPLETED"
	OrderCanceled       OrderStatus = "CANCELED"
	OrderExpired        OrderStatus = "EXPIRED"
	OrderRefunding      OrderStatus = "REFUNDING"
	OrderRefunded       OrderStatus = "REFUNDED"
)

type OrderEvent string

const (
	OrderEventPaySuccess    OrderEvent = "PAY_SUCCESS"
	OrderEventUserCancel    OrderEvent = "USER_CANCEL"
	OrderEventTimeout       OrderEvent = "TIMEOUT"
	OrderEventComplete      OrderEvent = "COMPLETE"
	OrderEventApplyRefund   OrderEvent = "APPLY_REFUND"
	OrderEventRefundSuccess OrderEvent = "REFUND_SUCCESS"
	OrderEventRefundFail    OrderEvent = "REFUND_FAIL"
)

var orderTransitions = map[OrderStatus]map[OrderEvent]OrderStatus{
	OrderPendingPayment: {
		OrderEventPaySuccess: OrderPaid,
		OrderEventUserCancel: OrderCanceled,
		OrderEventTimeout:    OrderExpired,
	},
	OrderPaid: {
		OrderEventComplete:    OrderCompleted,
		OrderEventApplyRefund: OrderRefunding,
	},
	OrderRefunding: {
		OrderEventRefundSuccess: OrderRefunded,
		OrderEventRefundFail:    OrderPaid,
	},
}

func (s OrderStatus) CanTransition(event OrderEvent) bool {
	_, ok := orderTransitions[s][event]
	return ok
}

type Order struct {
	OrderNo          string
	UserID           int64
	SessionID        int64
	CinemaID         int64
	MovieID          int64
	Status           OrderStatus
	TotalCents       int64
	DiscountCents    int64
	CouponCents      int64
	PaidCents        int64
	CouponInstanceID *int64
	ExpireAt         time.Time
	Version          int32
	Items            []OrderItem
	CreatedAt        time.Time
	PaidAt           *time.Time
}

func (o *Order) Transition(event OrderEvent) error {
	next, ok := orderTransitions[o.Status][event]
	if !ok {
		return ErrInvalidTransition
	}
	o.Status = next
	return nil
}

func (o *Order) IsExpired(now time.Time) bool {
	return o.Status == OrderPendingPayment && now.After(o.ExpireAt)
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
