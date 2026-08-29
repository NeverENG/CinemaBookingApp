package service

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

// SeatMapSvc 用户侧查询：场次列表 + 座位图实时占用。
type SeatMapSvc struct {
	sessions port.SessionRepo
	seats    port.SeatRepo
	locks    port.SeatLockRepo
	movies   port.MovieRepo
	halls    port.HallRepo
	cinemas  port.CinemaRepo
}

func NewSeatMapSvc(
	sessions port.SessionRepo,
	seats port.SeatRepo,
	locks port.SeatLockRepo,
	movies port.MovieRepo,
	halls port.HallRepo,
	cinemas ...port.CinemaRepo,
) *SeatMapSvc {
	var cinemaRepo port.CinemaRepo
	if len(cinemas) > 0 {
		cinemaRepo = cinemas[0]
	}
	return &SeatMapSvc{sessions: sessions, seats: seats, locks: locks, movies: movies, halls: halls, cinemas: cinemaRepo}
}

type SessionView struct {
	ID             int64                `json:"id"`
	MovieID        int64                `json:"movie_id"`
	MovieTitle     string               `json:"movie_title"`
	CinemaID       int64                `json:"cinema_id"`
	CinemaName     string               `json:"cinema_name"`
	HallID         int64                `json:"hall_id"`
	HallName       string               `json:"hall_name"`
	StartTime      time.Time            `json:"start_time"`
	EndTime        time.Time            `json:"end_time"`
	BasePriceCents int64                `json:"base_price_cents"`
	Status         domain.SessionStatus `json:"status"`
	RemainingSeats int                  `json:"remaining_seats"`
}

type SeatView struct {
	SeatID int64  `json:"seat_id"`
	RowNo  int    `json:"row_no"`
	ColNo  int    `json:"col_no"`
	SeatNo string `json:"seat_no"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type SeatMapView struct {
	Session    SessionView `json:"session"`
	Seats      []SeatView  `json:"seats"`
	ServerTime time.Time   `json:"server_time"`
}

// GetSeatMap 座位图：disabled / booked / locked / available。
// 前端视图只是乐观展示，最终并发以锁座事务为准。
func (s *SeatMapSvc) GetSeatMap(ctx context.Context, sessionID int64) (*SeatMapView, error) {
	session, err := s.sessions.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	view, err := s.sessionView(ctx, session)
	if err != nil {
		return nil, err
	}

	seats, err := s.seats.ListByHallID(ctx, session.HallID)
	if err != nil {
		return nil, err
	}
	locks, err := s.locks.ListActiveBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	occupied := make(map[int64]string, len(locks))
	for _, lock := range locks {
		switch lock.Status {
		case domain.SeatLockBooked:
			occupied[lock.SeatID] = "booked"
		case domain.SeatLockLocked:
			if lock.ExpiresAt.After(now) {
				occupied[lock.SeatID] = "locked"
			}
		}
	}

	seatViews := make([]SeatView, 0, len(seats))
	remainingSeats := 0
	for _, seat := range seats {
		status := "available"
		switch {
		case seat.Status == domain.SeatDisabled:
			status = "disabled"
		case occupied[seat.ID] == "booked":
			status = "booked"
		case occupied[seat.ID] == "locked":
			status = "locked"
		}
		if status == "available" && seat.Status == domain.SeatEnabled {
			remainingSeats++
		}
		seatViews = append(seatViews, SeatView{
			SeatID: seat.ID,
			RowNo:  seat.RowNo,
			ColNo:  seat.ColNo,
			SeatNo: seat.SeatNo,
			Type:   seat.Type,
			Status: status,
		})
	}
	view.RemainingSeats = remainingSeats
	return &SeatMapView{Session: *view, Seats: seatViews, ServerTime: now}, nil
}

func (s *SeatMapSvc) ListSessions(ctx context.Context, movieID, cinemaID int64) ([]SessionView, error) {
	sessions, err := s.sessions.ListByFilter(ctx, movieID, cinemaID)
	if err != nil {
		return nil, err
	}
	views := make([]SessionView, 0, len(sessions))
	for i := range sessions {
		v, err := s.sessionView(ctx, &sessions[i])
		if err != nil {
			return nil, err
		}
		v.RemainingSeats, err = s.remainingSeats(ctx, sessions[i].ID, sessions[i].HallID)
		if err != nil {
			return nil, err
		}
		views = append(views, *v)
	}
	return views, nil
}

func (s *SeatMapSvc) sessionView(ctx context.Context, session *domain.ShowSession) (*SessionView, error) {
	movie, err := s.movies.GetByID(ctx, session.MovieID)
	if err != nil {
		return nil, err
	}
	hall, err := s.halls.GetByID(ctx, session.HallID)
	if err != nil {
		return nil, err
	}
	cinemaName := ""
	if s.cinemas != nil {
		cinema, err := s.cinemas.GetByID(ctx, session.CinemaID)
		if err != nil {
			return nil, err
		}
		cinemaName = cinema.Name
	}
	return &SessionView{
		ID:             session.ID,
		MovieID:        session.MovieID,
		MovieTitle:     movie.Title,
		CinemaID:       session.CinemaID,
		CinemaName:     cinemaName,
		HallID:         session.HallID,
		HallName:       hall.Name,
		StartTime:      session.StartTime,
		EndTime:        session.EndTime,
		BasePriceCents: session.BasePriceCents,
		Status:         session.Status,
	}, nil
}

func (s *SeatMapSvc) remainingSeats(ctx context.Context, sessionID, hallID int64) (int, error) {
	seats, err := s.seats.ListByHallID(ctx, hallID)
	if err != nil {
		return 0, err
	}
	locks, err := s.locks.ListActiveBySessionID(ctx, sessionID)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	occupied := make(map[int64]struct{}, len(locks))
	for _, lock := range locks {
		if lock.Status == domain.SeatLockBooked || (lock.Status == domain.SeatLockLocked && lock.ExpiresAt.After(now)) {
			occupied[lock.SeatID] = struct{}{}
		}
	}
	remaining := 0
	for _, seat := range seats {
		if seat.Status == domain.SeatEnabled {
			if _, ok := occupied[seat.ID]; !ok {
				remaining++
			}
		}
	}
	return remaining, nil
}
