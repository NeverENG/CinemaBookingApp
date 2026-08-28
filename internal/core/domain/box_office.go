package domain

import "time"

type BoxOfficeBizType string

const (
	BoxOrderPaid   BoxOfficeBizType = "ORDER_PAID"
	BoxOrderRefund BoxOfficeBizType = "ORDER_REFUND"
)

// BoxOfficeEvent 票房事件：ledger 的唯一事实，daily 只是冗余聚合。
type BoxOfficeEvent struct {
	BizType     BoxOfficeBizType
	BizNo       string
	StatDate    time.Time
	CinemaID    int64
	MovieID     int64
	OrderDelta  int
	TicketDelta int
	GrossDelta  int64
	RefundDelta int64
}

// NewPaidEvent 支付成功事件：订单数 +1、票数 +N、票房 +实付。
func NewPaidEvent(order *Order, bizNo string, now time.Time) *BoxOfficeEvent {
	return &BoxOfficeEvent{
		BizType:     BoxOrderPaid,
		BizNo:       bizNo,
		StatDate:    statDate(now),
		CinemaID:    order.CinemaID,
		MovieID:     order.MovieID,
		OrderDelta:  1,
		TicketDelta: len(order.Items),
		GrossDelta:  order.PaidCents,
	}
}

// NewRefundEvent 退款成功事件：订单数 -1、票数 -N、退款额 +amount。
func NewRefundEvent(order *Order, amountCents int64, bizNo string, now time.Time) *BoxOfficeEvent {
	return &BoxOfficeEvent{
		BizType:     BoxOrderRefund,
		BizNo:       bizNo,
		StatDate:    statDate(now),
		CinemaID:    order.CinemaID,
		MovieID:     order.MovieID,
		OrderDelta:  -1,
		TicketDelta: -len(order.Items),
		RefundDelta: amountCents,
	}
}

// statDate UTC+8 自然日。
func statDate(t time.Time) time.Time {
	cst := time.FixedZone("CST", 8*3600)
	y, m, d := t.In(cst).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, cst)
}

// 看板只读模型。
type BoxOfficeTrendRow struct {
	Date        time.Time
	OrderCount  int
	TicketCount int
	GrossCents  int64
	RefundCents int64
	NetCents    int64
}

type BoxOfficeMovieRow struct {
	MovieID    int64
	MovieTitle string
	OrderCount int
	GrossCents int64
	NetCents   int64
}

type BoxOfficeCinemaRow struct {
	CinemaID   int64
	CinemaName string
	OrderCount int
	GrossCents int64
	NetCents   int64
}
