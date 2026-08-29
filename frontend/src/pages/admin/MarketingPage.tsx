import { Gift, Megaphone, Plus } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { Button, DemoBadge, Field, Input, Modal, PageHeader, Select, StatusBadge } from '../../components/ui'
import { useAdminBannersQuery, useAdminCouponsQuery, useCreateBannerMutation, useCreateCouponMutation } from '../../features/admin/hooks'
import { money } from '../../lib/format'

export function MarketingPage() {
  const [tab, setTab] = useState<'coupon' | 'banner'>('coupon')
  const [open, setOpen] = useState(false)
  return <div className="admin-page"><PageHeader eyebrow="MARKETING" title="运营内容" description="管理优惠券模板与首页运营位，所有操作保留清晰反馈。" actions={<Button onClick={() => setOpen(true)}><Plus size={15} />{tab === 'coupon' ? '新增优惠券' : '新增运营位'}</Button>} /><div className="tab-switcher"><button className={tab === 'coupon' ? 'active' : ''} onClick={() => setTab('coupon')}><Gift size={15} />优惠券模板</button><button className={tab === 'banner' ? 'active' : ''} onClick={() => setTab('banner')}><Megaphone size={15} />运营位</button></div>{tab === 'coupon' ? <CouponPanel /> : <BannerPanel />}{tab === 'coupon' ? <CouponModal open={open} onClose={() => setOpen(false)} /> : <BannerModal open={open} onClose={() => setOpen(false)} />}</div>
}

function CouponPanel() {
  const query = useAdminCouponsQuery()
  return <div className="admin-table-card"><div className="table-card-heading"><span>{query.isDemo && <DemoBadge />}</span><span>{(query.data ?? []).length} 个模板</span></div><table className="admin-table"><thead><tr><th>名称</th><th>类型</th><th>兑换</th><th>有效期</th><th>状态</th><th>库存</th></tr></thead><tbody>{(query.data ?? []).map((coupon) => <tr key={coupon.id}><td><strong>{coupon.name}</strong><span className="table-subtext">满 {money(coupon.minSpendCents)} 可用</span></td><td>{coupon.type === 'FIXED' ? `立减 ${money(coupon.valueCents)}` : `${coupon.percentBp / 100} 折`}</td><td>{coupon.redeemable ? `${coupon.redeemPoints} 积分` : '不可兑换'}</td><td>{coupon.validDays} 天</td><td><StatusBadge status={coupon.status} /></td><td>{coupon.totalQty.toLocaleString()}</td></tr>)}</tbody></table></div>
}

function BannerPanel() {
  const query = useAdminBannersQuery()
  return <div className="banner-admin-grid">{query.isDemo && <div className="demo-strip"><DemoBadge /> 当前运营位接口没有返回数据</div>}{(query.data ?? []).length === 0 ? <div className="empty-admin"><Megaphone size={22} /><strong>还没有运营位</strong><span>运营位不是用户首页的主流程，可以按需补充。</span></div> : (query.data ?? []).map((banner) => <article className="banner-admin-card" key={banner.id}><div><strong>{banner.title}</strong><span>{banner.imageUrl}</span></div><StatusBadge status={banner.enabled ? 'ACTIVE' : 'INACTIVE'} /></article>)}</div>
}

function CouponModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const create = useCreateCouponMutation()
  const [form, setForm] = useState({ name: '', type: 'FIXED', value_cents: '1000', percent_bp: '9000', min_spend_cents: '5000', redeemable: true, redeem_points: '1000', valid_days: '30', total_qty: '100', per_user_limit: '1' })
  async function submit(event: FormEvent) {
    event.preventDefault()
    await create.mutateAsync({ ...form, value_cents: Number(form.value_cents), percent_bp: Number(form.percent_bp), min_spend_cents: Number(form.min_spend_cents), redeem_points: Number(form.redeem_points), valid_days: Number(form.valid_days), total_qty: Number(form.total_qty), per_user_limit: Number(form.per_user_limit), max_discount_cents: 0 })
    onClose()
  }
  return <Modal open={open} title="新增优惠券模板" onClose={onClose}><form className="modal-form" onSubmit={submit}><Field label="模板名称"><Input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} placeholder="周末观影券" required /></Field><Field label="类型"><Select value={form.type} onChange={(event) => setForm({ ...form, type: event.target.value })}><option value="FIXED">满减券</option><option value="PERCENT">折扣券</option></Select></Field><div className="form-grid"><Field label="面额（分）"><Input type="number" value={form.value_cents} onChange={(event) => setForm({ ...form, value_cents: event.target.value })} /></Field><Field label="折扣（BP）"><Input type="number" value={form.percent_bp} onChange={(event) => setForm({ ...form, percent_bp: event.target.value })} /></Field></div><div className="form-grid"><Field label="使用门槛（分）"><Input type="number" value={form.min_spend_cents} onChange={(event) => setForm({ ...form, min_spend_cents: event.target.value })} /></Field><Field label="有效期（天）"><Input type="number" value={form.valid_days} onChange={(event) => setForm({ ...form, valid_days: event.target.value })} /></Field></div><div className="form-grid"><Field label="兑换积分"><Input type="number" value={form.redeem_points} onChange={(event) => setForm({ ...form, redeem_points: event.target.value })} /></Field><Field label="发行总量"><Input type="number" value={form.total_qty} onChange={(event) => setForm({ ...form, total_qty: event.target.value })} /></Field></div><div className="modal-form-actions"><Button type="button" variant="secondary" onClick={onClose}>取消</Button><Button type="submit" disabled={create.isPending}>{create.isPending ? '保存中…' : '保存模板'}</Button></div></form></Modal>
}

function BannerModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const create = useCreateBannerMutation()
  const [form, setForm] = useState({ title: '', image_url: '', sort: '0' })
  async function submit(event: FormEvent) {
    event.preventDefault()
    await create.mutateAsync({ ...form, sort: Number(form.sort), enabled: true })
    onClose()
  }
  return <Modal open={open} title="新增运营位" onClose={onClose}><form className="modal-form" onSubmit={submit}><Field label="标题"><Input value={form.title} onChange={(event) => setForm({ ...form, title: event.target.value })} required /></Field><Field label="图片 URL"><Input value={form.image_url} onChange={(event) => setForm({ ...form, image_url: event.target.value })} placeholder="https://" required /></Field><Field label="排序"><Input type="number" value={form.sort} onChange={(event) => setForm({ ...form, sort: event.target.value })} /></Field><div className="modal-form-actions"><Button type="button" variant="secondary" onClick={onClose}>取消</Button><Button type="submit" disabled={create.isPending}>{create.isPending ? '保存中…' : '保存运营位'}</Button></div></form></Modal>
}
