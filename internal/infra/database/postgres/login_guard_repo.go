package postgres

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
)

type loginGuardRow struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	Scope       string     `gorm:"column:scope"`
	Username    string     `gorm:"column:username"`
	FailedCount int        `gorm:"column:failed_count"`
	LockedUntil *time.Time `gorm:"column:locked_until"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (loginGuardRow) TableName() string { return "login_guards" }

// LoginGuardRepo 实现 port.LoginGuardRepo。
type LoginGuardRepo struct {
	db *DB
}

var _ port.LoginGuardRepo = (*LoginGuardRepo)(nil)

func NewLoginGuardRepo(db *DB) *LoginGuardRepo {
	return &LoginGuardRepo{db: db}
}

func (r *LoginGuardRepo) Get(ctx context.Context, scope, username string) (*domain.LoginGuard, error) {
	var row loginGuardRow
	err := r.db.db(ctx).Where("scope = ? AND username = ?", scope, username).Limit(1).Find(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	return toDomainLoginGuard(row), nil
}

func (r *LoginGuardRepo) RecordFailure(ctx context.Context, scope, username string) (int, error) {
	var count int
	err := r.db.db(ctx).Raw(`
		WITH upsert AS (
			INSERT INTO login_guards (scope, username, failed_count)
			VALUES (?, ?, 1)
			ON CONFLICT (scope, username)
			DO UPDATE SET failed_count = login_guards.failed_count + 1, updated_at = now()
			RETURNING failed_count
		)
		SELECT failed_count FROM upsert
	`, scope, username).Scan(&count).Error
	return count, err
}

func (r *LoginGuardRepo) Lock(ctx context.Context, scope, username string, until time.Time) error {
	return r.db.db(ctx).Exec(`
		INSERT INTO login_guards (scope, username, failed_count, locked_until)
		VALUES (?, ?, 5, ?)
		ON CONFLICT (scope, username)
		DO UPDATE SET locked_until = EXCLUDED.locked_until, updated_at = now()
	`, scope, username, until).Error
}

func (r *LoginGuardRepo) Reset(ctx context.Context, scope, username string) error {
	return r.db.db(ctx).Model(&loginGuardRow{}).
		Where("scope = ? AND username = ?", scope, username).
		Updates(map[string]any{
			"failed_count": 0,
			"locked_until": nil,
			"updated_at":   time.Now(),
		}).Error
}

func toDomainLoginGuard(row loginGuardRow) *domain.LoginGuard {
	return &domain.LoginGuard{
		Scope:       row.Scope,
		Username:    row.Username,
		FailedCount: row.FailedCount,
		LockedUntil: row.LockedUntil,
	}
}
