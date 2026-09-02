import type { AdminNavItem, Role } from '../../types'

export const adminNavItems: AdminNavItem[] = [
  { label: '票房大盘', path: '/admin/dashboard', icon: 'dashboard', roles: ['SUPER_ADMIN', 'CINEMA_ADMIN', 'FINANCE'] },
  { label: '影片管理', path: '/admin/movies', icon: 'movies', roles: ['SUPER_ADMIN'] },
  { label: '影厅管理', path: '/admin/halls', icon: 'halls', roles: ['SUPER_ADMIN', 'CINEMA_ADMIN'] },
  { label: '场次排片', path: '/admin/sessions', icon: 'sessions', roles: ['SUPER_ADMIN', 'CINEMA_ADMIN'] },
  { label: '票券核销', path: '/admin/tickets', icon: 'tickets', roles: ['SUPER_ADMIN', 'CINEMA_ADMIN'] },
  { label: '运营内容', path: '/admin/marketing', icon: 'marketing', roles: ['SUPER_ADMIN'] },
  { label: '管理员账号', path: '/admin/admins', icon: 'admins', roles: ['SUPER_ADMIN'] },
  { label: '影院管理', path: '/admin/cinemas', icon: 'cinemas', roles: ['SUPER_ADMIN'] },
]

export function canAccess(role: Role | undefined, roles: Role[]) {
  return Boolean(role && roles.includes(role))
}
