package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type paymentCallbackRow struct {
	ID            int64                 `gorm:"column:id;primaryKey"`
	EventID       string                `gorm:"column:event_id"`
	TransactionNo string                `gorm:"column:transaction_no"`
	AmountCents   int64                 `gorm:"column:amount_cents"`
	Payload       string                `gorm:"column:payload"`
	Status        domain.CallbackStatus `gorm:"column:status"`
	RetryCount    int                   `gorm:"column:retry_count"`
	ProcessResult string                `gorm:"column:process_result"`
	CreatedAt     time.Time             `gorm:"column:created_at"`
	ProcessedAt   *time.Time            `gorm:"column:processed_at"`
}

func (paymentCallbackRow) TableName() string { return "payment_callbacks" }

// PaymentCallbackRepo 实现 port.PaymentCallbackRepo。
type PaymentCallbackRepo struct {
	db *DB
}

var _ port.PaymentCallbackRepo = (*PaymentCallbackRepo)(nil)

func NewPaymentCallbackRepo(db *DB) *PaymentCallbackRepo {
	return &PaymentCallbackRepo{db: db}
}

// InsertIfAbsent 靠 event_id 唯一约束去重：重复返回 (false, nil)。
func (r *PaymentCallbackRepo) InsertIfAbsent(ctx context.Context, cb *domain.PaymentCallback) (bool, error) {
	row := paymentCallbackRow{
		EventID:       cb.EventID,
		TransactionNo: cb.TransactionNo,
		AmountCents:   cb.AmountCents,
		Payload:       cb.Payload,
		Status:        domain.CallbackReceived,
		CreatedAt:     time.Now(),
	}
	if err := r.db.db(ctx).Create(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *PaymentCallbackRepo) MarkProcessed(ctx context.Context, eventID string) error {
	now := time.Now()
	return r.db.db(ctx).
		Model(&paymentCallbackRow{}).
		Where("event_id = ?", eventID).
		Updates(map[string]any{
			"status":       domain.CallbackProcessed,
			"processed_at": now,
		}).Error
}

func (r *PaymentCallbackRepo) MarkFailed(ctx context.Context, eventID, reason string) error {
	return r.db.db(ctx).
		Model(&paymentCallbackRow{}).
		Where("event_id = ?", eventID).
		Updates(map[string]any{
			"status":         domain.CallbackFailed,
			"process_result": reason,
		}).Error
}

func (r *PaymentCallbackRepo) GetByEventID(ctx context.Context, eventID string) (*domain.PaymentCallback, error) {
	var row paymentCallbackRow
	err := r.db.db(ctx).Where("event_id = ?", eventID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &domain.PaymentCallback{
		EventID:       row.EventID,
		TransactionNo: row.TransactionNo,
		AmountCents:   row.AmountCents,
		Payload:       row.Payload,
		Status:        row.Status,
		RetryCount:    row.RetryCount,
		CreatedAt:     row.CreatedAt,
		ProcessedAt:   row.ProcessedAt,
	}, nil
}

const maxCallbackRetries = 5

func (r *PaymentCallbackRepo) ListPending(ctx context.Context, limit int) ([]domain.PaymentCallback, error) {
	var rows []paymentCallbackRow
	if err := r.db.db(ctx).
		Where("status IN ? AND retry_count < ?", []domain.CallbackStatus{domain.CallbackReceived, domain.CallbackFailed}, maxCallbackRetries).
		Order("id").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	callbacks := make([]domain.PaymentCallback, 0, len(rows))
	for _, row := range rows {
		callbacks = append(callbacks, domain.PaymentCallback{
			EventID:       row.EventID,
			TransactionNo: row.TransactionNo,
			AmountCents:   row.AmountCents,
			Payload:       row.Payload,
			Status:        row.Status,
			RetryCount:    row.RetryCount,
			CreatedAt:     row.CreatedAt,
			ProcessedAt:   row.ProcessedAt,
		})
	}
	return callbacks, nil
}

func (r *PaymentCallbackRepo) IncrementRetry(ctx context.Context, eventID string) error {
	return r.db.db(ctx).
		Model(&paymentCallbackRow{}).
		Where("event_id = ?", eventID).
		Update("retry_count", gorm.Expr("retry_count + 1")).Error
}
