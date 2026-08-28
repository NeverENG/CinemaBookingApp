package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type RoleRepo interface {
	GetByID(ctx context.Context, id int64) (*domain.Role, error)
	GetByCode(ctx context.Context, code string) (*domain.Role, error)
	Ensure(ctx context.Context, roles []domain.Role) error
}
