package postgres

import (
	"context"
	"errors"
	"time"

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
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
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

func (r *AdminRepo) GetByID(ctx context.Context, id int64) (*domain.Admin, error) {
	var row adminRow
	err := r.db.db(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrAdminNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomainAdmin(row), nil
}

func (r *AdminRepo) GetByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	var row adminRow
	err := r.db.db(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrAdminNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomainAdmin(row), nil
}

func (r *AdminRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.db(ctx).Model(&adminRow{}).Where("deleted_at IS NULL").Count(&n).Error
	return n, err
}

func (r *AdminRepo) Create(ctx context.Context, admin *domain.Admin) error {
	row := &adminRow{
		Username:     admin.Username,
		PasswordHash: admin.PasswordHash,
		Nickname:     admin.Nickname,
		RoleID:       admin.RoleID,
		CinemaID:     admin.CinemaID,
		Status:       admin.Status,
	}
	if err := r.db.db(ctx).Create(row).Error; err != nil {
		return err
	}
	admin.ID = row.ID
	return nil
}

func (r *AdminRepo) UpdatePassword(ctx context.Context, adminID int64, passwordHash string) error {
	result := r.db.db(ctx).
		Model(&adminRow{}).
		Where("id = ? AND deleted_at IS NULL", adminID).
		Update("password_hash", passwordHash)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return domain.ErrAdminNotFound
	}
	return nil
}

func toDomainAdmin(row adminRow) *domain.Admin {
	return &domain.Admin{
		ID:           row.ID,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		Nickname:     row.Nickname,
		RoleID:       row.RoleID,
		CinemaID:     row.CinemaID,
		Status:       row.Status,
	}
}
