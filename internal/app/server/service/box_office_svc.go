package service

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

// BoxOfficeSvc 票房看板查询（含影院管理员数据隔离）。
type BoxOfficeSvc struct {
	box port.BoxOfficeRepo
}

func NewBoxOfficeSvc(box port.BoxOfficeRepo) *BoxOfficeSvc {
	return &BoxOfficeSvc{box: box}
}

type DashboardQuery struct {
	StartDate   time.Time
	EndDate     time.Time
	CinemaID    int64
	MovieID     int64
	Granularity string
}

func (s *BoxOfficeSvc) Trend(ctx context.Context, scope domain.AdminScope, q DashboardQuery) ([]domain.BoxOfficeTrendRow, error) {
	if err := applyBoxScope(&q, scope); err != nil {
		return nil, err
	}
	return s.box.Trend(ctx, port.BoxOfficeFilter{
		StartDate: q.StartDate,
		EndDate:   q.EndDate,
		CinemaID:  q.CinemaID,
		MovieID:   q.MovieID,
	}, q.Granularity)
}

func (s *BoxOfficeSvc) ByMovie(ctx context.Context, scope domain.AdminScope, q DashboardQuery) ([]domain.BoxOfficeMovieRow, error) {
	if err := applyBoxScope(&q, scope); err != nil {
		return nil, err
	}
	return s.box.ByMovie(ctx, port.BoxOfficeFilter{
		StartDate: q.StartDate,
		EndDate:   q.EndDate,
		CinemaID:  q.CinemaID,
		MovieID:   q.MovieID,
	})
}

func (s *BoxOfficeSvc) ByCinema(ctx context.Context, scope domain.AdminScope, q DashboardQuery) ([]domain.BoxOfficeCinemaRow, error) {
	if err := applyBoxScope(&q, scope); err != nil {
		return nil, err
	}
	return s.box.ByCinema(ctx, port.BoxOfficeFilter{
		StartDate: q.StartDate,
		EndDate:   q.EndDate,
		CinemaID:  q.CinemaID,
	})
}

func (s *BoxOfficeSvc) Summary(ctx context.Context, scope domain.AdminScope, q DashboardQuery) (*domain.BoxOfficeSummary, error) {
	if err := applyBoxScope(&q, scope); err != nil {
		return nil, err
	}
	return s.box.Summary(ctx, port.BoxOfficeFilter{
		StartDate: q.StartDate,
		EndDate:   q.EndDate,
		CinemaID:  q.CinemaID,
		MovieID:   q.MovieID,
	})
}

func (s *BoxOfficeSvc) Reconcile(ctx context.Context) error {
	return s.box.Rebuild(ctx)
}

// applyBoxScope 影院管理员只能看自己影院。
func applyBoxScope(q *DashboardQuery, scope domain.AdminScope) error {
	if scope.IsCinemaAdmin() {
		if scope.CinemaID == nil {
			return domain.ErrForbidden
		}
		q.CinemaID = *scope.CinemaID
	}
	return nil
}
