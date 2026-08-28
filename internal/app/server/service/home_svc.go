package service

import (
	"context"
	"sort"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

// HomeSvc 首页：banner + 热映推荐（售票数 + 评分，课程级热度分）。
type HomeSvc struct {
	banners port.BannerRepo
	movies  port.MovieRepo
	orders  port.OrderRepo
}

func NewHomeSvc(banners port.BannerRepo, movies port.MovieRepo, orders port.OrderRepo) *HomeSvc {
	return &HomeSvc{banners: banners, movies: movies, orders: orders}
}

type BannerView struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	ImageURL string `json:"image_url"`
}

type HotMovieView struct {
	MovieID  int64   `json:"movie_id"`
	Title    string  `json:"title"`
	CoverURL string  `json:"cover_url"`
	Trailer  string  `json:"trailer_url"`
	Rating   float64 `json:"rating"`
	Sold     int64   `json:"sold_count"`
}

type HomeView struct {
	Banners   []BannerView   `json:"banners"`
	HotMovies []HotMovieView `json:"hot_movies"`
}

func (s *HomeSvc) GetHome(ctx context.Context) (*HomeView, error) {
	banners, err := s.banners.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	movies, err := s.movies.List(ctx)
	if err != nil {
		return nil, err
	}

	movieIDs := make([]int64, 0, len(movies))
	for _, m := range movies {
		movieIDs = append(movieIDs, m.ID)
	}
	sold, err := s.orders.CountPaidByMovieIDs(ctx, movieIDs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	hot := make([]HotMovieView, 0, len(movies))
	type scored struct {
		view HotMovieView
		heat float64
	}
	var scoredList []scored
	for _, m := range movies {
		if m.Status != domain.MovieOnSale || m.ReleaseDate.After(now) {
			continue
		}
		heat := float64(sold[m.ID])*0.7 + m.Rating*10
		scoredList = append(scoredList, scored{
			view: HotMovieView{
				MovieID:  m.ID,
				Title:    m.Title,
				CoverURL: m.CoverURL,
				Trailer:  m.TrailerURL,
				Rating:   m.Rating,
				Sold:     sold[m.ID],
			},
			heat: heat,
		})
	}
	sort.SliceStable(scoredList, func(i, j int) bool {
		if scoredList[i].heat != scoredList[j].heat {
			return scoredList[i].heat > scoredList[j].heat
		}
		return false
	})
	for i := 0; i < len(scoredList) && i < 10; i++ {
		hot = append(hot, scoredList[i].view)
	}

	bannerViews := make([]BannerView, 0, len(banners))
	for _, b := range banners {
		bannerViews = append(bannerViews, BannerView{ID: b.ID, Title: b.Title, ImageURL: b.ImageURL})
	}
	return &HomeView{Banners: bannerViews, HotMovies: hot}, nil
}
