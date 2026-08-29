import { http } from '../http/client'
import type { CouponTemplate, PointsView } from '../../types'
import { normalizeCoupon, normalizePoints } from './normalize'

export const rewardsApi = {
  points: async (): Promise<PointsView> => normalizePoints(await http.get('/me/points').then((response) => response.data)),
  redeemable: async (): Promise<CouponTemplate[]> => {
    const raw = await http.get('/coupons/redeemable').then((response) => response.data)
    return (Array.isArray(raw) ? raw : []).map(normalizeCoupon)
  },
  exchange: (templateId: number, idempotencyKey: string) => http.post('/me/points/exchange', { template_id: templateId }, { headers: { 'Idempotency-Key': idempotencyKey } }).then((response) => response.data),
}
