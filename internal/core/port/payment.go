package port

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

// PaymentRepo 支付交易仓储接口。方法只接收 ctx，不接收 tx。
type PaymentRepo interface {
	CreateTransaction(ctx context.Context, tx *domain.PaymentTransaction) error
	GetByTransactionNo(ctx context.Context, transactionNo string) (*domain.PaymentTransaction, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*domain.PaymentTransaction, error)
	ListPendingOlderThan(ctx context.Context, before time.Time, limit int) ([]domain.PaymentTransaction, error)
	Transition(ctx context.Context, transactionNo string, from, to domain.PaymentStatus, version int32) error
}

// PaymentCallbackRepo 回调记录仓储：event_id 唯一是幂等键。
type PaymentCallbackRepo interface {
	// InsertIfAbsent 已存在返回 (false, nil)，首次插入返回 (true, nil)。
	InsertIfAbsent(ctx context.Context, cb *domain.PaymentCallback) (bool, error)
	GetByEventID(ctx context.Context, eventID string) (*domain.PaymentCallback, error)
	// ListPending 待重试回调（RECEIVED/FAILED 且重试次数未达上限）。
	ListPending(ctx context.Context, limit int) ([]domain.PaymentCallback, error)
	IncrementRetry(ctx context.Context, eventID string) error
	MarkProcessed(ctx context.Context, eventID string) error
	MarkFailed(ctx context.Context, eventID, reason string) error
}
