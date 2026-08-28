package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type HallRepo interface {
	Create(ctx context.Context, hall *domain.Hall) error
	GetByID(ctx context.Context, id int64) (*domain.Hall, error)
	Update(ctx context.Context, hall *domain.Hall) error
	ListByCinema(ctx context.Context, cinemaID int64) ([]domain.Hall, error)
}
