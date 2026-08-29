import { ArrowLeft, CalendarDays, CheckCircle2, CircleAlert, RefreshCw, Ticket } from 'lucide-react'
import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Button, DemoBadge, EmptyState, PageHeader, StatusBadge } from '../../components/ui'
import { useOrderQuery, useRefundMutation } from '../../features/order/hooks'
import { dateText, money, timeText } from '../../lib/format'
import { AppError } from '../../services/http/errors'

export function OrderDetailPage() {
  const { orderNo = '' } = useParams()
  const orderQuery = useOrderQuery(orderNo)
  const refundMutation = useRefundMutation(orderNo)
  const [reason, setReason] = useState('行程有变')
  const order = orderQuery.data

  if (!order) return <div className="content-container"><EmptyState title="订单不存在" description="请检查订单号或返回订单列表。" action={<Link className="button button-secondary" to="/orders">返回订单</Link>} /></div>

  async function refund() {
    if (!window.confirm('确认申请整单退款吗？退款成功后座位将重新释放。')) return
    try {
      await refundMutation.mutateAsync(reason)
      window.alert('退款申请已提交')
    } catch (cause) {
      window.alert(cause instanceof AppError ? cause.message : '退款申请失败')
    }
  }

  const canRefund = order.status === 'PAID'
  return <div className="content-container narrow-container"><Link className="back-link" to="/orders"><ArrowLeft size={15} />返回订单</Link><PageHeader eyebrow="ORDER DETAIL" title={order.movieTitle || '订单详情'} actions={<StatusBadge status={order.status} />} />{orderQuery.isDemo && <div className="demo-strip"><DemoBadge /> 当前展示演示订单数据</div>}<div className="detail-order-grid"><section className="order-detail-main"><div className="order-info-head"><div className="checkout-movie-icon"><Ticket size={19} /></div><div><strong>{order.cinemaName} · {order.hallName}</strong><span><CalendarDays size={14} /> {dateText(order.startTime)} {timeText(order.startTime)} 开场</span></div></div><div className="order-detail-section"><div className="section-label">座位与取票码</div>{order.items.map((item) => <div className="ticket-row" key={item.seatNo}><div><strong>{item.seatNo}</strong><span>{money(item.priceCents)}</span></div><span className="ticket-code">{item.ticketNo || '支付后生成'}</span></div>)}</div><div className="order-detail-section"><div className="section-label">状态流转</div><div className="order-timeline"><div className="timeline-item done"><CheckCircle2 size={17} /><div><strong>订单已创建</strong><span>座位已暂时锁定</span></div></div><div className={`timeline-item ${order.status === 'PAID' || order.status === 'COMPLETED' ? 'done' : ''}`}><CheckCircle2 size={17} /><div><strong>支付状态</strong><span>{order.status === 'PAID' ? '支付成功并完成出票' : '等待支付确认'}</span></div></div><div className={`timeline-item ${order.status === 'REFUNDED' ? 'done' : ''}`}><RefreshCw size={17} /><div><strong>售后状态</strong><span>{order.status === 'REFUNDED' ? '退款已完成' : '开场前支持整单退款'}</span></div></div></div></div></section><aside className="order-detail-side"><div className="price-line"><span>票面合计</span><strong>{money(order.totalCents)}</strong></div><div className="price-line"><span>优惠</span><strong>-{money(order.discountCents + order.couponCents)}</strong></div><div className="price-total"><span>实付金额</span><strong>{money(order.paidCents)}</strong></div>{canRefund && <><label className="field"><span className="field-label">退款原因</span><select className="input" value={reason} onChange={(event) => setReason(event.target.value)}><option>行程有变</option><option>误购场次</option><option>其他原因</option></select></label><Button variant="danger" onClick={refund} disabled={refundMutation.isPending}>{refundMutation.isPending ? '提交中…' : '申请退款'}</Button></>}{!canRefund && <div className="summary-tip"><CircleAlert size={14} />当前状态不支持退款操作。</div>}<span className="order-number">订单号：{order.orderNo}</span></aside></div></div>
}
