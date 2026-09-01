import { http } from '../http/client'
import { demoCoupons, demoDashboardCinemas, demoDashboardMovies, demoDashboardSummary, demoDashboardTrend, demoHalls, demoMovies } from '../../demo'
import type { Banner, CouponTemplate, DashboardCinemaRow, DashboardMovieRow, DashboardSummary, DashboardTrendRow, Hall, Movie, TicketVerification } from '../../types'
import { normalizeBanner, normalizeCinemaRanking, normalizeCoupon, normalizeHall, normalizeMovie, normalizeMovieRanking, normalizeSession, normalizeSummary, normalizeTicketVerification, normalizeTrend } from './normalize'

type Query = Record<string, string | number | undefined>
type Raw = Record<string, unknown>

function rawList(value: unknown): Raw[] {
  return Array.isArray(value) ? value as Raw[] : []
}

export const adminApi = {
  movies: {
    list: async (): Promise<Movie[]> => rawList(await http.get('/admin/movies').then((response) => response.data)).map(normalizeMovie),
    create: async (payload: Record<string, unknown>) => normalizeMovie(await http.post('/admin/movies', payload).then((response) => response.data)),
    update: async (id: number, payload: Record<string, unknown>) => normalizeMovie(await http.patch(`/admin/movies/${id}`, payload).then((response) => response.data)),
    setStatus: (id: number, status: string) => http.patch(`/admin/movies/${id}/status`, { status }).then((response) => response.data),
  },
  halls: {
    list: async (cinemaId: number): Promise<Hall[]> => rawList(await http.get('/admin/halls', { params: { cinema_id: cinemaId } }).then((response) => response.data)).map(normalizeHall),
    create: async (payload: Record<string, unknown>) => normalizeHall(await http.post('/admin/halls', payload).then((response) => response.data)),
    update: async (id: number, payload: Record<string, unknown>) => normalizeHall(await http.patch(`/admin/halls/${id}`, payload).then((response) => response.data)),
  },
  sessions: {
    create: async (payload: Record<string, unknown>) => normalizeSession(await http.post('/admin/sessions', payload).then((response) => response.data)),
    updatePrice: (id: number, payload: Record<string, unknown>) => http.patch(`/admin/sessions/${id}/price`, payload).then((response) => response.data),
    cancel: (id: number) => http.post(`/admin/sessions/${id}/cancel`).then((response) => response.data),
  },
  banners: {
    list: async (): Promise<Banner[]> => rawList(await http.get('/admin/banners').then((response) => response.data)).map(normalizeBanner),
    create: async (payload: Record<string, unknown>) => normalizeBanner(await http.post('/admin/banners', payload).then((response) => response.data)),
    update: async (id: number, payload: Record<string, unknown>) => normalizeBanner(await http.patch(`/admin/banners/${id}`, payload).then((response) => response.data)),
    delete: (id: number) => http.delete(`/admin/banners/${id}`).then((response) => response.data),
  },
  coupons: {
    list: async (): Promise<CouponTemplate[]> => rawList(await http.get('/admin/coupons/templates').then((response) => response.data)).map(normalizeCoupon),
    create: async (payload: Record<string, unknown>) => normalizeCoupon(await http.post('/admin/coupons/templates', payload).then((response) => response.data)),
    setStatus: (id: number, status: string) => http.patch(`/admin/coupons/templates/${id}/status`, { status }).then((response) => response.data),
  },
  admins: {
    create: (payload: Record<string, unknown>) => http.post('/admin/admins', payload).then((response) => response.data),
  },
  tickets: {
    verify: async (ticketNo: string): Promise<TicketVerification> => normalizeTicketVerification(await http.post('/admin/tickets/verify', { ticket_no: ticketNo }).then((response) => response.data)),
  },
  dashboard: {
    summary: async (params: Query): Promise<DashboardSummary> => normalizeSummary(await http.get('/admin/dashboard/box-office/summary', { params }).then((response) => response.data)),
    trend: async (params: Query): Promise<DashboardTrendRow[]> => rawList(await http.get('/admin/dashboard/box-office', { params }).then((response) => response.data)).map(normalizeTrend),
    byMovie: async (params: Query): Promise<DashboardMovieRow[]> => rawList(await http.get('/admin/dashboard/box-office/by-movie', { params }).then((response) => response.data)).map(normalizeMovieRanking),
    byCinema: async (params: Query): Promise<DashboardCinemaRow[]> => rawList(await http.get('/admin/dashboard/box-office/by-cinema', { params }).then((response) => response.data)).map(normalizeCinemaRanking),
    reconcile: () => http.post('/admin/dashboard/box-office/reconcile').then((response) => response.data),
  },
}

export const adminFallbacks = {
  movies: () => demoMovies,
  halls: () => demoHalls,
  coupons: () => demoCoupons,
  dashboard: {
    summary: () => demoDashboardSummary,
    trend: () => demoDashboardTrend,
    byMovie: () => demoDashboardMovies,
    byCinema: () => demoDashboardCinemas,
  },
}
