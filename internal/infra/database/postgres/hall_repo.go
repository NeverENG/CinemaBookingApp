package postgres

import (
	"context"
	"errors"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type hallRow struct {
	ID             int64             `gorm:"column:id;primaryKey"`
	CinemaID       int64             `gorm:"column:cinema_id"`
	Name           string            `gorm:"column:name"`
	SeatLayoutJSON string            `gorm:"column:seat_layout_json"`
	Status         domain.HallStatus `gorm:"column:status"`
}

func (hallRow) TableName() string { return "halls" }

// HallRepo 实现 port.HallRepo。
type HallRepo struct {
	db *DB
}

var _ port.HallRepo = (*HallRepo)(nil)

func NewHallRepo(db *DB) *HallRepo {
	return &HallRepo{db: db}
}

func (r *HallRepo) Create(ctx context.Context, hall *domain.Hall) error {
	return r.db.db(ctx).Create(toHallRow(hall)).Error
}

func (r *HallRepo) GetByID(ctx context.Context, id int64) (*domain.Hall, error) {
	var row hallRow
	err := r.db.db(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrHallNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomainHall(row), nil
}

func (r *HallRepo) Update(ctx context.Context, hall *domain.Hall) error {
	return r.db.db(ctx).Save(toHallRow(hall)).Error
}

func (r *HallRepo) ListByCinema(ctx context.Context, cinemaID int64) ([]domain.Hall, error) {
	var rows []hallRow
	if err := r.db.db(ctx).Where("cinema_id = ?", cinemaID).Order("id").Find(&rows).Error; err != nil {
		return nil, err
	}
	halls := make([]domain.Hall, 0, len(rows))
	for _, row := range rows {
		halls = append(halls, *toDomainHall(row))
	}
	return halls, nil
}

func toHallRow(h *domain.Hall) *hallRow {
	return &hallRow{
		ID:             h.ID,
		CinemaID:       h.CinemaID,
		Name:           h.Name,
		SeatLayoutJSON: h.SeatLayoutJSON,
		Status:         h.Status,
	}
}

func toDomainHall(row hallRow) *domain.Hall {
	return &domain.Hall{
		ID:             row.ID,
		CinemaID:       row.CinemaID,
		Name:           row.Name,
		SeatLayoutJSON: row.SeatLayoutJSON,
		Status:         row.Status,
	}
}
