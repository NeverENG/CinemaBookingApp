package service

import (
	"context"
	"time"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/NeverENG/CinemaBookingApp/internal/core/port"
	"github.com/NeverENG/CinemaBookingApp/internal/pkg/uid"
)

// PointsSvc 用户积分查询与兑换。
type PointsSvc struct {
	tx      port.TxManager
	points  port.PointsRepo
	coupons port.UserCouponRepo
}

func NewPointsSvc(tx port.TxManager, points port.PointsRepo, coupons port.UserCouponRepo) *PointsSvc {
	return &PointsSvc{tx: tx, points: points, coupons: coupons}
}

type PointsView struct {
	Balance int                   `json:"balance"`
	Ledger  []domain.PointsLedger `json:"ledger"`
}

func (s *PointsSvc) GetPoints(ctx context.Context, userID int64) (*PointsView, error) {
	balance, err := s.points.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	ledger, err := s.points.GetRecentLedger(ctx, userID, 20)
	if err != nil {
		return nil, err
	}
	return &PointsView{Balance: balance, Ledger: ledger}, nil
}

type ExchangeResult struct {
	CouponNo     string `json:"coupon_no"`
	BalanceAfter int    `json:"balance_after"`
}

// Exchange 积分兑换优惠券：单事务内 校验模板 → 扣积分（负流水）→ 发券。
// 幂等由 points_ledger UNIQUE(EXCHANGE, exchange_no) 保证，无新增表。
func (s *PointsSvc) Exchange(ctx context.Context, userID int64, templateID int64) (*ExchangeResult, error) {
	var res *ExchangeResult
	err := s.tx.Run(ctx, func(txCtx context.Context) error {
		tpl, err := s.coupons.GetTemplateByID(txCtx, templateID)
		if err != nil {
			return err
		}
		if !tpl.Redeemable || tpl.RedeemPoints <= 0 {
			return domain.ErrCouponNotAvailable
		}
		balance, err := s.points.Exchange(txCtx, userID, tpl.RedeemPoints, uid.ExchangeNo())
		if err != nil {
			return err
		}
		validDays := tpl.ValidDays
		if validDays <= 0 {
			validDays = 30
		}
		coupon := &domain.UserCoupon{
			CouponNo:   uid.CouponNo(),
			UserID:     userID,
			TemplateID: templateID,
			Status:     domain.CouponUnused,
			ExpireAt:   time.Now().Add(time.Duration(validDays) * 24 * time.Hour),
		}
		if err := s.coupons.CreateInstance(txCtx, coupon); err != nil {
			return err
		}
		res = &ExchangeResult{CouponNo: coupon.CouponNo, BalanceAfter: balance}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *PointsSvc) ListRedeemable(ctx context.Context) ([]domain.CouponTemplate, error) {
	return s.coupons.ListRedeemableTemplates(ctx)
}
