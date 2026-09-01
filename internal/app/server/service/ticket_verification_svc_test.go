package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type fakeTicketRepo struct {
	orders map[string]*domain.Order
}

func (f *fakeTicketRepo) GetOrderByTicketNo(ctx context.Context, ticketNo string) (*domain.Order, error) {
	for _, order := range f.orders {
		for _, item := range order.Items {
			if item.TicketNo == ticketNo {
				return order, nil
			}
		}
	}
	return nil, domain.ErrTicketNotFound
}

func (f *fakeTicketRepo) MarkTicketUsed(ctx context.Context, ticketNo string, usedAt time.Time) (bool, error) {
	for _, order := range f.orders {
		for index := range order.Items {
			if order.Items[index].TicketNo != ticketNo {
				continue
			}
			if order.Items[index].UsedAt != nil {
				return false, nil
			}
			order.Items[index].UsedAt = &usedAt
			return true, nil
		}
	}
	return false, domain.ErrTicketNotFound
}

func (f *fakeTicketRepo) CountUnusedTickets(ctx context.Context, orderNo string) (int64, error) {
	order := f.orders[orderNo]
	if order == nil {
		return 0, domain.ErrOrderNotFound
	}
	var count int64
	for _, item := range order.Items {
		if item.UsedAt == nil {
			count++
		}
	}
	return count, nil
}

func TestTicketVerificationCompletesSingleTicketAndIsIdempotent(t *testing.T) {
	order := &domain.Order{
		OrderNo:  "O1",
		CinemaID: 10,
		MovieID:  20,
		Status:   domain.OrderPaid,
		Version:  1,
		Items:    []domain.OrderItem{{TicketNo: "TK1", SeatNo: "A1"}},
	}
	tickets := &fakeTicketRepo{orders: map[string]*domain.Order{"O1": order}}
	svc := NewTicketVerificationSvc(fakeTxManager{}, tickets, &fakeOrderRepo{orders: tickets.orders}, &fakeOperationLogRepo{})
	scope := domain.AdminScope{AdminID: 8, Role: domain.RoleCinemaAdmin, CinemaID: int64Ptr(10)}

	result, err := svc.Verify(context.Background(), scope, " TK1 ")
	if err != nil {
		t.Fatalf("verify ticket: %v", err)
	}
	if result.AlreadyUsed || result.OrderStatus != domain.OrderCompleted || order.Status != domain.OrderCompleted {
		t.Fatalf("unexpected first verification: %+v order=%s", result, order.Status)
	}

	result, err = svc.Verify(context.Background(), scope, "TK1")
	if err != nil {
		t.Fatalf("verify duplicate ticket: %v", err)
	}
	if !result.AlreadyUsed || result.OrderStatus != domain.OrderCompleted {
		t.Fatalf("unexpected duplicate verification: %+v", result)
	}
}

func TestTicketVerificationCompletesOnlyAfterAllTickets(t *testing.T) {
	order := &domain.Order{
		OrderNo:  "O2",
		CinemaID: 10,
		Status:   domain.OrderPaid,
		Version:  1,
		Items:    []domain.OrderItem{{TicketNo: "TK2A", SeatNo: "A1"}, {TicketNo: "TK2B", SeatNo: "A2"}},
	}
	tickets := &fakeTicketRepo{orders: map[string]*domain.Order{"O2": order}}
	orders := &fakeOrderRepo{orders: tickets.orders}
	svc := NewTicketVerificationSvc(fakeTxManager{}, tickets, orders, &fakeOperationLogRepo{})
	scope := domain.AdminScope{AdminID: 8, Role: domain.RoleSuperAdmin}

	first, err := svc.Verify(context.Background(), scope, "TK2A")
	if err != nil {
		t.Fatalf("verify first ticket: %v", err)
	}
	if first.OrderStatus != domain.OrderPaid || order.Status != domain.OrderPaid {
		t.Fatalf("order should remain paid after partial verification: %+v", first)
	}
	second, err := svc.Verify(context.Background(), scope, "TK2B")
	if err != nil {
		t.Fatalf("verify second ticket: %v", err)
	}
	if second.OrderStatus != domain.OrderCompleted || order.Status != domain.OrderCompleted {
		t.Fatalf("order should complete after all tickets: %+v", second)
	}
}

func TestTicketVerificationEnforcesCinemaScope(t *testing.T) {
	order := &domain.Order{
		OrderNo:  "O3",
		CinemaID: 10,
		Status:   domain.OrderPaid,
		Items:    []domain.OrderItem{{TicketNo: "TK3", SeatNo: "A1"}},
	}
	tickets := &fakeTicketRepo{orders: map[string]*domain.Order{"O3": order}}
	svc := NewTicketVerificationSvc(fakeTxManager{}, tickets, &fakeOrderRepo{orders: tickets.orders}, &fakeOperationLogRepo{})
	_, err := svc.Verify(context.Background(), domain.AdminScope{AdminID: 9, Role: domain.RoleCinemaAdmin, CinemaID: int64Ptr(99)}, "TK3")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
