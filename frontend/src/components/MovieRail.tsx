import { ArrowRight, Play, Star } from 'lucide-react'
import type { SyntheticEvent } from 'react'
import { Link } from 'react-router-dom'
import { useCinema } from '../app/providers'
import type { Movie } from '../types'

export function MovieRail({ movies }: { movies: Movie[] }) {
  const { cinemaId } = useCinema()

  function handleImageError(event: SyntheticEvent<HTMLImageElement>) {
    event.currentTarget.style.display = 'none'
    event.currentTarget.parentElement?.classList.add('image-missing')
  }

  return <div className={`movie-rail ${movies.length <= 6 ? 'movie-rail-fit' : ''}`} data-count={movies.length} role="list">
    {movies.map((movie) => <article className="movie-rail-card" key={movie.id} role="listitem">
      <Link className="movie-rail-poster-link" to={`/movies/${movie.id}?cinema_id=${cinemaId}`} aria-label={`查看${movie.title}`}>
        <div className={`movie-rail-poster ${movie.coverUrl ? '' : 'image-missing'}`}>
          {movie.coverUrl && <img src={movie.coverUrl} alt={movie.title} onError={handleImageError} />}
          <span className="movie-rail-play"><Play size={15} fill="currentColor" /></span>
          <span className="movie-rail-rating"><Star size={11} fill="currentColor" /> {movie.rating.toFixed(1)}</span>
        </div>
      </Link>
      <div className="movie-rail-copy">
        <div className="movie-rail-title-row"><h3>{movie.title}</h3><Link to={`/movies/${movie.id}?cinema_id=${cinemaId}`} aria-label={`了解${movie.title}`}><ArrowRight size={15} /></Link></div>
        <p>{movie.genre || '电影'} · {movie.durationMinutes || '--'} 分钟</p>
        <Link className="movie-rail-action" to={`/movies/${movie.id}?cinema_id=${cinemaId}`}>查看详情 <ArrowRight size={13} /></Link>
      </div>
    </article>)}
  </div>
}
