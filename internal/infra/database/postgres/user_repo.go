package postgres

import (
	"context"
	"errors"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type userRow struct {
	ID                int64  `gorm:"column:id;primaryKey"`
	Username          string `gorm:"column:username"`
	PasswordHash      string `gorm:"column:password_hash"`
	Nickname          string `gorm:"column:nickname"`
	MembershipLevelID int64  `gorm:"column:membership_level_id"`
	Status            string `gorm:"column:status"`
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
	err := r.db.db(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:                row.ID,
		Username:          row.Username,
		PasswordHash:      row.PasswordHash,
		Nickname:          row.Nickname,
		MembershipLevelID: row.MembershipLevelID,
		Status:            row.Status,
	}, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var row userRow
	err := r.db.db(ctx).Where("username = ?", username).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:                row.ID,
		Username:          row.Username,
		PasswordHash:      row.PasswordHash,
		Nickname:          row.Nickname,
		MembershipLevelID: row.MembershipLevelID,
		Status:            row.Status,
	}, nil
}
