import { AlertCircle, CheckCircle2, Edit3, Plus, Ticket, XCircle } from 'lucide-react'
import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { useAuth } from '../../app/providers'
import { Button, DemoBadge, EmptyState, Field, Input, Modal, PageHeader, Select, StatusBadge } from '../../components/ui'
import { useAdminHallsQuery, useAdminMoviesQuery, useAdminSessionsQuery, useCancelSessionMutation, useCreateSessionMutation, useUpdateSessionPriceMutation } from '../../features/admin/hooks'
import { useCinemaQuery } from '../../features/catalog/hooks'
import { dateText, money, timeText } from '../../lib/format'
import { AppError } from '../../services/http/errors'

interface SessionForm {
  movie_id: string
  hall_id: string
  start_time: string
  end_time: string
  base_price_cents: string
  vip_price_cents: string
}

const emptySessionForm: SessionForm = { movie_id: '', hall_id: '', start_time: '', end_time: '', base_price_cents: '5000', vip_price_cents: '' }

interface Feedback {
  kind: 'success' | 'error'
  message: string
}

function dateTimeLocal(value: Date) {
  const offset = value.getTimezoneOffset()
  return new Date(value.getTime() - offset * 60_000).toISOString().slice(0, 16)
}

function defaultSessionTimes(durationMinutes = 120) {
  const start = new Date()
  start.setSeconds(0, 0)
  start.setMinutes(start.getMinutes() < 30 ? 30 : 60, 0, 0)
  const end = new Date(start.getTime() + Math.max(durationMinutes, 60) * 60_000)
  return { start_time: dateTimeLocal(start), end_time: dateTimeLocal(end) }
}

function serializePriceRules(vipPrice: string) {
  const value = vipPrice.trim()
  return value ? JSON.stringify({ VIP: Number(value) }) : '{}'
}

function vipPriceOf(priceRulesJson?: string) {
  if (!priceRulesJson) return ''
  try {
    const rules = JSON.parse(priceRulesJson) as { VIP?: unknown }
    return rules.VIP === undefined || rules.VIP === null ? '' : String(rules.VIP)
  } catch {
    return ''
  }
}

function messageOf(cause: unknown, fallback: string) {
  if (!(cause instanceof AppError)) return fallback
  const knownMessages: Record<string, string> = {
    'session time conflict': '该影厅时间段已有其他场次，请更换时间或影厅',
    'session invalid': '场次信息无效，请检查影片、影厅、时间和票价',
    forbidden: '当前账号没有管理该影院场次的权限',
    'session locked for change': '场次距离开场不足 30 分钟，不能再修改',
  }
  return knownMessages[cause.message] ?? cause.message
}

export function SessionsPage() {
  const { session } = useAuth()
  const cinemasQuery = useCinemaQuery()
  const allCinemas = cinemasQuery.data ?? []
  const isCinemaAdmin = session?.role === 'CINEMA_ADMIN'
  const boundCinemaId = isCinemaAdmin ? session.cinemaId ?? 0 : 0
  const cinemas = boundCinemaId > 0 ? allCinemas.filter((cinema) => cinema.id === boundCinemaId) : allCinemas
  const moviesQuery = useAdminMoviesQuery()
  const movies = moviesQuery.data ?? []
  const [cinemaId, setCinemaId] = useState(1)
  const hallsQuery = useAdminHallsQuery(cinemaId)
  const halls = useMemo(() => (hallsQuery.data ?? []).filter((hall) => hall.cinemaId === cinemaId), [cinemaId, hallsQuery.data])
  const [open, setOpen] = useState(false)
  const [priceOpen, setPriceOpen] = useState(false)
  const [priceSessionId, setPriceSessionId] = useState(0)
  const [price, setPrice] = useState<string | number>('')
  const [form, setForm] = useState<SessionForm>(emptySessionForm)
  const [feedback, setFeedback] = useState<Feedback | null>(null)
  const query = useAdminSessionsQuery(cinemaId)
  const create = useCreateSessionMutation()
  const updatePrice = useUpdateSessionPriceMutation()
  const cancel = useCancelSessionMutation()
  const sessions = useMemo(() => query.data ?? [], [query.data])

  useEffect(() => {
    if (boundCinemaId > 0 && cinemaId !== boundCinemaId) {
      setCinemaId(boundCinemaId)
      return
    }
    if (cinemas.length > 0 && !cinemas.some((cinema) => cinema.id === cinemaId)) setCinemaId(cinemas[0].id)
  }, [cinemaId, cinemas, boundCinemaId])

  useEffect(() => {
    if (movies.length > 0 && !movies.some((movie) => String(movie.id) === form.movie_id)) setForm((current) => ({ ...current, movie_id: String(movies[0].id) }))
  }, [form.movie_id, movies])

  useEffect(() => {
    if (halls.length > 0 && !halls.some((hall) => String(hall.id) === form.hall_id)) setForm((current) => ({ ...current, hall_id: String(halls[0].id) }))
  }, [form.hall_id, halls])

  function openCreate() {
    setFeedback(null)
    setForm({ ...emptySessionForm, ...defaultSessionTimes(movies[0]?.durationMinutes), movie_id: movies[0] ? String(movies[0].id) : '', hall_id: halls[0] ? String(halls[0].id) : '' })
    setOpen(true)
  }

  function openPriceEditor(sessionId: number, currentPrice: number, priceRulesJson?: string) {
    setFeedback(null)
    setPriceSessionId(sessionId)
    setPrice(String(currentPrice))
    setForm((current) => ({ ...current, vip_price_cents: vipPriceOf(priceRulesJson) }))
    setPriceOpen(true)
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    setFeedback(null)
    const startTime = new Date(form.start_time)
    const endTime = new Date(form.end_time)
    if (Number.isNaN(startTime.getTime()) || Number.isNaN(endTime.getTime())) {
      setFeedback({ kind: 'error', message: '请填写有效的开始和结束时间' })
      return
    }
    if (!(endTime.getTime() > startTime.getTime())) {
      setFeedback({ kind: 'error', message: '结束时间必须晚于开始时间' })
      return
    }
    try {
      await create.mutateAsync({ cinema_id: cinemaId, movie_id: Number(form.movie_id), hall_id: Number(form.hall_id), start_time: startTime.toISOString(), end_time: endTime.toISOString(), base_price_cents: Number(form.base_price_cents), price_rules: serializePriceRules(form.vip_price_cents) })
      setOpen(false)
      setFeedback({ kind: 'success', message: '场次保存成功，已同步到售票端' })
    } catch (cause) {
      setFeedback({ kind: 'error', message: messageOf(cause, '场次保存失败，请稍后重试') })
    }
  }

  async function submitPrice(event: FormEvent) {
    event.preventDefault()
    setFeedback(null)
    try {
      await updatePrice.mutateAsync({ id: priceSessionId, data: { base_price_cents: Number(price), price_rules: serializePriceRules(form.vip_price_cents) } })
      setPriceOpen(false)
      setFeedback({ kind: 'success', message: '场次票价保存成功' })
    } catch (cause) {
      setFeedback({ kind: 'error', message: messageOf(cause, '票价保存失败，请稍后重试') })
    }
  }

  async function cancelSession(id: number) {
    if (!window.confirm('确认取消这个场次吗？已支付订单会自动退款。')) return
    setFeedback(null)
    try {
      await cancel.mutateAsync(id)
      setFeedback({ kind: 'success', message: '场次已取消，相关座位和订单状态已同步' })
    } catch (cause) {
      setFeedback({ kind: 'error', message: messageOf(cause, '取消场次失败，请稍后重试') })
    }
  }

  const demo = query.isDemo || cinemasQuery.isDemo || moviesQuery.isDemo || hallsQuery.isDemo
  const pageFeedback = !open && !priceOpen ? feedback : null
  return <div className="admin-page"><PageHeader eyebrow="OPERATIONS" title="场次排片" description="为影院排片、设置基础票价，并观察可售状态。" actions={<div className="dashboard-actions">{demo && <DemoBadge />}<Button onClick={openCreate}><Plus size={15} />新增场次</Button></div>} />{pageFeedback && <div className={pageFeedback.kind === 'success' ? 'form-success' : 'form-error'} role={pageFeedback.kind === 'success' ? 'status' : 'alert'}>{pageFeedback.kind === 'success' ? <CheckCircle2 size={14} /> : <AlertCircle size={14} />}{pageFeedback.message}</div>}{query.error && <div className="form-error" role="alert"><AlertCircle size={14} />场次加载失败：{query.error.message}</div>}<div className="admin-toolbar"><label className="inline-filter"><span>影院</span><Select value={cinemaId} disabled={isCinemaAdmin} onChange={(event) => setCinemaId(Number(event.target.value))}>{cinemas.map((cinema) => <option value={cinema.id} key={cinema.id}>{cinema.name}</option>)}</Select></label><span className="toolbar-count">{sessions.length} 个场次</span></div>{query.isPending ? <div className="empty-admin"><span>正在加载场次…</span></div> : sessions.length === 0 ? <div className="empty-admin"><EmptyState icon={Ticket} title="当前影院还没有场次" description="点击右上角新增场次，保存后会显示在这里。" /></div> : <div className="session-admin-list">{sessions.map((session) => { const vipPrice = vipPriceOf(session.priceRulesJson); return <article className="session-admin-card" key={session.id}><div className="session-admin-date"><strong>{timeText(session.startTime)}</strong><span>{dateText(session.startTime)}</span></div><div className="session-admin-copy"><strong>{session.movieTitle || `影片 ${session.movieId}`}</strong><span>{session.hallName || `影厅 ${session.hallId}`} · 预计 {timeText(session.endTime)} 散场</span></div><div className="session-admin-price"><strong>{money(session.basePriceCents)}</strong><span>基础票价{vipPrice ? ` · VIP ${money(Number(vipPrice))}` : ''}</span></div><div className="session-admin-status"><StatusBadge status={session.status} /><span><Ticket size={13} />余 {session.remainingSeats ?? '--'} 座</span></div><div className="session-admin-actions">{session.status === 'OPEN' || session.status === 'SOLD_OUT' ? <><button onClick={() => openPriceEditor(session.id, session.basePriceCents, session.priceRulesJson)} title="修改票价"><Edit3 size={15} />改价</button><button onClick={() => cancelSession(session.id)} title="取消场次"><XCircle size={15} />取消</button></> : null}</div></article> })}</div>}<Modal open={open} title="新增场次" onClose={() => setOpen(false)}><form className="modal-form" onSubmit={submit}><Field label="影院"><Select value={cinemaId} disabled={isCinemaAdmin} onChange={(event) => setCinemaId(Number(event.target.value))}>{cinemas.map((cinema) => <option value={cinema.id} key={cinema.id}>{cinema.name}</option>)}</Select></Field><Field label="影片"><Select value={form.movie_id} onChange={(event) => setForm({ ...form, movie_id: event.target.value })} required>{movies.map((movie) => <option key={movie.id} value={movie.id}>{movie.title}</option>)}</Select></Field><Field label="影厅"><Select value={form.hall_id} onChange={(event) => setForm({ ...form, hall_id: event.target.value })} required>{halls.map((hall) => <option key={hall.id} value={hall.id}>{hall.name}</option>)}</Select></Field><div className="form-grid"><Field label="开始时间"><Input type="datetime-local" value={form.start_time} onChange={(event) => setForm({ ...form, start_time: event.target.value })} required /></Field><Field label="结束时间"><Input type="datetime-local" min={form.start_time || undefined} value={form.end_time} onChange={(event) => setForm({ ...form, end_time: event.target.value })} required /></Field></div><div className="form-grid"><Field label="基础票价（分）"><Input type="number" value={form.base_price_cents} onChange={(event) => setForm({ ...form, base_price_cents: event.target.value })} min={1} required /></Field><Field label="VIP 票价（分）"><Input type="number" value={form.vip_price_cents} onChange={(event) => setForm({ ...form, vip_price_cents: event.target.value })} min={1} placeholder="留空则使用基础价" /></Field></div>{feedback?.kind === 'error' && <div className="form-error" role="alert"><AlertCircle size={14} />{feedback.message}</div>}<div className="modal-form-actions"><Button type="button" variant="secondary" onClick={() => setOpen(false)}>取消</Button><Button type="submit" disabled={create.isPending || movies.length === 0 || halls.length === 0}>{create.isPending ? '保存中…' : '保存场次'}</Button></div></form></Modal><Modal open={priceOpen} title="修改场次票价" onClose={() => setPriceOpen(false)}><form className="modal-form" onSubmit={submitPrice}><div className="form-grid"><Field label="基础票价（分）"><Input type="number" min={1} value={price} onChange={(event) => setPrice(Number(event.target.value))} required /></Field><Field label="VIP 票价（分）"><Input type="number" min={1} value={form.vip_price_cents} onChange={(event) => setForm({ ...form, vip_price_cents: event.target.value })} placeholder="留空则使用基础价" /></Field></div>{feedback?.kind === 'error' && <div className="form-error" role="alert"><AlertCircle size={14} />{feedback.message}</div>}<div className="modal-form-actions"><Button type="button" variant="secondary" onClick={() => setPriceOpen(false)}>取消</Button><Button type="submit" disabled={updatePrice.isPending}>{updatePrice.isPending ? '保存中…' : '保存票价'}</Button></div></form></Modal></div>
}
