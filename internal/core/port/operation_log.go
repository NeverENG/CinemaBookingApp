package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type OperationLogRepo interface {
	Create(ctx context.Context, log *domain.OperationLog) error
}
