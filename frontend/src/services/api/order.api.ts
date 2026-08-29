import { http } from '../http/client'
import type { Order, Payment, Refund } from '../../types'
import { normalizeOrder, normalizePayment, normalizeRefund } from './normalize'

function rawList(value: unknown): Record<string, unknown>[] {
  return Array.isArray(value) ? value as Record<string, unknown>[] : []
}

export const orderApi = {
  list: async (): Promise<Order[]> => rawList(await http.get('/orders').then((response) => response.data)).map(normalizeOrder),
  get: async (orderNo: string): Promise<Order> => normalizeOrder(await http.get(`/orders/${encodeURIComponent(orderNo)}`).then((response) => response.data)),
  refund: async (orderNo: string, reason: string, idempotencyKey: string): Promise<Refund> => normalizeRefund(await http.post(`/orders/${encodeURIComponent(orderNo)}/refund`, { reason }, { headers: { 'Idempotency-Key': idempotencyKey } }).then((response) => response.data)),
  change: (orderNo: string, payload: { new_session_id: number; new_seat_ids: number[] }, idempotencyKey: string) => http.post(`/orders/${encodeURIComponent(orderNo)}/change`, payload, { headers: { 'Idempotency-Key': idempotencyKey } }).then((response) => response.data),
  payment: async (orderNo: string): Promise<Payment> => normalizePayment(await http.get(`/payments/order/${encodeURIComponent(orderNo)}`).then((response) => response.data)),
  createPayment: async (orderNo: string, idempotencyKey: string): Promise<Payment & { mockCallbackUrl?: string }> => {
    const raw = await http.post(`/payments`, { order_no: orderNo }, { headers: { 'Idempotency-Key': idempotencyKey } }).then((response) => response.data)
    return { ...normalizePayment(raw), mockCallbackUrl: typeof raw === 'object' && raw !== null && 'mock_callback_url' in raw ? String(raw.mock_callback_url) : undefined }
  },
  mockPay: (transactionNo: string, idempotencyKey: string) => http.post('/payments/mock-pay', { transaction_no: transactionNo }, { headers: { 'Idempotency-Key': idempotencyKey } }).then((response) => response.data),
}
