package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

// OrderRepo 订单仓储接口。方法只接收 ctx，不接收 tx：
// 事务由 service 通过 TxManager 启动，infra 实现从 ctx 取出事务。
type OrderRepo interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	GetOrderByNo(ctx context.Context, orderNo string) (*domain.Order, error)
	Transition(ctx context.Context, orderNo string, from, to domain.OrderStatus, version int32) error
	IssueTickets(ctx context.Context, orderNo string, tickets []domain.OrderItem) error
	ExpirePendingBySessionID(ctx context.Context, sessionID int64) ([]string, error)
	ListOrdersByUserID(ctx context.Context, userID int64) ([]domain.Order, error)
}
