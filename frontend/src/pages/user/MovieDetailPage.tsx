import { ArrowLeft, CalendarDays, Clock3, Film, MapPin, Play, Star } from 'lucide-react'
import { useState } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { useCinema } from '../../app/providers'
import { useMovieQuery, useSessionQuery } from '../../features/catalog/hooks'
import { dateText, money, timeText } from '../../lib/format'
import { Button, DemoBadge, EmptyState, PageHeader, StatusBadge } from '../../components/ui'

export function MovieDetailPage() {
  const { movieId: rawMovieId } = useParams()
  const movieId = Number(rawMovieId)
  const [params] = useSearchParams()
  const { cinemaId } = useCinema()
  const navigate = useNavigate()
  const [showTrailer, setShowTrailer] = useState(false)
  const selectedCinemaId = Number(params.get('cinema_id')) || cinemaId
  const movieQuery = useMovieQuery(movieId)
  const currentSessionsQuery = useSessionQuery(movieId, selectedCinemaId)
  const allSessionsQuery = useSessionQuery(movieId)
  const movie = movieQuery.data
  const currentSessions = currentSessionsQuery.data ?? []
  const allSessions = allSessionsQuery.data ?? []
  const sessions = currentSessions.length > 0 ? currentSessions : allSessions
  const groupedSessions = sessions.reduce<Record<string, typeof sessions>>((groups, session) => {
    const key = `${session.cinemaId}-${session.cinemaName}`
    groups[key] = groups[key] ? [...groups[key], session] : [session]
    return groups
  }, {})

  if (!movie) return <div className="content-container"><EmptyState icon={Film} title="电影不存在" description="返回热映列表继续浏览。" action={<Link className="button button-secondary" to="/">返回热映</Link>} /></div>

  return <div className="content-container"><button className="back-link" onClick={() => navigate(-1)}><ArrowLeft size={15} />返回</button><div className="movie-detail"><div className={`movie-detail-poster ${movie.coverUrl ? '' : 'image-missing'}`}>{movie.coverUrl && <img src={movie.coverUrl} alt={movie.title} />}</div><div className="movie-detail-copy"><span className="eyebrow">NOW SHOWING</span><h1>{movie.title}</h1><div className="detail-tags"><span><Star size={14} fill="currentColor" /> {movie.rating.toFixed(1)}</span><span>{movie.genre || '电影'}</span><span><Clock3 size={14} /> {movie.durationMinutes} 分钟</span></div><p>{movie.description || '这部电影正在当前片单中，选择一场合适的放映时间开始观影。'}</p><div className="movie-detail-actions"><Button onClick={() => setShowTrailer(true)} variant="secondary"><Play size={15} />查看预告</Button><span className="detail-note"><CalendarDays size={15} /> 选择场次后进入选座</span></div></div></div><section className="showtime-section"><PageHeader eyebrow="SHOWTIMES" title="选择场次" description={currentSessions.length > 0 ? '当前影院有可购场次。' : '当前影院暂无场次，以下为其他影院的可购场次。'} actions={currentSessions.length === 0 ? <Link className="text-link" to="/cinemas">切换影院</Link> : undefined} />{(movieQuery.isDemo || currentSessionsQuery.isDemo) && <div className="demo-strip"><DemoBadge /> 当前页面展示演示数据</div>}{Object.entries(groupedSessions).length === 0 ? <EmptyState icon={CalendarDays} title="暂无可购场次" description="可以稍后再来，或换一部正在上映的电影。" /> : <div className="showtime-groups">{Object.entries(groupedSessions).map(([groupKey, group]) => <div className="showtime-group" key={groupKey}><div className="showtime-group-heading"><div><MapPin size={15} /><strong>{group[0].cinemaName}</strong>{group[0].cinemaId === selectedCinemaId && <span className="current-label">当前影院</span>}</div><span>{dateText(group[0].startTime)}</span></div><div className="showtime-list">{group.map((session) => <Link className="showtime-card" to={`/sessions/${session.id}/seats`} key={session.id}><div className="showtime-time"><strong>{timeText(session.startTime)}</strong><span>预计 {timeText(session.endTime)} 散场</span></div><div className="showtime-hall"><strong>{session.hallName || '标准厅'}</strong><span>普通厅 · {session.remainingSeats ? `余 ${session.remainingSeats} 座` : '可购'}</span></div><div className="showtime-price"><strong>{money(session.basePriceCents)}</strong><StatusBadge status={session.status} /></div><span className="showtime-arrow">选座 →</span></Link>)}</div></div>)}</div>}</section>{showTrailer && <div className="modal-backdrop" role="presentation" onClick={() => setShowTrailer(false)}><div className="trailer-modal" onClick={(event) => event.stopPropagation()}><button className="modal-close" onClick={() => setShowTrailer(false)}>×</button>{movie.trailerUrl ? <iframe title={`${movie.title}预告片`} src={movie.trailerUrl} allowFullScreen /> : <div className="trailer-placeholder"><Film size={28} /><strong>预告片资源待接入</strong><span>当前影片只有外部资源地址，后端返回后将在这里播放。</span></div>}</div></div>}</div>
}
