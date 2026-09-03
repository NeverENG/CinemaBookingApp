package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type seatLockRow struct {
	ID         int64                 `gorm:"column:id;primaryKey"`
	SessionID  int64                 `gorm:"column:session_id"`
	SeatID     int64                 `gorm:"column:seat_id"`
	UserID     int64                 `gorm:"column:user_id"`
	OrderNo    string                `gorm:"column:order_no"`
	LockToken  string                `gorm:"column:lock_token"`
	Status     domain.SeatLockStatus `gorm:"column:status"`
	ExpiresAt  time.Time             `gorm:"column:expires_at"`
	ReleasedAt *time.Time            `gorm:"column:released_at"`
	CreatedAt  time.Time             `gorm:"column:created_at"`
}

func (seatLockRow) TableName() string { return "seat_locks" }

// SeatLockRepo 实现 port.SeatLockRepo。
type SeatLockRepo struct {
	db *DB
}

var _ port.SeatLockRepo = (*SeatLockRepo)(nil)

func NewSeatLockRepo(db *DB) *SeatLockRepo {
	return &SeatLockRepo{db: db}
}

// CreateLocks 依赖部分唯一索引 (session_id, seat_id)
// WHERE status IN ('LOCKED','BOOKED') 防超卖；唯一键冲突转领域错误。
func (r *SeatLockRepo) CreateLocks(ctx context.Context, locks []domain.SeatLock) error {
	if len(locks) == 0 {
		return nil
	}
	rows := make([]seatLockRow, 0, len(locks))
	for _, lock := range locks {
		rows = append(rows, seatLockRow{
			SessionID: lock.SessionID,
			SeatID:    lock.SeatID,
			UserID:    lock.UserID,
			OrderNo:   lock.OrderNo,
			LockToken: lock.LockToken,
			Status:    lock.Status,
			ExpiresAt: lock.ExpiresAt,
			CreatedAt: time.Now(),
		})
	}
	if err := r.db.db(ctx).Create(&rows).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || isUniqueViolation(err) {
			return domain.ErrSeatLockConflict
		}
		return err
	}
	return nil
}

func (r *SeatLockRepo) ReleaseExpiredBySeats(ctx context.Context, sessionID int64, seatIDs []int64) error {
	if len(seatIDs) == 0 {
		return nil
	}
	return r.db.db(ctx).Model(&seatLockRow{}).
		Where("session_id = ? AND seat_id IN ? AND status = ? AND expires_at <= ?", sessionID, seatIDs, domain.SeatLockLocked, time.Now()).
		Updates(map[string]any{"status": domain.SeatLockExpired, "released_at": time.Now()}).Error
}

func (r *SeatLockRepo) MarkBookedByOrderNo(ctx context.Context, orderNo string) error {
	return r.db.db(ctx).
		Model(&seatLockRow{}).
		Where("order_no = ? AND status = ?", orderNo, domain.SeatLockLocked).
		Update("status", domain.SeatLockBooked).Error
}

func (r *SeatLockRepo) ReleaseByOrderNo(ctx context.Context, orderNo string, status domain.SeatLockStatus) error {
	return r.db.db(ctx).
		Model(&seatLockRow{}).
		Where("order_no = ? AND status IN ?", orderNo, []domain.SeatLockStatus{domain.SeatLockLocked, domain.SeatLockBooked}).
		Updates(map[string]any{
			"status":      status,
			"released_at": time.Now(),
		}).Error
}

func (r *SeatLockRepo) ReleaseBySessionID(ctx context.Context, sessionID int64, status domain.SeatLockStatus) error {
	return r.db.db(ctx).
		Model(&seatLockRow{}).
		Where("session_id = ? AND status IN ?", sessionID, []domain.SeatLockStatus{domain.SeatLockLocked, domain.SeatLockBooked}).
		Updates(map[string]any{
			"status":      status,
			"released_at": time.Now(),
		}).Error
}

func (r *SeatLockRepo) ListActiveBySessionID(ctx context.Context, sessionID int64) ([]domain.SeatLock, error) {
	var rows []seatLockRow
	if err := r.db.db(ctx).
		Where("session_id = ? AND status IN ?", sessionID, []domain.SeatLockStatus{domain.SeatLockLocked, domain.SeatLockBooked}).
		Order("id").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	locks := make([]domain.SeatLock, 0, len(rows))
	for _, row := range rows {
		locks = append(locks, domain.SeatLock{
			ID:         row.ID,
			SessionID:  row.SessionID,
			SeatID:     row.SeatID,
			UserID:     row.UserID,
			OrderNo:    row.OrderNo,
			LockToken:  row.LockToken,
			Status:     row.Status,
			ExpiresAt:  row.ExpiresAt,
			ReleasedAt: row.ReleasedAt,
			CreatedAt:  row.CreatedAt,
		})
	}
	return locks, nil
}
