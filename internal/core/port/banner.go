package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type BannerRepo interface {
	Create(ctx context.Context, banner *domain.Banner) error
	Update(ctx context.Context, banner *domain.Banner) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]domain.Banner, error)
	ListEnabled(ctx context.Context) ([]domain.Banner, error)
}
