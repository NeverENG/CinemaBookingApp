import { ArrowRight, LocateFixed, MapPin } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useCinema } from '../../app/providers'
import { useCinemaQuery } from '../../features/catalog/hooks'
import { useDebouncedValue } from '../../hooks/useDebouncedValue'
import { DemoBadge, EmptyState, Input, PageHeader } from '../../components/ui'

export function CinemasPage() {
  const [params] = useSearchParams()
  const { cinemaId, setCinemaId } = useCinema()
  const [input, setInput] = useState('')
  const keyword = useDebouncedValue(input, 300)
  const query = useCinemaQuery(keyword)

  useEffect(() => {
    const id = Number(params.get('cinema_id'))
    if (id > 0) setCinemaId(id)
  }, [params, setCinemaId])

  const cinemas = query.data ?? []
  return <div className="content-container narrow-container"><PageHeader eyebrow="CINEMAS" title="选择影院" description="先选影院，再查看它正在上映的电影和可购场次。" /><div className="cinema-search"><MapPin size={17} /><Input value={input} onChange={(event) => setInput(event.target.value)} placeholder="搜索影院名称、城市或地址" /><button className="location-button" title="使用当前位置"><LocateFixed size={16} /></button></div>{query.isDemo && <div className="demo-strip"><DemoBadge /> 当前使用演示影院数据</div>}<div className="cinema-list">{cinemas.map((cinema) => <article className={`cinema-card ${cinema.id === cinemaId ? 'selected' : ''}`} key={cinema.id}><div className="cinema-card-icon"><MapPin size={19} /></div><div className="cinema-card-copy"><div><strong>{cinema.name}</strong>{cinema.id === cinemaId && <span className="current-label">当前影院</span>}</div><p>{cinema.city} · {cinema.address}</p>{cinema.distanceKm ? <small>距离你 {cinema.distanceKm.toFixed(1)} km</small> : null}</div><Link className="button button-secondary button-sm" to={`/?cinema_id=${cinema.id}`} onClick={() => setCinemaId(cinema.id)}>{cinema.id === cinemaId ? '查看热映' : '选择'}<ArrowRight size={14} /></Link></article>)}{cinemas.length === 0 && <EmptyState icon={MapPin} title="没有找到影院" description="试试其他城市或影院名称。" />}</div></div>
}
