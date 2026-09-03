package service

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/uid"
)

// RefundSvc 用户主动退款：申请（PENDING）→ Mock 回调成功（REFUNDED）。
type RefundSvc struct {
	tx       port.TxManager
	orders   port.OrderRepo
	refunds  port.RefundRepo
	payments port.PaymentRepo
	locks    port.SeatLockRepo
	points   port.PointsRepo
	sessions port.SessionRepo
	box      port.BoxOfficeRepo
}

func NewRefundSvc(
	tx port.TxManager,
	orders port.OrderRepo,
	refunds port.RefundRepo,
	payments port.PaymentRepo,
	locks port.SeatLockRepo,
	points port.PointsRepo,
	sessions port.SessionRepo,
	box port.BoxOfficeRepo,
) *RefundSvc {
	return &RefundSvc{
		tx:       tx,
		orders:   orders,
		refunds:  refunds,
		payments: payments,
		locks:    locks,
		points:   points,
		sessions: sessions,
		box:      box,
	}
}

type ApplyRefundInput struct {
	OrderNo string
	Reason  string
}

// ApplyRefund 开场前整单退款：校验归属/可退 → 创建退款单 PENDING → 订单 REFUNDING。
// 重复申请返回已有退款单（一单一退幂等）。
func (s *RefundSvc) ApplyRefund(ctx context.Context, userID int64, in ApplyRefundInput) (*domain.Refund, error) {
	var refund *domain.Refund
	err := s.tx.Run(ctx, func(txCtx context.Context) error {
		// Serialize concurrent refund requests for the same order. This keeps
		// the status check and refund creation atomic at the business level.
		order, err := s.orders.GetOrderForUpdate(txCtx, in.OrderNo)
		if err != nil {
			return err
		}
		if order.UserID != userID {
			return domain.ErrForbidden
		}
		if existing, err := s.refunds.GetByOrderNo(txCtx, in.OrderNo); err == nil {
			refund = existing
			return nil
		} else if !errors.Is(err, domain.ErrRefundNotFound) {
			return err
		}
		if order.Status != domain.OrderPaid || hasUsedTicket(order) {
			return domain.ErrOrderNotRefundable
		}
		session, err := s.sessions.GetSessionByID(txCtx, order.SessionID)
		if err != nil {
			return err
		}
		if !session.StartTime.After(time.Now()) {
			return domain.ErrOrderNotRefundable
		}

		refund = &domain.Refund{
			RefundNo:         uid.RefundNo(),
			OrderNo:          order.OrderNo,
			UserID:           order.UserID,
			AmountCents:      order.PaidCents,
			Reason:           in.Reason,
			Status:           domain.RefundPending,
			ExternalRefundNo: uid.ExternalRefundNo(),
		}
		if err := s.refunds.Create(txCtx, refund); err != nil {
			return err
		}
		if err := order.Transition(domain.OrderEventApplyRefund); err != nil {
			return err
		}
		return s.orders.Transition(txCtx, order.OrderNo, domain.OrderPaid, domain.OrderRefunding, order.Version)
	})
	if err != nil {
		return nil, err
	}
	return refund, nil
}

// HandleMockCallback Mock 退款回调：单事务完成 订单 REFUNDED + 交易 REFUNDED +
// 锁 RELEASED + 积分扣回 + 退款 SUCCESS。重复回调幂等。
func (s *RefundSvc) HandleMockCallback(ctx context.Context, refundNo string) error {
	return s.tx.Run(ctx, func(txCtx context.Context) error {
		refund, err := s.refunds.GetByRefundNo(txCtx, refundNo)
		if err != nil {
			return err
		}
		if refund.Status == domain.RefundSuccess {
			return nil // 已处理，幂等成功
		}
		if refund.Status != domain.RefundPending {
			return domain.ErrInvalidTransition
		}

		order, err := s.orders.GetOrderByNo(txCtx, refund.OrderNo)
		if err != nil {
			return err
		}
		if order.Status != domain.OrderRefunding {
			return domain.ErrInvalidTransition
		}
		if err := order.Transition(domain.OrderEventRefundSuccess); err != nil {
			return err
		}
		if err := s.orders.Transition(txCtx, order.OrderNo, domain.OrderRefunding, domain.OrderRefunded, order.Version); err != nil {
			return err
		}

		payment, err := s.payments.GetByOrderNo(txCtx, refund.OrderNo)
		if err != nil && !errors.Is(err, domain.ErrPaymentNotFound) {
			return err
		}
		if err == nil && payment.Status == domain.PaymentSuccess {
			if err := payment.Transition(domain.PaymentEventRefunded); err != nil {
				return err
			}
			if err := s.payments.Transition(txCtx, payment.TransactionNo, domain.PaymentSuccess, domain.PaymentRefunded, payment.Version); err != nil {
				return err
			}
		}

		if err := s.locks.ReleaseByOrderNo(txCtx, refund.OrderNo, domain.SeatLockReleased); err != nil {
			return err
		}
		if err := s.sessions.RecalcStatus(txCtx, order.SessionID); err != nil {
			return err
		}
		if err := s.points.ReclaimOnRefund(txCtx, refund.UserID, refund.AmountCents, refund.RefundNo); err != nil {
			return err
		}
		if err := s.box.Record(txCtx, domain.NewRefundEvent(order, refund.AmountCents, refund.RefundNo, time.Now())); err != nil {
			return err
		}
		return s.refunds.MarkSuccess(txCtx, refund.RefundNo)
	})
}

func (s *RefundSvc) HandleMockCallbackForUser(ctx context.Context, userID int64, refundNo string) error {
	refund, err := s.refunds.GetByRefundNo(ctx, refundNo)
	if err != nil {
		return err
	}
	if refund.UserID != userID {
		return domain.ErrForbidden
	}
	return s.HandleMockCallback(ctx, refundNo)
}
