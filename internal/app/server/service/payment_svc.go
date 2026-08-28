package service

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/uid"
)

const mockPayChannel = "MOCK_PAY"

// PaymentSvc 支付用例：创建支付交易 + 处理回调（幂等）。
type PaymentSvc struct {
	tx        port.TxManager
	payments  port.PaymentRepo
	callbacks port.PaymentCallbackRepo
	orders    port.OrderRepo
	locks     port.SeatLockRepo
	coupons   port.UserCouponRepo
	points    port.PointsRepo
	box       port.BoxOfficeRepo
}

func NewPaymentSvc(
	tx port.TxManager,
	payments port.PaymentRepo,
	callbacks port.PaymentCallbackRepo,
	orders port.OrderRepo,
	locks port.SeatLockRepo,
	coupons port.UserCouponRepo,
	points port.PointsRepo,
	box port.BoxOfficeRepo,
) *PaymentSvc {
	return &PaymentSvc{
		tx:        tx,
		payments:  payments,
		callbacks: callbacks,
		orders:    orders,
		locks:     locks,
		coupons:   coupons,
		points:    points,
		box:       box,
	}
}

type CreatePaymentInput struct {
	OrderNo string
}

// CreatePayment 为待支付订单创建支付交易（PENDING）。
// 已存在交易则直接返回（用户重试幂等）；一单一付由 (biz_type, biz_no) 唯一约束兜底。
func (s *PaymentSvc) CreatePayment(ctx context.Context, in CreatePaymentInput) (*domain.PaymentTransaction, error) {
	var payment *domain.PaymentTransaction
	err := s.tx.Run(ctx, func(txCtx context.Context) error {
		order, err := s.orders.GetOrderByNo(txCtx, in.OrderNo)
		if err != nil {
			return err
		}
		if order.Status != domain.OrderPendingPayment {
			return domain.ErrInvalidTransition
		}
		if order.IsExpired(time.Now()) {
			return domain.ErrOrderExpired
		}

		existing, err := s.payments.GetByOrderNo(txCtx, in.OrderNo)
		if err == nil {
			payment = existing
			return nil
		}
		if !errors.Is(err, domain.ErrPaymentNotFound) {
			return err
		}

		payment = &domain.PaymentTransaction{
			TransactionNo: uid.TransactionNo(),
			OrderNo:       order.OrderNo,
			UserID:        order.UserID,
			AmountCents:   order.PaidCents,
			Channel:       mockPayChannel,
			Status:        domain.PaymentPending,
			Version:       1,
			CreatedAt:     time.Now(),
		}
		return s.payments.CreateTransaction(txCtx, payment)
	})
	if err != nil {
		return nil, err
	}
	return payment, nil
}

type MockCallbackInput struct {
	EventID       string
	TransactionNo string
	AmountCents   int64
	Payload       string
}

// HandleMockCallback 模拟网关回调：event_id 幂等 + 单事务完成出票链路。
// 事务内：回调落库 → 交易 SUCCESS → 订单 PAID → 锁座 BOOKED →
// 生成取票码 → 优惠券 USED → 回调 PROCESSED。
func (s *PaymentSvc) HandleMockCallback(ctx context.Context, in MockCallbackInput) error {
	return s.tx.Run(ctx, func(txCtx context.Context) error {
		cb := &domain.PaymentCallback{
			EventID:       in.EventID,
			TransactionNo: in.TransactionNo,
			AmountCents:   in.AmountCents,
			Payload:       in.Payload,
			CreatedAt:     time.Now(),
		}
		inserted, err := s.callbacks.InsertIfAbsent(txCtx, cb)
		if err != nil {
			return err
		}
		if !inserted {
			existing, err := s.callbacks.GetByEventID(txCtx, cb.EventID)
			if err != nil {
				return err
			}
			if existing.Status == domain.CallbackProcessed {
				return nil // 已处理：按幂等成功返回
			}
			// RECEIVED/FAILED：继续处理（重试场景）
		}

		payment, err := s.payments.GetByTransactionNo(txCtx, in.TransactionNo)
		if err != nil {
			return err
		}
		if payment.Status != domain.PaymentPending {
			return domain.ErrInvalidTransition
		}
		if payment.AmountCents != in.AmountCents {
			return domain.ErrPaymentAmountMismatch
		}
		if err := payment.Transition(domain.PaymentEventSuccess); err != nil {
			return err
		}
		if err := s.payments.Transition(txCtx, payment.TransactionNo, domain.PaymentPending, domain.PaymentSuccess, payment.Version); err != nil {
			return err
		}

		order, err := s.orders.GetOrderByNo(txCtx, payment.OrderNo)
		if err != nil {
			return err
		}
		if order.Status != domain.OrderPendingPayment {
			return domain.ErrInvalidTransition
		}
		if order.PaidCents != in.AmountCents {
			return domain.ErrPaymentAmountMismatch
		}
		if err := order.Transition(domain.OrderEventPaySuccess); err != nil {
			return err
		}
		if err := s.orders.Transition(txCtx, order.OrderNo, domain.OrderPendingPayment, domain.OrderPaid, order.Version); err != nil {
			return err
		}

		if err := s.locks.MarkBookedByOrderNo(txCtx, order.OrderNo); err != nil {
			return err
		}

		tickets := make([]domain.OrderItem, 0, len(order.Items))
		for _, item := range order.Items {
			tickets = append(tickets, domain.OrderItem{
				OrderNo:  order.OrderNo,
				SeatID:   item.SeatID,
				TicketNo: uid.TicketNo(),
			})
		}
		if err := s.orders.IssueTickets(txCtx, order.OrderNo, tickets); err != nil {
			return err
		}

		if err := s.coupons.MarkUsedByOrderNo(txCtx, order.OrderNo); err != nil {
			return err
		}

		if err := s.points.GrantOnPaid(txCtx, order.UserID, order.PaidCents, order.OrderNo); err != nil {
			return err
		}
		if err := s.box.Record(txCtx, domain.NewPaidEvent(order, order.OrderNo, time.Now())); err != nil {
			return err
		}

		return s.callbacks.MarkProcessed(txCtx, cb.EventID)
	})
}

const maxCallbackRetryAttempts = 5

// RetryCallbacks 定时任务：重试未处理/失败的回调，返回失败数量。
func (s *PaymentSvc) RetryCallbacks(ctx context.Context, limit int) (int, error) {
	callbacks, err := s.callbacks.ListPending(ctx, limit)
	if err != nil {
		return 0, err
	}
	failed := 0
	for _, cb := range callbacks {
		if cb.RetryCount >= maxCallbackRetryAttempts {
			continue
		}
		if err := s.HandleMockCallback(ctx, MockCallbackInput{
			EventID:       cb.EventID,
			TransactionNo: cb.TransactionNo,
			AmountCents:   cb.AmountCents,
			Payload:       cb.Payload,
		}); err != nil {
			failed++
			if err := s.callbacks.IncrementRetry(ctx, cb.EventID); err != nil {
				return failed, err
			}
		}
	}
	return failed, nil
}

// MockPay 模拟支付页确认：生成回调事件并走完整回调链路（幂等）。
func (s *PaymentSvc) MockPay(ctx context.Context, userID int64, transactionNo string) error {
	payment, err := s.payments.GetByTransactionNo(ctx, transactionNo)
	if err != nil {
		return err
	}
	if payment.UserID != userID {
		return domain.ErrForbidden
	}
	return s.HandleMockCallback(ctx, MockCallbackInput{
		EventID:       uid.EventID(),
		TransactionNo: transactionNo,
		AmountCents:   payment.AmountCents,
		Payload:       "mock-pay",
	})
}

// GetByOrder 查询订单支付信息（轮询用），校验归属。
func (s *PaymentSvc) GetByOrder(ctx context.Context, userID int64, orderNo string) (*domain.PaymentTransaction, error) {
	payment, err := s.payments.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if payment.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return payment, nil
}
