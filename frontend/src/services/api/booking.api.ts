import { http } from '../http/client'
import { demoSeatMap } from '../../demo'
import type { CreateOrderResult, SeatMapView } from '../../types'
import { normalizeSeatMap } from './normalize'

export const bookingApi = {
  seatMap: async (sessionId: number): Promise<SeatMapView> => normalizeSeatMap(await http.get(`/sessions/${sessionId}/seats`).then((response) => response.data)),
  createOrder: (payload: { session_id: number; seat_ids: number[]; coupon_no?: string }, idempotencyKey: string) => http.post<CreateOrderResult>('/orders', payload, { headers: { 'Idempotency-Key': idempotencyKey } }).then((response) => response.data),
}

export const bookingFallbacks = {
  seatMap: () => demoSeatMap,
}
