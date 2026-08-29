import type { AuthSession, Banner, CouponTemplate, DashboardCinemaRow, DashboardMovieRow, DashboardSummary, DashboardTrendRow, Hall, HomeView, Movie, Order, OrderItem, Payment, PointsLedger, PointsView, Refund, Seat, SeatMapView, Session } from '../../types'

type Raw = Record<string, unknown> | null | undefined

function read<T>(raw: Raw, ...keys: string[]): T | undefined {
  if (!raw) return undefined
  for (const key of keys) {
    if (raw[key] !== undefined && raw[key] !== null) return raw[key] as T
  }
  return undefined
}

function number(raw: Raw, ...keys: string[]) {
  const value = read<number | string>(raw, ...keys)
  return value === undefined ? 0 : Number(value)
}

function string(raw: Raw, ...keys: string[]) {
  return String(read<string | number>(raw, ...keys) ?? '')
}

function boolean(raw: Raw, ...keys: string[]) {
  return Boolean(read<boolean | string | number>(raw, ...keys))
}

export function normalizeAuth(raw: Raw): AuthSession {
  return {
    token: string(raw, 'token', 'Token'),
    userId: number(raw, 'userId', 'user_id', 'UserID'),
    role: string(raw, 'role', 'Role') as AuthSession['role'],
  }
}

export function normalizeMovie(raw: Raw): Movie {
  return {
    id: number(raw, 'id', 'ID', 'movieId', 'movie_id', 'MovieID'),
    title: string(raw, 'title', 'Title'),
    coverUrl: string(raw, 'coverUrl', 'cover_url', 'CoverURL'),
    backdropUrl: string(raw, 'backdropUrl', 'backdrop_url', 'BackdropURL'),
    trailerUrl: string(raw, 'trailerUrl', 'trailer_url', 'TrailerURL', 'Trailer'),
    description: string(raw, 'description', 'Description'),
    durationMinutes: number(raw, 'durationMinutes', 'duration_minutes', 'DurationMinutes'),
    genre: string(raw, 'genre', 'Genre'),
    releaseDate: string(raw, 'releaseDate', 'release_date', 'ReleaseDate'),
    rating: number(raw, 'rating', 'Rating'),
    soldCount: number(raw, 'soldCount', 'sold_count', 'Sold'),
    status: string(raw, 'status', 'Status'),
  }
}

export function normalizeBanner(raw: Raw): Banner {
  return {
    id: number(raw, 'id', 'ID'),
    title: string(raw, 'title', 'Title'),
    imageUrl: string(raw, 'imageUrl', 'image_url', 'ImageURL'),
    sort: number(raw, 'sort', 'Sort'),
    enabled: boolean(raw, 'enabled', 'Enabled'),
  }
}

export function normalizeCinema(raw: Raw) {
  return {
    id: number(raw, 'id', 'ID', 'cinema_id', 'CinemaID'),
    name: string(raw, 'name', 'Name', 'cinema_name', 'CinemaName'),
    city: string(raw, 'city', 'City'),
    address: string(raw, 'address', 'Address'),
    distanceKm: number(raw, 'distanceKm', 'distance_km', 'DistanceKm'),
  }
}

export function normalizeSession(raw: Raw): Session {
  return {
    id: number(raw, 'id', 'ID'),
    movieId: number(raw, 'movieId', 'movie_id', 'MovieID'),
    movieTitle: string(raw, 'movieTitle', 'movie_title', 'MovieTitle'),
    cinemaId: number(raw, 'cinemaId', 'cinema_id', 'CinemaID'),
    cinemaName: string(raw, 'cinemaName', 'cinema_name', 'CinemaName') || 'LTerm 影院',
    hallId: number(raw, 'hallId', 'hall_id', 'HallID'),
    hallName: string(raw, 'hallName', 'hall_name', 'HallName'),
    startTime: string(raw, 'startTime', 'start_time', 'StartTime'),
    endTime: string(raw, 'endTime', 'end_time', 'EndTime'),
    basePriceCents: number(raw, 'basePriceCents', 'base_price_cents', 'BasePriceCents'),
    status: string(raw, 'status', 'Status'),
    remainingSeats: number(raw, 'remainingSeats', 'remaining_seats', 'RemainingSeats'),
  }
}

export function normalizeSeat(raw: Raw): Seat {
  const status = string(raw, 'status', 'Status').toLowerCase() as Seat['status']
  return {
    seatId: number(raw, 'seatId', 'seat_id', 'SeatID'),
    rowNo: number(raw, 'rowNo', 'row_no', 'RowNo'),
    colNo: number(raw, 'colNo', 'col_no', 'ColNo'),
    seatNo: string(raw, 'seatNo', 'seat_no', 'SeatNo'),
    type: string(raw, 'type', 'Type'),
    status: ['available', 'locked', 'booked', 'disabled'].includes(status) ? status : 'available',
  }
}

export function normalizeSeatMap(raw: Raw): SeatMapView {
  const session = read<Raw>(raw, 'session', 'Session')
  const seats = read<Raw[]>(raw, 'seats', 'Seats') ?? []
  return {
    session: normalizeSession(session),
    seats: seats.map(normalizeSeat),
    serverTime: string(raw, 'serverTime', 'server_time', 'ServerTime'),
  }
}

function normalizeOrderItem(raw: Raw): OrderItem {
  return {
    id: number(raw, 'id', 'ID'),
    seatNo: string(raw, 'seatNo', 'seat_no', 'SeatNo'),
    priceCents: number(raw, 'priceCents', 'price_cents', 'PriceCents'),
    ticketNo: string(raw, 'ticketNo', 'ticket_no', 'TicketNo') || undefined,
    usedAt: read<string | null>(raw, 'usedAt', 'used_at', 'UsedAt'),
  }
}

export function normalizeOrder(raw: Raw): Order {
  const items = read<Raw[]>(raw, 'items', 'Items') ?? []
  return {
    orderNo: string(raw, 'orderNo', 'order_no', 'OrderNo'),
    userId: number(raw, 'userId', 'user_id', 'UserID'),
    sessionId: number(raw, 'sessionId', 'session_id', 'SessionID'),
    cinemaId: number(raw, 'cinemaId', 'cinema_id', 'CinemaID'),
    movieId: number(raw, 'movieId', 'movie_id', 'MovieID'),
    movieTitle: string(raw, 'movieTitle', 'movie_title', 'MovieTitle'),
    cinemaName: string(raw, 'cinemaName', 'cinema_name', 'CinemaName') || 'LTerm 影院',
    hallName: string(raw, 'hallName', 'hall_name', 'HallName'),
    startTime: string(raw, 'startTime', 'start_time', 'StartTime'),
    status: string(raw, 'status', 'Status'),
    totalCents: number(raw, 'totalCents', 'total_cents', 'TotalCents'),
    discountCents: number(raw, 'discountCents', 'discount_cents', 'DiscountCents'),
    couponCents: number(raw, 'couponCents', 'coupon_cents', 'CouponCents'),
    paidCents: number(raw, 'paidCents', 'paid_cents', 'PaidCents'),
    expireAt: string(raw, 'expireAt', 'expire_at', 'ExpireAt') || undefined,
    createdAt: string(raw, 'createdAt', 'created_at', 'CreatedAt') || undefined,
    paidAt: read<string | null>(raw, 'paidAt', 'paid_at', 'PaidAt'),
    items: items.map(normalizeOrderItem),
  }
}

export function normalizePayment(raw: Raw): Payment {
  return {
    transactionNo: string(raw, 'transactionNo', 'transaction_no', 'TransactionNo'),
    orderNo: string(raw, 'orderNo', 'order_no', 'OrderNo'),
    amountCents: number(raw, 'amountCents', 'amount_cents', 'AmountCents'),
    channel: string(raw, 'channel', 'Channel'),
    status: string(raw, 'status', 'Status'),
  }
}

export function normalizeRefund(raw: Raw): Refund {
  return {
    refundNo: string(raw, 'refundNo', 'refund_no', 'RefundNo'),
    orderNo: string(raw, 'orderNo', 'order_no', 'OrderNo'),
    amountCents: number(raw, 'amountCents', 'amount_cents', 'AmountCents'),
    reason: string(raw, 'reason', 'Reason'),
    status: string(raw, 'status', 'Status'),
  }
}

export function normalizePoints(raw: Raw): PointsView {
  const ledger = read<Raw[]>(raw, 'ledger', 'Ledger') ?? []
  return {
    balance: number(raw, 'balance', 'Balance'),
    ledger: ledger.map((item) => ({
      id: number(item, 'id', 'ID'),
      changePoints: number(item, 'changePoints', 'change_points', 'ChangePoints'),
      balanceAfter: number(item, 'balanceAfter', 'balance_after', 'BalanceAfter'),
      bizType: string(item, 'bizType', 'biz_type', 'BizType'),
      bizNo: string(item, 'bizNo', 'biz_no', 'BizNo'),
      createdAt: string(item, 'createdAt', 'created_at', 'CreatedAt'),
    } satisfies PointsLedger)),
  }
}

export function normalizeCoupon(raw: Raw): CouponTemplate {
  return {
    id: number(raw, 'id', 'ID'),
    name: string(raw, 'name', 'Name'),
    type: string(raw, 'type', 'Type'),
    valueCents: number(raw, 'valueCents', 'value_cents', 'ValueCents'),
    percentBp: number(raw, 'percentBp', 'percent_bp', 'PercentBp'),
    minSpendCents: number(raw, 'minSpendCents', 'min_spend_cents', 'MinSpendCents'),
    maxDiscountCents: number(raw, 'maxDiscountCents', 'max_discount_cents', 'MaxDiscountCents'),
    redeemable: boolean(raw, 'redeemable', 'Redeemable'),
    redeemPoints: number(raw, 'redeemPoints', 'redeem_points', 'RedeemPoints'),
    validDays: number(raw, 'validDays', 'valid_days', 'ValidDays'),
    totalQty: number(raw, 'totalQty', 'total_qty', 'TotalQty'),
    perUserLimit: number(raw, 'perUserLimit', 'per_user_limit', 'PerUserLimit'),
    status: string(raw, 'status', 'Status'),
  }
}

export function normalizeHall(raw: Raw): Hall {
  return {
    id: number(raw, 'id', 'ID'),
    cinemaId: number(raw, 'cinemaId', 'cinema_id', 'CinemaID'),
    name: string(raw, 'name', 'Name'),
    seatLayoutJson: string(raw, 'seatLayoutJson', 'seat_layout_json', 'SeatLayoutJSON'),
    status: string(raw, 'status', 'Status'),
  }
}

export function normalizeHome(raw: Raw): HomeView {
  const banners = read<Raw[]>(raw, 'banners', 'Banners') ?? []
  const hotMovies = read<Raw[]>(raw, 'hotMovies', 'hot_movies', 'HotMovies') ?? []
  return { banners: banners.map(normalizeBanner), hotMovies: hotMovies.map(normalizeMovie) }
}

export function normalizeSummary(raw: Raw): DashboardSummary {
  return {
    orderCount: number(raw, 'orderCount', 'order_count', 'OrderCount'),
    ticketCount: number(raw, 'ticketCount', 'ticket_count', 'TicketCount'),
    grossCents: number(raw, 'grossCents', 'gross_cents', 'GrossCents'),
    refundCents: number(raw, 'refundCents', 'refund_cents', 'RefundCents'),
    netCents: number(raw, 'netCents', 'net_cents', 'NetCents'),
  }
}

export function normalizeTrend(raw: Raw): DashboardTrendRow {
  return {
    date: string(raw, 'date', 'Date'),
    orderCount: number(raw, 'orderCount', 'order_count', 'OrderCount'),
    ticketCount: number(raw, 'ticketCount', 'ticket_count', 'TicketCount'),
    grossCents: number(raw, 'grossCents', 'gross_cents', 'GrossCents'),
    refundCents: number(raw, 'refundCents', 'refund_cents', 'RefundCents'),
    netCents: number(raw, 'netCents', 'net_cents', 'NetCents'),
  }
}

export function normalizeMovieRanking(raw: Raw): DashboardMovieRow {
  return {
    movieId: number(raw, 'movieId', 'movie_id', 'MovieID'),
    movieTitle: string(raw, 'movieTitle', 'movie_title', 'MovieTitle'),
    orderCount: number(raw, 'orderCount', 'order_count', 'OrderCount'),
    grossCents: number(raw, 'grossCents', 'gross_cents', 'GrossCents'),
    netCents: number(raw, 'netCents', 'net_cents', 'NetCents'),
  }
}

export function normalizeCinemaRanking(raw: Raw): DashboardCinemaRow {
  return {
    cinemaId: number(raw, 'cinemaId', 'cinema_id', 'CinemaID'),
    cinemaName: string(raw, 'cinemaName', 'cinema_name', 'CinemaName'),
    orderCount: number(raw, 'orderCount', 'order_count', 'OrderCount'),
    grossCents: number(raw, 'grossCents', 'gross_cents', 'GrossCents'),
    netCents: number(raw, 'netCents', 'net_cents', 'NetCents'),
  }
}
