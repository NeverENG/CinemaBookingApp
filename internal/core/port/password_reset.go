package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type PasswordResetRepo interface {
	Create(ctx context.Context, code *domain.PasswordResetCode) error
	// FindUnusedByEmail 取该邮箱最新未使用且未过期的验证码。
	FindUnusedByEmail(ctx context.Context, email string) (*domain.PasswordResetCode, error)
	MarkUsed(ctx context.Context, id int64) error
}
