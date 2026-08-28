package service

import (
	"context"
	"strconv"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

// AdminBannerSvc banner 运营位管理（图片只存 URL）。
type AdminBannerSvc struct {
	banners port.BannerRepo
	logs    port.OperationLogRepo
}

func NewAdminBannerSvc(banners port.BannerRepo, logs port.OperationLogRepo) *AdminBannerSvc {
	return &AdminBannerSvc{banners: banners, logs: logs}
}

type BannerInput struct {
	Title    string
	ImageURL string
	Sort     int
	Enabled  bool
}

func (s *AdminBannerSvc) Create(ctx context.Context, adminID int64, in BannerInput) (*domain.Banner, error) {
	banner := &domain.Banner{
		Title:     in.Title,
		ImageURL:  in.ImageURL,
		Sort:      in.Sort,
		Enabled:   in.Enabled,
		CreatedAt: time.Now(),
	}
	if err := banner.Validate(); err != nil {
		return nil, err
	}
	if err := s.banners.Create(ctx, banner); err != nil {
		return nil, err
	}
	return banner, s.log(ctx, adminID, "CREATE_BANNER", "banner", strconv.FormatInt(banner.ID, 10), banner)
}

func (s *AdminBannerSvc) Update(ctx context.Context, adminID, bannerID int64, in BannerInput) (*domain.Banner, error) {
	banner := &domain.Banner{
		ID:       bannerID,
		Title:    in.Title,
		ImageURL: in.ImageURL,
		Sort:     in.Sort,
		Enabled:  in.Enabled,
	}
	if err := banner.Validate(); err != nil {
		return nil, err
	}
	if err := s.banners.Update(ctx, banner); err != nil {
		return nil, err
	}
	return banner, s.log(ctx, adminID, "UPDATE_BANNER", "banner", strconv.FormatInt(bannerID, 10), banner)
}

func (s *AdminBannerSvc) Delete(ctx context.Context, adminID, bannerID int64) error {
	if err := s.banners.Delete(ctx, bannerID); err != nil {
		return err
	}
	return s.log(ctx, adminID, "DELETE_BANNER", "banner", strconv.FormatInt(bannerID, 10), nil)
}

func (s *AdminBannerSvc) List(ctx context.Context) ([]domain.Banner, error) {
	return s.banners.List(ctx)
}

func (s *AdminBannerSvc) log(ctx context.Context, adminID int64, action, targetType, targetID string, detail any) error {
	return s.logs.Create(ctx, &domain.OperationLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detail,
	})
}
