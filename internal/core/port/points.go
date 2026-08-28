package port

import (
	"context"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
)

// PointsRepo 积分仓储：余额以流水为准，users.points_balance 是冗余快照。
type PointsRepo interface {
	// GrantOnPaid 支付成功赠送积分（幂等：ORDER_PAID + order_no）。
	GrantOnPaid(ctx context.Context, userID int64, paidCents int64, orderNo string) error
	// ReclaimOnRefund 退款按比例扣回积分（幂等：ORDER_REFUND + refund_no），余额不足则扣到 0。
	ReclaimOnRefund(ctx context.Context, userID int64, refundCents int64, refundNo string) error
	// Exchange 兑换扣积分：余额不足返回 ErrInsufficientPoints；返回兑换后余额。
	Exchange(ctx context.Context, userID int64, points int, exchangeNo string) (int, error)
	GetBalance(ctx context.Context, userID int64) (int, error)
	GetRecentLedger(ctx context.Context, userID int64, limit int) ([]domain.PointsLedger, error)
	// ListBalanceMismatches 余额与流水不一致的用户（对账用）。
	ListBalanceMismatches(ctx context.Context) ([]int64, error)
}
