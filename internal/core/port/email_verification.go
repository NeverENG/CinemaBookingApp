package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type EmailVerificationRepo interface {
	Create(ctx context.Context, code *domain.EmailVerificationCode) error
	FindUnusedByEmail(ctx context.Context, email string) (*domain.EmailVerificationCode, error)
	MarkUsed(ctx context.Context, id int64) error
}
