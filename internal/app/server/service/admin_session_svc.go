package service

import (
	"context"
	"strconv"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
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
	logs     port.OperationLogRepo
}

func NewAdminSessionSvc(
	sessions port.SessionRepo,
	movies port.MovieRepo,
	halls port.HallRepo,
	locks port.SeatLockRepo,
	orders port.OrderRepo,
	coupons port.UserCouponRepo,
	logs port.OperationLogRepo,
) *AdminSessionSvc {
	return &AdminSessionSvc{
		sessions: sessions,
		movies:   movies,
		halls:    halls,
		locks:    locks,
		orders:   orders,
		coupons:  coupons,
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

func (s *AdminSessionSvc) Create(ctx context.Context, adminID int64, in SessionInput) (*domain.ShowSession, error) {
	if in.HallID <= 0 || in.MovieID <= 0 || in.BasePriceCents <= 0 ||
		!in.EndTime.After(in.StartTime) {
		return nil, domain.ErrSessionInvalid
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
		Status:         domain.SessionOpen,
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, s.log(ctx, adminID, "CREATE_SESSION", "session", strconv.FormatInt(session.ID, 10), session)
}

// UpdatePrice 开场前 30 分钟锁定改价。
func (s *AdminSessionSvc) UpdatePrice(ctx context.Context, adminID, sessionID int64, basePriceCents int64, priceRulesJSON string) error {
	session, err := s.sessions.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if !s.canChange(session) {
		return domain.ErrSessionLockedForChange
	}
	if basePriceCents <= 0 {
		return domain.ErrSessionInvalid
	}
	if err := s.sessions.UpdatePrice(ctx, sessionID, basePriceCents, priceRulesJSON); err != nil {
		return err
	}
	return s.log(ctx, adminID, "UPDATE_SESSION_PRICE", "session", strconv.FormatInt(sessionID, 10), map[string]any{
		"base_price_cents": basePriceCents,
		"price_rules":      priceRulesJSON,
	})
}

// Cancel 取消场次：释放锁、过期待支付订单、解锁优惠券。
func (s *AdminSessionSvc) Cancel(ctx context.Context, adminID, sessionID int64) error {
	session, err := s.sessions.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
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
	return s.log(ctx, adminID, "CANCEL_SESSION", "session", strconv.FormatInt(sessionID, 10), nil)
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
