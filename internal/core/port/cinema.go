package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type CinemaRepo interface {
	GetByID(ctx context.Context, id int64) (*domain.Cinema, error)
	List(ctx context.Context, keyword, city string) ([]domain.Cinema, error)
}
