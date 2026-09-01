package service

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

// ChangeTicketSvc 改签：同影片跨场次 = 退旧单 + 订新单（多退少补）。
type ChangeTicketSvc struct {
	orders     port.OrderRepo
	sessions   port.SessionRepo
	orderSvc   *OrderSvc
	paymentSvc *PaymentSvc
	refundSvc  *RefundSvc
}

func NewChangeTicketSvc(
	orders port.OrderRepo,
	sessions port.SessionRepo,
	orderSvc *OrderSvc,
	paymentSvc *PaymentSvc,
	refundSvc *RefundSvc,
) *ChangeTicketSvc {
	return &ChangeTicketSvc{orders: orders, sessions: sessions, orderSvc: orderSvc, paymentSvc: paymentSvc, refundSvc: refundSvc}
}

type ChangeTicketResult struct {
	NewOrderNo        string `json:"new_order_no"`
	NewPaidCents      int64  `json:"new_paid_cents"`
	RefundNo          string `json:"refund_no"`
	RefundAmountCents int64  `json:"refund_amount_cents"`
}

func (s *ChangeTicketSvc) Change(
	ctx context.Context,
	userID int64,
	orderNo string,
	newSessionID int64,
	newSeatIDs []int64,
) (*ChangeTicketResult, error) {
	order, err := s.orders.GetOrderByNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, domain.ErrForbidden
	}
	if order.Status != domain.OrderPaid || hasUsedTicket(order) {
		return nil, domain.ErrOrderNotRefundable
	}
	if len(order.Items) > 0 && len(newSeatIDs) != len(order.Items) {
		return nil, domain.ErrChangeSeatCount
	}
	oldSession, err := s.sessions.GetSessionByID(ctx, order.SessionID)
	if err != nil {
		return nil, err
	}
	if !oldSession.StartTime.After(time.Now()) {
		return nil, domain.ErrOrderNotRefundable
	}
	newSession, err := s.sessions.GetSessionByID(ctx, newSessionID)
	if err != nil {
		return nil, err
	}
	if !newSession.CanBook(time.Now()) {
		return nil, domain.ErrSessionNotBookable
	}
	if newSession.MovieID != oldSession.MovieID {
		return nil, domain.ErrChangeMovieMismatch
	}

	// 1. 订新单
	newOrder, err := s.orderSvc.CreateOrder(ctx, CreateOrderInput{
		UserID:    userID,
		SessionID: newSessionID,
		SeatIDs:   newSeatIDs,
	})
	if err != nil {
		return nil, err
	}

	// 2. 支付新单；失败取消新单，原单不受影响
	payment, err := s.paymentSvc.CreatePayment(ctx, CreatePaymentInput{UserID: userID, OrderNo: newOrder.OrderNo})
	if err != nil {
		_ = s.orderSvc.CancelPending(ctx, userID, newOrder.OrderNo)
		return nil, err
	}
	if err := s.paymentSvc.MockPay(ctx, userID, payment.TransactionNo); err != nil {
		_ = s.orderSvc.CancelPending(ctx, userID, newOrder.OrderNo)
		return nil, err
	}

	// 3. 退原单；失败则退回新单，恢复原状
	refund, err := s.refundSvc.ApplyRefund(ctx, userID, ApplyRefundInput{OrderNo: orderNo, Reason: "change_ticket"})
	if err != nil {
		s.rollbackNewOrder(ctx, userID, newOrder.OrderNo)
		return nil, err
	}
	if err := s.refundSvc.HandleMockCallback(ctx, refund.RefundNo); err != nil {
		s.rollbackNewOrder(ctx, userID, newOrder.OrderNo)
		return nil, err
	}

	return &ChangeTicketResult{
		NewOrderNo:        newOrder.OrderNo,
		NewPaidCents:      newOrder.PaidCents,
		RefundNo:          refund.RefundNo,
		RefundAmountCents: refund.AmountCents,
	}, nil
}

// rollbackNewOrder 新单已支付但原单退款失败时，把新单也退掉。
func (s *ChangeTicketSvc) rollbackNewOrder(ctx context.Context, userID int64, newOrderNo string) {
	refund, err := s.refundSvc.ApplyRefund(ctx, userID, ApplyRefundInput{OrderNo: newOrderNo, Reason: "change_rollback"})
	if err != nil {
		return
	}
	_ = s.refundSvc.HandleMockCallback(ctx, refund.RefundNo)
}
