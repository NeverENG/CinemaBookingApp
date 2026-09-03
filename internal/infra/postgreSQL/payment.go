package postgresql

import (
	"context"
	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"gorm.io/gorm"
)

type PaymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepo(db *gorm.DB) PaymentRepo {
	return PaymentRepo{db: db}
}

func (r *PaymentRepo) CreatePayment(ctx context.Context, payment domain.PaymentTransaction) error {
	err := r.db.WithContext(ctx).Exec("INSERT INTO payments (transaction_no, order_no, amount_cents, status) VALUES ($1, $2, $3, $4)", payment.TransactionNo, payment.OrderNo, payment.AmountCents, payment.Status)
	return err.Error
}

func (r *PaymentRepo) GetPayment(ctx context.Context, id string) (domain.PaymentTransaction, error) {
	var payment domain.PaymentTransaction
	err := r.db.WithContext(ctx).Exec("SELECT * FROM payments WHERE id = $1", id).Scan(&payment)
	return payment, err.Error
}
