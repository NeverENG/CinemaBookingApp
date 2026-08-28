package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pointsLedgerRow struct {
	ID           int64                `gorm:"column:id;primaryKey"`
	UserID       int64                `gorm:"column:user_id"`
	ChangePoints int                  `gorm:"column:change_points"`
	BalanceAfter int                  `gorm:"column:balance_after"`
	BizType      domain.PointsBizType `gorm:"column:biz_type"`
	BizNo        string               `gorm:"column:biz_no"`
	CreatedAt    time.Time            `gorm:"column:created_at"`
}

func (pointsLedgerRow) TableName() string { return "points_ledger" }

// PointsRepo 实现 port.PointsRepo。
type PointsRepo struct {
	db *DB
}

var _ port.PointsRepo = (*PointsRepo)(nil)

func NewPointsRepo(db *DB) *PointsRepo {
	return &PointsRepo{db: db}
}

func (r *PointsRepo) GrantOnPaid(ctx context.Context, userID int64, paidCents int64, orderNo string) error {
	change := int(paidCents / 100 * domain.PointsPerYuan)
	if change <= 0 {
		return nil
	}
	return r.apply(ctx, userID, change, domain.PointsOrderPaid, orderNo, true)
}

func (r *PointsRepo) ReclaimOnRefund(ctx context.Context, userID int64, refundCents int64, refundNo string) error {
	change := -(int(refundCents / 100 * domain.PointsPerYuan))
	if change == 0 {
		return nil
	}
	return r.apply(ctx, userID, change, domain.PointsOrderRefund, refundNo, false)
}

// Exchange 兑换扣积分：校验余额 → 写负流水（幂等）→ 更新余额，返回兑换后余额。
func (r *PointsRepo) Exchange(ctx context.Context, userID int64, points int, exchangeNo string) (int, error) {
	if points <= 0 {
		return 0, domain.ErrInsufficientPoints
	}
	db := r.db.db(ctx)
	var user userRow
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		return 0, err
	}
	if user.PointsBalance < points {
		return 0, domain.ErrInsufficientPoints
	}
	balance := user.PointsBalance - points

	row := pointsLedgerRow{
		UserID:       userID,
		ChangePoints: -points,
		BalanceAfter: balance,
		BizType:      domain.PointsExchange,
		BizNo:        exchangeNo,
		CreatedAt:    time.Now(),
	}
	if err := db.Create(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return user.PointsBalance, nil // 重复兑换请求：已扣过，幂等成功
		}
		return 0, err
	}
	if err := db.Model(&userRow{}).
		Where("id = ?", userID).
		Update("points_balance", balance).Error; err != nil {
		return 0, err
	}
	return balance, nil
}

// apply 行锁用户 → 计算余额 → 写流水（唯一键幂等）→ 更新用户冗余快照。
func (r *PointsRepo) apply(ctx context.Context, userID int64, change int, bizType domain.PointsBizType, bizNo string, earned bool) error {
	db := r.db.db(ctx)
	var user userRow
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		return err
	}

	actual := change
	if user.PointsBalance+actual < 0 {
		actual = -user.PointsBalance
	}
	if actual == 0 {
		return nil
	}
	balance := user.PointsBalance + actual

	row := pointsLedgerRow{
		UserID:       userID,
		ChangePoints: actual,
		BalanceAfter: balance,
		BizType:      bizType,
		BizNo:        bizNo,
		CreatedAt:    time.Now(),
	}
	if err := db.Create(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil // 已记账，幂等成功
		}
		return err
	}

	updates := map[string]any{"points_balance": balance}
	if earned {
		updates["total_earned_points"] = gorm.Expr("total_earned_points + ?", actual)
	} else {
		updates["total_reclaimed_points"] = gorm.Expr("total_reclaimed_points + ?", -actual)
	}
	return db.Model(&userRow{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *PointsRepo) GetBalance(ctx context.Context, userID int64) (int, error) {
	var user userRow
	if err := r.db.db(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return 0, err
	}
	return user.PointsBalance, nil
}

func (r *PointsRepo) GetRecentLedger(ctx context.Context, userID int64, limit int) ([]domain.PointsLedger, error) {
	var rows []pointsLedgerRow
	if err := r.db.db(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	ledger := make([]domain.PointsLedger, 0, len(rows))
	for _, row := range rows {
		ledger = append(ledger, domain.PointsLedger{
			ID:           row.ID,
			UserID:       row.UserID,
			ChangePoints: row.ChangePoints,
			BalanceAfter: row.BalanceAfter,
			BizType:      row.BizType,
			BizNo:        row.BizNo,
			CreatedAt:    row.CreatedAt,
		})
	}
	return ledger, nil
}
