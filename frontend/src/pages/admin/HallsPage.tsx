import { Armchair, LayoutGrid, Plus } from 'lucide-react'
import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { useAuth } from '../../app/providers'
import { Button, DemoBadge, Field, Input, Modal, PageHeader, Select, StatusBadge } from '../../components/ui'
import { useAdminHallsQuery, useCreateHallMutation, useUpdateHallMutation } from '../../features/admin/hooks'
import { useCinemaQuery } from '../../features/catalog/hooks'
import type { Hall } from '../../types'

interface SeatLayout {
  rows: number
  cols: number
  seat_types: Record<string, string>
  disabled: string[]
}

type EditMode = 'standard' | 'vip' | 'disabled'

function rowLabel(row: number) {
  let value = row
  let label = ''
  while (value > 0) {
    const remainder = (value - 1) % 26
    label = String.fromCharCode(65 + remainder) + label
    value = Math.floor((value - 1) / 26)
  }
  return label
}

function rowNumber(label: string) {
  return [...label].reduce((value, character) => value * 26 + character.charCodeAt(0) - 64, 0)
}

function seatNo(row: number, col: number) {
  return `${rowLabel(row)}${col}`
}

function parseLayout(raw: string): SeatLayout {
  try {
    const value = JSON.parse(raw) as Partial<SeatLayout>
    const rows = Number(value.rows)
    const cols = Number(value.cols)
    const seatTypes = value.seat_types && typeof value.seat_types === 'object' ? value.seat_types : {}
    const disabled = Array.isArray(value.disabled) ? value.disabled.filter((item): item is string => typeof item === 'string') : []
    return {
      rows: Number.isInteger(rows) && rows > 0 ? Math.min(rows, 50) : 8,
      cols: Number.isInteger(cols) && cols > 0 ? Math.min(cols, 50) : 10,
      seat_types: Object.fromEntries(Object.entries(seatTypes).filter(([, type]) => typeof type === 'string')),
      disabled,
    }
  } catch {
    return { rows: 8, cols: 10, seat_types: {}, disabled: [] }
  }
}

function trimLayout(layout: SeatLayout): SeatLayout {
  const inBounds = (value: string) => {
    const match = value.match(/^([A-Z]+)(\d+)$/)
    if (!match) return false
    const row = rowNumber(match[1])
    const col = Number(match[2])
    return row >= 1 && row <= layout.rows && col >= 1 && col <= layout.cols
  }
  return {
    ...layout,
    seat_types: Object.fromEntries(Object.entries(layout.seat_types).filter(([no]) => inBounds(no))),
    disabled: layout.disabled.filter(inBounds),
  }
}

function applySeatMode(layout: SeatLayout, no: string, mode: EditMode): SeatLayout {
  const disabled = new Set(layout.disabled)
  const seatTypes = { ...layout.seat_types }
  if (mode === 'disabled') {
    disabled.add(no)
    delete seatTypes[no]
  } else {
    disabled.delete(no)
    if (mode === 'vip') seatTypes[no] = 'VIP'
    else delete seatTypes[no]
  }
  return { ...layout, seat_types: seatTypes, disabled: [...disabled] }
}

export function HallsPage() {
  const { session } = useAuth()
  const cinemasQuery = useCinemaQuery()
  const cinemas = cinemasQuery.data ?? []
  const isCinemaAdmin = session?.role === 'CINEMA_ADMIN'
  const boundCinemaId = isCinemaAdmin ? session.cinemaId ?? 0 : 0
  const availableCinemas = boundCinemaId > 0 ? cinemas.filter((cinema) => cinema.id === boundCinemaId) : cinemas
  const [cinemaId, setCinemaId] = useState(1)
  const [open, setOpen] = useState(false)
  const [editingHall, setEditingHall] = useState<Hall | null>(null)
  const [name, setName] = useState('')
  const [layout, setLayout] = useState<SeatLayout>(() => parseLayout(''))
  const [editMode, setEditMode] = useState<EditMode>('standard')
  const query = useAdminHallsQuery(cinemaId)
  const create = useCreateHallMutation()
  const update = useUpdateHallMutation()
  const halls = useMemo(() => (query.data ?? []).filter((hall) => hall.cinemaId === cinemaId), [cinemaId, query.data])

  useEffect(() => {
    if (boundCinemaId > 0 && cinemaId !== boundCinemaId) {
      setCinemaId(boundCinemaId)
      return
    }
    if (availableCinemas.length > 0 && !availableCinemas.some((cinema) => cinema.id === cinemaId)) setCinemaId(availableCinemas[0].id)
  }, [availableCinemas, boundCinemaId, cinemaId])

  function openCreate() {
    setEditingHall(null)
    setName('')
    setLayout(parseLayout(''))
    setEditMode('standard')
    setOpen(true)
  }

  function openEdit(hall: Hall) {
    setEditingHall(hall)
    setCinemaId(hall.cinemaId)
    setName(hall.name)
    setLayout(parseLayout(hall.seatLayoutJson))
    setEditMode('standard')
    setOpen(true)
  }

  function toggleSeat(no: string) {
    setLayout((current) => applySeatMode(current, no, editMode))
  }

  function resize(nextRows: number, nextCols: number) {
    setLayout((current) => trimLayout({ ...current, rows: nextRows, cols: nextCols }))
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    const nextLayout = trimLayout(layout)
    const data = { cinema_id: cinemaId, name, seat_layout: JSON.stringify(nextLayout) }
    if (editingHall) await update.mutateAsync({ id: editingHall.id, data })
    else await create.mutateAsync(data)
    setOpen(false)
    setEditingHall(null)
  }

  const saving = create.isPending || update.isPending
  return <div className="admin-page"><PageHeader eyebrow="OPERATIONS" title="影厅管理" description="配置影厅名称与座位布局，供排片和选座使用。" actions={<Button onClick={openCreate}><Plus size={15} />新增影厅</Button>} /><div className="admin-toolbar"><label className="inline-filter"><span>影院</span><Select value={cinemaId} disabled={isCinemaAdmin} onChange={(event) => setCinemaId(Number(event.target.value))}>{availableCinemas.map((cinema) => <option value={cinema.id} key={cinema.id}>{cinema.name}</option>)}</Select></label>{(query.isDemo || cinemasQuery.isDemo) && <DemoBadge />}</div><div className="hall-grid">{halls.map((hall) => { const parsed = parseLayout(hall.seatLayoutJson); const disabledCount = parsed.disabled.length; return <article className="hall-card" key={hall.id}><div className="hall-card-top"><div className="hall-icon"><Armchair size={20} /></div><StatusBadge status={hall.status} /></div><h2>{hall.name}</h2><p>{parsed.rows} 排 × {parsed.cols} 列 · {parsed.rows * parsed.cols - disabledCount} 个可用座位</p><div className="hall-layout-preview"><LayoutGrid size={16} /><span>{disabledCount > 0 ? `${disabledCount} 个停用座位` : '完整座位布局'}</span></div><button className="text-link" onClick={() => openEdit(hall)}>编辑布局</button></article> })}</div><Modal open={open} title={editingHall ? '编辑影厅布局' : '新增影厅'} onClose={() => setOpen(false)}><form className="modal-form" onSubmit={submit}><Field label="影院"><Select value={cinemaId} disabled={isCinemaAdmin} onChange={(event) => setCinemaId(Number(event.target.value))}>{availableCinemas.map((cinema) => <option value={cinema.id} key={cinema.id}>{cinema.name}</option>)}</Select></Field><Field label="影厅名称"><Input value={name} onChange={(event) => setName(event.target.value)} placeholder="例如 2号厅" required /></Field><div className="form-grid"><Field label="排数"><Input type="number" min={1} max={50} value={layout.rows} onChange={(event) => resize(Number(event.target.value), layout.cols)} required /></Field><Field label="列数"><Input type="number" min={1} max={50} value={layout.cols} onChange={(event) => resize(layout.rows, Number(event.target.value))} required /></Field></div><div className="layout-editor-tools"><span>点击座位设置</span><button type="button" className={editMode === 'standard' ? 'active' : ''} onClick={() => setEditMode('standard')}><i className="layout-dot standard" />普通座</button><button type="button" className={editMode === 'vip' ? 'active' : ''} onClick={() => setEditMode('vip')}><i className="layout-dot vip" />VIP座</button><button type="button" className={editMode === 'disabled' ? 'active' : ''} onClick={() => setEditMode('disabled')}><i className="layout-dot disabled" />停用</button></div><div className="layout-editor-grid" style={{ gridTemplateColumns: `repeat(${layout.cols}, minmax(22px, 1fr))` }}>{Array.from({ length: layout.rows }, (_, rowIndex) => Array.from({ length: layout.cols }, (_, colIndex) => { const no = seatNo(rowIndex + 1, colIndex + 1); const disabled = layout.disabled.includes(no); const vip = layout.seat_types[no] === 'VIP'; return <button type="button" key={no} className={`layout-editor-seat ${disabled ? 'disabled' : vip ? 'vip' : ''}`} onClick={() => toggleSeat(no)} title={`${no} · ${disabled ? '停用' : vip ? 'VIP' : '普通'}`}>{colIndex + 1}</button> }))}</div><div className="layout-editor-summary"><span>已配置 {layout.rows * layout.cols - layout.disabled.length} 个座位</span><span>{Object.values(layout.seat_types).filter((type) => type === 'VIP').length} 个 VIP</span><span>{layout.disabled.length} 个停用</span></div><div className="modal-form-actions"><Button type="button" variant="secondary" onClick={() => setOpen(false)}>取消</Button><Button type="submit" disabled={saving}>{saving ? '保存中…' : editingHall ? '保存布局' : '保存影厅'}</Button></div></form></Modal></div>
}
