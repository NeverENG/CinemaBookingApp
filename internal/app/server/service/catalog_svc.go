package service

import (
	"context"
	"strings"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

type CatalogSvc struct {
	movies  port.MovieRepo
	cinemas port.CinemaRepo
	orders  port.OrderRepo
}

func NewCatalogSvc(movies port.MovieRepo, cinemas port.CinemaRepo, orders port.OrderRepo) *CatalogSvc {
	return &CatalogSvc{movies: movies, cinemas: cinemas, orders: orders}
}

type MovieView struct {
	ID              int64              `json:"id"`
	Title           string             `json:"title"`
	CoverURL        string             `json:"cover_url"`
	BackdropURL     string             `json:"backdrop_url,omitempty"`
	TrailerURL      string             `json:"trailer_url"`
	Description     string             `json:"description"`
	DurationMinutes int                `json:"duration_minutes"`
	Genre           string             `json:"genre"`
	ReleaseDate     string             `json:"release_date"`
	Rating          float64            `json:"rating"`
	SoldCount       int64              `json:"sold_count"`
	Status          domain.MovieStatus `json:"status"`
}

func (s *CatalogSvc) ListMovies(ctx context.Context, keyword, status string) ([]MovieView, error) {
	movies, err := s.movies.List(ctx)
	if err != nil {
		return nil, err
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = string(domain.MovieOnSale)
	}
	keyword = strings.ToLower(strings.TrimSpace(keyword))

	filtered := make([]domain.Movie, 0, len(movies))
	for _, movie := range movies {
		if string(movie.Status) != status {
			continue
		}
		if keyword != "" && !strings.Contains(strings.ToLower(movie.Title+" "+movie.Genre+" "+movie.Description), keyword) {
			continue
		}
		filtered = append(filtered, movie)
	}

	sold, err := s.soldCounts(ctx, filtered)
	if err != nil {
		return nil, err
	}
	views := make([]MovieView, 0, len(filtered))
	for _, movie := range filtered {
		views = append(views, toMovieView(movie, sold[movie.ID]))
	}
	return views, nil
}

func (s *CatalogSvc) GetMovie(ctx context.Context, movieID int64) (*MovieView, error) {
	movie, err := s.movies.GetByID(ctx, movieID)
	if err != nil {
		return nil, err
	}
	if movie.Status != domain.MovieOnSale {
		return nil, domain.ErrMovieNotFound
	}
	sold, err := s.soldCounts(ctx, []domain.Movie{*movie})
	if err != nil {
		return nil, err
	}
	view := toMovieView(*movie, sold[movie.ID])
	return &view, nil
}

type CinemaView struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	City      string  `json:"city"`
	Address   string  `json:"address"`
	Longitude float64 `json:"longitude,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
}

func (s *CatalogSvc) ListCinemas(ctx context.Context, keyword, city string) ([]CinemaView, error) {
	cinemas, err := s.cinemas.List(ctx, strings.TrimSpace(keyword), strings.TrimSpace(city))
	if err != nil {
		return nil, err
	}
	views := make([]CinemaView, 0, len(cinemas))
	for _, cinema := range cinemas {
		views = append(views, CinemaView{
			ID:        cinema.ID,
			Name:      cinema.Name,
			City:      cinema.City,
			Address:   cinema.Address,
			Longitude: cinema.Longitude,
			Latitude:  cinema.Latitude,
		})
	}
	return views, nil
}

func (s *CatalogSvc) soldCounts(ctx context.Context, movies []domain.Movie) (map[int64]int64, error) {
	ids := make([]int64, 0, len(movies))
	for _, movie := range movies {
		ids = append(ids, movie.ID)
	}
	return s.orders.CountPaidByMovieIDs(ctx, ids)
}

func toMovieView(movie domain.Movie, sold int64) MovieView {
	return MovieView{
		ID:              movie.ID,
		Title:           movie.Title,
		CoverURL:        movie.CoverURL,
		TrailerURL:      movie.TrailerURL,
		Description:     movie.Description,
		DurationMinutes: movie.DurationMinutes,
		Genre:           movie.Genre,
		ReleaseDate:     movie.ReleaseDate.Format("2006-01-02"),
		Rating:          movie.Rating,
		SoldCount:       sold,
		Status:          movie.Status,
	}
}
