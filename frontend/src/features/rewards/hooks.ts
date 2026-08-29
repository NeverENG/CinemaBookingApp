import { useMutation, useQueryClient } from '@tanstack/react-query'
import { rewardsApi } from '../../services/api'
import { demoCoupons, demoPoints } from '../../demo'
import { useServerQuery } from '../../hooks/server/useServerQuery'
import { randomUUID } from '../../lib/ids'

export function usePointsQuery() {
  return useServerQuery({ queryKey: ['points'], queryFn: rewardsApi.points, fallback: () => demoPoints, staleTime: 30_000 })
}

export function useRedeemableCouponsQuery() {
  return useServerQuery({ queryKey: ['redeemable-coupons'], queryFn: rewardsApi.redeemable, fallback: () => demoCoupons, staleTime: 120_000 })
}

export function useExchangeCouponMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (templateId: number) => rewardsApi.exchange(templateId, randomUUID()),
    retry: false,
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['points'] }),
  })
}
