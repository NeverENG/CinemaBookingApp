import { ChevronDown, MapPin } from 'lucide-react'
import { useCinema } from '../app/providers'
import { useCinemaQuery } from '../features/catalog/hooks'
import { DemoBadge, Select } from './ui'

export function CinemaPicker({ compact = false }: { compact?: boolean }) {
  const { cinemaId, setCinemaId } = useCinema()
  const cinemas = useCinemaQuery().data ?? []
  const current = cinemas.find((cinema) => cinema.id === cinemaId) ?? cinemas[0]
  return <div className={`cinema-picker ${compact ? 'cinema-picker-compact' : ''}`}>
    <MapPin size={15} />
    <div className="cinema-picker-copy"><small>当前影院</small><strong>{current?.name ?? '选择影院'}</strong></div>
    <Select aria-label="切换影院" value={current?.id ?? cinemaId} onChange={(event) => setCinemaId(Number(event.target.value))}>{cinemas.map((cinema) => <option key={cinema.id} value={cinema.id}>{cinema.name}</option>)}</Select>
    <ChevronDown className="cinema-picker-chevron" size={15} />
    {current?.distanceKm && !compact ? <span className="distance-label">{current.distanceKm.toFixed(1)} km</span> : null}
    {!compact && cinemas.length === 0 && <DemoBadge />}
  </div>
}
