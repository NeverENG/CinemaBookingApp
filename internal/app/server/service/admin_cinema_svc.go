package service

import (
	"context"
	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"strconv"
	"strings"
)

type AdminCinemaSvc struct {
	cinemas port.CinemaRepo
	logs    port.OperationLogRepo
}

func NewAdminCinemaSvc(cinemas port.CinemaRepo, logs port.OperationLogRepo) *AdminCinemaSvc {
	return &AdminCinemaSvc{cinemas: cinemas, logs: logs}
}

type CinemaInput struct {
	Name, City, Address string
	Longitude, Latitude float64
}

func (s *AdminCinemaSvc) List(ctx context.Context, scope domain.AdminScope) ([]domain.Cinema, error) {
	if scope.Role != domain.RoleSuperAdmin {
		return nil, domain.ErrForbidden
	}
	return s.cinemas.ListAll(ctx)
}
func (s *AdminCinemaSvc) Create(ctx context.Context, scope domain.AdminScope, in CinemaInput) (*domain.Cinema, error) {
	if scope.Role != domain.RoleSuperAdmin {
		return nil, domain.ErrForbidden
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.City) == "" || strings.TrimSpace(in.Address) == "" {
		return nil, domain.ErrInvalidInput
	}
	c := &domain.Cinema{Name: strings.TrimSpace(in.Name), City: strings.TrimSpace(in.City), Address: strings.TrimSpace(in.Address), Longitude: in.Longitude, Latitude: in.Latitude, Status: domain.CinemaActive}
	if err := s.cinemas.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, s.log(ctx, scope.AdminID, "CREATE_CINEMA", "cinema", strconv.FormatInt(c.ID, 10), c)
}
func (s *AdminCinemaSvc) Update(ctx context.Context, scope domain.AdminScope, id int64, in CinemaInput) (*domain.Cinema, error) {
	if scope.Role != domain.RoleSuperAdmin {
		return nil, domain.ErrForbidden
	}
	c, e := s.cinemas.GetByID(ctx, id)
	if e != nil {
		return nil, e
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.City) == "" || strings.TrimSpace(in.Address) == "" {
		return nil, domain.ErrInvalidInput
	}
	c.Name = strings.TrimSpace(in.Name)
	c.City = strings.TrimSpace(in.City)
	c.Address = strings.TrimSpace(in.Address)
	c.Longitude = in.Longitude
	c.Latitude = in.Latitude
	if e = s.cinemas.Update(ctx, c); e != nil {
		return nil, e
	}
	return c, s.log(ctx, scope.AdminID, "UPDATE_CINEMA", "cinema", strconv.FormatInt(id, 10), c)
}
func (s *AdminCinemaSvc) SetStatus(ctx context.Context, scope domain.AdminScope, id int64, status domain.CinemaStatus) error {
	if scope.Role != domain.RoleSuperAdmin {
		return domain.ErrForbidden
	}
	if status != domain.CinemaActive && status != domain.CinemaInactive {
		return domain.ErrInvalidInput
	}
	if e := s.cinemas.SetStatus(ctx, id, status); e != nil {
		return e
	}
	return s.log(ctx, scope.AdminID, "SET_CINEMA_STATUS", "cinema", strconv.FormatInt(id, 10), map[string]any{"status": status})
}
func (s *AdminCinemaSvc) log(ctx context.Context, id int64, a, t, i string, d any) error {
	return s.logs.Create(ctx, &domain.OperationLog{AdminID: id, Action: a, TargetType: t, TargetID: i, Detail: d})
}
