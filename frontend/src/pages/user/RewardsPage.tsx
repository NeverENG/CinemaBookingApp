import { ArrowUpRight, BadgePercent, Crown, Gift, History, Sparkles } from 'lucide-react'
import { useState } from 'react'
import { Button, DemoBadge, EmptyState, PageHeader, Panel } from '../../components/ui'
import { useExchangeCouponMutation, usePointsQuery, useRedeemableCouponsQuery } from '../../features/rewards/hooks'
import { money } from '../../lib/format'
import { AppError } from '../../services/http/errors'

const levels = [{ name: '青铜', min: 0, discount: '原价' }, { name: '白银', min: 1000, discount: '95 折' }, { name: '黄金', min: 5000, discount: '9 折' }, { name: '钻石', min: 20000, discount: '85 折' }]

export function RewardsPage() {
  const pointsQuery = usePointsQuery()
  const couponsQuery = useRedeemableCouponsQuery()
  const exchange = useExchangeCouponMutation()
  const [message, setMessage] = useState('')
  const balance = pointsQuery.data?.balance ?? 0
  const currentLevel = [...levels].reverse().find((level) => balance >= level.min) ?? levels[0]
  const nextLevel = levels.find((level) => level.min > balance)
  const progress = nextLevel ? Math.min(100, Math.round(((balance - currentLevel.min) / (nextLevel.min - currentLevel.min)) * 100)) : 100

  async function redeem(templateId: number) {
    setMessage('')
    try {
      const result = await exchange.mutateAsync(templateId)
      setMessage(`兑换成功，优惠券编号：${String(result.coupon_no ?? result)}`)
    } catch (cause) {
      setMessage(cause instanceof AppError ? cause.message : '兑换失败')
    }
  }

  return <div className="content-container narrow-container"><PageHeader eyebrow="MEMBERSHIP" title="积分与会员" description="购票获得积分，兑换下一次观影的优惠。" />{(pointsQuery.isDemo || couponsQuery.isDemo) && <div className="demo-strip"><DemoBadge /> 当前展示会员演示数据</div>}<div className="reward-overview"><div className="points-balance"><span>当前积分</span><strong>{balance.toLocaleString()}</strong><small>积分是观影赠送值，可兑换优惠券</small></div><div className="level-progress"><div className="level-heading"><div><span className="eyebrow">CURRENT LEVEL</span><strong><Crown size={16} />{currentLevel.name}会员</strong></div><span>{currentLevel.discount}</span></div><div className="progress-track"><span style={{ width: `${progress}%` }} /></div><small>{nextLevel ? `再获得 ${nextLevel.min - balance} 积分升级为${nextLevel.name}` : '已达到最高会员等级'}</small></div></div><div className="reward-grid"><Panel title="兑换优惠券" action={<BadgePercent size={17} />}>{message && <div className="form-success">{message}</div>}<div className="coupon-grid">{(couponsQuery.data ?? []).filter((coupon) => coupon.redeemable).map((coupon) => <article className="reward-coupon" key={coupon.id}><div className="coupon-value">{coupon.type === 'FIXED' ? money(coupon.valueCents) : `${coupon.percentBp / 100}折`}</div><div className="coupon-copy"><strong>{coupon.name}</strong><span>满 {money(coupon.minSpendCents)} 可用 · {coupon.validDays} 天有效</span></div><Button size="sm" variant="secondary" onClick={() => redeem(coupon.id)} disabled={exchange.isPending}>{coupon.redeemPoints.toLocaleString()} 积分兑换</Button></article>)}</div></Panel><Panel title="积分流水" action={<History size={17} />}>{(pointsQuery.data?.ledger ?? []).length === 0 ? <EmptyState icon={History} title="还没有积分记录" /> : <div className="ledger-list">{pointsQuery.data?.ledger.map((item) => <div className="ledger-row" key={`${item.bizType}-${item.bizNo}`}><div><strong>{item.bizType === 'EXCHANGE' ? '兑换优惠券' : '购票赠送'}</strong><span>{item.bizNo}</span></div><strong className={item.changePoints > 0 ? 'positive' : 'negative'}>{item.changePoints > 0 ? '+' : ''}{item.changePoints}</strong></div>)}</div>}</Panel></div><div className="reward-note"><Gift size={16} /><span>等级只升不降，退款会按比例扣回当前积分余额，不影响已获得的会员等级。</span><ArrowUpRight size={15} /></div><Sparkles className="page-watermark" size={120} /></div>
}
