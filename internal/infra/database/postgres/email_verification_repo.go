package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type emailVerificationCodeRow struct {
	ID        int64      `gorm:"column:id;primaryKey"`
	Email     string     `gorm:"column:email"`
	CodeHash  string     `gorm:"column:code_hash"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (emailVerificationCodeRow) TableName() string { return "email_verification_codes" }

type EmailVerificationRepo struct {
	db *DB
}

var _ port.EmailVerificationRepo = (*EmailVerificationRepo)(nil)

func NewEmailVerificationRepo(db *DB) *EmailVerificationRepo {
	return &EmailVerificationRepo{db: db}
}

func (r *EmailVerificationRepo) Create(ctx context.Context, code *domain.EmailVerificationCode) error {
	row := &emailVerificationCodeRow{
		Email:     code.Email,
		CodeHash:  code.CodeHash,
		ExpiresAt: code.ExpiresAt,
		CreatedAt: time.Now(),
	}
	if err := r.db.db(ctx).Create(row).Error; err != nil {
		return err
	}
	code.ID = row.ID
	return nil
}

func (r *EmailVerificationRepo) FindUnusedByEmail(ctx context.Context, email string) (*domain.EmailVerificationCode, error) {
	var row emailVerificationCodeRow
	err := r.db.db(ctx).
		Where("email = ? AND used_at IS NULL AND expires_at > ?", email, time.Now()).
		Order("id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrVerificationCodeInvalid
	}
	if err != nil {
		return nil, err
	}
	return &domain.EmailVerificationCode{
		ID:        row.ID,
		Email:     row.Email,
		CodeHash:  row.CodeHash,
		ExpiresAt: row.ExpiresAt,
		UsedAt:    row.UsedAt,
	}, nil
}

func (r *EmailVerificationRepo) MarkUsed(ctx context.Context, id int64) error {
	result := r.db.db(ctx).
		Model(&emailVerificationCodeRow{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return domain.ErrVerificationCodeInvalid
	}
	return nil
}
