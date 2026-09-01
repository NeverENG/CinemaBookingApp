package service

import (
	"context"
	"strconv"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/uid"
)

// AdminCouponSvc 优惠券模板管理 + 给用户发券。
type AdminCouponSvc struct {
	tx      port.TxManager
	coupons port.UserCouponRepo
	users   port.UserRepo
	logs    port.OperationLogRepo
}

func NewAdminCouponSvc(tx port.TxManager, coupons port.UserCouponRepo, users port.UserRepo, logs port.OperationLogRepo) *AdminCouponSvc {
	return &AdminCouponSvc{tx: tx, coupons: coupons, users: users, logs: logs}
}

type CouponTemplateInput struct {
	Name             string
	Type             string
	ValueCents       int64
	PercentBp        int
	MinSpendCents    int64
	MaxDiscountCents int64
	Redeemable       bool
	RedeemPoints     int
	ValidDays        int
	TotalQty         int
	PerUserLimit     int
}

func (s *AdminCouponSvc) CreateTemplate(ctx context.Context, scope domain.AdminScope, in CouponTemplateInput) (*domain.CouponTemplate, error) {
	if scope.Role != domain.RoleSuperAdmin {
		return nil, domain.ErrForbidden
	}
	tpl := &domain.CouponTemplate{
		Name:             in.Name,
		Type:             in.Type,
		ValueCents:       in.ValueCents,
		PercentBp:        in.PercentBp,
		MinSpendCents:    in.MinSpendCents,
		MaxDiscountCents: in.MaxDiscountCents,
		Redeemable:       in.Redeemable,
		RedeemPoints:     in.RedeemPoints,
		ValidDays:        in.ValidDays,
		TotalQty:         in.TotalQty,
		PerUserLimit:     in.PerUserLimit,
		Status:           "ACTIVE",
	}
	if tpl.ValidDays <= 0 {
		tpl.ValidDays = 30
	}
	if tpl.PerUserLimit <= 0 {
		tpl.PerUserLimit = 1
	}
	if err := tpl.Validate(); err != nil {
		return nil, err
	}
	if err := s.coupons.CreateTemplate(ctx, tpl); err != nil {
		return nil, err
	}
	return tpl, s.log(ctx, scope.AdminID, "CREATE_COUPON_TEMPLATE", "coupon_template", strconv.FormatInt(tpl.ID, 10), tpl)
}

func (s *AdminCouponSvc) ListTemplates(ctx context.Context, scope domain.AdminScope) ([]domain.CouponTemplate, error) {
	if scope.Role != domain.RoleSuperAdmin {
		return nil, domain.ErrForbidden
	}
	return s.coupons.ListTemplates(ctx)
}

func (s *AdminCouponSvc) SetTemplateStatus(ctx context.Context, scope domain.AdminScope, templateID int64, status string) error {
	if scope.Role != domain.RoleSuperAdmin {
		return domain.ErrForbidden
	}
	if status != "ACTIVE" && status != "PAUSED" {
		return domain.ErrInvalidInput
	}
	if err := s.coupons.SetTemplateStatus(ctx, templateID, status); err != nil {
		return err
	}
	return s.log(ctx, scope.AdminID, "SET_COUPON_TEMPLATE_STATUS", "coupon_template", strconv.FormatInt(templateID, 10), map[string]string{"status": status})
}

// IssueToUser 给指定用户发一张券。
func (s *AdminCouponSvc) IssueToUser(ctx context.Context, scope domain.AdminScope, userID, templateID int64) (*domain.UserCoupon, error) {
	if scope.Role != domain.RoleSuperAdmin {
		return nil, domain.ErrForbidden
	}
	var issued *domain.UserCoupon
	err := s.tx.Run(ctx, func(txCtx context.Context) error {
		if _, err := s.users.GetUserByID(txCtx, userID); err != nil {
			return err
		}
		tpl, err := s.coupons.GetTemplateByID(txCtx, templateID)
		if err != nil {
			return err
		}
		if tpl.Status != "ACTIVE" {
			return domain.ErrCouponNotAvailable
		}
		validDays := tpl.ValidDays
		if validDays <= 0 {
			validDays = 30
		}
		issued = &domain.UserCoupon{
			CouponNo:   uid.CouponNo(),
			UserID:     userID,
			TemplateID: templateID,
			Status:     domain.CouponUnused,
			ExpireAt:   time.Now().Add(time.Duration(validDays) * 24 * time.Hour),
		}
		if err := s.coupons.CreateInstance(txCtx, issued); err != nil {
			return err
		}
		return s.log(txCtx, scope.AdminID, "ISSUE_COUPON", "user", strconv.FormatInt(userID, 10), map[string]any{
			"template_id": templateID,
			"coupon_no":   issued.CouponNo,
		})
	})
	if err != nil {
		return nil, err
	}
	return issued, nil
}

func (s *AdminCouponSvc) log(ctx context.Context, adminID int64, action, targetType, targetID string, detail any) error {
	return s.logs.Create(ctx, &domain.OperationLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detail,
	})
}
