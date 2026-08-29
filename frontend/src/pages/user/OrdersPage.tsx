import { ArrowRight, ClipboardList } from 'lucide-react'
import { Link } from 'react-router-dom'
import { DemoBadge, EmptyState, PageHeader, StatusBadge } from '../../components/ui'
import { useOrdersQuery } from '../../features/order/hooks'
import { dateText, money, timeText } from '../../lib/format'

export function OrdersPage() {
  const query = useOrdersQuery()
  const orders = query.data ?? []
  return <div className="content-container narrow-container"><PageHeader eyebrow="ORDERS" title="我的订单" description="查看购票、支付和退款状态。" />{query.isDemo && <div className="demo-strip"><DemoBadge /> 后端暂不可用，当前展示演示订单</div>}{query.error && <div className="form-error">订单加载失败：{query.error.message}</div>}{orders.length > 0 ? <div className="order-list">{orders.map((order) => <article className="order-card" key={order.orderNo}><div className="order-card-top"><div><span className="eyebrow">{dateText(order.startTime)}</span><h2>{order.movieTitle || '电影场次'}</h2></div><StatusBadge status={order.status} /></div><div className="order-card-meta"><span>{order.cinemaName} · {order.hallName}</span><span>{timeText(order.startTime)} · {order.items.map((item) => item.seatNo).join('、')}</span><strong>{money(order.paidCents)}</strong></div><div className="order-card-bottom"><span>订单号 {order.orderNo}</span><Link className="text-link" to={`/orders/${encodeURIComponent(order.orderNo)}`}>查看详情 <ArrowRight size={15} /></Link></div></article>)}</div> : <EmptyState icon={ClipboardList} title="还没有订单" description="完成一次购票后，你的订单会按时间顺序出现在这里。" action={<Link className="button button-secondary" to="/">去看热映</Link>} />}</div>
}
