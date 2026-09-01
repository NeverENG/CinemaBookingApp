package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type AdminRepo interface {
	GetByID(ctx context.Context, id int64) (*domain.Admin, error)
	GetByUsername(ctx context.Context, username string) (*domain.Admin, error)
	Count(ctx context.Context) (int64, error)
	Create(ctx context.Context, admin *domain.Admin) error
	UpdatePassword(ctx context.Context, adminID int64, passwordHash string) error
}
