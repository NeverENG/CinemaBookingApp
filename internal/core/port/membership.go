package port

import "context"

// MembershipRepo 会员等级：支付后按累计积分升级。
type MembershipRepo interface {
	// UpgradeIfNeeded 检查并升级，返回是否发生变更。
	UpgradeIfNeeded(ctx context.Context, userID int64) (bool, error)
}
