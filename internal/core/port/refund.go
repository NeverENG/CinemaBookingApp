package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type RefundRepo interface {
	Create(ctx context.Context, refund *domain.Refund) error
	GetByRefundNo(ctx context.Context, refundNo string) (*domain.Refund, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*domain.Refund, error)
	MarkSuccess(ctx context.Context, refundNo string) error
}
