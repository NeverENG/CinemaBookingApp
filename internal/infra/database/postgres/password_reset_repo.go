package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type passwordResetCodeRow struct {
	ID        int64      `gorm:"column:id;primaryKey"`
	Email     string     `gorm:"column:email"`
	CodeHash  string     `gorm:"column:code_hash"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (passwordResetCodeRow) TableName() string { return "password_reset_codes" }

// PasswordResetRepo 实现 port.PasswordResetRepo。
type PasswordResetRepo struct {
	db *DB
}

var _ port.PasswordResetRepo = (*PasswordResetRepo)(nil)

func NewPasswordResetRepo(db *DB) *PasswordResetRepo {
	return &PasswordResetRepo{db: db}
}

func (r *PasswordResetRepo) Create(ctx context.Context, code *domain.PasswordResetCode) error {
	return r.db.db(ctx).Create(&passwordResetCodeRow{
		Email:     code.Email,
		CodeHash:  code.CodeHash,
		ExpiresAt: code.ExpiresAt,
		CreatedAt: time.Now(),
	}).Error
}

func (r *PasswordResetRepo) FindUnusedByEmail(ctx context.Context, email string) (*domain.PasswordResetCode, error) {
	var row passwordResetCodeRow
	err := r.db.db(ctx).
		Where("email = ? AND used_at IS NULL AND expires_at > ?", email, time.Now()).
		Order("id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrResetCodeInvalid
	}
	if err != nil {
		return nil, err
	}
	return &domain.PasswordResetCode{
		ID:        row.ID,
		Email:     row.Email,
		CodeHash:  row.CodeHash,
		ExpiresAt: row.ExpiresAt,
		UsedAt:    row.UsedAt,
	}, nil
}

func (r *PasswordResetRepo) MarkUsed(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.db(ctx).
		Model(&passwordResetCodeRow{}).
		Where("id = ?", id).
		Update("used_at", now).Error
}
