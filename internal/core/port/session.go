package port

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type SessionRepo interface {
	GetSessionByID(ctx context.Context, id int64) (*domain.ShowSession, error)
	Create(ctx context.Context, session *domain.ShowSession) error
	UpdatePrice(ctx context.Context, id int64, basePriceCents int64, priceRulesJSON string) error
	Cancel(ctx context.Context, id int64) error
	ListOverlapping(ctx context.Context, hallID int64, start, end time.Time) ([]domain.ShowSession, error)
}
