import { Armchair, LayoutGrid, Plus } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { Button, DemoBadge, Field, Input, Modal, PageHeader, Select, StatusBadge, TextArea } from '../../components/ui'
import { useAdminHallsQuery, useCreateHallMutation } from '../../features/admin/hooks'

export function HallsPage() {
  const [cinemaId, setCinemaId] = useState(1)
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [layout, setLayout] = useState('{"rows":8,"cols":10}')
  const query = useAdminHallsQuery(cinemaId)
  const create = useCreateHallMutation()

  async function submit(event: FormEvent) {
    event.preventDefault()
    await create.mutateAsync({ cinema_id: cinemaId, name, seat_layout: layout })
    setName('')
    setOpen(false)
  }

  return <div className="admin-page"><PageHeader eyebrow="OPERATIONS" title="影厅管理" description="配置影厅名称与座位布局，供排片和选座使用。" actions={<Button onClick={() => setOpen(true)}><Plus size={15} />新增影厅</Button>} /><div className="admin-toolbar"><label className="inline-filter"><span>影院</span><Select value={cinemaId} onChange={(event) => setCinemaId(Number(event.target.value))}><option value={1}>LTerm 光影中心</option><option value={2}>LTerm 北岸影院</option></Select></label>{query.isDemo && <DemoBadge />}</div><div className="hall-grid">{(query.data ?? []).map((hall) => <article className="hall-card" key={hall.id}><div className="hall-card-top"><div className="hall-icon"><Armchair size={20} /></div><StatusBadge status={hall.status} /></div><h2>{hall.name}</h2><p>影院 ID {hall.cinemaId}</p><div className="hall-layout-preview"><LayoutGrid size={16} /><span>{hall.seatLayoutJson || '座位布局待配置'}</span></div><button className="text-link">编辑布局</button></article>)}</div><Modal open={open} title="新增影厅" onClose={() => setOpen(false)}><form className="modal-form" onSubmit={submit}><Field label="影院"><Select value={cinemaId} onChange={(event) => setCinemaId(Number(event.target.value))}><option value={1}>LTerm 光影中心</option><option value={2}>LTerm 北岸影院</option></Select></Field><Field label="影厅名称"><Input value={name} onChange={(event) => setName(event.target.value)} placeholder="例如 2号厅" required /></Field><Field label="座位布局 JSON" hint="例如：{&quot;rows&quot;:8,&quot;cols&quot;:10}"><TextArea value={layout} onChange={(event) => setLayout(event.target.value)} rows={4} required /></Field><div className="modal-form-actions"><Button type="button" variant="secondary" onClick={() => setOpen(false)}>取消</Button><Button type="submit" disabled={create.isPending}>{create.isPending ? '保存中…' : '保存影厅'}</Button></div></form></Modal></div>
}
