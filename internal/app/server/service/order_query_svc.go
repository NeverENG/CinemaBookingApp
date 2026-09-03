package service

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

type OrderQuerySvc struct {
	orders   port.OrderRepo
	sessions port.SessionRepo
	movies   port.MovieRepo
	halls    port.HallRepo
	cinemas  port.CinemaRepo
}

func NewOrderQuerySvc(
	orders port.OrderRepo,
	sessions port.SessionRepo,
	movies port.MovieRepo,
	halls port.HallRepo,
	cinemas port.CinemaRepo,
) *OrderQuerySvc {
	return &OrderQuerySvc{orders: orders, sessions: sessions, movies: movies, halls: halls, cinemas: cinemas}
}

type OrderView struct {
	OrderNo       string             `json:"order_no"`
	UserID        int64              `json:"user_id"`
	SessionID     int64              `json:"session_id"`
	CinemaID      int64              `json:"cinema_id"`
	MovieID       int64              `json:"movie_id"`
	MovieTitle    string             `json:"movie_title"`
	CinemaName    string             `json:"cinema_name"`
	HallName      string             `json:"hall_name"`
	StartTime     time.Time          `json:"start_time"`
	Status        domain.OrderStatus `json:"status"`
	TotalCents    int64              `json:"total_cents"`
	DiscountCents int64              `json:"discount_cents"`
	CouponCents   int64              `json:"coupon_cents"`
	PaidCents     int64              `json:"paid_cents"`
	ExpireAt      time.Time          `json:"expire_at"`
	CreatedAt     time.Time          `json:"created_at"`
	PaidAt        *time.Time         `json:"paid_at"`
	CanRefund     bool               `json:"can_refund"`
	CanChange     bool               `json:"can_change"`
	Items         []OrderItemView    `json:"items"`
}

type OrderItemView struct {
	ID         int64      `json:"id"`
	SeatID     int64      `json:"seat_id"`
	SeatNo     string     `json:"seat_no"`
	PriceCents int64      `json:"price_cents"`
	TicketNo   string     `json:"ticket_no,omitempty"`
	UsedAt     *time.Time `json:"used_at"`
}

func (s *OrderQuerySvc) Get(ctx context.Context, userID int64, orderNo string) (*OrderView, error) {
	order, err := s.orders.GetOrderByNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return s.toView(ctx, order)
}

func (s *OrderQuerySvc) List(ctx context.Context, userID int64) ([]OrderView, error) {
	orders, err := s.orders.ListOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	views := make([]OrderView, 0, len(orders))
	for i := range orders {
		view, err := s.toView(ctx, &orders[i])
		if err != nil {
			return nil, err
		}
		views = append(views, *view)
	}
	return views, nil
}

func (s *OrderQuerySvc) toView(ctx context.Context, order *domain.Order) (*OrderView, error) {
	session, err := s.sessions.GetSessionByID(ctx, order.SessionID)
	if err != nil {
		return nil, err
	}
	movie, err := s.movies.GetByID(ctx, order.MovieID)
	if err != nil {
		return nil, err
	}
	hall, err := s.halls.GetByID(ctx, session.HallID)
	if err != nil {
		return nil, err
	}
	cinema, err := s.cinemas.GetByID(ctx, order.CinemaID)
	if err != nil {
		return nil, err
	}
	items := make([]OrderItemView, 0, len(order.Items))
	hasUsed := false
	for _, item := range order.Items {
		if item.UsedAt != nil {
			hasUsed = true
		}
		items = append(items, OrderItemView{
			ID:         item.ID,
			SeatID:     item.SeatID,
			SeatNo:     item.SeatNo,
			PriceCents: item.PriceCents,
			TicketNo:   item.TicketNo,
			UsedAt:     item.UsedAt,
		})
	}
	canRefund := order.Status == domain.OrderPaid && !hasUsed && session.StartTime.After(time.Now())
	return &OrderView{
		OrderNo:       order.OrderNo,
		UserID:        order.UserID,
		SessionID:     order.SessionID,
		CinemaID:      order.CinemaID,
		MovieID:       order.MovieID,
		MovieTitle:    movie.Title,
		CinemaName:    cinema.Name,
		HallName:      hall.Name,
		StartTime:     session.StartTime,
		Status:        order.Status,
		TotalCents:    order.TotalCents,
		DiscountCents: order.DiscountCents,
		CouponCents:   order.CouponCents,
		PaidCents:     order.PaidCents,
		ExpireAt:      order.ExpireAt,
		CreatedAt:     order.CreatedAt,
		PaidAt:        order.PaidAt,
		CanRefund:     canRefund,
		CanChange:     canRefund,
		Items:         items,
	}, nil
}
