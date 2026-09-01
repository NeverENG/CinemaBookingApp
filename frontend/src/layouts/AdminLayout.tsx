import { Activity, Armchair, BadgeCheck, ChevronRight, Clapperboard, LayoutDashboard, LogOut, Megaphone, Menu, Settings2, Ticket, UsersRound, X } from 'lucide-react'
import { useState } from 'react'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { roleLabel } from '../auth'
import { useAuth } from '../app/providers'
import { adminNavItems, canAccess } from '../app/router/route-registry'
import { Avatar } from '../components/ui'

const icons = { dashboard: LayoutDashboard, movies: Clapperboard, halls: Armchair, sessions: Ticket, tickets: BadgeCheck, marketing: Megaphone, admins: UsersRound }

export function AdminLayout() {
  const { session, signOut } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [open, setOpen] = useState(false)
  const visibleItems = adminNavItems.filter((item) => canAccess(session?.role, item.roles))

  function logout() {
    signOut()
    navigate('/login?mode=admin')
  }

  return <div className="admin-shell">
    <aside className={`admin-sidebar ${open ? 'sidebar-open' : ''}`}><div className="admin-brand"><span className="brand-mark"><Activity size={17} /></span><div><strong>LTerm</strong><small>Control Room</small></div><button className="mobile-close" onClick={() => setOpen(false)}><X size={18} /></button></div><div className="admin-scope"><span>当前身份</span><strong>{roleLabel(session?.role)}</strong><small>{session?.role === 'CINEMA_ADMIN' ? `绑定影院 ${session.cinemaId ?? '--'}` : '全局工作空间'}</small></div><nav className="admin-nav">{visibleItems.map((item) => { const Icon = icons[item.icon as keyof typeof icons] ?? Settings2; return <NavLink key={item.path} to={item.path} onClick={() => setOpen(false)} className={({ isActive }) => isActive ? 'active' : ''}><Icon size={17} /><span>{item.label}</span><ChevronRight className="nav-chevron" size={14} /></NavLink> })}</nav><div className="admin-sidebar-bottom"><button className="sidebar-action" onClick={() => navigate('/admin/profile')}><Settings2 size={16} />账户设置</button><button className="sidebar-action" onClick={logout}><LogOut size={16} />退出登录</button></div></aside>
    <div className="admin-content"><header className="admin-topbar"><button className="mobile-menu" onClick={() => setOpen(true)}><Menu size={19} /></button><div><span className="eyebrow">LTerm / Control Room</span><strong>{visibleItems.find((item) => location.pathname.startsWith(item.path))?.label ?? '票房大盘'}</strong></div><div className="admin-topbar-right"><div className="admin-status"><span className="status-dot" />系统运行正常</div><Avatar name={roleLabel(session?.role)} /></div></header><main className="admin-main"><Outlet /></main></div>
  </div>
}
