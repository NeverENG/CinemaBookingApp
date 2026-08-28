package biz

import "github.com/NeverENG/CinemaBookingApp/internal/core/domain"

// CalcOrderAmount 计算订单金额：票面合计 + 优惠券抵扣（本期无会员折扣）。
// 返回 total（票面合计）、discount（会员折扣，本期恒 0）、coupon（券抵扣）、paid（实付）。
func CalcOrderAmount(seatPrices []int64, couponDiscount int64) (total, discount, coupon, paid int64, err error) {
	for _, p := range seatPrices {
		if p < 0 {
			return 0, 0, 0, 0, domain.ErrMoneyInvalid
		}
		total += p
	}
	if couponDiscount < 0 || couponDiscount > total {
		return 0, 0, 0, 0, domain.ErrMoneyInvalid
	}
	discount = 0
	coupon = couponDiscount
	paid = total - coupon
	return total, discount, coupon, paid, nil
}
