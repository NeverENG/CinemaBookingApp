package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

// seatLayout 影厅布局：行列 + 座位类型覆盖 + 禁用座位。
type seatLayout struct {
	Rows      int               `json:"rows"`
	Cols      int               `json:"cols"`
	SeatTypes map[string]string `json:"seat_types,omitempty"`
	Disabled  []string          `json:"disabled,omitempty"`
}

func parseSeatLayout(raw string) (*seatLayout, error) {
	var l seatLayout
	if err := json.Unmarshal([]byte(raw), &l); err != nil {
		return nil, domain.ErrSeatLayoutInvalid
	}
	if l.Rows <= 0 || l.Cols <= 0 || l.Rows > 50 || l.Cols > 50 {
		return nil, domain.ErrSeatLayoutInvalid
	}
	return &l, nil
}

func buildSeats(l *seatLayout) []domain.Seat {
	disabled := make(map[string]bool, len(l.Disabled))
	for _, no := range l.Disabled {
		disabled[no] = true
	}
	seats := make([]domain.Seat, 0, l.Rows*l.Cols)
	for row := 1; row <= l.Rows; row++ {
		letter := string(rune('A' + row - 1))
		for col := 1; col <= l.Cols; col++ {
			no := fmt.Sprintf("%s%d", letter, col)
			typ := "STANDARD"
			if t, ok := l.SeatTypes[no]; ok {
				typ = t
			}
			status := domain.SeatEnabled
			if disabled[no] {
				status = domain.SeatDisabled
			}
			seats = append(seats, domain.Seat{
				RowNo:  row,
				ColNo:  col,
				SeatNo: no,
				Type:   typ,
				Status: status,
			})
		}
	}
	return seats
}

// AdminHallSvc 影厅管理：保存布局时 diff 同步 seats。
type AdminHallSvc struct {
	halls port.HallRepo
	seats port.SeatRepo
	logs  port.OperationLogRepo
}

func NewAdminHallSvc(halls port.HallRepo, seats port.SeatRepo, logs port.OperationLogRepo) *AdminHallSvc {
	return &AdminHallSvc{halls: halls, seats: seats, logs: logs}
}

type HallInput struct {
	CinemaID   int64
	Name       string
	SeatLayout string
}

func (s *AdminHallSvc) Create(ctx context.Context, adminID int64, in HallInput) (*domain.Hall, error) {
	layout, err := parseSeatLayout(in.SeatLayout)
	if err != nil {
		return nil, err
	}
	hall := &domain.Hall{
		CinemaID:       in.CinemaID,
		Name:           in.Name,
		SeatLayoutJSON: in.SeatLayout,
		Status:         domain.HallActive,
	}
	if err := hall.Validate(); err != nil {
		return nil, err
	}
	if err := s.halls.Create(ctx, hall); err != nil {
		return nil, err
	}
	if err := s.seats.SyncSeats(ctx, hall.ID, buildSeats(layout)); err != nil {
		return nil, err
	}
	return hall, s.log(ctx, adminID, "CREATE_HALL", "hall", strconv.FormatInt(hall.ID, 10), hall)
}

func (s *AdminHallSvc) Update(ctx context.Context, adminID, hallID int64, in HallInput) (*domain.Hall, error) {
	layout, err := parseSeatLayout(in.SeatLayout)
	if err != nil {
		return nil, err
	}
	hall, err := s.halls.GetByID(ctx, hallID)
	if err != nil {
		return nil, err
	}
	hall.Name = in.Name
	hall.SeatLayoutJSON = in.SeatLayout
	if err := hall.Validate(); err != nil {
		return nil, err
	}
	if err := s.halls.Update(ctx, hall); err != nil {
		return nil, err
	}
	if err := s.seats.SyncSeats(ctx, hall.ID, buildSeats(layout)); err != nil {
		return nil, err
	}
	return hall, s.log(ctx, adminID, "UPDATE_HALL", "hall", strconv.FormatInt(hall.ID, 10), hall)
}

func (s *AdminHallSvc) ListByCinema(ctx context.Context, cinemaID int64) ([]domain.Hall, error) {
	return s.halls.ListByCinema(ctx, cinemaID)
}

func (s *AdminHallSvc) log(ctx context.Context, adminID int64, action, targetType, targetID string, detail any) error {
	return s.logs.Create(ctx, &domain.OperationLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detail,
	})
}
