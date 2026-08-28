package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type SeatRepo interface {
	ListSeatsByIDs(ctx context.Context, ids []int64) ([]domain.Seat, error)
	ListByHallID(ctx context.Context, hallID int64) ([]domain.Seat, error)
	// SyncSeats 按布局 diff 同步座位：新增、更新、未出现的置 DISABLED（保留 ID）。
	SyncSeats(ctx context.Context, hallID int64, seats []domain.Seat) error
}
