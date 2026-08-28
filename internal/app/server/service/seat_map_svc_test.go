package service

import (
	"context"
	"testing"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

func TestGetSeatMapStatuses(t *testing.T) {
	now := time.Now()
	sessions := &fakeSessionRepo{sessions: map[int64]*domain.ShowSession{
		10: {ID: 10, CinemaID: 100, HallID: 1000, MovieID: 9, StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour), BasePriceCents: 5000, Status: domain.SessionOpen},
	}}
	movies := &fakeMovieRepo{movies: map[int64]*domain.Movie{
		9: {ID: 9, Title: "沙丘3"},
	}}
	halls := &fakeHallRepo{halls: map[int64]*domain.Hall{
		1000: {ID: 1000, CinemaID: 100, Name: "1号厅"},
	}}
	seats := &fakeSeatRepo{seats: map[int64]domain.Seat{
		1: {ID: 1, HallID: 1000, RowNo: 1, ColNo: 1, SeatNo: "A1", Type: "STANDARD", Status: domain.SeatEnabled},
		2: {ID: 2, HallID: 1000, RowNo: 1, ColNo: 2, SeatNo: "A2", Type: "STANDARD", Status: domain.SeatEnabled},
		3: {ID: 3, HallID: 1000, RowNo: 1, ColNo: 3, SeatNo: "A3", Type: "STANDARD", Status: domain.SeatDisabled},
		4: {ID: 4, HallID: 1000, RowNo: 2, ColNo: 1, SeatNo: "B1", Type: "VIP", Status: domain.SeatEnabled},
	}}
	locks := &fakeSeatLockRepo{active: []domain.SeatLock{
		{SessionID: 10, SeatID: 1, Status: domain.SeatLockLocked, ExpiresAt: now.Add(10 * time.Minute)},
		{SessionID: 10, SeatID: 2, Status: domain.SeatLockBooked},
		{SessionID: 10, SeatID: 4, Status: domain.SeatLockLocked, ExpiresAt: now.Add(-time.Minute)}, // 已过期，视为可售
	}}
	svc := NewSeatMapSvc(sessions, seats, locks, movies, halls)

	view, err := svc.GetSeatMap(context.Background(), 10)
	if err != nil {
		t.Fatalf("get seat map: %v", err)
	}
	if view.Session.MovieTitle != "沙丘3" || view.Session.HallName != "1号厅" {
		t.Fatalf("unexpected session view: %+v", view.Session)
	}
	statusBySeat := make(map[string]string, len(view.Seats))
	for _, s := range view.Seats {
		statusBySeat[s.SeatNo] = s.Status
	}
	want := map[string]string{
		"A1": "locked",
		"A2": "booked",
		"A3": "disabled",
		"B1": "available",
	}
	for seatNo, status := range want {
		if statusBySeat[seatNo] != status {
			t.Fatalf("seat %s: expected %s, got %s", seatNo, status, statusBySeat[seatNo])
		}
	}
}

func TestListSessionsEnriched(t *testing.T) {
	now := time.Now()
	sessions := &fakeSessionRepo{sessions: map[int64]*domain.ShowSession{
		10: {ID: 10, CinemaID: 100, HallID: 1000, MovieID: 9, StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour), BasePriceCents: 5000, Status: domain.SessionOpen},
		11: {ID: 11, CinemaID: 100, HallID: 1000, MovieID: 9, StartTime: now.Add(6 * time.Hour), EndTime: now.Add(8 * time.Hour), BasePriceCents: 6000, Status: domain.SessionCanceled},
	}}
	movies := &fakeMovieRepo{movies: map[int64]*domain.Movie{
		9: {ID: 9, Title: "沙丘3"},
	}}
	halls := &fakeHallRepo{halls: map[int64]*domain.Hall{
		1000: {ID: 1000, CinemaID: 100, Name: "1号厅"},
	}}
	svc := NewSeatMapSvc(sessions, &fakeSeatRepo{}, &fakeSeatLockRepo{}, movies, halls)

	views, err := svc.ListSessions(context.Background(), 9, 0)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(views) != 1 || views[0].ID != 10 {
		t.Fatalf("expected only open session, got %+v", views)
	}
}
