package domain

import "time"

type PointsBizType string

const (
	PointsOrderPaid   PointsBizType = "ORDER_PAID"
	PointsOrderRefund PointsBizType = "ORDER_REFUND"
)

// PointsPerYuan 每实付 1 元赠送 1 积分（可配置）。
const PointsPerYuan = 1

// PointsLedger 积分流水（只增）：UNIQUE(biz_type, biz_no) 保证幂等。
type PointsLedger struct {
	ID           int64
	UserID       int64
	ChangePoints int
	BalanceAfter int
	BizType      PointsBizType
	BizNo        string
	CreatedAt    time.Time
}
