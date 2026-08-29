import { Check, Clock3, LoaderCircle, Ticket, XCircle } from 'lucide-react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { DemoBadge, EmptyState } from '../../components/ui'
import { useOrderQuery, usePaymentQuery } from '../../features/order/hooks'
import { dateText, money, timeText } from '../../lib/format'

export function PaymentPage() {
  const { orderNo = '' } = useParams()
  const [params] = useSearchParams()
  const orderQuery = useOrderQuery(orderNo)
  const paymentQuery = usePaymentQuery(orderNo)
  const order = orderQuery.data
  const demo = params.get('demo') === '1' || orderQuery.isDemo || paymentQuery.isDemo
  const success = order?.status === 'PAID' || order?.status === 'COMPLETED' || paymentQuery.data?.status.toUpperCase() === 'SUCCESS'
  const failed = order?.status === 'CANCELED' || order?.status === 'EXPIRED'

  if (!order) return <div className="content-container"><EmptyState title="正在确认订单" description="支付状态同步后会自动显示结果。" /></div>

  return <div className="content-container narrow-container payment-result-page">{demo && <div className="demo-strip"><DemoBadge /> 当前为支付链路演示</div>}<div className={`payment-result ${success ? 'result-success' : failed ? 'result-failed' : 'result-pending'}`}>{success ? <div className="result-icon"><Check size={30} /></div> : failed ? <div className="result-icon"><XCircle size={30} /></div> : <div className="result-icon"><LoaderCircle className="spin" size={30} /></div>}<span className="eyebrow">{success ? 'PAYMENT COMPLETE' : failed ? 'PAYMENT FAILED' : 'PROCESSING'}</span><h1>{success ? '购票成功' : failed ? '支付未完成' : '正在确认支付'}</h1><p>{success ? '支付回调已完成，座位已经为你保留。' : failed ? '订单已结束，可以返回重新选择场次。' : '请不要关闭页面，系统正在等待支付回调。'}</p>{!success && !failed && <div className="processing-note"><Clock3 size={15} />状态每 3 秒自动同步</div>}</div><section className="ticket-result-card"><div className="ticket-result-heading"><div className="checkout-movie-icon"><Ticket size={19} /></div><div><strong>{order.movieTitle || '电影场次'}</strong><span>{order.cinemaName} · {order.hallName}</span></div><span className="ticket-order-no">{order.orderNo}</span></div><div className="ticket-result-meta"><div><small>观影时间</small><strong>{dateText(order.startTime)} {timeText(order.startTime)}</strong></div><div><small>座位</small><strong>{order.items.map((item) => item.seatNo).join('、') || '待同步'}</strong></div><div><small>实付金额</small><strong>{money(order.paidCents)}</strong></div></div>{success && <div className="ticket-codes"><div className="section-label">取票码</div>{order.items.map((item) => <span key={item.seatNo}>{item.ticketNo || '待生成'}</span>)}</div>}</section><div className="result-actions"><Link className="button button-primary" to={`/orders/${encodeURIComponent(order.orderNo)}`}>查看订单</Link><Link className="button button-secondary" to="/">返回热映</Link></div></div>
}
