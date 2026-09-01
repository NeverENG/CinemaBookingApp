package service

import (
	"context"
	"strconv"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

// AdminMovieSvc 影片管理用例（含操作审计）。
type AdminMovieSvc struct {
	movies port.MovieRepo
	logs   port.OperationLogRepo
}

func NewAdminMovieSvc(movies port.MovieRepo, logs port.OperationLogRepo) *AdminMovieSvc {
	return &AdminMovieSvc{movies: movies, logs: logs}
}

type MovieInput struct {
	Title           string
	CoverURL        string
	TrailerURL      string
	Description     string
	DurationMinutes int
	Genre           string
	ReleaseDate     time.Time
	Rating          float64
}

func (s *AdminMovieSvc) Create(ctx context.Context, scope domain.AdminScope, in MovieInput) (*domain.Movie, error) {
	if scope.Role != domain.RoleSuperAdmin {
		return nil, domain.ErrForbidden
	}
	movie := &domain.Movie{
		Title:           in.Title,
		CoverURL:        in.CoverURL,
		TrailerURL:      in.TrailerURL,
		Description:     in.Description,
		DurationMinutes: in.DurationMinutes,
		Genre:           in.Genre,
		ReleaseDate:     in.ReleaseDate,
		Rating:          in.Rating,
		Status:          domain.MovieOnSale,
		CreatedAt:       time.Now(),
	}
	if err := movie.Validate(); err != nil {
		return nil, err
	}
	if err := s.movies.Create(ctx, movie); err != nil {
		return nil, err
	}
	return movie, s.log(ctx, scope.AdminID, "CREATE_MOVIE", "movie", strconv.FormatInt(movie.ID, 10), movie)
}

func (s *AdminMovieSvc) Update(ctx context.Context, scope domain.AdminScope, movieID int64, in MovieInput) (*domain.Movie, error) {
	if scope.Role != domain.RoleSuperAdmin {
		return nil, domain.ErrForbidden
	}
	movie, err := s.movies.GetByID(ctx, movieID)
	if err != nil {
		return nil, err
	}
	movie.Title = in.Title
	movie.CoverURL = in.CoverURL
	movie.TrailerURL = in.TrailerURL
	movie.Description = in.Description
	movie.DurationMinutes = in.DurationMinutes
	movie.Genre = in.Genre
	movie.ReleaseDate = in.ReleaseDate
	movie.Rating = in.Rating
	if err := movie.Validate(); err != nil {
		return nil, err
	}
	if err := s.movies.Update(ctx, movie); err != nil {
		return nil, err
	}
	return movie, s.log(ctx, scope.AdminID, "UPDATE_MOVIE", "movie", strconv.FormatInt(movie.ID, 10), movie)
}

func (s *AdminMovieSvc) SetStatus(ctx context.Context, scope domain.AdminScope, movieID int64, status domain.MovieStatus) error {
	if scope.Role != domain.RoleSuperAdmin {
		return domain.ErrForbidden
	}
	if _, err := s.movies.GetByID(ctx, movieID); err != nil {
		return err
	}
	if err := s.movies.SetStatus(ctx, movieID, status); err != nil {
		return err
	}
	return s.log(ctx, scope.AdminID, "SET_MOVIE_STATUS", "movie", strconv.FormatInt(movieID, 10), map[string]string{"status": string(status)})
}

func (s *AdminMovieSvc) List(ctx context.Context) ([]domain.Movie, error) {
	return s.movies.List(ctx)
}

func (s *AdminMovieSvc) log(ctx context.Context, adminID int64, action, targetType, targetID string, detail any) error {
	return s.logs.Create(ctx, &domain.OperationLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detail,
	})
}
