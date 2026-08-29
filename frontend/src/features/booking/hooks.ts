import { useQueryClient } from '@tanstack/react-query'
import { bookingApi, bookingFallbacks } from '../../services/api'
import { useServerQuery } from '../../hooks/server/useServerQuery'
import { useMutation } from '@tanstack/react-query'
import { randomUUID } from '../../lib/ids'

export function useSeatMapQuery(sessionId: number) {
  return useServerQuery({ queryKey: ['seat-map', sessionId], queryFn: () => bookingApi.seatMap(sessionId), fallback: bookingFallbacks.seatMap, enabled: sessionId > 0, staleTime: 0, refetchInterval: 5000 })
}

export function useCreateOrderMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: { session_id: number; seat_ids: number[]; coupon_no?: string }) => bookingApi.createOrder(payload, randomUUID()),
    retry: false,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['seat-map'] })
      void queryClient.invalidateQueries({ queryKey: ['sessions'] })
    },
  })
}
