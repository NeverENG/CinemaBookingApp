package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type boxOfficeLedgerRow struct {
	ID          int64                   `gorm:"column:id;primaryKey"`
	BizType     domain.BoxOfficeBizType `gorm:"column:biz_type"`
	BizNo       string                  `gorm:"column:biz_no"`
	StatDate    time.Time               `gorm:"column:stat_date"`
	CinemaID    int64                   `gorm:"column:cinema_id"`
	MovieID     int64                   `gorm:"column:movie_id"`
	OrderDelta  int                     `gorm:"column:order_delta"`
	TicketDelta int                     `gorm:"column:ticket_delta"`
	GrossDelta  int64                   `gorm:"column:gross_delta"`
	RefundDelta int64                   `gorm:"column:refund_delta"`
	CreatedAt   time.Time               `gorm:"column:created_at"`
}

func (boxOfficeLedgerRow) TableName() string { return "box_office_ledger" }

type dailyBoxOfficeRow struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	StatDate    time.Time `gorm:"column:stat_date"`
	CinemaID    int64     `gorm:"column:cinema_id"`
	MovieID     int64     `gorm:"column:movie_id"`
	OrderCount  int       `gorm:"column:order_count"`
	TicketCount int       `gorm:"column:ticket_count"`
	GrossCents  int64     `gorm:"column:gross_cents"`
	RefundCents int64     `gorm:"column:refund_cents"`
	NetCents    int64     `gorm:"column:net_cents"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (dailyBoxOfficeRow) TableName() string { return "daily_box_office" }

// BoxOfficeRepo 实现 port.BoxOfficeRepo。
type BoxOfficeRepo struct {
	db *DB
}

var _ port.BoxOfficeRepo = (*BoxOfficeRepo)(nil)

func NewBoxOfficeRepo(db *DB) *BoxOfficeRepo {
	return &BoxOfficeRepo{db: db}
}

// Record 写 ledger（唯一键幂等）→ upsert daily。
func (r *BoxOfficeRepo) Record(ctx context.Context, event *domain.BoxOfficeEvent) error {
	db := r.db.db(ctx)
	ledger := boxOfficeLedgerRow{
		BizType:     event.BizType,
		BizNo:       event.BizNo,
		StatDate:    event.StatDate,
		CinemaID:    event.CinemaID,
		MovieID:     event.MovieID,
		OrderDelta:  event.OrderDelta,
		TicketDelta: event.TicketDelta,
		GrossDelta:  event.GrossDelta,
		RefundDelta: event.RefundDelta,
		CreatedAt:   time.Now(),
	}
	if err := db.Create(&ledger).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil // 已记录，幂等
		}
		return err
	}

	netDelta := event.GrossDelta - event.RefundDelta
	daily := dailyBoxOfficeRow{
		StatDate:    event.StatDate,
		CinemaID:    event.CinemaID,
		MovieID:     event.MovieID,
		OrderCount:  event.OrderDelta,
		TicketCount: event.TicketDelta,
		GrossCents:  event.GrossDelta,
		RefundCents: event.RefundDelta,
		NetCents:    netDelta,
		UpdatedAt:   time.Now(),
	}
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "stat_date"}, {Name: "cinema_id"}, {Name: "movie_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"order_count":  gorm.Expr("daily_box_office.order_count + ?", event.OrderDelta),
			"ticket_count": gorm.Expr("daily_box_office.ticket_count + ?", event.TicketDelta),
			"gross_cents":  gorm.Expr("daily_box_office.gross_cents + ?", event.GrossDelta),
			"refund_cents": gorm.Expr("daily_box_office.refund_cents + ?", event.RefundDelta),
			"net_cents":    gorm.Expr("daily_box_office.net_cents + ?", netDelta),
			"updated_at":   time.Now(),
		}),
	}).Create(&daily).Error
}

func (r *BoxOfficeRepo) Trend(ctx context.Context, filter port.BoxOfficeFilter, granularity string) ([]domain.BoxOfficeTrendRow, error) {
	switch granularity {
	case "week", "month":
	default:
		granularity = "day"
	}
	query := `
		SELECT date_trunc('` + granularity + `', stat_date) AS date,
		       COALESCE(SUM(order_count), 0)  AS order_count,
		       COALESCE(SUM(ticket_count), 0) AS ticket_count,
		       COALESCE(SUM(gross_cents), 0)  AS gross_cents,
		       COALESCE(SUM(refund_cents), 0) AS refund_cents,
		       COALESCE(SUM(net_cents), 0)    AS net_cents
		FROM daily_box_office
		WHERE stat_date >= ? AND stat_date < ?
		  AND (? = 0 OR cinema_id = ?) AND (? = 0 OR movie_id = ?)
		GROUP BY 1 ORDER BY 1`
	var rows []domain.BoxOfficeTrendRow
	if err := r.db.db(ctx).Raw(query,
		filter.StartDate, filter.EndDate,
		filter.CinemaID, filter.CinemaID,
		filter.MovieID, filter.MovieID,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *BoxOfficeRepo) ByMovie(ctx context.Context, filter port.BoxOfficeFilter) ([]domain.BoxOfficeMovieRow, error) {
	query := `
		SELECT d.movie_id,
		       COALESCE(m.title, '') AS movie_title,
		       COALESCE(SUM(d.order_count), 0) AS order_count,
		       COALESCE(SUM(d.gross_cents), 0) AS gross_cents,
		       COALESCE(SUM(d.net_cents), 0)   AS net_cents
		FROM daily_box_office d
		LEFT JOIN movies m ON m.id = d.movie_id
		WHERE d.stat_date >= ? AND d.stat_date < ?
		  AND (? = 0 OR d.cinema_id = ?)
		GROUP BY d.movie_id, m.title
		ORDER BY gross_cents DESC`
	var rows []domain.BoxOfficeMovieRow
	if err := r.db.db(ctx).Raw(query,
		filter.StartDate, filter.EndDate,
		filter.CinemaID, filter.CinemaID,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *BoxOfficeRepo) ByCinema(ctx context.Context, filter port.BoxOfficeFilter) ([]domain.BoxOfficeCinemaRow, error) {
	query := `
		SELECT d.cinema_id,
		       COALESCE(c.name, '') AS cinema_name,
		       COALESCE(SUM(d.order_count), 0) AS order_count,
		       COALESCE(SUM(d.gross_cents), 0) AS gross_cents,
		       COALESCE(SUM(d.net_cents), 0)   AS net_cents
		FROM daily_box_office d
		LEFT JOIN cinemas c ON c.id = d.cinema_id
		WHERE d.stat_date >= ? AND d.stat_date < ?
		  AND (? = 0 OR d.cinema_id = ?)
		GROUP BY d.cinema_id, c.name
		ORDER BY gross_cents DESC`
	var rows []domain.BoxOfficeCinemaRow
	if err := r.db.db(ctx).Raw(query,
		filter.StartDate, filter.EndDate,
		filter.CinemaID, filter.CinemaID,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// Rebuild 对账：由 ledger 全量重建 daily。
func (r *BoxOfficeRepo) Rebuild(ctx context.Context) error {
	db := r.db.db(ctx)
	if err := db.Exec("DELETE FROM daily_box_office").Error; err != nil {
		return err
	}
	return db.Exec(`
		INSERT INTO daily_box_office
			(stat_date, cinema_id, movie_id, order_count, ticket_count, gross_cents, refund_cents, net_cents, updated_at)
		SELECT stat_date, cinema_id, movie_id,
		       SUM(order_delta), SUM(ticket_delta), SUM(gross_delta), SUM(refund_delta),
		       SUM(gross_delta) - SUM(refund_delta), now()
		FROM box_office_ledger
		GROUP BY stat_date, cinema_id, movie_id`).Error
}
