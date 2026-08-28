package domain

import "time"

// OrderItem 订单明细（一张票对应一行）。
type OrderItem struct {
	ID         int64
	OrderNo    string
	SessionID  int64
	SeatID     int64
	SeatNo     string
	PriceCents int64
	TicketNo   string
	UsedAt     *time.Time
}
