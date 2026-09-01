package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/uid"
)

const sessionChangeLockMinutes = 30

// AdminSessionSvc 场次排期/改价/取消。
type AdminSessionSvc struct {
	sessions port.SessionRepo
	movies   port.MovieRepo
	halls    port.HallRepo
	locks    port.SeatLockRepo
	orders   port.OrderRepo
	coupons  port.UserCouponRepo
	refunds  port.RefundRepo
	payments port.PaymentRepo
	points   port.PointsRepo
	box      port.BoxOfficeRepo
	logs     port.OperationLogRepo
}

func NewAdminSessionSvc(
	sessions port.SessionRepo,
	movies port.MovieRepo,
	halls port.HallRepo,
	locks port.SeatLockRepo,
	orders port.OrderRepo,
	coupons port.UserCouponRepo,
	refunds port.RefundRepo,
	payments port.PaymentRepo,
	points port.PointsRepo,
	box port.BoxOfficeRepo,
	logs port.OperationLogRepo,
) *AdminSessionSvc {
	return &AdminSessionSvc{
		sessions: sessions,
		movies:   movies,
		halls:    halls,
		locks:    locks,
		orders:   orders,
		coupons:  coupons,
		refunds:  refunds,
		payments: payments,
		points:   points,
		box:      box,
		logs:     logs,
	}
}

type SessionInput struct {
	CinemaID       int64
	HallID         int64
	MovieID        int64
	StartTime      time.Time
	EndTime        time.Time
	BasePriceCents int64
	PriceRulesJSON string
}

func (s *AdminSessionSvc) Create(ctx context.Context, scope domain.AdminScope, in SessionInput) (*domain.ShowSession, error) {
	if !scope.CanManageCinema(in.CinemaID) {
		return nil, domain.ErrForbidden
	}
	if in.HallID <= 0 || in.MovieID <= 0 || in.BasePriceCents <= 0 ||
		!in.EndTime.After(in.StartTime) {
		return nil, domain.ErrSessionInvalid
	}
	priceRulesJSON, err := normalizePriceRulesJSON(in.PriceRulesJSON)
	if err != nil {
		return nil, err
	}
	if _, err := s.movies.GetByID(ctx, in.MovieID); err != nil {
		return nil, err
	}
	hall, err := s.halls.GetByID(ctx, in.HallID)
	if err != nil {
		return nil, err
	}
	if hall.CinemaID != in.CinemaID {
		return nil, domain.ErrSessionInvalid
	}

	overlaps, err := s.sessions.ListOverlapping(ctx, in.HallID, in.StartTime, in.EndTime)
	if err != nil {
		return nil, err
	}
	if len(overlaps) > 0 {
		return nil, domain.ErrSessionTimeConflict
	}

	session := &domain.ShowSession{
		CinemaID:       in.CinemaID,
		HallID:         in.HallID,
		MovieID:        in.MovieID,
		StartTime:      in.StartTime,
		EndTime:        in.EndTime,
		BasePriceCents: in.BasePriceCents,
		PriceRulesJSON: priceRulesJSON,
		Status:         domain.SessionOpen,
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, s.log(ctx, scope.AdminID, "CREATE_SESSION", "session", strconv.FormatInt(session.ID, 10), session)
}

// UpdatePrice 开场前 30 分钟锁定改价。
func (s *AdminSessionSvc) UpdatePrice(ctx context.Context, scope domain.AdminScope, sessionID int64, basePriceCents int64, priceRulesJSON string) error {
	session, err := s.sessions.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if !scope.CanManageCinema(session.CinemaID) {
		return domain.ErrForbidden
	}
	if !s.canChange(session) {
		return domain.ErrSessionLockedForChange
	}
	if basePriceCents <= 0 {
		return domain.ErrSessionInvalid
	}
	priceRulesJSON, err = normalizePriceRulesJSON(priceRulesJSON)
	if err != nil {
		return err
	}
	if err := s.sessions.UpdatePrice(ctx, sessionID, basePriceCents, priceRulesJSON); err != nil {
		return err
	}
	return s.log(ctx, scope.AdminID, "UPDATE_SESSION_PRICE", "session", strconv.FormatInt(sessionID, 10), map[string]any{
		"base_price_cents": basePriceCents,
		"price_rules":      priceRulesJSON,
	})
}

// Cancel 取消场次：释放锁、过期待支付订单、解锁优惠券。
func (s *AdminSessionSvc) Cancel(ctx context.Context, scope domain.AdminScope, sessionID int64) error {
	session, err := s.sessions.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if !scope.CanManageCinema(session.CinemaID) {
		return domain.ErrForbidden
	}
	if !s.canChange(session) {
		return domain.ErrSessionLockedForChange
	}
	if err := s.sessions.Cancel(ctx, sessionID); err != nil {
		return err
	}
	if err := s.locks.ReleaseBySessionID(ctx, sessionID, domain.SeatLockReleased); err != nil {
		return err
	}
	orderNos, err := s.orders.ExpirePendingBySessionID(ctx, sessionID)
	if err != nil {
		return err
	}
	for _, no := range orderNos {
		if err := s.coupons.UnlockByOrderNo(ctx, no); err != nil {
			return err
		}
	}

	// 已支付订单自动退款（整单，Mock 即时成功）
	paidOrders, err := s.orders.ListPaidBySessionID(ctx, sessionID)
	if err != nil {
		return err
	}
	for _, order := range paidOrders {
		if err := s.refundOrder(ctx, order); err != nil {
			return err
		}
	}
	return s.log(ctx, scope.AdminID, "CANCEL_SESSION", "session", strconv.FormatInt(sessionID, 10), nil)
}

func (s *AdminSessionSvc) refundOrder(ctx context.Context, order domain.Order) error {
	refund := &domain.Refund{
		RefundNo:         uid.RefundNo(),
		OrderNo:          order.OrderNo,
		UserID:           order.UserID,
		AmountCents:      order.PaidCents,
		Reason:           "session_canceled",
		Status:           domain.RefundSuccess,
		ExternalRefundNo: uid.ExternalRefundNo(),
	}
	if err := s.refunds.Create(ctx, refund); err != nil {
		return err
	}
	if err := s.points.ReclaimOnRefund(ctx, order.UserID, refund.AmountCents, refund.RefundNo); err != nil {
		return err
	}
	if err := s.box.Record(ctx, domain.NewRefundEvent(&order, refund.AmountCents, refund.RefundNo, time.Now())); err != nil {
		return err
	}

	// 订单 PAID -> REFUNDING -> REFUNDED（乐观锁版本递增）
	if err := order.Transition(domain.OrderEventApplyRefund); err != nil {
		return err
	}
	if err := s.orders.Transition(ctx, order.OrderNo, domain.OrderPaid, domain.OrderRefunding, order.Version); err != nil {
		return err
	}
	if err := order.Transition(domain.OrderEventRefundSuccess); err != nil {
		return err
	}
	if err := s.orders.Transition(ctx, order.OrderNo, domain.OrderRefunding, domain.OrderRefunded, order.Version+1); err != nil {
		return err
	}

	// 支付交易 SUCCESS -> REFUNDED
	payment, err := s.payments.GetByOrderNo(ctx, order.OrderNo)
	if err != nil && !errors.Is(err, domain.ErrPaymentNotFound) {
		return err
	}
	if err == nil && payment.Status == domain.PaymentSuccess {
		if err := payment.Transition(domain.PaymentEventRefunded); err != nil {
			return err
		}
		if err := s.payments.Transition(ctx, payment.TransactionNo, domain.PaymentSuccess, domain.PaymentRefunded, payment.Version); err != nil {
			return err
		}
	}
	return nil
}

// canChange 场次未开场且未结束/取消时允许修改（预留 30 分钟锁定期）。
func (s *AdminSessionSvc) canChange(session *domain.ShowSession) bool {
	if session.Status == domain.SessionClosed || session.Status == domain.SessionCanceled {
		return false
	}
	return time.Now().Add(sessionChangeLockMinutes * time.Minute).Before(session.StartTime)
}

func (s *AdminSessionSvc) log(ctx context.Context, adminID int64, action, targetType, targetID string, detail any) error {
	return s.logs.Create(ctx, &domain.OperationLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detail,
	})
}
