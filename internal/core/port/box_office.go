package port

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

// BoxOfficeFilter 看板查询过滤。
type BoxOfficeFilter struct {
	StartDate time.Time
	EndDate   time.Time
	CinemaID  int64 // 0 = 全部
	MovieID   int64 // 0 = 全部
}

type BoxOfficeRepo interface {
	// Record 写 ledger（唯一键幂等）并 upsert daily 聚合。
	Record(ctx context.Context, event *domain.BoxOfficeEvent) error
	// Trend 按日/周/月聚合趋势。
	Trend(ctx context.Context, filter BoxOfficeFilter, granularity string) ([]domain.BoxOfficeTrendRow, error)
	ByMovie(ctx context.Context, filter BoxOfficeFilter) ([]domain.BoxOfficeMovieRow, error)
	ByCinema(ctx context.Context, filter BoxOfficeFilter) ([]domain.BoxOfficeCinemaRow, error)
	Summary(ctx context.Context, filter BoxOfficeFilter) (*domain.BoxOfficeSummary, error)
	// Rebuild 对账：由 ledger 全量重建 daily。
	Rebuild(ctx context.Context) error
}
