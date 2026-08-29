package port

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

// LoginGuardRepo 登录失败与锁定记录。Get 无记录时返回 (nil, nil)。
type LoginGuardRepo interface {
	Get(ctx context.Context, scope, username string) (*domain.LoginGuard, error)
	// RecordFailure 原子累加失败次数并返回最新次数。
	RecordFailure(ctx context.Context, scope, username string) (int, error)
	// Lock 写入锁定截止时间（记录不存在时创建）。
	Lock(ctx context.Context, scope, username string, until time.Time) error
	// Reset 登录成功后清零失败次数与锁定。
	Reset(ctx context.Context, scope, username string) error
}
