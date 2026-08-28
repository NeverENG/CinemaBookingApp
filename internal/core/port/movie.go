package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type MovieRepo interface {
	Create(ctx context.Context, movie *domain.Movie) error
	GetByID(ctx context.Context, id int64) (*domain.Movie, error)
	Update(ctx context.Context, movie *domain.Movie) error
	List(ctx context.Context) ([]domain.Movie, error)
	SetStatus(ctx context.Context, id int64, status domain.MovieStatus) error
}
