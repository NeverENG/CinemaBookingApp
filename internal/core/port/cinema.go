package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type CinemaRepo interface {
	GetByID(ctx context.Context, id int64) (*domain.Cinema, error)
	List(ctx context.Context, keyword, city string) ([]domain.Cinema, error)
	ListAll(ctx context.Context) ([]domain.Cinema, error)
	Create(ctx context.Context, cinema *domain.Cinema) error
	Update(ctx context.Context, cinema *domain.Cinema) error
	SetStatus(ctx context.Context, id int64, status domain.CinemaStatus) error
}
