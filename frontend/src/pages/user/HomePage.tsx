import { ArrowRight, CalendarDays, MapPinned, Search, Sparkles } from 'lucide-react'
import { Link, useLocation, useSearchParams } from 'react-router-dom'
import { useCinema } from '../../app/providers'
import { CinemaPicker } from '../../components/CinemaPicker'
import { FilmHero } from '../../components/FilmHero'
import { MovieCard } from '../../components/MovieCard'
import { MovieRail } from '../../components/MovieRail'
import { DemoBadge, EmptyState, LoadingBlock } from '../../components/ui'
import { useHomeQuery, useSessionQuery } from '../../features/catalog/hooks'
import { dateText, timeText } from '../../lib/format'

const genreOptions = ['全部', '科幻', '动画', '悬疑', '剧情', '动作', '喜剧']

export function HomePage() {
  const { cinemaId } = useCinema()
  const location = useLocation()
  const [params, setParams] = useSearchParams()
  const isRecommendation = location.pathname === '/recommend'
  const homeQuery = useHomeQuery()
  const sessionQuery = useSessionQuery(undefined, cinemaId)
  const movies = homeQuery.data?.hotMovies ?? []
  const sessions = sessionQuery.data ?? []
  const activeGenre = params.get('genre') ?? ''
  const filteredMovies = activeGenre ? movies.filter((movie) => movie.genre.includes(activeGenre)) : movies
  const cinemaMovieIds = new Set(sessions.map((session) => session.movieId))
  const cinemaMovies = filteredMovies.filter((movie) => cinemaMovieIds.has(movie.id))
  const displayMovies = isRecommendation ? [...filteredMovies].sort((left, right) => right.rating - left.rating) : [...cinemaMovies, ...filteredMovies.filter((movie) => !cinemaMovieIds.has(movie.id))].slice(0, 4)
  const recommendationMovies = [...filteredMovies].sort((left, right) => right.rating - left.rating || right.soldCount - left.soldCount).slice(0, 6)
  const displayMovieIds = new Set(filteredMovies.map((movie) => movie.id))
  const displaySessions = activeGenre ? sessions.filter((session) => displayMovieIds.has(session.movieId)) : sessions

  function selectGenre(genre: string) {
    const next = new URLSearchParams(params)
    if (genre === '全部') next.delete('genre')
    else next.set('genre', genre)
    setParams(next, { replace: true })
  }

  return <div className="home-page"><div className="content-container home-container">
    <div className="home-context-bar"><div><span className="home-context-label">LTERM CINEMA / 2026</span><strong>{isRecommendation ? '给你推荐一些好电影' : '今天，去看一场电影。'}</strong></div><div className="home-context-actions"><span className="home-date"><CalendarDays size={15} /> {dateText(new Date())}</span><CinemaPicker /></div></div>
    <div className="home-catalog-bar"><div className="home-catalog-title"><span>EXPLORE FILMS</span><strong>按类型看片</strong></div><div className="home-category-list" aria-label="电影分类">{genreOptions.map((genre) => <button type="button" key={genre} className={genre === (activeGenre || '全部') ? 'active' : ''} onClick={() => selectGenre(genre)}>{genre}</button>)}</div><Link className="home-search-link" to="/search"><Search size={14} />搜索电影</Link></div>
    {homeQuery.isPending ? <div className="hero-skeleton" /> : <FilmHero movies={displayMovies.length > 0 ? displayMovies : filteredMovies} banners={homeQuery.data?.banners} cinemaId={cinemaId} isDemo={homeQuery.isDemo} />}
    {(homeQuery.isDemo || sessionQuery.isDemo) && <div className="demo-strip"><DemoBadge /> 后端接口暂未返回内容，当前展示可交互演示数据</div>}
    <section className="home-section home-recommendation-section"><div className="home-section-heading"><div><span className="home-section-kicker">{isRecommendation ? 'CURATED FOR YOU' : 'RECOMMENDED'}</span><h2>{isRecommendation ? '今天看什么' : '推荐电影'}</h2><p>{activeGenre ? `正在浏览「${activeGenre}」类型的精选片单。` : '按评分、热度和观影情绪，先挑一部值得看的。'}</p></div><Link className="section-arrow-link" to="/recommend">更多推荐 <ArrowRight size={15} /></Link></div>{homeQuery.isPending ? <LoadingBlock lines={4} /> : recommendationMovies.length === 0 ? <EmptyState icon={Sparkles} title="暂时没有这个分类的推荐" description="换一个类型，或者搜索你想看的电影。" action={<Link className="button button-secondary" to="/search">去搜索电影</Link>} /> : <MovieRail movies={recommendationMovies} />}</section>
    <section className="home-section"><div className="home-section-heading"><div><span className="home-section-kicker">{isRecommendation ? 'NOW SHOWING' : 'CINEMA NOW'}</span><h2>{isRecommendation ? '影院热映' : '正在影院上映'}</h2><p>{isRecommendation ? '选一部喜欢的电影，再选择影院和场次。' : '当前影院有场次的影片优先展示，也可以继续探索全城热映。'}</p></div><Link className="section-arrow-link" to="/search">查看全部 <ArrowRight size={15} /></Link></div>{homeQuery.isPending ? <LoadingBlock lines={4} /> : displayMovies.length === 0 ? <EmptyState icon={Sparkles} title="当前分类暂时没有热映影片" description="可以切换类型、搜索其他电影，或更换影院。" action={<Link className="button button-secondary" to="/cinemas">选择其他影院</Link>} /> : <div className="movie-wall" data-count={displayMovies.length}>{displayMovies.map((movie) => <MovieCard key={movie.id} movie={movie} />)}</div>}</section>
    <section className="home-schedule"><div className="home-section-heading compact"><div><span className="home-section-kicker">SHOWTIMES</span><h2>最近可购场次</h2></div><Link className="section-arrow-link" to="/cinemas">切换影院 <ArrowRight size={15} /></Link></div>{displaySessions.length === 0 ? <div className="schedule-empty"><MapPinned size={18} /><span>当前影院暂无符合条件的场次，去搜索其他电影或影院。</span><Link to="/search">开始搜索</Link></div> : <div className="schedule-list">{displaySessions.slice(0, 5).map((session) => <Link className="schedule-row" to={`/sessions/${session.id}/seats`} key={session.id}><div className="schedule-film"><strong>{session.movieTitle || '电影场次'}</strong><span>{session.hallName} · {dateText(session.startTime)}</span></div><div className="schedule-time"><strong>{timeText(session.startTime)}</strong><span>¥{(session.basePriceCents / 100).toFixed(0)} 起</span></div><span className="schedule-cta">选座 <ArrowRight size={15} /></span></Link>)}</div>}</section>
  </div></div>
}
