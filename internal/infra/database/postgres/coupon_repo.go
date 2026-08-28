package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"gorm.io/gorm"
)

type userCouponRow struct {
	ID         int64                   `gorm:"column:id;primaryKey"`
	CouponNo   string                  `gorm:"column:coupon_no"`
	UserID     int64                   `gorm:"column:user_id"`
	TemplateID int64                   `gorm:"column:template_id"`
	Status     domain.UserCouponStatus `gorm:"column:status"`
	OrderNo    string                  `gorm:"column:order_no"`
	ExpireAt   time.Time               `gorm:"column:expire_at"`
}

func (userCouponRow) TableName() string { return "user_coupons" }

type couponTemplateRow struct {
	ID               int64  `gorm:"column:id;primaryKey"`
	Name             string `gorm:"column:name"`
	Type             string `gorm:"column:type"`
	ValueCents       int64  `gorm:"column:value_cents"`
	PercentBp        int    `gorm:"column:percent_bp"`
	MinSpendCents    int64  `gorm:"column:min_spend_cents"`
	MaxDiscountCents int64  `gorm:"column:max_discount_cents"`
	Redeemable       bool   `gorm:"column:redeemable"`
	RedeemPoints     int    `gorm:"column:redeem_points"`
	ValidDays        int    `gorm:"column:valid_days"`
	TotalQty         int    `gorm:"column:total_qty"`
	PerUserLimit     int    `gorm:"column:per_user_limit"`
	Status           string `gorm:"column:status"`
}

func (couponTemplateRow) TableName() string { return "coupon_templates" }

// UserCouponRepo 实现 port.UserCouponRepo。
type UserCouponRepo struct {
	db *DB
}

var _ port.UserCouponRepo = (*UserCouponRepo)(nil)

func NewUserCouponRepo(db *DB) *UserCouponRepo {
	return &UserCouponRepo{db: db}
}

func (r *UserCouponRepo) GetByCouponNo(ctx context.Context, couponNo string) (*domain.UserCoupon, error) {
	var row userCouponRow
	err := r.db.db(ctx).Where("coupon_no = ?", couponNo).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCouponNotAvailable
	}
	if err != nil {
		return nil, err
	}
	return &domain.UserCoupon{
		ID:         row.ID,
		CouponNo:   row.CouponNo,
		UserID:     row.UserID,
		TemplateID: row.TemplateID,
		Status:     row.Status,
		OrderNo:    row.OrderNo,
		ExpireAt:   row.ExpireAt,
	}, nil
}

func (r *UserCouponRepo) GetTemplateByID(ctx context.Context, templateID int64) (*domain.CouponTemplate, error) {
	var row couponTemplateRow
	err := r.db.db(ctx).Where("id = ?", templateID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCouponNotAvailable
	}
	if err != nil {
		return nil, err
	}
	return &domain.CouponTemplate{
		ID:               row.ID,
		Name:             row.Name,
		Type:             row.Type,
		ValueCents:       row.ValueCents,
		PercentBp:        row.PercentBp,
		MinSpendCents:    row.MinSpendCents,
		MaxDiscountCents: row.MaxDiscountCents,
		Redeemable:       row.Redeemable,
		RedeemPoints:     row.RedeemPoints,
		ValidDays:        row.ValidDays,
		TotalQty:         row.TotalQty,
		PerUserLimit:     row.PerUserLimit,
		Status:           row.Status,
	}, nil
}

func (r *UserCouponRepo) CreateTemplate(ctx context.Context, template *domain.CouponTemplate) error {
	row := &couponTemplateRow{
		Name:             template.Name,
		Type:             template.Type,
		ValueCents:       template.ValueCents,
		PercentBp:        template.PercentBp,
		MinSpendCents:    template.MinSpendCents,
		MaxDiscountCents: template.MaxDiscountCents,
		Redeemable:       template.Redeemable,
		RedeemPoints:     template.RedeemPoints,
		ValidDays:        template.ValidDays,
		TotalQty:         template.TotalQty,
		PerUserLimit:     template.PerUserLimit,
		Status:           template.Status,
	}
	if err := r.db.db(ctx).Create(row).Error; err != nil {
		return err
	}
	template.ID = row.ID
	return nil
}

func (r *UserCouponRepo) ListTemplates(ctx context.Context) ([]domain.CouponTemplate, error) {
	var rows []couponTemplateRow
	if err := r.db.db(ctx).Order("id").Find(&rows).Error; err != nil {
		return nil, err
	}
	templates := make([]domain.CouponTemplate, 0, len(rows))
	for _, row := range rows {
		templates = append(templates, toCouponTemplate(row))
	}
	return templates, nil
}

func (r *UserCouponRepo) SetTemplateStatus(ctx context.Context, templateID int64, status string) error {
	return r.db.db(ctx).
		Model(&couponTemplateRow{}).
		Where("id = ?", templateID).
		Update("status", status).Error
}

func (r *UserCouponRepo) ListRedeemableTemplates(ctx context.Context) ([]domain.CouponTemplate, error) {
	var rows []couponTemplateRow
	if err := r.db.db(ctx).
		Where("redeemable = ? AND status = ?", true, "ACTIVE").
		Order("redeem_points, id").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	templates := make([]domain.CouponTemplate, 0, len(rows))
	for _, row := range rows {
		templates = append(templates, toCouponTemplate(row))
	}
	return templates, nil
}

func toCouponTemplate(row couponTemplateRow) domain.CouponTemplate {
	return domain.CouponTemplate{
		ID:               row.ID,
		Name:             row.Name,
		Type:             row.Type,
		ValueCents:       row.ValueCents,
		PercentBp:        row.PercentBp,
		MinSpendCents:    row.MinSpendCents,
		MaxDiscountCents: row.MaxDiscountCents,
		Redeemable:       row.Redeemable,
		RedeemPoints:     row.RedeemPoints,
		ValidDays:        row.ValidDays,
		TotalQty:         row.TotalQty,
		PerUserLimit:     row.PerUserLimit,
		Status:           row.Status,
	}
}

func (r *UserCouponRepo) CreateInstance(ctx context.Context, coupon *domain.UserCoupon) error {
	row := &userCouponRow{
		CouponNo:   coupon.CouponNo,
		TemplateID: coupon.TemplateID,
		UserID:     coupon.UserID,
		Status:     coupon.Status,
		OrderNo:    coupon.OrderNo,
		ExpireAt:   coupon.ExpireAt,
	}
	if err := r.db.db(ctx).Create(row).Error; err != nil {
		return err
	}
	coupon.ID = row.ID
	return nil
}

// LockForOrder 条件 UPDATE：只有 UNUSED 的券能锁到，并发下仅一单成功。
func (r *UserCouponRepo) LockForOrder(ctx context.Context, couponNo, orderNo string) error {
	res := r.db.db(ctx).
		Model(&userCouponRow{}).
		Where("coupon_no = ? AND status = ?", couponNo, domain.CouponUnused).
		Updates(map[string]any{
			"status":   domain.CouponLocked,
			"order_no": orderNo,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return domain.ErrCouponNotAvailable
	}
	return nil
}

func (r *UserCouponRepo) UnlockByOrderNo(ctx context.Context, orderNo string) error {
	return r.db.db(ctx).
		Model(&userCouponRow{}).
		Where("order_no = ? AND status = ?", orderNo, domain.CouponLocked).
		Updates(map[string]any{
			"status":   domain.CouponUnused,
			"order_no": nil,
		}).Error
}

func (r *UserCouponRepo) MarkUsedByOrderNo(ctx context.Context, orderNo string) error {
	return r.db.db(ctx).
		Model(&userCouponRow{}).
		Where("order_no = ? AND status = ?", orderNo, domain.CouponLocked).
		Updates(map[string]any{
			"status":  domain.CouponUsed,
			"used_at": time.Now(),
		}).Error
}
