import { ArrowUpRight, Clock3, Star } from 'lucide-react'
import type { SyntheticEvent } from 'react'
import { Link } from 'react-router-dom'
import type { Movie } from '../types'
import { useCinema } from '../app/providers'

export function MovieCard({ movie, compact = false }: { movie: Movie; compact?: boolean }) {
  const { cinemaId } = useCinema()
  const hasCover = Boolean(movie.coverUrl)
  function handleImageError(event: SyntheticEvent<HTMLImageElement>) {
    event.currentTarget.style.display = 'none'
    event.currentTarget.parentElement?.classList.add('image-missing')
  }

  return <article className={`movie-card ${compact ? 'movie-card-compact' : ''}`}>
    <Link to={`/movies/${movie.id}?cinema_id=${cinemaId}`} className="movie-poster-link">
      <div className={`movie-poster ${hasCover ? '' : 'image-missing'}`}>{hasCover && <img src={movie.coverUrl} alt={movie.title} onError={handleImageError} />}<div className="movie-poster-overlay"><span>查看详情</span><ArrowUpRight size={14} /></div><span className="movie-rating"><Star size={12} fill="currentColor" /> {movie.rating.toFixed(1)}</span></div>
    </Link>
    <div className="movie-card-content"><div className="movie-card-title-row"><h3>{movie.title}</h3><Link className="icon-link" to={`/movies/${movie.id}?cinema_id=${cinemaId}`} aria-label={`查看${movie.title}`}><ArrowUpRight size={16} /></Link></div><p>{movie.genre || '电影'} · {movie.durationMinutes || '--'} 分钟</p>{!compact && <div className="movie-meta"><span><Clock3 size={13} /> 今日有场次</span><Link to={`/movies/${movie.id}?cinema_id=${cinemaId}`}>查看场次</Link></div>}</div>
  </article>
}
