package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type cinemaRow struct {
	ID        int64               `gorm:"column:id;primaryKey"`
	Name      string              `gorm:"column:name"`
	City      string              `gorm:"column:city"`
	Address   string              `gorm:"column:address"`
	Longitude float64             `gorm:"column:longitude"`
	Latitude  float64             `gorm:"column:latitude"`
	Status    domain.CinemaStatus `gorm:"column:status"`
	CreatedAt time.Time           `gorm:"column:created_at"`
}

func (cinemaRow) TableName() string { return "cinemas" }

type CinemaRepo struct {
	db *DB
}

var _ port.CinemaRepo = (*CinemaRepo)(nil)

func NewCinemaRepo(db *DB) *CinemaRepo {
	return &CinemaRepo{db: db}
}

func (r *CinemaRepo) GetByID(ctx context.Context, id int64) (*domain.Cinema, error) {
	var row cinemaRow
	err := r.db.db(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCinemaNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomainCinema(row), nil
}

func (r *CinemaRepo) List(ctx context.Context, keyword, city string) ([]domain.Cinema, error) {
	q := r.db.db(ctx).Where("status = ?", domain.CinemaActive)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name ILIKE ? OR city ILIKE ? OR address ILIKE ?", like, like, like)
	}
	if city != "" {
		q = q.Where("city = ?", city)
	}

	var rows []cinemaRow
	if err := q.Order("city, name, id").Find(&rows).Error; err != nil {
		return nil, err
	}
	cinemas := make([]domain.Cinema, 0, len(rows))
	for _, row := range rows {
		cinemas = append(cinemas, *toDomainCinema(row))
	}
	return cinemas, nil
}

func (r *CinemaRepo) ListAll(ctx context.Context) ([]domain.Cinema, error) {
	var rows []cinemaRow
	if err := r.db.db(ctx).Order("city, name, id").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Cinema, 0, len(rows))
	for _, row := range rows {
		result = append(result, *toDomainCinema(row))
	}
	return result, nil
}

func (r *CinemaRepo) Create(ctx context.Context, cinema *domain.Cinema) error {
	row := &cinemaRow{Name: cinema.Name, City: cinema.City, Address: cinema.Address, Longitude: cinema.Longitude, Latitude: cinema.Latitude, Status: cinema.Status}
	if err := r.db.db(ctx).Create(row).Error; err != nil {
		return err
	}
	cinema.ID, cinema.CreatedAt = row.ID, row.CreatedAt
	return nil
}

func (r *CinemaRepo) Update(ctx context.Context, cinema *domain.Cinema) error {
	result := r.db.db(ctx).Model(&cinemaRow{}).Where("id = ?", cinema.ID).Updates(map[string]any{"name": cinema.Name, "city": cinema.City, "address": cinema.Address, "longitude": cinema.Longitude, "latitude": cinema.Latitude})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return domain.ErrCinemaNotFound
	}
	return nil
}

func (r *CinemaRepo) SetStatus(ctx context.Context, id int64, status domain.CinemaStatus) error {
	result := r.db.db(ctx).Model(&cinemaRow{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return domain.ErrCinemaNotFound
	}
	return nil
}

func toDomainCinema(row cinemaRow) *domain.Cinema {
	return &domain.Cinema{
		ID:        row.ID,
		Name:      row.Name,
		City:      row.City,
		Address:   row.Address,
		Longitude: row.Longitude,
		Latitude:  row.Latitude,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
	}
}
