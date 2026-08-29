import { ArrowDownRight, ArrowUpRight, BarChart3, CircleDollarSign, FileCheck2, RefreshCw, Ticket, TrendingUp } from 'lucide-react'
import { lazy, Suspense, useMemo, useState } from 'react'
import { useAuth } from '../../app/providers'
import { Button, DemoBadge, MetricCard, PageHeader, Panel } from '../../components/ui'
import { useDashboardQueries, useReconcileMutation } from '../../features/admin/hooks'
import { money } from '../../lib/format'
import type { DashboardTrendRow } from '../../types'
import type { EChartsOption } from 'echarts'

const Chart = lazy(() => import('../../components/Chart').then(({ Chart: page }) => ({ default: page })))

function dateValue(date: Date) {
  return date.toISOString().slice(0, 10)
}

export function DashboardPage() {
  const { session } = useAuth()
  const today = new Date()
  const weekAgo = new Date(today)
  weekAgo.setDate(today.getDate() - 6)
  const [filters, setFilters] = useState({ start_date: dateValue(weekAgo), end_date: dateValue(today), granularity: 'day' })
  const [draft, setDraft] = useState(filters)
  const queries = useDashboardQueries(filters)
  const reconcile = useReconcileMutation()
  const summary = queries.summary.data ?? { orderCount: 0, ticketCount: 0, grossCents: 0, refundCents: 0, netCents: 0 }
  const trend = queries.trend.data ?? []
  const chartOption = useMemo(() => buildChartOption(trend), [trend])

  function applyFilters() {
    setFilters(draft)
  }

  return (
    <div className="admin-page">
      <PageHeader
        eyebrow="OVERVIEW"
        title="票房大盘"
        description={session?.role === 'CINEMA_ADMIN' ? '当前展示绑定影院的经营数据。' : '追踪订单、出票与净票房的变化趋势。'}
        actions={<div className="dashboard-actions">{(queries.summary.isDemo || queries.trend.isDemo) && <DemoBadge />}{session?.role === 'SUPER_ADMIN' && <Button variant="secondary" size="sm" onClick={() => reconcile.mutate()} disabled={reconcile.isPending}><RefreshCw size={14} className={reconcile.isPending ? 'spin' : ''} />{reconcile.isPending ? '对账中…' : '重建聚合'}</Button>}</div>}
      />
      <Panel className="dashboard-filters">
        <div className="filter-fields">
          <label className="field"><span className="field-label">开始日期</span><input className="input" type="date" value={draft.start_date} onChange={(event) => setDraft({ ...draft, start_date: event.target.value })} /></label>
          <label className="field"><span className="field-label">结束日期</span><input className="input" type="date" value={draft.end_date} onChange={(event) => setDraft({ ...draft, end_date: event.target.value })} /></label>
          <label className="field"><span className="field-label">统计粒度</span><select className="input select" value={draft.granularity} onChange={(event) => setDraft({ ...draft, granularity: event.target.value })}><option value="day">按日</option><option value="week">按周</option><option value="month">按月</option></select></label>
          <Button variant="primary" size="sm" onClick={applyFilters}>更新数据</Button>
        </div>
        <div className="scope-note"><span className="status-dot" />数据口径：已支付订单，退款后计算净票房</div>
      </Panel>
      <div className="metric-grid admin-metrics">
        <MetricCard label="支付订单" value={summary.orderCount.toLocaleString()} note="统计区间内" icon={FileCheck2} tone="default" />
        <MetricCard label="出票数" value={summary.ticketCount.toLocaleString()} note="张" icon={Ticket} tone="gold" />
        <MetricCard label="总票房" value={money(summary.grossCents)} note="支付成功金额" icon={CircleDollarSign} tone="default" />
        <MetricCard label="退款额" value={money(summary.refundCents)} note="已完成退款" icon={ArrowDownRight} tone="red" />
        <MetricCard label="净票房" value={money(summary.netCents)} note="总票房 - 退款额" icon={TrendingUp} tone="green" />
      </div>
      <div className="dashboard-grid">
        <Panel title="票房趋势" action={<span className="panel-note"><span className="chart-dot gold" />总票房 <span className="chart-dot green" />净票房</span>}>
          <Suspense fallback={<div className="chart route-loading">加载图表中…</div>}><Chart option={chartOption} /></Suspense>
        </Panel>
        <Panel title="电影排行" action={<BarChart3 size={17} />}>
          <div className="ranking-list">{(queries.movies.data ?? []).slice(0, 5).map((movie, index) => <div className="ranking-row" key={movie.movieId}><span className="ranking-index">{String(index + 1).padStart(2, '0')}</span><div className="ranking-copy"><strong>{movie.movieTitle || '未命名影片'}</strong><span>{movie.orderCount} 个订单</span></div><strong>{money(movie.netCents)}</strong><ArrowUpRight size={14} /></div>)}</div>
        </Panel>
      </div>
      <div className="dashboard-grid dashboard-grid-bottom">
        <Panel title="影院排行">
          <div className="ranking-list">{(queries.cinemas.data ?? []).map((cinema, index) => <div className="ranking-row" key={cinema.cinemaId}><span className="ranking-index">{String(index + 1).padStart(2, '0')}</span><div className="ranking-copy"><strong>{cinema.cinemaName}</strong><span>{cinema.orderCount} 个订单</span></div><strong>{money(cinema.netCents)}</strong><ArrowUpRight size={14} /></div>)}</div>
        </Panel>
        <Panel title="数据说明">
          <div className="data-note-list">
            <div><span className="data-note-icon"><FileCheck2 size={15} /></span><p><strong>订单口径</strong><span>支付成功订单计入，退款成功后扣减。</span></p></div>
            <div><span className="data-note-icon"><RefreshCw size={15} /></span><p><strong>聚合机制</strong><span>事实流水与日聚合表保持一致。</span></p></div>
            <div><span className="data-note-icon"><BarChart3 size={15} /></span><p><strong>数据范围</strong><span>当前角色只能访问授权影院数据。</span></p></div>
          </div>
        </Panel>
      </div>
    </div>
  )
}

function buildChartOption(rows: DashboardTrendRow[]): EChartsOption {
  const labels = rows.map((row) => row.date.slice(5, 10))
  return {
    color: ['#caa45b', '#6cb18d'],
    tooltip: { trigger: 'axis', valueFormatter: (value: unknown) => money(Number(value)) },
    grid: { left: 12, right: 18, top: 28, bottom: 20, containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: labels, axisLine: { lineStyle: { color: '#343d47' } }, axisLabel: { color: '#85909c' } },
    yAxis: { type: 'value', splitLine: { lineStyle: { color: '#26303a' } }, axisLabel: { color: '#85909c', formatter: (value: number) => `¥${Math.round(value / 100)}` } },
    series: [
      { name: '总票房', type: 'line', smooth: true, symbol: 'none', data: rows.map((row) => row.grossCents), areaStyle: { color: 'rgba(202,164,91,0.08)' } },
      { name: '净票房', type: 'line', smooth: true, symbol: 'none', data: rows.map((row) => row.netCents), areaStyle: { color: 'rgba(108,177,141,0.07)' } },
    ],
  }
}
