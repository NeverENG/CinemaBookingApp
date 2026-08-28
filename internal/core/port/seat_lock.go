package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

// SeatLockRepo 座位锁仓储接口。
// CreateLocks 依赖数据库部分唯一索引 (session_id, seat_id)
// WHERE status IN ('LOCKED','BOOKED') 防超卖。
type SeatLockRepo interface {
	CreateLocks(ctx context.Context, locks []domain.SeatLock) error
	MarkBookedByOrderNo(ctx context.Context, orderNo string) error
	ReleaseByOrderNo(ctx context.Context, orderNo string, status domain.SeatLockStatus) error
	ReleaseBySessionID(ctx context.Context, sessionID int64, status domain.SeatLockStatus) error
	ListActiveBySessionID(ctx context.Context, sessionID int64) ([]domain.SeatLock, error)
}
