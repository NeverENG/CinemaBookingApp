import { useMutation, useQueryClient } from '@tanstack/react-query'
import { adminApi, adminFallbacks } from '../../services/api'
import { useServerQuery } from '../../hooks/server/useServerQuery'

export function useAdminMoviesQuery() {
  return useServerQuery({ queryKey: ['admin', 'movies'], queryFn: adminApi.movies.list, fallback: adminFallbacks.movies, staleTime: 30_000 })
}

export function useAdminHallsQuery(cinemaId: number) {
  return useServerQuery({ queryKey: ['admin', 'halls', cinemaId], queryFn: () => adminApi.halls.list(cinemaId), fallback: adminFallbacks.halls, enabled: cinemaId > 0, staleTime: 30_000 })
}

export function useAdminCouponsQuery() {
  return useServerQuery({ queryKey: ['admin', 'coupons'], queryFn: adminApi.coupons.list, fallback: adminFallbacks.coupons, staleTime: 30_000 })
}

export function useAdminBannersQuery() {
  return useServerQuery({ queryKey: ['admin', 'banners'], queryFn: adminApi.banners.list, fallback: () => [], staleTime: 30_000 })
}

export function useAdminAccountsQuery() {
  return useServerQuery({ queryKey: ['admin', 'admins'], queryFn: adminApi.admins.list, fallback: adminFallbacks.admins, staleTime: 30_000 })
}

export function useDashboardQueries(params: Record<string, string | number | undefined>) {
  return {
    summary: useServerQuery({ queryKey: ['admin', 'dashboard', 'summary', params], queryFn: () => adminApi.dashboard.summary(params), fallback: adminFallbacks.dashboard.summary, staleTime: 30_000 }),
    trend: useServerQuery({ queryKey: ['admin', 'dashboard', 'trend', params], queryFn: () => adminApi.dashboard.trend(params), fallback: adminFallbacks.dashboard.trend, staleTime: 30_000 }),
    movies: useServerQuery({ queryKey: ['admin', 'dashboard', 'movies', params], queryFn: () => adminApi.dashboard.byMovie(params), fallback: adminFallbacks.dashboard.byMovie, staleTime: 30_000 }),
    cinemas: useServerQuery({ queryKey: ['admin', 'dashboard', 'cinemas', params], queryFn: () => adminApi.dashboard.byCinema(params), fallback: adminFallbacks.dashboard.byCinema, staleTime: 30_000 }),
  }
}

export function useCreateAdminMovieMutation() {
  const queryClient = useQueryClient()
  return useMutation({ mutationFn: (payload: Record<string, unknown>) => adminApi.movies.create(payload), retry: false, onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['admin', 'movies'] }) })
}

export function useUpdateAdminMovieMutation() {
  const queryClient = useQueryClient()
  return useMutation({ mutationFn: (payload: { id: number; data: Record<string, unknown> }) => adminApi.movies.update(payload.id, payload.data), retry: false, onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['admin', 'movies'] }) })
}

export function useSetMovieStatusMutation() {
  const queryClient = useQueryClient()
  return useMutation({ mutationFn: (payload: { id: number; status: string }) => adminApi.movies.setStatus(payload.id, payload.status), retry: false, onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['admin', 'movies'] }) })
}

export function useCreateHallMutation() {
  const queryClient = useQueryClient()
  return useMutation({ mutationFn: (payload: Record<string, unknown>) => adminApi.halls.create(payload), retry: false, onSuccess: (_, payload) => void queryClient.invalidateQueries({ queryKey: ['admin', 'halls', Number(payload.cinema_id)] }) })
}

export function useUpdateHallMutation() {
  const queryClient = useQueryClient()
  return useMutation({ mutationFn: (payload: { id: number; data: Record<string, unknown> }) => adminApi.halls.update(payload.id, payload.data), retry: false, onSuccess: (_, payload) => void queryClient.invalidateQueries({ queryKey: ['admin', 'halls', Number(payload.data.cinema_id)] }) })
}

export function useCreateSessionMutation() {
  const queryClient = useQueryClient()
  return useMutation({ mutationFn: (payload: Record<string, unknown>) => adminApi.sessions.create(payload), retry: false, onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['sessions'] }) })
}

export function useUpdateSessionPriceMutation() {
  const queryClient = useQueryClient()
  return useMutation({ mutationFn: (payload: { id: number; data: Record<string, unknown> }) => adminApi.sessions.updatePrice(payload.id, payload.data), retry: false, onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['sessions'] }) })
}

export function useCancelSessionMutation() {
  const queryClient = useQueryClient()
  return useMutation({ mutationFn: (id: number) => adminApi.sessions.cancel(id), retry: false, onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['sessions'] }) })
}

export function useCreateCouponMutation() {
  const queryClient = useQueryClient()
  return useMutation({ mutationFn: (payload: Record<string, unknown>) => adminApi.coupons.create(payload), retry: false, onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['admin', 'coupons'] }) })
}

export function useCreateBannerMutation() {
  const queryClient = useQueryClient()
  return useMutation({ mutationFn: (payload: Record<string, unknown>) => adminApi.banners.create(payload), retry: false, onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['admin', 'banners'] }) })
}

export function useCreateAdminMutation() {
  const queryClient = useQueryClient()
  return useMutation({ mutationFn: (payload: Record<string, unknown>) => adminApi.admins.create(payload), retry: false, onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['admin', 'admins'] }) })
}

export function useVerifyTicketMutation() {
  return useMutation({ mutationFn: (ticketNo: string) => adminApi.tickets.verify(ticketNo), retry: false })
}

export function useReconcileMutation() {
  return useMutation({ mutationFn: adminApi.dashboard.reconcile, retry: false })
}
