package postgres

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm/clause"
)

type membershipLevelRow struct {
	ID             int64  `gorm:"column:id;primaryKey"`
	LevelCode      string `gorm:"column:level_code"`
	Name           string `gorm:"column:name"`
	MinTotalPoints int    `gorm:"column:min_total_points"`
	DiscountBp     int    `gorm:"column:discount_bp"`
	Status         string `gorm:"column:status"`
}

func (membershipLevelRow) TableName() string { return "membership_levels" }

type membershipLevelLogRow struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	UserID      int64     `gorm:"column:user_id"`
	FromLevelID *int64    `gorm:"column:from_level_id"`
	ToLevelID   int64     `gorm:"column:to_level_id"`
	ChangeType  string    `gorm:"column:change_type"`
	Reason      string    `gorm:"column:reason"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (membershipLevelLogRow) TableName() string { return "membership_level_logs" }

// MembershipRepo 实现 port.MembershipRepo。
type MembershipRepo struct {
	db *DB
}

var _ port.MembershipRepo = (*MembershipRepo)(nil)

func NewMembershipRepo(db *DB) *MembershipRepo {
	return &MembershipRepo{db: db}
}

// UpgradeIfNeeded 按累计赠送积分取最高可升级别；变化则更新用户并写日志。
func (r *MembershipRepo) UpgradeIfNeeded(ctx context.Context, userID int64) (bool, error) {
	db := r.db.db(ctx)
	var user userRow
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		return false, err
	}

	var level membershipLevelRow
	if err := db.
		Where("min_total_points <= ? AND status = ?", user.TotalEarnedPoints, "ACTIVE").
		Order("min_total_points DESC").
		First(&level).Error; err != nil {
		return false, err
	}
	if user.MembershipLevelID == level.ID {
		return false, nil
	}

	fromID := user.MembershipLevelID
	if err := db.Model(&userRow{}).
		Where("id = ?", userID).
		Update("membership_level_id", level.ID).Error; err != nil {
		return false, err
	}
	var fromPtr *int64
	if fromID > 0 {
		v := fromID
		fromPtr = &v
	}
	if err := db.Create(&membershipLevelLogRow{
		UserID:      userID,
		FromLevelID: fromPtr,
		ToLevelID:   level.ID,
		ChangeType:  "UPGRADE",
		Reason:      "total_earned_points",
		CreatedAt:   time.Now(),
	}).Error; err != nil {
		return false, err
	}
	return true, nil
}
