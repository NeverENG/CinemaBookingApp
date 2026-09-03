package postgresql

import (
	"context"
	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"gorm.io/gorm"
)

type OrderRepo struct {
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) GetOrderByID(ctx context.Context, id string) (domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).Exec("SELECT * FROM orders WHERE id = ?", id).Scan(&order)
	return order, err.Error
}

func (r *OrderRepo) TransOrderTo(ctx context.Context, from, to domain.OrderStatus) error {
	err := r.db.WithContext(ctx).Exec("UPDATE orders SET status = ? WHERE status = ?", to, from)
	return err.Error
}

func (r *OrderRepo) CreateOrder(ctx context.Context, order *domain.Order) error {
	err := r.db.WithContext(ctx).Create(order)
	return err.Error
}

func (r *OrderRepo) ListOrdersByUserID(ctx context.Context, userID int64) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.WithContext(ctx).Exec("SELECT * FROM orders WHERE user_id = ?", userID).Scan(&orders)
	return orders, err.Error
}
