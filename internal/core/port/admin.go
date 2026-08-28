package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type AdminRepo interface {
	GetByUsername(ctx context.Context, username string) (*domain.Admin, error)
	Count(ctx context.Context) (int64, error)
	Create(ctx context.Context, admin *domain.Admin) error
}
