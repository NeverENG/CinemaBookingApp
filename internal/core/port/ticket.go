package port

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

// TicketRepo 提供票券核销需要的原子读写能力。
// 票券事实仍然保存在 order_items，避免为了核销再引入一张重复表。
type TicketRepo interface {
	GetOrderByTicketNo(ctx context.Context, ticketNo string) (*domain.Order, error)
	MarkTicketUsed(ctx context.Context, ticketNo string, usedAt time.Time) (bool, error)
	CountUnusedTickets(ctx context.Context, orderNo string) (int64, error)
}
