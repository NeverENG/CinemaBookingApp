package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type showSessionRow struct {
	ID             int64                `gorm:"column:id;primaryKey"`
	CinemaID       int64                `gorm:"column:cinema_id"`
	HallID         int64                `gorm:"column:hall_id"`
	MovieID        int64                `gorm:"column:movie_id"`
	StartTime      time.Time            `gorm:"column:start_time"`
	EndTime        time.Time            `gorm:"column:end_time"`
	BasePriceCents int64                `gorm:"column:base_price_cents"`
	Status         domain.SessionStatus `gorm:"column:status"`
}

func (showSessionRow) TableName() string { return "show_sessions" }

// SessionRepo 实现 port.SessionRepo。
type SessionRepo struct {
	db *DB
}

var _ port.SessionRepo = (*SessionRepo)(nil)

func NewSessionRepo(db *DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) GetSessionByID(ctx context.Context, id int64) (*domain.ShowSession, error) {
	var row showSessionRow
	err := r.db.db(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &domain.ShowSession{
		ID:             row.ID,
		CinemaID:       row.CinemaID,
		HallID:         row.HallID,
		MovieID:        row.MovieID,
		StartTime:      row.StartTime,
		EndTime:        row.EndTime,
		BasePriceCents: row.BasePriceCents,
		Status:         row.Status,
	}, nil
}

func (r *SessionRepo) Create(ctx context.Context, session *domain.ShowSession) error {
	return r.db.db(ctx).Create(&showSessionRow{
		CinemaID:       session.CinemaID,
		HallID:         session.HallID,
		MovieID:        session.MovieID,
		StartTime:      session.StartTime,
		EndTime:        session.EndTime,
		BasePriceCents: session.BasePriceCents,
		Status:         session.Status,
	}).Error
}

func (r *SessionRepo) UpdatePrice(ctx context.Context, id int64, basePriceCents int64, priceRulesJSON string) error {
	return r.db.db(ctx).
		Model(&showSessionRow{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"base_price_cents": basePriceCents,
			"price_rules_json": priceRulesJSON,
		}).Error
}

func (r *SessionRepo) Cancel(ctx context.Context, id int64) error {
	res := r.db.db(ctx).
		Model(&showSessionRow{}).
		Where("id = ? AND status IN ?", id, []domain.SessionStatus{domain.SessionOpen, domain.SessionSoldOut}).
		Update("status", domain.SessionCanceled)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (r *SessionRepo) ListOverlapping(ctx context.Context, hallID int64, start, end time.Time) ([]domain.ShowSession, error) {
	var rows []showSessionRow
	err := r.db.db(ctx).
		Where("hall_id = ? AND status IN ? AND start_time < ? AND end_time > ?",
			hallID, []domain.SessionStatus{domain.SessionOpen, domain.SessionSoldOut}, end, start).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	sessions := make([]domain.ShowSession, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, domain.ShowSession{
			ID:             row.ID,
			CinemaID:       row.CinemaID,
			HallID:         row.HallID,
			MovieID:        row.MovieID,
			StartTime:      row.StartTime,
			EndTime:        row.EndTime,
			BasePriceCents: row.BasePriceCents,
			Status:         row.Status,
		})
	}
	return sessions, nil
}

func (r *SessionRepo) ListByFilter(ctx context.Context, movieID, cinemaID int64) ([]domain.ShowSession, error) {
	q := r.db.db(ctx)
	if movieID > 0 {
		q = q.Where("movie_id = ?", movieID)
	}
	if cinemaID > 0 {
		q = q.Where("cinema_id = ?", cinemaID)
	}
	var rows []showSessionRow
	if err := q.
		Where("status IN ?", []domain.SessionStatus{domain.SessionOpen, domain.SessionSoldOut}).
		Order("start_time").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	sessions := make([]domain.ShowSession, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, domain.ShowSession{
			ID:             row.ID,
			CinemaID:       row.CinemaID,
			HallID:         row.HallID,
			MovieID:        row.MovieID,
			StartTime:      row.StartTime,
			EndTime:        row.EndTime,
			BasePriceCents: row.BasePriceCents,
			Status:         row.Status,
		})
	}
	return sessions, nil
}
