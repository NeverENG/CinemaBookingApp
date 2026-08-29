import type {
  Cinema,
  CouponTemplate,
  DashboardCinemaRow,
  DashboardMovieRow,
  DashboardSummary,
  DashboardTrendRow,
  Hall,
  HomeView,
  Movie,
  Order,
  PointsView,
  Seat,
  SeatMapView,
  Session,
} from './types'

const coverSources: Record<string, string> = {
  dune: 'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=780&h=1040&q=85',
  nezha: 'https://images.unsplash.com/photo-1536440136628-849c177e76a1?auto=format&fit=crop&w=780&h=1040&q=85',
  echo: 'https://images.unsplash.com/photo-1485846234645-a62644f84728?auto=format&fit=crop&w=780&h=1040&q=85',
  midnight: 'https://images.unsplash.com/photo-1440404653325-ab127d49abc1?auto=format&fit=crop&w=780&h=1040&q=85',
}

const cover = (seed: string) => coverSources[seed] ?? `https://picsum.photos/seed/lterm-${seed}/520/760`

export const demoCinemas: Cinema[] = [
  { id: 1, name: 'LTerm 光影中心', city: '北京', address: '朝阳区演示路 2 号', distanceKm: 1.2 },
  { id: 2, name: 'LTerm 北岸影院', city: '北京', address: '海淀区北岸路 18 号', distanceKm: 6.8 },
  { id: 3, name: 'LTerm 南山店', city: '深圳', address: '南山区科苑路 9 号', distanceKm: 8.4 },
]

export const demoMovies: Movie[] = [
  { id: 1, title: '沙丘：预言', coverUrl: cover('dune'), trailerUrl: '', description: '在浩瀚沙海与权力暗涌之间，一场关于信仰、家族与未来的史诗正在展开。', durationMinutes: 166, genre: '科幻 · 剧情', releaseDate: '2026-08-15', rating: 9.2, soldCount: 128, status: 'ON_SALE' },
  { id: 2, title: '哪吒：重生', coverUrl: cover('nezha'), trailerUrl: '', description: '少年与城市的命运交错，东方神话在霓虹与机械中重新点燃。', durationMinutes: 144, genre: '动画 · 奇幻', releaseDate: '2026-08-10', rating: 9.4, soldCount: 116, status: 'ON_SALE' },
  { id: 3, title: '星际回声', coverUrl: cover('echo'), trailerUrl: '', description: '一段来自深空的信号，唤醒了人类关于归途最遥远的想象。', durationMinutes: 128, genre: '科幻 · 冒险', releaseDate: '2026-08-20', rating: 8.7, soldCount: 84, status: 'ON_SALE' },
  { id: 4, title: '午夜放映室', coverUrl: cover('midnight'), trailerUrl: '', description: '每一场午夜电影，都藏着一封写给未眠之人的信。', durationMinutes: 112, genre: '悬疑 · 文艺', releaseDate: '2026-08-01', rating: 8.4, soldCount: 63, status: 'ON_SALE' },
]

export const demoHome: HomeView = { banners: [], hotMovies: demoMovies }

const start = new Date()
start.setHours(19, 0, 0, 0)

export const demoSessions: Session[] = [
  { id: 201, movieId: 1, movieTitle: demoMovies[0].title, cinemaId: 1, cinemaName: demoCinemas[0].name, hallId: 1, hallName: '星河厅', startTime: new Date(start.getTime() + 86400000).toISOString(), endTime: new Date(start.getTime() + 86400000 + 166 * 60000).toISOString(), basePriceCents: 5000, status: 'OPEN', remainingSeats: 42 },
  { id: 202, movieId: 1, movieTitle: demoMovies[0].title, cinemaId: 1, cinemaName: demoCinemas[0].name, hallId: 2, hallName: 'IMAX 厅', startTime: new Date(start.getTime() + 86400000 + 3 * 3600000).toISOString(), endTime: new Date(start.getTime() + 86400000 + 3 * 3600000 + 166 * 60000).toISOString(), basePriceCents: 6800, status: 'OPEN', remainingSeats: 18 },
  { id: 203, movieId: 2, movieTitle: demoMovies[1].title, cinemaId: 2, cinemaName: demoCinemas[1].name, hallId: 3, hallName: '3号厅', startTime: new Date(start.getTime() + 2 * 86400000).toISOString(), endTime: new Date(start.getTime() + 2 * 86400000 + 144 * 60000).toISOString(), basePriceCents: 4500, status: 'OPEN', remainingSeats: 36 },
  { id: 204, movieId: 3, movieTitle: demoMovies[2].title, cinemaId: 1, cinemaName: demoCinemas[0].name, hallId: 1, hallName: '星河厅', startTime: new Date(start.getTime() + 3 * 86400000).toISOString(), endTime: new Date(start.getTime() + 3 * 86400000 + 128 * 60000).toISOString(), basePriceCents: 5200, status: 'OPEN', remainingSeats: 27 },
]

export function createDemoSeats(rows = 8, cols = 10): Seat[] {
  const seats: Seat[] = []
  let id = 1
  for (let row = 1; row <= rows; row += 1) {
    for (let col = 1; col <= cols; col += 1) {
      const disabled = row === 1 && (col === 1 || col === cols)
      const booked = (row === 3 && col === 4) || (row === 4 && col === 5) || (row === 6 && col === 8)
      const locked = row === 5 && col === 3
      seats.push({ seatId: id, rowNo: row, colNo: col, seatNo: `${String.fromCharCode(64 + row)}${col}`, type: row <= 2 ? 'VIP' : 'STANDARD', status: disabled ? 'disabled' : booked ? 'booked' : locked ? 'locked' : 'available' })
      id += 1
    }
  }
  return seats
}

export const demoSeatMap: SeatMapView = { session: demoSessions[0], seats: createDemoSeats(), serverTime: new Date().toISOString() }

export const demoOrder: Order = {
  orderNo: 'LT202608290001', sessionId: demoSessions[0].id, movieId: demoMovies[0].id, movieTitle: demoMovies[0].title, cinemaName: demoSessions[0].cinemaName, hallName: demoSessions[0].hallName, startTime: demoSessions[0].startTime, status: 'PAID', totalCents: 10000, discountCents: 0, couponCents: 500, paidCents: 9500,
  items: [{ seatNo: 'D5', priceCents: 5000, ticketNo: 'TK-AX9Q1' }, { seatNo: 'D6', priceCents: 5000, ticketNo: 'TK-AX9Q2' }],
}

export const demoPoints: PointsView = {
  balance: 2860,
  ledger: [
    { changePoints: 95, balanceAfter: 2860, bizType: 'ORDER_PAID', bizNo: 'LT202608290001', createdAt: new Date().toISOString() },
    { changePoints: 120, balanceAfter: 2765, bizType: 'ORDER_PAID', bizNo: 'LT202608210012', createdAt: new Date(Date.now() - 345600000).toISOString() },
    { changePoints: -1000, balanceAfter: 2645, bizType: 'EXCHANGE', bizNo: 'EX20260801', createdAt: new Date(Date.now() - 950400000).toISOString() },
  ],
}

export const demoCoupons: CouponTemplate[] = [
  { id: 1, name: '新客观影立减券', type: 'FIXED', valueCents: 1500, percentBp: 0, minSpendCents: 6000, maxDiscountCents: 1500, redeemable: true, redeemPoints: 1000, validDays: 30, totalQty: 500, perUserLimit: 1, status: 'ACTIVE' },
  { id: 2, name: '周末 9 折券', type: 'PERCENT', valueCents: 0, percentBp: 9000, minSpendCents: 8000, maxDiscountCents: 2000, redeemable: true, redeemPoints: 1500, validDays: 14, totalQty: 300, perUserLimit: 2, status: 'ACTIVE' },
  { id: 3, name: '午夜场专享券', type: 'FIXED', valueCents: 800, percentBp: 0, minSpendCents: 4000, maxDiscountCents: 800, redeemable: false, redeemPoints: 0, validDays: 7, totalQty: 800, perUserLimit: 1, status: 'ACTIVE' },
]

export const demoHalls: Hall[] = [
  { id: 1, cinemaId: 1, name: '星河厅', seatLayoutJson: '{"rows":8,"cols":10}', status: 'ACTIVE' },
  { id: 2, cinemaId: 1, name: 'IMAX 厅', seatLayoutJson: '{"rows":6,"cols":12}', status: 'ACTIVE' },
  { id: 3, cinemaId: 2, name: '3号厅', seatLayoutJson: '{"rows":8,"cols":10}', status: 'MAINTENANCE' },
]

export const demoDashboardSummary: DashboardSummary = { orderCount: 486, ticketCount: 722, grossCents: 3826800, refundCents: 126000, netCents: 3700800 }

export const demoDashboardTrend: DashboardTrendRow[] = Array.from({ length: 7 }, (_, index) => {
  const date = new Date()
  date.setDate(date.getDate() - (6 - index))
  const tickets = 70 + index * 11
  const gross = tickets * 5200
  const refund = index === 3 ? 18000 : index === 5 ? 9000 : 0
  return { date: date.toISOString(), orderCount: Math.round(tickets * 0.68), ticketCount: tickets, grossCents: gross, refundCents: refund, netCents: gross - refund }
})

export const demoDashboardMovies: DashboardMovieRow[] = demoMovies.map((movie, index) => ({ movieId: movie.id, movieTitle: movie.title, orderCount: 118 - index * 19, grossCents: 960000 - index * 132000, netCents: 930000 - index * 126000 }))
export const demoDashboardCinemas: DashboardCinemaRow[] = [
  { cinemaId: 1, cinemaName: demoCinemas[0].name, orderCount: 318, grossCents: 2600000, netCents: 2520000 },
  { cinemaId: 2, cinemaName: demoCinemas[1].name, orderCount: 168, grossCents: 1226800, netCents: 1180800 },
]
