package port

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

// OrderRepo 订单仓储接口。方法只接收 ctx，不接收 tx：
// 事务由 service 通过 TxManager 启动，infra 实现从 ctx 取出事务。
type OrderRepo interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	GetOrderByNo(ctx context.Context, orderNo string) (*domain.Order, error)
	GetOrderForUpdate(ctx context.Context, orderNo string) (*domain.Order, error)
	Transition(ctx context.Context, orderNo string, from, to domain.OrderStatus, version int32) error
	ListExpiredPending(ctx context.Context, now time.Time) ([]domain.Order, error)
	ExpirePendingBySessionID(ctx context.Context, sessionID int64) ([]string, error)
	ListPaidBySessionID(ctx context.Context, sessionID int64) ([]domain.Order, error)
	CountPaidByMovieIDs(ctx context.Context, movieIDs []int64) (map[int64]int64, error)
	ListOrdersByUserID(ctx context.Context, userID int64) ([]domain.Order, error)
}
