package domain

import "time"

type SessionStatus string

const (
	SessionOpen     SessionStatus = "OPEN"
	SessionSoldOut  SessionStatus = "SOLD_OUT"
	SessionClosed   SessionStatus = "CLOSED"
	SessionCanceled SessionStatus = "CANCELED"
)

// ShowSession 场次：某影厅某影片某时间的放映。
type ShowSession struct {
	ID             int64
	CinemaID       int64
	HallID         int64
	MovieID        int64
	StartTime      time.Time
	EndTime        time.Time
	BasePriceCents int64
	Status         SessionStatus
}

// CanBook 校验：场次 OPEN 且未开场。
func (s *ShowSession) CanBook(now time.Time) bool {
	return s.Status == SessionOpen && now.Before(s.StartTime)
}
