package service

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

type fakeBannerRepo struct {
	banners map[int64]*domain.Banner
}

func (f *fakeBannerRepo) Create(ctx context.Context, banner *domain.Banner) error {
	if f.banners == nil {
		f.banners = make(map[int64]*domain.Banner)
	}
	banner.ID = int64(len(f.banners) + 1)
	f.banners[banner.ID] = banner
	return nil
}

func (f *fakeBannerRepo) Update(ctx context.Context, banner *domain.Banner) error {
	if _, ok := f.banners[banner.ID]; !ok {
		return domain.ErrBannerNotFound
	}
	f.banners[banner.ID] = banner
	return nil
}

func (f *fakeBannerRepo) Delete(ctx context.Context, id int64) error {
	if _, ok := f.banners[id]; !ok {
		return domain.ErrBannerNotFound
	}
	delete(f.banners, id)
	return nil
}

func (f *fakeBannerRepo) List(ctx context.Context) ([]domain.Banner, error) {
	out := make([]domain.Banner, 0, len(f.banners))
	for _, b := range f.banners {
		out = append(out, *b)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Sort != out[j].Sort {
			return out[i].Sort < out[j].Sort
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (f *fakeBannerRepo) ListEnabled(ctx context.Context) ([]domain.Banner, error) {
	out := make([]domain.Banner, 0)
	for _, b := range f.banners {
		if b.Enabled {
			out = append(out, *b)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Sort != out[j].Sort {
			return out[i].Sort < out[j].Sort
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func TestGetHome(t *testing.T) {
	now := time.Now()
	banners := &fakeBannerRepo{banners: map[int64]*domain.Banner{
		1: {ID: 1, Title: "b1", ImageURL: "https://a/1.png", Sort: 1, Enabled: true},
		2: {ID: 2, Title: "b2", ImageURL: "https://a/2.png", Sort: 0, Enabled: true},
		3: {ID: 3, Title: "b3", ImageURL: "https://a/3.png", Sort: 5, Enabled: false},
	}}
	movies := &fakeMovieRepo{movies: map[int64]*domain.Movie{
		1: {ID: 1, Title: "热片A", CoverURL: "a.png", Rating: 8.0, Status: domain.MovieOnSale, ReleaseDate: now.Add(-10 * 24 * time.Hour)},
		2: {ID: 2, Title: "热片B", CoverURL: "b.png", Rating: 9.0, Status: domain.MovieOnSale, ReleaseDate: now.Add(-20 * 24 * time.Hour)},
		3: {ID: 3, Title: "下架片", Rating: 10, Status: domain.MovieOffSale, ReleaseDate: now.Add(-30 * 24 * time.Hour)},
		4: {ID: 4, Title: "未上映", Rating: 10, Status: domain.MovieOnSale, ReleaseDate: now.Add(24 * time.Hour)},
	}}
	orders := &fakeOrderRepo{orders: map[string]*domain.Order{
		"O1": {OrderNo: "O1", MovieID: 1, Status: domain.OrderPaid},
		"O2": {OrderNo: "O2", MovieID: 1, Status: domain.OrderPaid},
		"O3": {OrderNo: "O3", MovieID: 2, Status: domain.OrderPaid},
		"O4": {OrderNo: "O4", MovieID: 1, Status: domain.OrderPendingPayment},
	}}
	svc := NewHomeSvc(banners, movies, orders)

	view, err := svc.GetHome(context.Background())
	if err != nil {
		t.Fatalf("get home: %v", err)
	}
	if len(view.Banners) != 2 || view.Banners[0].Title != "b2" || view.Banners[1].Title != "b1" {
		t.Fatalf("unexpected banners: %+v", view.Banners)
	}
	if len(view.HotMovies) != 2 {
		t.Fatalf("expected 2 hot movies, got %d", len(view.HotMovies))
	}
	// 热度：B = 5*0.7 + 90 = 93.5 > A = 10*0.7 + 80 = 87
	if view.HotMovies[0].MovieID != 2 {
		t.Fatalf("expected movie 2 first, got %+v", view.HotMovies)
	}
	if view.HotMovies[0].Sold != 1 || view.HotMovies[1].Sold != 2 {
		t.Fatalf("unexpected sold counts: %+v", view.HotMovies)
	}
}

func TestAdminBannerCreateInvalid(t *testing.T) {
	banners := &fakeBannerRepo{}
	logs := &fakeOperationLogRepo{}
	svc := NewAdminBannerSvc(banners, logs)

	_, err := svc.Create(context.Background(), 1, BannerInput{})
	if !errors.Is(err, domain.ErrBannerInvalid) {
		t.Fatalf("expected ErrBannerInvalid, got %v", err)
	}
	if len(logs.logs) != 0 {
		t.Fatal("invalid create should not log")
	}
}
