package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type paymentRow struct {
	TransactionNo   string               `gorm:"column:transaction_no;primaryKey"`
	BizType         string               `gorm:"column:biz_type"`
	BizNo           string               `gorm:"column:biz_no"`
	UserID          int64                `gorm:"column:user_id"`
	AmountCents     int64                `gorm:"column:amount_cents"`
	Channel         string               `gorm:"column:channel"`
	Status          domain.PaymentStatus `gorm:"column:status"`
	ExternalTradeNo string               `gorm:"column:external_trade_no"`
	Version         int32                `gorm:"column:version"`
	CreatedAt       time.Time            `gorm:"column:created_at"`
	PaidAt          *time.Time           `gorm:"column:paid_at"`
	ClosedAt        *time.Time           `gorm:"column:closed_at"`
}

func (paymentRow) TableName() string { return "payment_transactions" }

// PaymentRepo 实现 port.PaymentRepo。
type PaymentRepo struct {
	db *DB
}

var _ port.PaymentRepo = (*PaymentRepo)(nil)

func NewPaymentRepo(db *DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) CreateTransaction(ctx context.Context, tx *domain.PaymentTransaction) error {
	return r.db.db(ctx).Create(toPaymentRow(tx)).Error
}

func (r *PaymentRepo) GetByTransactionNo(ctx context.Context, transactionNo string) (*domain.PaymentTransaction, error) {
	var row paymentRow
	err := r.db.db(ctx).Where("transaction_no = ?", transactionNo).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomainPayment(row), nil
}

func (r *PaymentRepo) GetByOrderNo(ctx context.Context, orderNo string) (*domain.PaymentTransaction, error) {
	var row paymentRow
	err := r.db.db(ctx).Where("biz_type = ? AND biz_no = ?", "ORDER_PAY", orderNo).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomainPayment(row), nil
}

func (r *PaymentRepo) ListPendingOlderThan(ctx context.Context, before time.Time, limit int) ([]domain.PaymentTransaction, error) {
	var rows []paymentRow
	if err := r.db.db(ctx).
		Where("status = ? AND created_at < ?", domain.PaymentPending, before).
		Order("created_at, transaction_no").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	txs := make([]domain.PaymentTransaction, 0, len(rows))
	for _, row := range rows {
		txs = append(txs, *toDomainPayment(row))
	}
	return txs, nil
}

func (r *PaymentRepo) Transition(ctx context.Context, transactionNo string, from, to domain.PaymentStatus, version int32) error {
	res := r.db.db(ctx).
		Model(&paymentRow{}).
		Where("transaction_no = ? AND status = ? AND version = ?", transactionNo, from, version).
		Updates(map[string]any{
			"status":  to,
			"version": gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return domain.ErrInvalidTransition
	}
	return nil
}

func (r *PaymentRepo) ListOrderPaymentMismatches(ctx context.Context) ([]string, error) {
	var items []string
	rows, err := r.db.db(ctx).Raw(`
		SELECT 'PAY_WITHOUT_ORDER:' || p.biz_no FROM payment_transactions p
		LEFT JOIN orders o ON o.order_no = p.biz_no
		WHERE p.status = 'SUCCESS'
		  AND (o.order_no IS NULL OR o.status NOT IN ('PAID','REFUNDING','REFUNDED','COMPLETED'))
		UNION ALL
		SELECT 'ORDER_WITHOUT_PAY:' || o.order_no FROM orders o
		WHERE o.status IN ('PAID','REFUNDING','REFUNDED')
		  AND NOT EXISTS (
		      SELECT 1 FROM payment_transactions p
		      WHERE p.biz_no = o.order_no AND p.status IN ('SUCCESS','REFUNDED')
		  )`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func toPaymentRow(tx *domain.PaymentTransaction) *paymentRow {
	return &paymentRow{
		TransactionNo:   tx.TransactionNo,
		BizType:         "ORDER_PAY",
		BizNo:           tx.OrderNo,
		UserID:          tx.UserID,
		AmountCents:     tx.AmountCents,
		Channel:         tx.Channel,
		Status:          tx.Status,
		ExternalTradeNo: tx.ExternalTradeNo,
		Version:         tx.Version,
		CreatedAt:       tx.CreatedAt,
		PaidAt:          tx.PaidAt,
		ClosedAt:        tx.ClosedAt,
	}
}

func toDomainPayment(row paymentRow) *domain.PaymentTransaction {
	return &domain.PaymentTransaction{
		TransactionNo:   row.TransactionNo,
		OrderNo:         row.BizNo,
		UserID:          row.UserID,
		AmountCents:     row.AmountCents,
		Channel:         row.Channel,
		Status:          row.Status,
		ExternalTradeNo: row.ExternalTradeNo,
		Version:         row.Version,
		CreatedAt:       row.CreatedAt,
		PaidAt:          row.PaidAt,
		ClosedAt:        row.ClosedAt,
	}
}
