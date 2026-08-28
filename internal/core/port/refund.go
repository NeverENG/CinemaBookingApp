package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type RefundRepo interface {
	Create(ctx context.Context, refund *domain.Refund) error
	MarkSuccess(ctx context.Context, refundNo string) error
}
