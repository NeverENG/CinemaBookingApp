package postgres

import (
	"context"
	"errors"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type adminRow struct {
	ID           int64  `gorm:"column:id;primaryKey"`
	Username     string `gorm:"column:username"`
	PasswordHash string `gorm:"column:password_hash"`
	Nickname     string `gorm:"column:nickname"`
	RoleID       int64  `gorm:"column:role_id"`
	CinemaID     *int64 `gorm:"column:cinema_id"`
	Status       string `gorm:"column:status"`
}

func (adminRow) TableName() string { return "admins" }

// AdminRepo 实现 port.AdminRepo。
type AdminRepo struct {
	db *DB
}

var _ port.AdminRepo = (*AdminRepo)(nil)

func NewAdminRepo(db *DB) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) GetByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	var row adminRow
	err := r.db.db(ctx).Where("username = ?", username).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrAdminNotFound
	}
	if err != nil {
		return nil, err
	}
	return &domain.Admin{
		ID:           row.ID,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		Nickname:     row.Nickname,
		RoleID:       row.RoleID,
		CinemaID:     row.CinemaID,
		Status:       row.Status,
	}, nil
}

func (r *AdminRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.db(ctx).Model(&adminRow{}).Count(&n).Error
	return n, err
}

func (r *AdminRepo) Create(ctx context.Context, admin *domain.Admin) error {
	return r.db.db(ctx).Create(&adminRow{
		Username:     admin.Username,
		PasswordHash: admin.PasswordHash,
		Nickname:     admin.Nickname,
		RoleID:       admin.RoleID,
		CinemaID:     admin.CinemaID,
		Status:       admin.Status,
	}).Error
}
