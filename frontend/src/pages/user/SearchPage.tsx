import { Film, MapPin, Search as SearchIcon, Sparkles, TrendingUp, X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useCinema } from '../../app/providers'
import { useCinemaQuery, useMovieSearchQuery } from '../../features/catalog/hooks'
import { MovieCard } from '../../components/MovieCard'
import { useDebouncedValue } from '../../hooks/useDebouncedValue'
import { DemoBadge, EmptyState, Input } from '../../components/ui'

const popularKeywords = ['科幻', '动画', '悬疑', '喜剧', 'IMAX']

export function SearchPage() {
  const [params, setParams] = useSearchParams()
  const [input, setInput] = useState(params.get('q') ?? '')
  const [scope, setScope] = useState<'all' | 'movie' | 'cinema'>('all')
  const keyword = useDebouncedValue(input.trim(), 300)
  const { cinemaId } = useCinema()
  const moviesQuery = useMovieSearchQuery(keyword)
  const cinemasQuery = useCinemaQuery(keyword)

  useEffect(() => {
    if (keyword !== (params.get('q') ?? '')) {
      const next = new URLSearchParams(params)
      if (keyword) next.set('q', keyword)
      else next.delete('q')
      setParams(next, { replace: true })
    }
  }, [keyword, params, setParams])

  function setKeyword(next: string) {
    setInput(next)
  }

  const movies = moviesQuery.data ?? []
  const cinemas = cinemasQuery.data ?? []
  const isDemo = moviesQuery.isDemo || cinemasQuery.isDemo

  return <div className="content-container search-page">
    <section className="search-intro"><span className="home-section-kicker">SEARCH THE SCREEN</span><h1>找到你想看的电影</h1><p>从片名、类型或影院开始。先找到电影，再决定在哪里看。</p><div className="search-large search-large-modern"><SearchIcon size={20} /><Input autoFocus type="search" value={input} onChange={(event) => setKeyword(event.target.value)} placeholder="搜索电影、类型、导演或影院" />{input && <button type="button" className="search-clear" onClick={() => setKeyword('')} aria-label="清空搜索"><X size={15} /></button>}<span className="search-shortcut">⌘ K</span></div><div className="search-hotwords"><span>大家都在搜</span>{popularKeywords.map((word) => <button type="button" key={word} onClick={() => setKeyword(word)}>{word}</button>)}</div></section>
    {!keyword ? <section className="search-discover"><div className="home-section-heading"><div><span className="home-section-kicker">DISCOVER</span><h2>从一部电影开始</h2><p>热门片单、影院与场次，都从这里进入。</p></div><TrendingUp size={19} /></div><div className="discover-grid"><Link to="/recommend" className="discover-card discover-card-dark"><Sparkles size={18} /><strong>编辑推荐</strong><span>不确定看什么？先看我们挑好的。</span><Arrow /></Link><Link to="/cinemas" className="discover-card"><MapPin size={18} /><strong>附近影院</strong><span>按城市、距离和当前影院筛选。</span><Arrow /></Link><Link to="/" className="discover-card"><Film size={18} /><strong>影院热映</strong><span>直接查看当前影院正在上映的影片。</span><Arrow /></Link></div></section> : <>
      <div className="search-result-head"><div><span className="home-section-kicker">RESULTS</span><h2>关于“{keyword}”的结果</h2></div>{isDemo && <DemoBadge />}</div>
      <div className="search-scope-tabs"><button className={scope === 'all' ? 'active' : ''} onClick={() => setScope('all')}>全部 <span>{movies.length + cinemas.length}</span></button><button className={scope === 'movie' ? 'active' : ''} onClick={() => setScope('movie')}><Film size={14} />电影 <span>{movies.length}</span></button><button className={scope === 'cinema' ? 'active' : ''} onClick={() => setScope('cinema')}><MapPin size={14} />影院 <span>{cinemas.length}</span></button></div>
      <div className={`search-results search-results-${scope}`}>
        {(scope === 'all' || scope === 'movie') && <section className="search-result-section"><div className="search-section-label"><Film size={15} /><strong>影片</strong><span>{movies.length}</span></div>{movies.length === 0 ? <EmptyState icon={Film} title="没有找到相关电影" description="可以尝试更短的片名、类型，或切换关键词。" /> : <div className="search-movie-grid">{movies.map((movie) => <MovieCard key={movie.id} movie={movie} compact />)}</div>}</section>}
        {(scope === 'all' || scope === 'cinema') && <section className="search-result-section"><div className="search-section-label"><MapPin size={15} /><strong>影院</strong><span>{cinemas.length}</span></div>{cinemas.length === 0 ? <EmptyState icon={MapPin} title="没有找到相关影院" description="可以搜索影院名称或城市。" /> : <div className="cinema-result-list">{cinemas.map((cinema) => <Link className={`cinema-result ${cinema.id === cinemaId ? 'current' : ''}`} to={`/cinemas?cinema_id=${cinema.id}`} key={cinema.id}><div><strong>{cinema.name}</strong><span>{cinema.city} · {cinema.address}</span></div><span>{cinema.id === cinemaId ? '当前影院' : '查看'} <Arrow /></span></Link>)}</div>}</section>}
      </div>
    </>}
  </div>
}

function Arrow() {
  return <span className="inline-arrow">→</span>
}
