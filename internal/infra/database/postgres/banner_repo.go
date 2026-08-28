package postgres

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

type bannerRow struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Title     string    `gorm:"column:title"`
	ImageURL  string    `gorm:"column:image_url"`
	Sort      int       `gorm:"column:sort"`
	Enabled   bool      `gorm:"column:enabled"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (bannerRow) TableName() string { return "banners" }

// BannerRepo 实现 port.BannerRepo。
type BannerRepo struct {
	db *DB
}

var _ port.BannerRepo = (*BannerRepo)(nil)

func NewBannerRepo(db *DB) *BannerRepo {
	return &BannerRepo{db: db}
}

func (r *BannerRepo) Create(ctx context.Context, banner *domain.Banner) error {
	row := toBannerRow(banner)
	if err := r.db.db(ctx).Create(row).Error; err != nil {
		return err
	}
	banner.ID = row.ID
	return nil
}

func (r *BannerRepo) Update(ctx context.Context, banner *domain.Banner) error {
	res := r.db.db(ctx).
		Model(&bannerRow{}).
		Where("id = ?", banner.ID).
		Updates(map[string]any{
			"title":      banner.Title,
			"image_url":  banner.ImageURL,
			"sort":       banner.Sort,
			"enabled":    banner.Enabled,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return domain.ErrBannerNotFound
	}
	return nil
}

func (r *BannerRepo) Delete(ctx context.Context, id int64) error {
	res := r.db.db(ctx).Where("id = ?", id).Delete(&bannerRow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return domain.ErrBannerNotFound
	}
	return nil
}

func (r *BannerRepo) List(ctx context.Context) ([]domain.Banner, error) {
	var rows []bannerRow
	if err := r.db.db(ctx).Order("sort, id").Find(&rows).Error; err != nil {
		return nil, err
	}
	return toDomainBanners(rows), nil
}

func (r *BannerRepo) ListEnabled(ctx context.Context) ([]domain.Banner, error) {
	var rows []bannerRow
	if err := r.db.db(ctx).Where("enabled = ?", true).Order("sort, id").Find(&rows).Error; err != nil {
		return nil, err
	}
	return toDomainBanners(rows), nil
}

func toBannerRow(b *domain.Banner) *bannerRow {
	return &bannerRow{
		ID:        b.ID,
		Title:     b.Title,
		ImageURL:  b.ImageURL,
		Sort:      b.Sort,
		Enabled:   b.Enabled,
		CreatedAt: b.CreatedAt,
	}
}

func toDomainBanners(rows []bannerRow) []domain.Banner {
	banners := make([]domain.Banner, 0, len(rows))
	for _, row := range rows {
		banners = append(banners, domain.Banner{
			ID:        row.ID,
			Title:     row.Title,
			ImageURL:  row.ImageURL,
			Sort:      row.Sort,
			Enabled:   row.Enabled,
			CreatedAt: row.CreatedAt,
		})
	}
	return banners
}
