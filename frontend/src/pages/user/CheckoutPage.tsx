import { ArrowRight, CheckCircle2, Clock3, Ticket } from 'lucide-react'
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom'
import { Button, DemoBadge, EmptyState, PageHeader } from '../../components/ui'
import { useOrderQuery } from '../../features/order/hooks'
import { money, timeText } from '../../lib/format'
import { useAsyncLock } from '../../hooks/useAsyncLock'
import { useCreatePaymentMutation, useMockPayMutation } from '../../features/order/hooks'
import type { CreateOrderResult } from '../../types'

export function CheckoutPage() {
  const [params] = useSearchParams()
  const orderNo = params.get('order') ?? ''
  const location = useLocation()
  const navigate = useNavigate()
  const orderQuery = useOrderQuery(orderNo)
  const createPayment = useCreatePaymentMutation()
  const mockPay = useMockPayMutation()
  const { run, isLocked } = useAsyncLock()
  const created = (location.state as { createOrder?: CreateOrderResult } | null)?.createOrder
  const order = orderQuery.data

  async function pay() {
    if (!order) return
    await run(async () => {
      if (orderQuery.isDemo) {
        navigate(`/payment/${encodeURIComponent(order.orderNo)}?demo=1`)
        return
      }
      try {
        const payment = await createPayment.mutateAsync(order.orderNo)
        await mockPay.mutateAsync(payment.transactionNo)
        navigate(`/payment/${encodeURIComponent(order.orderNo)}`)
      } catch (cause) {
        window.alert(cause instanceof Error ? cause.message : '支付发起失败')
      }
    })
  }

  if (!order) return <div className="content-container"><EmptyState title="订单不存在" description="返回选座页重新创建订单。" /></div>

  return <div className="content-container narrow-container"><PageHeader eyebrow="CHECKOUT" title="确认订单" description="确认场次和座位后，进入模拟支付。" />{orderQuery.isDemo && <div className="demo-strip"><DemoBadge /> 当前为演示订单，不会产生真实支付</div>}<div className="checkout-grid"><section className="checkout-main"><div className="checkout-movie"><div className="checkout-movie-icon"><Ticket size={20} /></div><div><strong>{order.movieTitle || '电影场次'}</strong><span>{order.cinemaName} · {order.hallName}</span><span>{timeText(order.startTime)} 开场</span></div></div><div className="checkout-block"><div className="checkout-block-heading"><strong>座位</strong><span>{created?.seat_nos?.length ?? order.items.length} 张票</span></div><div className="checkout-seats">{created?.seat_nos?.map((seat) => <span key={seat}>{seat}</span>) ?? order.items.map((item) => <span key={item.seatNo}>{item.seatNo}</span>)}</div></div><div className="checkout-block"><div className="checkout-block-heading"><strong>支付方式</strong><span>当前为 Mock 支付</span></div><div className="mock-payment"><div className="mock-payment-icon"><CheckCircle2 size={20} /></div><div><strong>模拟支付</strong><span>用于演示支付回调、订单状态与出票流程</span></div><span className="payment-radio" /></div></div></section><aside className="checkout-summary"><div className="summary-kicker"><Clock3 size={15} />支付有效期 15 分钟</div><div className="price-line"><span>票面合计</span><strong>{money(order.totalCents || order.paidCents)}</strong></div><div className="price-line"><span>优惠</span><strong>-{money((order.discountCents || 0) + (order.couponCents || 0))}</strong></div><div className="price-total"><span>实付金额</span><strong>{money(created?.paid_cents ?? order.paidCents)}</strong></div><Button size="lg" onClick={pay} disabled={isLocked || createPayment.isPending || mockPay.isPending}>{isLocked || createPayment.isPending || mockPay.isPending ? '支付处理中…' : '确认模拟支付'}<ArrowRight size={16} /></Button><p className="summary-tip">点击后会完整触发支付交易与回调链路。</p></aside></div></div>
}
