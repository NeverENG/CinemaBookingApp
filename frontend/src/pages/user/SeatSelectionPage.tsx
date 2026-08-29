import { ArrowLeft, Check, Clock3, Info, RefreshCw } from 'lucide-react'
import { useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useAuth } from '../../app/providers'
import { SeatMap } from '../../components/SeatMap'
import { Button, DemoBadge, EmptyState, StatusBadge } from '../../components/ui'
import { useCreateOrderMutation, useSeatMapQuery } from '../../features/booking/hooks'
import { useAsyncLock } from '../../hooks/useAsyncLock'
import { money, timeText } from '../../lib/format'
import { AppError } from '../../services/http/errors'
import type { Seat } from '../../types'

export function SeatSelectionPage() {
  const { sessionId: rawSessionId } = useParams()
  const sessionId = Number(rawSessionId)
  const navigate = useNavigate()
  const { session: auth } = useAuth()
  const seatMapQuery = useSeatMapQuery(sessionId)
  const createOrder = useCreateOrderMutation()
  const { run, isLocked } = useAsyncLock()
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const seatMap = seatMapQuery.data
  const selectedSeats = useMemo(() => seatMap?.seats.filter((seat) => selectedIds.includes(seat.seatId)) ?? [], [seatMap?.seats, selectedIds])
  const price = selectedSeats.length * (seatMap?.session.basePriceCents || 5000)

  if (!seatMap) return <div className="content-container"><EmptyState title="座位图加载中" description="正在同步当前场次的座位状态。" /></div>

  function toggleSeat(seat: Seat) {
    setSelectedIds((current) => current.includes(seat.seatId) ? current.filter((id) => id !== seat.seatId) : current.length >= 6 ? current : [...current, seat.seatId])
  }

  async function continueBooking() {
    if (selectedSeats.length === 0) return
    if (!auth) {
      navigate(`/login?redirect=${encodeURIComponent(`/sessions/${sessionId}/seats`)}`)
      return
    }
    await run(async () => {
      try {
        const result = await createOrder.mutateAsync({ session_id: sessionId, seat_ids: selectedIds })
        navigate(`/checkout?order=${encodeURIComponent(result.order_no)}`, { state: { createOrder: result } })
      } catch (cause) {
        if (cause instanceof AppError && cause.kind === 'conflict') {
          window.alert('座位状态已变化，请刷新后重新选择。')
          void seatMapQuery.refetch()
        } else {
          window.alert(cause instanceof AppError ? cause.message : '订单创建失败')
        }
      }
    })
  }

  return <div className="content-container seat-page"><div className="booking-steps"><span className="complete"><Check size={14} />01 场次</span><span className="active">02 选座</span><span>03 支付</span></div><div className="seat-heading"><div><Link className="back-link" to={`/movies/${seatMap.session.movieId}`}><ArrowLeft size={15} />返回场次</Link><h1>{seatMap.session.movieTitle || '选择座位'}</h1><p>{seatMap.session.hallName} · {timeText(seatMap.session.startTime)} 开场</p></div><div className="seat-sync"><RefreshCw size={14} className={seatMapQuery.isFetching ? 'spin' : ''} />每 5 秒同步</div></div>{seatMapQuery.isDemo && <div className="demo-strip"><DemoBadge /> 座位图暂未连接服务端，当前展示交互演示</div>}<div className="seat-layout"><section className="seat-stage"><SeatMap seats={seatMap.seats} selectedIds={selectedIds} onToggle={toggleSeat} /><div className="seat-legend"><span><i className="legend-seat available" />可选</span><span><i className="legend-seat selected" />已选</span><span><i className="legend-seat locked" />已锁</span><span><i className="legend-seat booked" />已售</span></div></section><aside className="seat-summary"><div className="summary-kicker"><Clock3 size={15} />订单将在创建后保留 15 分钟</div><div><span className="summary-label">当前场次</span><strong>{seatMap.session.hallName}</strong><span>{timeText(seatMap.session.startTime)} · {money(seatMap.session.basePriceCents)} / 座</span></div><div className="selected-seat-summary"><span className="summary-label">已选座位 ({selectedSeats.length}/6)</span>{selectedSeats.length > 0 ? <div className="selected-seat-list">{selectedSeats.map((seat) => <span key={seat.seatId}>{seat.seatNo}</span>)}</div> : <p>点击座位图选择位置</p>}</div><div className="summary-total"><span>预计应付</span><strong>{money(price)}</strong></div><Button size="lg" disabled={selectedSeats.length === 0 || isLocked || createOrder.isPending} onClick={continueBooking}>{isLocked || createOrder.isPending ? '创建订单中…' : '锁座并继续'}<Check size={16} /></Button><div className="summary-tip"><Info size={14} />最终价格与座位状态以服务端校验为准。</div><StatusBadge status={seatMap.session.status} /></aside></div></div>
}
