import { useQueryClient, useMutation } from '@tanstack/react-query'
import { orderApi } from '../../services/api'
import { demoOrder } from '../../demo'
import { useServerQuery } from '../../hooks/server/useServerQuery'
import { randomUUID } from '../../lib/ids'

export function useOrdersQuery() {
  return useServerQuery({ queryKey: ['orders'], queryFn: orderApi.list, fallback: () => [demoOrder], staleTime: 30_000 })
}

export function useOrderQuery(orderNo: string) {
  return useServerQuery({ queryKey: ['order', orderNo], queryFn: () => orderApi.get(orderNo), fallback: () => ({ ...demoOrder, orderNo: orderNo || demoOrder.orderNo }), enabled: orderNo.length > 0, staleTime: 0, refetchInterval: 5000 })
}

export function usePaymentQuery(orderNo: string) {
  return useServerQuery({ queryKey: ['payment', orderNo], queryFn: () => orderApi.payment(orderNo), fallback: () => ({ transactionNo: 'TX-DEMO-001', orderNo, amountCents: demoOrder.paidCents, channel: 'MOCK', status: 'SUCCESS' }), enabled: orderNo.length > 0, staleTime: 0, refetchInterval: 3000 })
}

export function useCancelOrderMutation(orderNo: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => orderApi.cancel(orderNo),
    retry: false,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['order', orderNo] })
      void queryClient.invalidateQueries({ queryKey: ['orders'] })
      void queryClient.invalidateQueries({ queryKey: ['payment', orderNo] })
      void queryClient.invalidateQueries({ queryKey: ['sessions'] })
      void queryClient.invalidateQueries({ queryKey: ['seat-map'] })
    },
  })
}

export function useCreatePaymentMutation() {
  return useMutation({ mutationFn: (orderNo: string) => orderApi.createPayment(orderNo, randomUUID()), retry: false })
}

export function useMockPayMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (transactionNo: string) => orderApi.mockPay(transactionNo, randomUUID()),
    retry: false,
    onSuccess: (_, transactionNo) => {
      void queryClient.invalidateQueries({ queryKey: ['payment'] })
      void queryClient.invalidateQueries({ queryKey: ['order'] })
      void queryClient.invalidateQueries({ queryKey: ['orders'] })
      void queryClient.invalidateQueries({ queryKey: ['payment', transactionNo] })
    },
  })
}

export function useRefundMutation(orderNo: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (reason: string) => {
      const refund = await orderApi.refund(orderNo, reason, randomUUID())
      if (refund.status === 'PENDING') await orderApi.mockRefund(refund.refundNo)
      return refund
    },
    retry: false,
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['order', orderNo] }),
  })
}

export function useChangeOrderMutation(orderNo: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: { new_session_id: number; new_seat_ids: number[] }) => orderApi.change(orderNo, payload, randomUUID()),
    retry: false,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['order', orderNo] })
      void queryClient.invalidateQueries({ queryKey: ['orders'] })
      void queryClient.invalidateQueries({ queryKey: ['seat-map'] })
      void queryClient.invalidateQueries({ queryKey: ['sessions'] })
    },
  })
}
