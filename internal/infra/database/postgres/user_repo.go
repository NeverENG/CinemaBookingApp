package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type userRow struct {
	ID                   int64      `gorm:"column:id;primaryKey"`
	Username             string     `gorm:"column:username"`
	Email                string     `gorm:"column:email"`
	PasswordHash         string     `gorm:"column:password_hash"`
	Nickname             string     `gorm:"column:nickname"`
	MembershipLevelID    int64      `gorm:"column:membership_level_id"`
	PointsBalance        int        `gorm:"column:points_balance"`
	TotalEarnedPoints    int        `gorm:"column:total_earned_points"`
	TotalReclaimedPoints int        `gorm:"column:total_reclaimed_points"`
	Status               string     `gorm:"column:status"`
	EmailVerifiedAt      *time.Time `gorm:"column:email_verified_at"`
	DeletedAt            *time.Time `gorm:"column:deleted_at"`
}

func (userRow) TableName() string { return "users" }

// UserRepo 实现 port.UserRepo。
type UserRepo struct {
	db *DB
}

var _ port.UserRepo = (*UserRepo)(nil)

func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	var row userRow
	err := r.db.db(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:                row.ID,
		Username:          row.Username,
		Email:             row.Email,
		PasswordHash:      row.PasswordHash,
		Nickname:          row.Nickname,
		MembershipLevelID: row.MembershipLevelID,
		Status:            row.Status,
		EmailVerifiedAt:   row.EmailVerifiedAt,
	}, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var row userRow
	err := r.db.db(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:                row.ID,
		Username:          row.Username,
		Email:             row.Email,
		PasswordHash:      row.PasswordHash,
		Nickname:          row.Nickname,
		MembershipLevelID: row.MembershipLevelID,
		Status:            row.Status,
		EmailVerifiedAt:   row.EmailVerifiedAt,
	}, nil
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	row := &userRow{
		Username:        user.Username,
		Email:           user.Email,
		PasswordHash:    user.PasswordHash,
		Nickname:        user.Nickname,
		Status:          user.Status,
		EmailVerifiedAt: user.EmailVerifiedAt,
	}
	if err := r.db.db(ctx).Create(row).Error; err != nil {
		return err
	}
	user.ID = row.ID
	return nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	return r.db.db(ctx).
		Model(&userRow{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("password_hash", passwordHash).Error
}
