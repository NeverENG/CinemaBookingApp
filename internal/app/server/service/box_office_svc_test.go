package service

import (
	"context"
	"testing"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

type fakeBoxOfficeRepo struct {
	events          []*domain.BoxOfficeEvent
	lastFilter      port.BoxOfficeFilter
	lastGranularity string
}

func (f *fakeBoxOfficeRepo) Record(ctx context.Context, event *domain.BoxOfficeEvent) error {
	f.events = append(f.events, event)
	return nil
}

func (f *fakeBoxOfficeRepo) Trend(ctx context.Context, filter port.BoxOfficeFilter, granularity string) ([]domain.BoxOfficeTrendRow, error) {
	f.lastFilter = filter
	f.lastGranularity = granularity
	return nil, nil
}

func (f *fakeBoxOfficeRepo) ByMovie(ctx context.Context, filter port.BoxOfficeFilter) ([]domain.BoxOfficeMovieRow, error) {
	f.lastFilter = filter
	return nil, nil
}

func (f *fakeBoxOfficeRepo) ByCinema(ctx context.Context, filter port.BoxOfficeFilter) ([]domain.BoxOfficeCinemaRow, error) {
	f.lastFilter = filter
	return nil, nil
}

func (f *fakeBoxOfficeRepo) Summary(ctx context.Context, filter port.BoxOfficeFilter) (*domain.BoxOfficeSummary, error) {
	f.lastFilter = filter
	return &domain.BoxOfficeSummary{OrderCount: 1}, nil
}

func (f *fakeBoxOfficeRepo) Rebuild(ctx context.Context) error {
	return nil
}

func TestDashboardCinemaAdminScope(t *testing.T) {
	box := &fakeBoxOfficeRepo{}
	svc := NewBoxOfficeSvc(box)
	scope := domain.AdminScope{AdminID: 2, Role: domain.RoleCinemaAdmin, CinemaID: int64Ptr(5)}

	_, err := svc.Trend(context.Background(), scope, DashboardQuery{Granularity: "week"})
	if err != nil {
		t.Fatalf("trend: %v", err)
	}
	if box.lastFilter.CinemaID != 5 {
		t.Fatalf("expected cinema scope 5, got %d", box.lastFilter.CinemaID)
	}
	if box.lastGranularity != "week" {
		t.Fatalf("expected granularity week, got %s", box.lastGranularity)
	}
}

func TestDashboardSuperAdminKeepsFilter(t *testing.T) {
	box := &fakeBoxOfficeRepo{}
	svc := NewBoxOfficeSvc(box)
	scope := domain.AdminScope{AdminID: 1, Role: domain.RoleSuperAdmin}

	_, err := svc.ByCinema(context.Background(), scope, DashboardQuery{CinemaID: 7})
	if err != nil {
		t.Fatalf("by cinema: %v", err)
	}
	if box.lastFilter.CinemaID != 7 {
		t.Fatalf("expected cinema 7 kept, got %d", box.lastFilter.CinemaID)
	}
}
