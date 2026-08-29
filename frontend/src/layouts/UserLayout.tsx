import { Film, LogIn, LogOut, Search, Ticket, WalletCards } from 'lucide-react'
import { FormEvent, useEffect, useRef, useState } from 'react'
import { Link, NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { roleLabel } from '../auth'
import { useAuth, useCinema } from '../app/providers'
import { Avatar } from '../components/ui'
import { CinemaPicker } from '../components/CinemaPicker'

export function UserLayout() {
  const { session, signOut } = useAuth()
  const { setCinemaId } = useCinema()
  const navigate = useNavigate()
  const location = useLocation()
  const searchInputRef = useRef<HTMLInputElement>(null)
  const initialQuery = new URLSearchParams(location.search).get('q') ?? ''
  const [keyword, setKeyword] = useState(initialQuery)

  useEffect(() => {
    const nextKeyword = new URLSearchParams(location.search).get('q') ?? ''
    setKeyword((current) => current === nextKeyword ? current : nextKeyword)
  }, [location.search])

  useEffect(() => {
    const nextCinemaId = Number(new URLSearchParams(location.search).get('cinema_id'))
    if (Number.isInteger(nextCinemaId) && nextCinemaId > 0) setCinemaId(nextCinemaId)
  }, [location.search, setCinemaId])

  useEffect(() => {
    function focusSearch(event: KeyboardEvent) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault()
        searchInputRef.current?.focus()
      }
    }
    window.addEventListener('keydown', focusSearch)
    return () => window.removeEventListener('keydown', focusSearch)
  }, [])

  function submitSearch(event: FormEvent) {
    event.preventDefault()
    const value = keyword.trim()
    if (value) navigate(`/search?q=${encodeURIComponent(value)}`)
  }

  return <div className="user-shell">
    <header className="user-header">
      <div className="header-inner">
        <Link className="brand" to="/"><span className="brand-mark"><Film size={17} /></span><span>LTerm</span></Link>
        <form className="global-search" role="search" onSubmit={submitSearch}><Search size={16} /><input ref={searchInputRef} type="search" value={keyword} onChange={(event) => setKeyword(event.target.value)} placeholder="搜索电影、类型或影院" aria-label="搜索电影、类型或影院" /></form>
        <nav className="user-nav"><NavLink to="/" end>热映</NavLink><NavLink to="/recommend">推荐</NavLink><NavLink to="/cinemas">影院</NavLink>{session && <NavLink to="/orders"><Ticket size={15} />订单</NavLink>}{session && <NavLink to="/rewards"><WalletCards size={15} />会员</NavLink>}</nav>
        <div className="header-actions"><CinemaPicker compact />{session ? <div className="user-menu"><Avatar name={roleLabel(session.role)} size="sm" /><button onClick={() => { signOut(); navigate('/') }} title="退出登录"><LogOut size={15} /></button></div> : <Link className="login-link" to="/login"><LogIn size={15} />登录</Link>}</div>
      </div>
    </header>
    <main className="user-main"><Outlet /></main>
    <footer className="user-footer"><span>© 2026 LTerm Cinema</span><span>把时间留给值得的电影</span></footer>
  </div>
}
