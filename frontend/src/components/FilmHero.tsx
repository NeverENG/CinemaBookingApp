import { ChevronLeft, ChevronRight, Play, PlayCircle, Star } from 'lucide-react'
import { useEffect, useMemo, useState, type CSSProperties, type SyntheticEvent } from 'react'
import { Link } from 'react-router-dom'
import type { Banner, Movie } from '../types'

interface FilmHeroProps {
  movies: Movie[]
  banners?: Banner[]
  cinemaId: number
  isDemo?: boolean
}

function heatText(soldCount: number) {
  if (soldCount >= 10000) return `${(soldCount / 10000).toFixed(1)} 万人想看`
  return `${soldCount} 人想看`
}

export function FilmHero({ movies, banners = [], cinemaId, isDemo = false }: FilmHeroProps) {
  const featuredMovies = useMemo(() => movies.slice(0, 4), [movies])
  const [activeIndex, setActiveIndex] = useState(0)
  const activeMovie = featuredMovies[activeIndex]

  useEffect(() => {
    setActiveIndex((current) => Math.min(current, Math.max(featuredMovies.length - 1, 0)))
  }, [featuredMovies.length])

  useEffect(() => {
    if (featuredMovies.length < 2) return undefined
    const timer = window.setInterval(() => setActiveIndex((current) => (current + 1) % featuredMovies.length), 5200)
    return () => window.clearInterval(timer)
  }, [featuredMovies.length])

  if (!activeMovie) return null

  const activeBanner = banners.length > 0 ? banners[activeIndex % banners.length] : undefined
  const heroImage = activeBanner?.imageUrl || activeMovie.backdropUrl || activeMovie.coverUrl
  const heroStyle = heroImage ? { '--hero-image': `url("${heroImage}")` } as CSSProperties : undefined

  function move(step: number) {
    setActiveIndex((current) => (current + step + featuredMovies.length) % featuredMovies.length)
  }

  function handlePosterError(event: SyntheticEvent<HTMLImageElement>) {
    event.currentTarget.style.display = 'none'
    event.currentTarget.parentElement?.classList.add('image-missing')
  }

  return <section className="film-hero" style={heroStyle}>
    <div className="film-hero-backdrop" />
    <div className="film-hero-noise" />
    <div className="film-hero-content">
      <Link className={`film-hero-poster ${activeMovie.coverUrl ? '' : 'image-missing'}`} to={`/movies/${activeMovie.id}?cinema_id=${cinemaId}`} aria-label={`查看${activeMovie.title}详情`}>{activeMovie.coverUrl && <img key={activeMovie.id} src={activeMovie.coverUrl} alt={activeMovie.title} onError={handlePosterError} />}<span className="film-hero-poster-play"><PlayCircle size={24} /><small>查看预告</small></span></Link>
      <div className="film-hero-copy">
        <div className="film-hero-kicker"><span>{activeBanner?.title || 'FEATURED FILM'}</span><i />{isDemo ? '演示片单' : '影院热映'}</div>
        <h1>{activeMovie.title}</h1>
        <div className="film-hero-meta"><span className="film-rating"><Star size={14} fill="currentColor" /> {activeMovie.rating.toFixed(1)}</span><span>{activeMovie.genre || '剧情'}</span><span>{activeMovie.durationMinutes || '--'} 分钟</span></div>
        <p>{activeMovie.description || '一部值得在大银幕上观看的电影。选择影院与场次，留出今晚的时间。'}</p>
        <div className="film-hero-actions"><Link className="button button-film" to={`/movies/${activeMovie.id}?cinema_id=${cinemaId}`}><Play size={15} fill="currentColor" />查看场次</Link><Link className="film-hero-text-link" to={`/movies/${activeMovie.id}?cinema_id=${cinemaId}`}>了解这部电影 <ChevronRight size={15} /></Link></div>
        <Link className="film-hero-player-strip" to={`/movies/${activeMovie.id}?cinema_id=${cinemaId}#trailer`}><span className="film-hero-player-icon"><Play size={12} fill="currentColor" /></span><span><strong>播放预告片</strong><small>进入影片详情，了解更多</small></span><i><span /></i></Link>
      </div>
      <div className="film-hero-rail">
        <div className="film-hero-rail-heading"><span>本周精选</span><span>{String(activeIndex + 1).padStart(2, '0')} / {String(featuredMovies.length).padStart(2, '0')}</span></div>
        <div className="film-hero-thumbnails">{featuredMovies.map((movie, index) => <button type="button" className={`film-hero-thumb ${index === activeIndex ? 'active' : ''}`} key={movie.id} onClick={() => setActiveIndex(index)} aria-label={`查看${movie.title}`} aria-pressed={index === activeIndex}>{movie.coverUrl && <img src={movie.coverUrl} alt="" onError={handlePosterError} />}<span><strong>{movie.title}</strong><small>{heatText(movie.soldCount)} · {movie.rating.toFixed(1)} 分</small></span></button>)}</div>
        <div className="film-hero-controls"><button onClick={() => move(-1)} aria-label="上一部"><ChevronLeft size={16} /></button><button onClick={() => move(1)} aria-label="下一部"><ChevronRight size={16} /></button></div>
      </div>
    </div>
    <div className="film-hero-footer"><span>把时间留给值得的电影</span><div className="film-hero-dots">{featuredMovies.map((movie, index) => <button type="button" key={movie.id} className={index === activeIndex ? 'active' : ''} onClick={() => setActiveIndex(index)} aria-label={`第${index + 1}部精选`} aria-pressed={index === activeIndex} />)}</div><span>当前影院 · 选场次即可购票</span></div>
  </section>
}
