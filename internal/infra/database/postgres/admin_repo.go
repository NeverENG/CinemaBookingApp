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
	ID           int64      `gorm:"column:id;primaryKey"`
	Username     string     `gorm:"column:username"`
	PasswordHash string     `gorm:"column:password_hash"`
	Nickname     string     `gorm:"column:nickname"`
	RoleID       int64      `gorm:"column:role_id"`
	CinemaID     *int64     `gorm:"column:cinema_id"`
	Status       string     `gorm:"column:status"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

type adminListRow struct {
	ID         int64     `gorm:"column:id"`
	Username   string    `gorm:"column:username"`
	Nickname   string    `gorm:"column:nickname"`
	RoleID     int64     `gorm:"column:role_id"`
	RoleCode   string    `gorm:"column:role_code"`
	CinemaID   *int64    `gorm:"column:cinema_id"`
	CinemaName string    `gorm:"column:cinema_name"`
	Status     string    `gorm:"column:status"`
	CreatedAt  time.Time `gorm:"column:created_at"`
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

func (r *AdminRepo) List(ctx context.Context) ([]domain.Admin, error) {
	var rows []adminListRow
	err := r.db.db(ctx).
		Table("admins AS a").
		Select(`a.id, a.username, a.nickname, a.role_id, roles.code AS role_code,
			a.cinema_id, COALESCE(cinemas.name, '') AS cinema_name, a.status, a.created_at`).
		Joins("JOIN roles ON roles.id = a.role_id").
		Joins("LEFT JOIN cinemas ON cinemas.id = a.cinema_id").
		Where("a.deleted_at IS NULL").
		Order("a.created_at DESC, a.id DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	admins := make([]domain.Admin, 0, len(rows))
	for _, row := range rows {
		admins = append(admins, domain.Admin{
			ID:         row.ID,
			Username:   row.Username,
			Nickname:   row.Nickname,
			RoleID:     row.RoleID,
			RoleCode:   row.RoleCode,
			CinemaID:   row.CinemaID,
			CinemaName: row.CinemaName,
			Status:     row.Status,
			CreatedAt:  row.CreatedAt,
		})
	}
	return admins, nil
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
		CreatedAt:    row.CreatedAt,
	}
}
