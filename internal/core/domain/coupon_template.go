package domain

// CouponType 优惠券类型：满减 / 折扣。
const (
	CouponTypeFixed   = "FIXED"
	CouponTypePercent = "PERCENT"
)

// CouponTemplate 优惠券模板（发行规则）。
type CouponTemplate struct {
	ID               int64
	Name             string
	Type             string
	ValueCents       int64 // FIXED 面额（分）
	PercentBp        int   // PERCENT 折扣万分比（9000 = 9 折）
	MinSpendCents    int64 // 使用门槛
	MaxDiscountCents int64 // 折扣封顶，0 = 不封顶
	Redeemable       bool  // 是否可积分兑换
	RedeemPoints     int   // 兑换所需积分
	ValidDays        int   // 发放后有效期（天）
	TotalQty         int
	PerUserLimit     int
	Status           string
}

// DiscountCents 计算本券可抵扣金额（不能超过订单合计）。
func (t *CouponTemplate) DiscountCents(totalCents int64) (int64, error) {
	if totalCents < t.MinSpendCents {
		return 0, ErrCouponNotAvailable
	}
	var d int64
	switch t.Type {
	case CouponTypeFixed:
		d = t.ValueCents
	case CouponTypePercent:
		d = totalCents * int64(t.PercentBp) / 10000
	default:
		return 0, ErrCouponNotAvailable
	}
	if t.MaxDiscountCents > 0 && d > t.MaxDiscountCents {
		d = t.MaxDiscountCents
	}
	if d > totalCents {
		d = totalCents
	}
	return d, nil
}

// Validate 模板校验。
func (t *CouponTemplate) Validate() error {
	if t.Name == "" || t.Type == "" {
		return ErrCouponNotAvailable
	}
	switch t.Type {
	case CouponTypeFixed:
		if t.ValueCents <= 0 {
			return ErrCouponNotAvailable
		}
	case CouponTypePercent:
		if t.PercentBp <= 0 || t.PercentBp > 10000 {
			return ErrCouponNotAvailable
		}
	default:
		return ErrCouponNotAvailable
	}
	if t.MinSpendCents < 0 || t.TotalQty < 0 || t.PerUserLimit <= 0 || t.RedeemPoints < 0 {
		return ErrCouponNotAvailable
	}
	return nil
}
