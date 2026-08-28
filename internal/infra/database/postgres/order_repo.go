package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

// orderRow 是 orders 表的 GORM 行模型（只存在于 infra，domain 不打 tag）。
type orderRow struct {
	OrderNo          string             `gorm:"column:order_no;primaryKey"`
	UserID           int64              `gorm:"column:user_id"`
	SessionID        int64              `gorm:"column:session_id"`
	CinemaID         int64              `gorm:"column:cinema_id"`
	MovieID          int64              `gorm:"column:movie_id"`
	Status           domain.OrderStatus `gorm:"column:status"`
	TotalCents       int64              `gorm:"column:total_cents"`
	DiscountCents    int64              `gorm:"column:discount_cents"`
	CouponCents      int64              `gorm:"column:coupon_cents"`
	PaidCents        int64              `gorm:"column:paid_cents"`
	CouponInstanceID *int64             `gorm:"column:coupon_instance_id"`
	ExpireAt         time.Time          `gorm:"column:expire_at"`
	Version          int32              `gorm:"column:version"`
	CreatedAt        time.Time          `gorm:"column:created_at"`
	PaidAt           *time.Time         `gorm:"column:paid_at"`
}

func (orderRow) TableName() string { return "orders" }

type orderItemRow struct {
	ID         int64      `gorm:"column:id;primaryKey"`
	OrderNo    string     `gorm:"column:order_no"`
	SessionID  int64      `gorm:"column:session_id"`
	SeatID     int64      `gorm:"column:seat_id"`
	SeatNo     string     `gorm:"column:seat_no"`
	PriceCents int64      `gorm:"column:price_cents"`
	TicketNo   string     `gorm:"column:ticket_no"`
	UsedAt     *time.Time `gorm:"column:used_at"`
}

func (orderItemRow) TableName() string { return "order_items" }

// OrderRepo 实现 port.OrderRepo。
type OrderRepo struct {
	db *DB
}

var _ port.OrderRepo = (*OrderRepo)(nil)

func NewOrderRepo(db *DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) CreateOrder(ctx context.Context, order *domain.Order) error {
	db := r.db.db(ctx)
	if err := db.Create(toOrderRow(order)).Error; err != nil {
		return err
	}
	if len(order.Items) == 0 {
		return nil
	}
	rows := make([]orderItemRow, 0, len(order.Items))
	for _, item := range order.Items {
		rows = append(rows, toOrderItemRow(item))
	}
	return db.Create(&rows).Error
}

func (r *OrderRepo) GetOrderByNo(ctx context.Context, orderNo string) (*domain.Order, error) {
	var row orderRow
	err := r.db.db(ctx).Where("order_no = ?", orderNo).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	var itemRows []orderItemRow
	if err := r.db.db(ctx).Where("order_no = ?", orderNo).Order("id").Find(&itemRows).Error; err != nil {
		return nil, err
	}
	return toDomainOrder(row, itemRows), nil
}

func (r *OrderRepo) Transition(ctx context.Context, orderNo string, from, to domain.OrderStatus, version int32) error {
	res := r.db.db(ctx).
		Model(&orderRow{}).
		Where("order_no = ? AND status = ? AND version = ?", orderNo, from, version).
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

func (r *OrderRepo) ListExpiredPending(ctx context.Context, now time.Time) ([]domain.Order, error) {
	var rows []orderRow
	if err := r.db.db(ctx).
		Where("status = ? AND expire_at < ?", domain.OrderPendingPayment, now).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	orders := make([]domain.Order, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, *toDomainOrder(row, nil))
	}
	return orders, nil
}

// ExpirePendingBySessionID 过期场次下全部待支付订单，返回受影响订单号。
func (r *OrderRepo) ExpirePendingBySessionID(ctx context.Context, sessionID int64) ([]string, error) {
	var orderNos []string
	err := r.db.db(ctx).
		Model(&orderRow{}).
		Where("session_id = ? AND status = ?", sessionID, domain.OrderPendingPayment).
		Pluck("order_no", &orderNos).Error
	if err != nil {
		return nil, err
	}
	if len(orderNos) == 0 {
		return orderNos, nil
	}
	now := time.Now()
	err = r.db.db(ctx).
		Model(&orderRow{}).
		Where("session_id = ? AND status = ?", sessionID, domain.OrderPendingPayment).
		Updates(map[string]any{
			"status":     domain.OrderExpired,
			"expired_at": now,
		}).Error
	return orderNos, err
}

func (r *OrderRepo) ListPaidBySessionID(ctx context.Context, sessionID int64) ([]domain.Order, error) {
	var rows []orderRow
	if err := r.db.db(ctx).
		Where("session_id = ? AND status = ?", sessionID, domain.OrderPaid).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	orders := make([]domain.Order, 0, len(rows))
	for _, row := range rows {
		var itemRows []orderItemRow
		if err := r.db.db(ctx).Where("order_no = ?", row.OrderNo).Order("id").Find(&itemRows).Error; err != nil {
			return nil, err
		}
		orders = append(orders, *toDomainOrder(row, itemRows))
	}
	return orders, nil
}

func (r *OrderRepo) CountPaidByMovieIDs(ctx context.Context, movieIDs []int64) (map[int64]int64, error) {
	if len(movieIDs) == 0 {
		return map[int64]int64{}, nil
	}
	type movieSold struct {
		MovieID int64
		Cnt     int64
	}
	var rows []movieSold
	err := r.db.db(ctx).
		Model(&orderRow{}).
		Select("movie_id, count(*) AS cnt").
		Where("status = ? AND movie_id IN ?", domain.OrderPaid, movieIDs).
		Group("movie_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	counts := make(map[int64]int64, len(rows))
	for _, row := range rows {
		counts[row.MovieID] = row.Cnt
	}
	return counts, nil
}

func (r *OrderRepo) ListOrdersByUserID(ctx context.Context, userID int64) ([]domain.Order, error) {
	var rows []orderRow
	if err := r.db.db(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []domain.Order{}, nil
	}

	orderNos := make([]string, 0, len(rows))
	for _, row := range rows {
		orderNos = append(orderNos, row.OrderNo)
	}
	var itemRows []orderItemRow
	if err := r.db.db(ctx).Where("order_no IN ?", orderNos).Order("id").Find(&itemRows).Error; err != nil {
		return nil, err
	}
	itemsByOrder := make(map[string][]orderItemRow)
	for _, item := range itemRows {
		itemsByOrder[item.OrderNo] = append(itemsByOrder[item.OrderNo], item)
	}

	orders := make([]domain.Order, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, *toDomainOrder(row, itemsByOrder[row.OrderNo]))
	}
	return orders, nil
}

func toOrderRow(o *domain.Order) *orderRow {
	return &orderRow{
		OrderNo:          o.OrderNo,
		UserID:           o.UserID,
		SessionID:        o.SessionID,
		CinemaID:         o.CinemaID,
		MovieID:          o.MovieID,
		Status:           o.Status,
		TotalCents:       o.TotalCents,
		DiscountCents:    o.DiscountCents,
		CouponCents:      o.CouponCents,
		PaidCents:        o.PaidCents,
		CouponInstanceID: o.CouponInstanceID,
		ExpireAt:         o.ExpireAt,
		Version:          o.Version,
		CreatedAt:        o.CreatedAt,
		PaidAt:           o.PaidAt,
	}
}

func toOrderItemRow(item domain.OrderItem) orderItemRow {
	return orderItemRow{
		OrderNo:    item.OrderNo,
		SessionID:  item.SessionID,
		SeatID:     item.SeatID,
		SeatNo:     item.SeatNo,
		PriceCents: item.PriceCents,
		TicketNo:   item.TicketNo,
		UsedAt:     item.UsedAt,
	}
}

func toDomainOrder(row orderRow, itemRows []orderItemRow) *domain.Order {
	items := make([]domain.OrderItem, 0, len(itemRows))
	for _, item := range itemRows {
		items = append(items, domain.OrderItem{
			ID:         item.ID,
			OrderNo:    item.OrderNo,
			SessionID:  item.SessionID,
			SeatID:     item.SeatID,
			SeatNo:     item.SeatNo,
			PriceCents: item.PriceCents,
			TicketNo:   item.TicketNo,
			UsedAt:     item.UsedAt,
		})
	}
	return &domain.Order{
		OrderNo:          row.OrderNo,
		UserID:           row.UserID,
		SessionID:        row.SessionID,
		CinemaID:         row.CinemaID,
		MovieID:          row.MovieID,
		Status:           row.Status,
		TotalCents:       row.TotalCents,
		DiscountCents:    row.DiscountCents,
		CouponCents:      row.CouponCents,
		PaidCents:        row.PaidCents,
		CouponInstanceID: row.CouponInstanceID,
		ExpireAt:         row.ExpireAt,
		Version:          row.Version,
		Items:            items,
		CreatedAt:        row.CreatedAt,
		PaidAt:           row.PaidAt,
	}
}
