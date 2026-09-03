package service

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/biz"
	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/uid"
)

const (
	orderTTL          = 15 * time.Minute
	seatStatusEnabled = "ENABLED"
)

// OrderSvc 下单用例编排：锁座 → 建单 → 锁券，一个用例一个事务。
type OrderSvc struct {
	tx       port.TxManager
	users    port.UserRepo
	sessions port.SessionRepo
	seats    port.SeatRepo
	locks    port.SeatLockRepo
	coupons  port.UserCouponRepo
	orders   port.OrderRepo
}

func NewOrderSvc(
	tx port.TxManager,
	users port.UserRepo,
	sessions port.SessionRepo,
	seats port.SeatRepo,
	locks port.SeatLockRepo,
	coupons port.UserCouponRepo,
	orders port.OrderRepo,
) *OrderSvc {
	return &OrderSvc{
		tx:       tx,
		users:    users,
		sessions: sessions,
		seats:    seats,
		locks:    locks,
		coupons:  coupons,
		orders:   orders,
	}
}

type CreateOrderInput struct {
	UserID    int64
	SessionID int64
	SeatIDs   []int64
	CouponNo  string // 可空：不用券
}

func (s *OrderSvc) CreateOrder(ctx context.Context, in CreateOrderInput) (*domain.Order, error) {
	if len(in.SeatIDs) == 0 {
		return nil, domain.ErrSeatNotAvailable
	}

	var order *domain.Order
	err := s.tx.Run(ctx, func(txCtx context.Context) error {
		if _, err := s.users.GetUserByID(txCtx, in.UserID); err != nil {
			return err
		}

		session, err := s.sessions.GetSessionForUpdate(txCtx, in.SessionID)
		if err != nil {
			return err
		}
		if !session.CanBook(time.Now()) {
			return domain.ErrSessionNotBookable
		}
		priceRules, err := parsePriceRules(session.PriceRulesJSON)
		if err != nil {
			return err
		}

		seats, err := s.seats.ListSeatsByIDs(txCtx, in.SeatIDs)
		if err != nil {
			return err
		}
		if len(seats) != len(in.SeatIDs) {
			return domain.ErrSeatNotAvailable
		}
		if err := s.locks.ReleaseExpiredBySeats(txCtx, session.ID, in.SeatIDs); err != nil {
			return err
		}
		prices := make([]int64, 0, len(seats))
		for _, seat := range seats {
			if seat.HallID != session.HallID || seat.Status != seatStatusEnabled {
				return domain.ErrSeatNotAvailable
			}
			price, err := priceForSeat(session, priceRules, seat)
			if err != nil {
				return err
			}
			prices = append(prices, price)
		}

		couponDiscount := int64(0)
		var coupon *domain.UserCoupon
		if in.CouponNo != "" {
			coupon, err = s.coupons.GetByCouponNo(txCtx, in.CouponNo)
			if err != nil {
				return err
			}
			if coupon.UserID != in.UserID || coupon.Status != domain.CouponUnused || time.Now().After(coupon.ExpireAt) {
				return domain.ErrCouponNotAvailable
			}
			template, err := s.coupons.GetTemplateByID(txCtx, coupon.TemplateID)
			if err != nil {
				return err
			}
			total, _, _, _, err := biz.CalcOrderAmount(prices, 0)
			if err != nil {
				return err
			}
			couponDiscount, err = template.DiscountCents(total)
			if err != nil {
				return err
			}
		}

		total, discount, couponAmt, _, err := biz.CalcOrderAmount(prices, couponDiscount)
		if err != nil {
			return err
		}

		now := time.Now()
		order = &domain.Order{
			OrderNo:   uid.OrderNo(),
			UserID:    in.UserID,
			SessionID: in.SessionID,
			CinemaID:  session.CinemaID,
			MovieID:   session.MovieID,
			Status:    domain.OrderPendingPayment,
			ExpireAt:  now.Add(orderTTL),
			Version:   1,
			CreatedAt: now,
		}
		if coupon != nil {
			couponID := coupon.ID
			order.CouponInstanceID = &couponID
		}
		if err := order.Settle(total, discount, couponAmt); err != nil {
			return err
		}

		items := make([]domain.OrderItem, 0, len(seats))
		for i, seat := range seats {
			items = append(items, domain.OrderItem{
				OrderNo:    order.OrderNo,
				SessionID:  session.ID,
				SeatID:     seat.ID,
				SeatNo:     seat.SeatNo,
				PriceCents: prices[i],
				TicketNo:   uid.TicketNo(),
			})
		}
		order.Items = items

		if err := s.orders.CreateOrder(txCtx, order); err != nil {
			return err
		}

		locks := make([]domain.SeatLock, 0, len(seats))
		for _, seat := range seats {
			locks = append(locks, domain.SeatLock{
				SessionID: session.ID,
				SeatID:    seat.ID,
				UserID:    in.UserID,
				OrderNo:   order.OrderNo,
				LockToken: uid.LockToken(),
				Status:    domain.SeatLockLocked,
				ExpiresAt: order.ExpireAt,
			})
		}
		if err := s.locks.CreateLocks(txCtx, locks); err != nil {
			return err
		}
		if err := s.sessions.RecalcStatus(txCtx, session.ID); err != nil {
			return err
		}

		if coupon != nil {
			if err := s.coupons.LockForOrder(txCtx, coupon.CouponNo, order.OrderNo); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

// GetOrder 查询订单详情（轮询用），校验归属。
func (s *OrderSvc) GetOrder(ctx context.Context, userID int64, orderNo string) (*domain.Order, error) {
	order, err := s.orders.GetOrderByNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return order, nil
}

func (s *OrderSvc) ListOrders(ctx context.Context, userID int64) ([]domain.Order, error) {
	return s.orders.ListOrdersByUserID(ctx, userID)
}

// ExpireOverdueOrders 定时任务：过期待支付订单 → EXPIRED，释放锁、解锁券。
func (s *OrderSvc) ExpireOverdueOrders(ctx context.Context, now time.Time) (int, error) {
	expired, err := s.orders.ListExpiredPending(ctx, now)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, candidate := range expired {
		err := s.tx.Run(ctx, func(txCtx context.Context) error {
			order, err := s.orders.GetOrderByNo(txCtx, candidate.OrderNo)
			if err != nil {
				if errors.Is(err, domain.ErrOrderNotFound) {
					return nil
				}
				return err
			}
			if order.Status != domain.OrderPendingPayment {
				return nil // 已被其他路径处理
			}
			if err := order.Transition(domain.OrderEventTimeout); err != nil {
				return err
			}
			if err := s.orders.Transition(txCtx, order.OrderNo, domain.OrderPendingPayment, domain.OrderExpired, order.Version); err != nil {
				return err
			}
			if err := s.locks.ReleaseByOrderNo(txCtx, order.OrderNo, domain.SeatLockReleased); err != nil {
				return err
			}
			if err := s.sessions.RecalcStatus(txCtx, order.SessionID); err != nil {
				return err
			}
			if err := s.coupons.UnlockByOrderNo(txCtx, order.OrderNo); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return processed, err
		}
		processed++
	}
	return processed, nil
}

// CancelPending 用户取消待支付订单（改签回滚用）：释放锁、解锁券。
func (s *OrderSvc) CancelPending(ctx context.Context, userID int64, orderNo string) error {
	return s.tx.Run(ctx, func(txCtx context.Context) error {
		order, err := s.orders.GetOrderByNo(txCtx, orderNo)
		if err != nil {
			return err
		}
		if order.UserID != userID {
			return domain.ErrForbidden
		}
		if order.Status == domain.OrderCanceled {
			return nil
		}
		if order.Status != domain.OrderPendingPayment {
			return domain.ErrInvalidTransition
		}
		if err := order.Transition(domain.OrderEventUserCancel); err != nil {
			return err
		}
		if err := s.orders.Transition(txCtx, orderNo, domain.OrderPendingPayment, domain.OrderCanceled, order.Version); err != nil {
			return err
		}
		if err := s.locks.ReleaseByOrderNo(txCtx, orderNo, domain.SeatLockReleased); err != nil {
			return err
		}
		if err := s.sessions.RecalcStatus(txCtx, order.SessionID); err != nil {
			return err
		}
		if err := s.coupons.UnlockByOrderNo(txCtx, orderNo); err != nil {
			return err
		}
		return nil
	})
}
