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
		letter := rowLabel(row)
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

func rowLabel(row int) string {
	label := ""
	for row > 0 {
		row--
		label = string(rune('A'+row%26)) + label
		row /= 26
	}
	return label
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

func (s *AdminHallSvc) Create(ctx context.Context, scope domain.AdminScope, in HallInput) (*domain.Hall, error) {
	if !scope.CanManageCinema(in.CinemaID) {
		return nil, domain.ErrForbidden
	}
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
	return hall, s.log(ctx, scope.AdminID, "CREATE_HALL", "hall", strconv.FormatInt(hall.ID, 10), hall)
}

func (s *AdminHallSvc) Update(ctx context.Context, scope domain.AdminScope, hallID int64, in HallInput) (*domain.Hall, error) {
	layout, err := parseSeatLayout(in.SeatLayout)
	if err != nil {
		return nil, err
	}
	hall, err := s.halls.GetByID(ctx, hallID)
	if err != nil {
		return nil, err
	}
	if hall.CinemaID != in.CinemaID || !scope.CanManageCinema(hall.CinemaID) {
		return nil, domain.ErrForbidden
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
	return hall, s.log(ctx, scope.AdminID, "UPDATE_HALL", "hall", strconv.FormatInt(hall.ID, 10), hall)
}

func (s *AdminHallSvc) ListByCinema(ctx context.Context, scope domain.AdminScope, cinemaID int64) ([]domain.Hall, error) {
	if scope.Role != domain.RoleSuperAdmin && !scope.CanManageCinema(cinemaID) {
		return nil, domain.ErrForbidden
	}
	if scope.IsCinemaAdmin() {
		cinemaID = *scope.CinemaID
	}
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
