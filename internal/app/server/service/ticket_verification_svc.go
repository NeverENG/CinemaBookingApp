package service

import (
	"context"
	"strings"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

// TicketVerificationSvc 票券核销：支付成功后由影院工作人员核销取票码。
type TicketVerificationSvc struct {
	tx      port.TxManager
	tickets port.TicketRepo
	orders  port.OrderRepo
	logs    port.OperationLogRepo
}

func NewTicketVerificationSvc(tx port.TxManager, tickets port.TicketRepo, orders port.OrderRepo, logs port.OperationLogRepo) *TicketVerificationSvc {
	return &TicketVerificationSvc{tx: tx, tickets: tickets, orders: orders, logs: logs}
}

type TicketVerificationResult struct {
	TicketNo    string             `json:"ticket_no"`
	OrderNo     string             `json:"order_no"`
	SeatNo      string             `json:"seat_no"`
	MovieID     int64              `json:"movie_id"`
	CinemaID    int64              `json:"cinema_id"`
	UsedAt      time.Time          `json:"used_at"`
	OrderStatus domain.OrderStatus `json:"order_status"`
	AlreadyUsed bool               `json:"already_used"`
}

func (s *TicketVerificationSvc) Verify(ctx context.Context, scope domain.AdminScope, ticketNo string) (*TicketVerificationResult, error) {
	ticketNo = strings.TrimSpace(ticketNo)
	if ticketNo == "" {
		return nil, domain.ErrInvalidInput
	}
	if scope.Role != domain.RoleSuperAdmin && scope.Role != domain.RoleCinemaAdmin {
		return nil, domain.ErrForbidden
	}

	var result *TicketVerificationResult
	err := s.tx.Run(ctx, func(txCtx context.Context) error {
		order, err := s.tickets.GetOrderByTicketNo(txCtx, ticketNo)
		if err != nil {
			return err
		}
		if scope.IsCinemaAdmin() && (scope.CinemaID == nil || *scope.CinemaID != order.CinemaID) {
			return domain.ErrForbidden
		}

		item := findOrderItem(order, ticketNo)
		if item == nil {
			return domain.ErrTicketNotFound
		}
		if item.UsedAt != nil {
			if order.Status != domain.OrderPaid && order.Status != domain.OrderCompleted {
				return domain.ErrTicketNotUsable
			}
			result = verificationResult(order, item, *item.UsedAt, true)
			return nil
		}
		if order.Status != domain.OrderPaid {
			return domain.ErrTicketNotUsable
		}

		now := time.Now()
		marked, err := s.tickets.MarkTicketUsed(txCtx, ticketNo, now)
		if err != nil {
			return err
		}
		if !marked {
			// 兼容并发请求：另一请求已完成核销时直接返回幂等成功。
			current, reloadErr := s.tickets.GetOrderByTicketNo(txCtx, ticketNo)
			if reloadErr != nil {
				return reloadErr
			}
			currentItem := findOrderItem(current, ticketNo)
			if currentItem == nil || currentItem.UsedAt == nil {
				return domain.ErrTicketNotUsable
			}
			if current.Status != domain.OrderPaid && current.Status != domain.OrderCompleted {
				return domain.ErrTicketNotUsable
			}
			result = verificationResult(current, currentItem, *currentItem.UsedAt, true)
			return nil
		}

		unused, err := s.tickets.CountUnusedTickets(txCtx, order.OrderNo)
		if err != nil {
			return err
		}
		if unused == 0 {
			if err := order.Transition(domain.OrderEventComplete); err != nil {
				return err
			}
			if err := s.orders.Transition(txCtx, order.OrderNo, domain.OrderPaid, domain.OrderCompleted, order.Version); err != nil {
				return err
			}
		}
		result = verificationResult(order, item, now, false)
		return s.logs.Create(txCtx, &domain.OperationLog{
			AdminID:    scope.AdminID,
			Action:     "VERIFY_TICKET",
			TargetType: "ticket",
			TargetID:   ticketNo,
			Detail: map[string]any{
				"order_no":     order.OrderNo,
				"seat_no":      item.SeatNo,
				"order_status": order.Status,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func findOrderItem(order *domain.Order, ticketNo string) *domain.OrderItem {
	for index := range order.Items {
		if order.Items[index].TicketNo == ticketNo {
			return &order.Items[index]
		}
	}
	return nil
}

func verificationResult(order *domain.Order, item *domain.OrderItem, usedAt time.Time, alreadyUsed bool) *TicketVerificationResult {
	return &TicketVerificationResult{
		TicketNo:    item.TicketNo,
		OrderNo:     order.OrderNo,
		SeatNo:      item.SeatNo,
		MovieID:     order.MovieID,
		CinemaID:    order.CinemaID,
		UsedAt:      usedAt,
		OrderStatus: order.Status,
		AlreadyUsed: alreadyUsed,
	}
}

func hasUsedTicket(order *domain.Order) bool {
	for _, item := range order.Items {
		if item.UsedAt != nil {
			return true
		}
	}
	return false
}
