export type Role = 'USER' | 'SUPER_ADMIN' | 'CINEMA_ADMIN' | 'FINANCE'

export interface AuthSession {
  token: string
  userId: number
  role: Role
  cinemaId?: number
}

export interface Cinema {
  id: number
  name: string
  city: string
  address: string
  distanceKm?: number
  status?: string
  longitude?: number
  latitude?: number
}

export interface AdminAccount {
  id: number
  username: string
  nickname: string
  role: Role
  cinemaId?: number
  cinemaName?: string
  status: string
  createdAt: string
}

export interface Movie {
  id: number
  title: string
  coverUrl: string
  backdropUrl?: string
  trailerUrl: string
  description: string
  durationMinutes: number
  genre: string
  releaseDate: string
  rating: number
  soldCount: number
  status: string
}

export interface Banner {
  id: number
  title: string
  imageUrl: string
  sort?: number
  enabled?: boolean
}

export interface HomeView {
  banners: Banner[]
  hotMovies: Movie[]
}

export interface Session {
  id: number
  movieId: number
  movieTitle: string
  cinemaId: number
  cinemaName: string
  hallId: number
  hallName: string
  startTime: string
  endTime: string
  basePriceCents: number
  priceRulesJson?: string
  status: string
  remainingSeats?: number
}

export type SeatStatus = 'available' | 'locked' | 'booked' | 'disabled' | 'selected'

export interface Seat {
  seatId: number
  rowNo: number
  colNo: number
  seatNo: string
  type: string
  status: SeatStatus
  priceCents?: number
}

export interface SeatMapView {
  session: Session
  seats: Seat[]
  serverTime: string
}

export interface OrderItem {
  id?: number
  seatNo: string
  priceCents: number
  ticketNo?: string
  usedAt?: string | null
}

export interface Order {
  orderNo: string
  userId?: number
  sessionId: number
  cinemaId?: number
  movieId?: number
  movieTitle: string
  cinemaName: string
  hallName: string
  startTime: string
  status: string
  totalCents: number
  discountCents: number
  couponCents: number
  paidCents: number
  expireAt?: string
  createdAt?: string
  paidAt?: string | null
  canRefund?: boolean
  canChange?: boolean
  items: OrderItem[]
}

export interface CreateOrderResult {
  order_no: string
  expire_at: string
  paid_cents: number
  seat_nos: string[]
}

export interface Payment {
  transactionNo: string
  orderNo: string
  amountCents: number
  channel: string
  status: string
  paidAt?: string | null
  closedAt?: string | null
}

export interface PointsLedger {
  id?: number
  changePoints: number
  balanceAfter: number
  bizType: string
  bizNo: string
  createdAt: string
}

export interface PointsView {
  balance: number
  ledger: PointsLedger[]
}

export interface CouponTemplate {
  id: number
  name: string
  type: string
  valueCents: number
  percentBp: number
  minSpendCents: number
  maxDiscountCents: number
  redeemable: boolean
  redeemPoints: number
  validDays: number
  totalQty: number
  perUserLimit: number
  status: string
}

export interface Refund {
  refundNo: string
  orderNo: string
  amountCents: number
  reason: string
  status: string
}

export interface TicketVerification {
  ticketNo: string
  orderNo: string
  seatNo: string
  movieId: number
  cinemaId: number
  usedAt: string
  orderStatus: string
  alreadyUsed: boolean
}

export interface ChangeTicketResult {
  newOrderNo: string
  newPaidCents: number
  refundNo: string
  refundAmountCents: number
}

export interface Hall {
  id: number
  cinemaId: number
  name: string
  seatLayoutJson: string
  status: string
}

export interface DashboardSummary {
  orderCount: number
  ticketCount: number
  grossCents: number
  refundCents: number
  netCents: number
}

export interface DashboardTrendRow {
  date: string
  orderCount: number
  ticketCount: number
  grossCents: number
  refundCents: number
  netCents: number
}

export interface DashboardMovieRow {
  movieId: number
  movieTitle: string
  orderCount: number
  grossCents: number
  netCents: number
}

export interface DashboardCinemaRow {
  cinemaId: number
  cinemaName: string
  orderCount: number
  grossCents: number
  netCents: number
}

export interface ApiEnvelope<T> {
  code: number
  msg: string
  data?: T
}

export interface AdminNavItem {
  label: string
  path: string
  icon: string
  roles: Role[]
}
