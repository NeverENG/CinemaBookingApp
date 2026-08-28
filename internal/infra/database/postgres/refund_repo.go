package postgres

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

type refundRow struct {
	ID               int64               `gorm:"column:id;primaryKey"`
	RefundNo         string              `gorm:"column:refund_no"`
	OrderNo          string              `gorm:"column:order_no"`
	UserID           int64               `gorm:"column:user_id"`
	AmountCents      int64               `gorm:"column:amount_cents"`
	Reason           string              `gorm:"column:reason"`
	Status           domain.RefundStatus `gorm:"column:status"`
	ExternalRefundNo string              `gorm:"column:external_refund_no"`
	RefundedAt       *time.Time          `gorm:"column:refunded_at"`
	CreatedAt        time.Time           `gorm:"column:created_at"`
}

func (refundRow) TableName() string { return "refunds" }

// RefundRepo 实现 port.RefundRepo。
type RefundRepo struct {
	db *DB
}

var _ port.RefundRepo = (*RefundRepo)(nil)

func NewRefundRepo(db *DB) *RefundRepo {
	return &RefundRepo{db: db}
}

func (r *RefundRepo) Create(ctx context.Context, refund *domain.Refund) error {
	now := time.Now()
	row := refundRow{
		RefundNo:         refund.RefundNo,
		OrderNo:          refund.OrderNo,
		UserID:           refund.UserID,
		AmountCents:      refund.AmountCents,
		Reason:           refund.Reason,
		Status:           refund.Status,
		ExternalRefundNo: refund.ExternalRefundNo,
		CreatedAt:        now,
	}
	if refund.Status == domain.RefundSuccess {
		row.RefundedAt = &now
	}
	return r.db.db(ctx).Create(&row).Error
}

func (r *RefundRepo) MarkSuccess(ctx context.Context, refundNo string) error {
	now := time.Now()
	return r.db.db(ctx).
		Model(&refundRow{}).
		Where("refund_no = ? AND status = ?", refundNo, domain.RefundPending).
		Updates(map[string]any{
			"status":      domain.RefundSuccess,
			"refunded_at": now,
		}).Error
}
