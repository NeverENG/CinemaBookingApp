import { Plus, Ticket } from 'lucide-react'
import { useMemo, useState, type FormEvent } from 'react'
import { Button, DemoBadge, Field, Input, Modal, PageHeader, Select, StatusBadge } from '../../components/ui'
import { useSessionQuery } from '../../features/catalog/hooks'
import { useCreateSessionMutation } from '../../features/admin/hooks'
import { dateText, money, timeText } from '../../lib/format'
import { demoMovies } from '../../demo'

export function SessionsPage() {
  const [cinemaId, setCinemaId] = useState(1)
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState({ movie_id: '1', hall_id: '1', start_time: '', end_time: '', base_price_cents: '5000' })
  const query = useSessionQuery(undefined, cinemaId)
  const create = useCreateSessionMutation()
  const sessions = useMemo(() => query.data ?? [], [query.data])

  async function submit(event: FormEvent) {
    event.preventDefault()
    await create.mutateAsync({ cinema_id: cinemaId, movie_id: Number(form.movie_id), hall_id: Number(form.hall_id), start_time: new Date(form.start_time).toISOString(), end_time: new Date(form.end_time).toISOString(), base_price_cents: Number(form.base_price_cents), price_rules: '' })
    setOpen(false)
  }

  return <div className="admin-page"><PageHeader eyebrow="OPERATIONS" title="场次排片" description="为影院排片、设置基础票价，并观察可售状态。" actions={<div className="dashboard-actions">{query.isDemo && <DemoBadge />}<Button onClick={() => setOpen(true)}><Plus size={15} />新增场次</Button></div>} /><div className="admin-toolbar"><label className="inline-filter"><span>影院</span><Select value={cinemaId} onChange={(event) => setCinemaId(Number(event.target.value))}><option value={1}>LTerm 光影中心</option><option value={2}>LTerm 北岸影院</option></Select></label><span className="toolbar-count">{sessions.length} 个场次</span></div><div className="session-admin-list">{sessions.map((session) => <article className="session-admin-card" key={session.id}><div className="session-admin-date"><strong>{timeText(session.startTime)}</strong><span>{dateText(session.startTime)}</span></div><div className="session-admin-copy"><strong>{session.movieTitle || demoMovies.find((movie) => movie.id === session.movieId)?.title || '电影'}</strong><span>{session.hallName || `影厅 ${session.hallId}`} · 预计 {timeText(session.endTime)} 散场</span></div><div className="session-admin-price"><strong>{money(session.basePriceCents)}</strong><span>基础票价</span></div><div className="session-admin-status"><StatusBadge status={session.status} /><span><Ticket size={13} />可售状态</span></div></article>)}</div><Modal open={open} title="新增场次" onClose={() => setOpen(false)}><form className="modal-form" onSubmit={submit}><Field label="影院"><Select value={cinemaId} onChange={(event) => setCinemaId(Number(event.target.value))}><option value={1}>LTerm 光影中心</option><option value={2}>LTerm 北岸影院</option></Select></Field><Field label="影片"><Select value={form.movie_id} onChange={(event) => setForm({ ...form, movie_id: event.target.value })}>{demoMovies.map((movie) => <option key={movie.id} value={movie.id}>{movie.title}</option>)}</Select></Field><Field label="影厅 ID"><Input type="number" value={form.hall_id} onChange={(event) => setForm({ ...form, hall_id: event.target.value })} required /></Field><div className="form-grid"><Field label="开始时间"><Input type="datetime-local" value={form.start_time} onChange={(event) => setForm({ ...form, start_time: event.target.value })} required /></Field><Field label="结束时间"><Input type="datetime-local" value={form.end_time} onChange={(event) => setForm({ ...form, end_time: event.target.value })} required /></Field></div><Field label="基础票价（分）"><Input type="number" value={form.base_price_cents} onChange={(event) => setForm({ ...form, base_price_cents: event.target.value })} min={1} required /></Field><div className="modal-form-actions"><Button type="button" variant="secondary" onClick={() => setOpen(false)}>取消</Button><Button type="submit" disabled={create.isPending}>{create.isPending ? '保存中…' : '保存场次'}</Button></div></form></Modal></div>
}
