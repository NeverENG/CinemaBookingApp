import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { isAdminRole } from '../../auth'
import { useAuth } from '../providers'
import type { Role } from '../../types'

export function RequireAuth() {
  const { session } = useAuth()
  const location = useLocation()
  if (!session) return <Navigate to={`/login?redirect=${encodeURIComponent(location.pathname + location.search)}`} replace />
  return <Outlet />
}

export function RequireRole({ roles, children }: { roles: Role[]; children?: React.ReactNode }) {
  const { session } = useAuth()
  if (!session) return <Navigate to="/login?mode=admin" replace />
  if (!isAdminRole(session.role) || !roles.includes(session.role)) return <Navigate to="/forbidden" replace />
  return children ?? <Outlet />
}
