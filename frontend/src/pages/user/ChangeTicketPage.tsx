import { ArrowLeft, Check, Clock3, RefreshCw, Ticket } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { SeatMap } from '../../components/SeatMap'
import { Button, DemoBadge, EmptyState, PageHeader, StatusBadge } from '../../components/ui'
import { useSeatMapQuery } from '../../features/booking/hooks'
import { useSessionQuery } from '../../features/catalog/hooks'
import { useChangeOrderMutation, useOrderQuery } from '../../features/order/hooks'
import { useAsyncLock } from '../../hooks/useAsyncLock'
import { dateText, money, timeText } from '../../lib/format'
import { AppError } from '../../services/http/errors'
import type { Seat } from '../../types'

export function ChangeTicketPage() {
  const { orderNo = '' } = useParams()
  const navigate = useNavigate()
  const orderQuery = useOrderQuery(orderNo)
  const order = orderQuery.data
  const sessionsQuery = useSessionQuery(order?.movieId, order?.cinemaId)
  const sessions = useMemo(() => (sessionsQuery.data ?? []).filter((session) => session.id !== order?.sessionId && new Date(session.startTime).getTime() > Date.now()), [order?.sessionId, sessionsQuery.data])
  const [sessionId, setSessionId] = useState(0)
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const seatMapQuery = useSeatMapQuery(sessionId)
  const change = useChangeOrderMutation(orderNo)
  const { run, isLocked } = useAsyncLock()
  const seatMap = seatMapQuery.data
  const selectedSeats = seatMap?.seats.filter((seat) => selectedIds.includes(seat.seatId)) ?? []
  const newPaidCents = selectedSeats.reduce((total, seat) => total + (seat.priceCents || seatMap?.session.basePriceCents || 0), 0)

  useEffect(() => {
    if (!sessionId && sessions.length > 0) setSessionId(sessions[0].id)
  }, [sessionId, sessions])

  if (!order) return <div className="content-container"><EmptyState title="订单不存在" description="请返回订单列表重新选择。" action={<Link className="button button-secondary" to="/orders">返回订单</Link>} /></div>

  const currentOrder = order
  const usedCount = currentOrder.items.filter((item) => item.usedAt).length
  const invalidOrder = currentOrder.status !== 'PAID' || usedCount > 0
  const demoMode = orderQuery.isDemo || sessionsQuery.isDemo || seatMapQuery.isDemo

  function toggleSeat(seat: Seat) {
    setSelectedIds((current) => current.includes(seat.seatId) ? current.filter((id) => id !== seat.seatId) : current.length >= currentOrder.items.length ? current : [...current, seat.seatId])
  }

  async function submitChange() {
    if (!seatMap || selectedIds.length !== currentOrder.items.length || invalidOrder || demoMode) return
    await run(async () => {
      try {
        const result = await change.mutateAsync({ new_session_id: sessionId, new_seat_ids: selectedIds })
        window.alert(`改签成功，原订单已退款：${result.refundNo}`)
        navigate(`/orders/${encodeURIComponent(result.newOrderNo)}`)
      } catch (cause) {
        window.alert(cause instanceof AppError ? cause.message : '改签失败，请稍后重试')
        void seatMapQuery.refetch()
      }
    })
  }

  return <div className="content-container seat-page change-ticket-page">
    <Link className="back-link" to={`/orders/${encodeURIComponent(orderNo)}`}><ArrowLeft size={15} />返回订单详情</Link>
    <PageHeader eyebrow="CHANGE TICKET" title="改签" description="选择同一影片的新场次和座位，系统会自动完成原单退款。" />
    {orderQuery.isDemo && <div className="demo-strip"><DemoBadge /> 当前为演示订单，服务端连接后才可提交改签</div>}
    {invalidOrder && <div className="form-error">当前订单已核销、已退款或不是已支付状态，不能改签。</div>}
    <section className="change-origin-card"><div className="checkout-movie-icon"><Ticket size={20} /></div><div><strong>{currentOrder.movieTitle || '电影场次'}</strong><span>{currentOrder.cinemaName} · {currentOrder.hallName} · {dateText(currentOrder.startTime)} {timeText(currentOrder.startTime)}</span></div><StatusBadge status={currentOrder.status} /></section>
    {sessions.length === 0 ? <EmptyState title="暂无可改签场次" description="当前影片没有其他可售场次，或原场次已临近开场。" /> : <div className="change-ticket-layout">
      <section className="change-session-panel"><div className="change-section-heading"><div><span className="eyebrow">01 / SHOWTIME</span><h2>选择新场次</h2></div><span>{sessions.length} 个可选场次</span></div><div className="change-session-list">{sessions.map((session) => <button type="button" className={`change-session-option ${session.id === sessionId ? 'active' : ''}`} key={session.id} onClick={() => { setSessionId(session.id); setSelectedIds([]) }}><span><strong>{dateText(session.startTime)}</strong><small>{timeText(session.startTime)} 开场 · {session.hallName}</small></span><span><b>{money(session.basePriceCents)}</b><StatusBadge status={session.status} /></span></button>)}</div></section>
      {seatMap ? <section className="change-seat-panel"><div className="change-section-heading"><div><span className="eyebrow">02 / SEATS</span><h2>重新选择座位</h2></div><span>{selectedIds.length} / {currentOrder.items.length} 张</span></div><div className="change-seat-map"><SeatMap seats={seatMap.seats} selectedIds={selectedIds} onToggle={toggleSeat} /></div><div className="seat-legend"><span><i className="legend-seat available" />可选</span><span><i className="legend-seat selected" />已选</span><span><i className="legend-seat locked" />已锁</span><span><i className="legend-seat booked" />已售</span></div></section> : <EmptyState title="座位图加载中" description="正在同步新场次的实时座位状态。" />}
      <aside className="change-ticket-summary"><div className="summary-kicker"><Clock3 size={15} />改签后自动支付新订单</div><div className="selected-seat-summary"><span className="summary-label">已选座位</span>{selectedIds.length > 0 ? <div className="selected-seat-list">{selectedSeats.map((seat) => <span key={seat.seatId}>{seat.seatNo} · {money(seat.priceCents || seatMap?.session.basePriceCents || 0)}</span>)}</div> : <p>请选择与原订单相同数量的座位</p>}</div><div className="summary-total"><span>新订单预计应付</span><strong>{money(newPaidCents)}</strong></div><Button size="lg" disabled={!seatMap || selectedIds.length !== currentOrder.items.length || invalidOrder || demoMode || isLocked || change.isPending} onClick={submitChange}>{isLocked || change.isPending ? '改签处理中…' : '确认改签'}<RefreshCw size={16} /></Button><div className="summary-tip"><Check size={14} />原订单退款、新订单支付会在一次操作中完成。</div></aside>
    </div>}
  </div>
}
