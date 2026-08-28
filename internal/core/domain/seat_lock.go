package domain

import "time"

type SeatLockStatus string

const (
	SeatLockLocked   SeatLockStatus = "LOCKED"
	SeatLockBooked   SeatLockStatus = "BOOKED"
	SeatLockReleased SeatLockStatus = "RELEASED"
	SeatLockExpired  SeatLockStatus = "EXPIRED"
)

// SeatLock 座位锁：下单时 LOCKED，支付成功 BOOKED，取消/退款 RELEASED。
type SeatLock struct {
	ID         int64
	SessionID  int64
	SeatID     int64
	UserID     int64
	OrderNo    string
	LockToken  string
	Status     SeatLockStatus
	ExpiresAt  time.Time
	ReleasedAt *time.Time
	CreatedAt  time.Time
}
