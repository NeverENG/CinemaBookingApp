package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type movieRow struct {
	ID              int64              `gorm:"column:id;primaryKey"`
	Title           string             `gorm:"column:title"`
	CoverURL        string             `gorm:"column:cover_url"`
	TrailerURL      string             `gorm:"column:trailer_url"`
	Description     string             `gorm:"column:description"`
	DurationMinutes int                `gorm:"column:duration_minutes"`
	Genre           string             `gorm:"column:genre"`
	ReleaseDate     time.Time          `gorm:"column:release_date"`
	Rating          float64            `gorm:"column:rating"`
	Status          domain.MovieStatus `gorm:"column:status"`
	CreatedAt       time.Time          `gorm:"column:created_at"`
}

func (movieRow) TableName() string { return "movies" }

// MovieRepo 实现 port.MovieRepo。
type MovieRepo struct {
	db *DB
}

var _ port.MovieRepo = (*MovieRepo)(nil)

func NewMovieRepo(db *DB) *MovieRepo {
	return &MovieRepo{db: db}
}

func (r *MovieRepo) Create(ctx context.Context, movie *domain.Movie) error {
	return r.db.db(ctx).Create(toMovieRow(movie)).Error
}

func (r *MovieRepo) GetByID(ctx context.Context, id int64) (*domain.Movie, error) {
	var row movieRow
	err := r.db.db(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrMovieNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomainMovie(row), nil
}

func (r *MovieRepo) Update(ctx context.Context, movie *domain.Movie) error {
	return r.db.db(ctx).Save(toMovieRow(movie)).Error
}

func (r *MovieRepo) List(ctx context.Context) ([]domain.Movie, error) {
	var rows []movieRow
	if err := r.db.db(ctx).Order("release_date DESC, id").Find(&rows).Error; err != nil {
		return nil, err
	}
	movies := make([]domain.Movie, 0, len(rows))
	for _, row := range rows {
		movies = append(movies, *toDomainMovie(row))
	}
	return movies, nil
}

func (r *MovieRepo) SetStatus(ctx context.Context, id int64, status domain.MovieStatus) error {
	return r.db.db(ctx).Model(&movieRow{}).Where("id = ?", id).Update("status", status).Error
}

func toMovieRow(m *domain.Movie) *movieRow {
	return &movieRow{
		ID:              m.ID,
		Title:           m.Title,
		CoverURL:        m.CoverURL,
		TrailerURL:      m.TrailerURL,
		Description:     m.Description,
		DurationMinutes: m.DurationMinutes,
		Genre:           m.Genre,
		ReleaseDate:     m.ReleaseDate,
		Rating:          m.Rating,
		Status:          m.Status,
		CreatedAt:       m.CreatedAt,
	}
}

func toDomainMovie(row movieRow) *domain.Movie {
	return &domain.Movie{
		ID:              row.ID,
		Title:           row.Title,
		CoverURL:        row.CoverURL,
		TrailerURL:      row.TrailerURL,
		Description:     row.Description,
		DurationMinutes: row.DurationMinutes,
		Genre:           row.Genre,
		ReleaseDate:     row.ReleaseDate,
		Rating:          row.Rating,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt,
	}
}
